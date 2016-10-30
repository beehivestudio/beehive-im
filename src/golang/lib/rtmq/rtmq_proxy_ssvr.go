package rtmq

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	RTMQ_HEAD_SIZE uint32 = uint32(binary.Size(RtmqHeader{}))
)

/* 错误类型 */
var (
	TCP_ERR_CONN_CLOSING   = errors.New("Use of closed network connection")
	TCP_ERR_WRITE_BLOCKING = errors.New("Write packet was blocking")
	TCP_ERR_READ_BLOCKING  = errors.New("Read packet was blocking")
)

/* TCP连接对象 */
type RtmqProxyConn struct {
	svr        *RtmqProxyServer
	conn       *net.TCPConn     /* the raw connection */
	extra      interface{}      /* to save extra data */
	closeOnce  sync.Once        /* close the conn, once, per instance */
	is_close   int32            /* close flag */
	send_chan  chan *RtmqPacket /* 普通消息发送队列 */
	mesg_chan  chan *RtmqPacket /* 系统消息发送队列 */
	recv_chan  chan *RtmqPacket /* 普通消息接收队列 */
	close_chan chan struct{}    /* close chanel */
}

/* "网络->主机"字节序 */
func rtmq_head_ntoh(p *RtmqPacket) *RtmqHeader {
	head := &RtmqHeader{}

	head.cmd = p.get_cmd()       /* CMD */
	head.nid = p.get_nid()       /* NID */
	head.flag = p.get_flag()     /* FLAG */
	head.length = p.get_len()    /* LENGTH */
	head.chksum = p.get_chksum() /* CHKSUM */

	return head
}

func (p *RtmqPacket) get_cmd() uint32 {
	return binary.BigEndian.Uint32(p.buff[0:4])
}

func (p *RtmqPacket) get_nid() uint32 {
	return binary.BigEndian.Uint32(p.buff[4:8])
}

func (p *RtmqPacket) get_flag() uint32 {
	return binary.BigEndian.Uint32(p.buff[8:12])
}

func (p *RtmqPacket) get_len() uint32 {
	return binary.BigEndian.Uint32(p.buff[12:16])
}

func (p *RtmqPacket) get_chksum() uint32 {
	return binary.BigEndian.Uint32(p.buff[16:20])
}

/* "主机->网络"字节序 */
func rtmq_head_hton(header *RtmqHeader, p *RtmqPacket) {
	binary.BigEndian.PutUint32(p.buff[0:4], header.cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.buff[4:8], header.nid)      /* NID */
	binary.BigEndian.PutUint32(p.buff[8:12], header.flag)    /* NID */
	binary.BigEndian.PutUint32(p.buff[12:16], header.length) /* LENGTH */
	binary.BigEndian.PutUint32(p.buff[16:20], header.chksum) /* CHKSUM */
}

/* 创建连接对象 */
func rtmq_proxy_conn_creat(conn *net.TCPConn, s *RtmqProxyServer) *RtmqProxyConn {
	return &RtmqProxyConn{
		svr:        s,
		conn:       conn,
		close_chan: make(chan struct{}),
		send_chan:  s.send_chan,
		mesg_chan:  make(chan *RtmqPacket, 100),
		recv_chan:  s.recv_chan,
	}
}

/* 获取扩展数据 */
func (c *RtmqProxyConn) GetExtraData() interface{} {
	return c.extra
}

/* 设置扩展数据 */
func (c *RtmqProxyConn) SetExtraData(extra interface{}) {
	c.extra = extra
}

// GetRawConn returns the raw net.TCPConn from the RtmqProxyConn
func (c *RtmqProxyConn) GetRawConn() *net.TCPConn {
	return c.conn
}

// Close closes the connection
func (c *RtmqProxyConn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.is_close, 1)
		close(c.close_chan)
		c.conn.Close()
		c.svr.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *RtmqProxyConn) IsClosed() bool {
	return atomic.LoadInt32(&c.is_close) == 1
}

/* 各协程启动一个 */
func (c *RtmqProxyConn) Do() {
	if !c.svr.OnConnect(c) {
		return
	}

	c.auth()

	go c.read_routine()
	go c.write_routine()
	go c.handle_routine()
}

/* 启动多个处理协程 */
func (c *RtmqProxyConn) DoPool(num uint32) {
	var i uint32

	if !c.svr.OnConnect(c) {
		return
	}

	c.auth()

	for i = 0; i < num; i++ {
		go c.handle_routine()
	}
	go c.read_routine()
	go c.write_routine()
}

/* 接收协程的处理流程 */
func (c *RtmqProxyConn) read_routine() {
	c.svr.waitGroup.Add(1)
	defer func() {
		recover()
		c.Close()
		c.svr.waitGroup.Done()
	}()

	for {
		select {
		case <-c.svr.exit_chan:
			return
		case <-c.close_chan:
			return
		default:
		}

		tp := &RtmqPacket{}

		/* 读取RTMQ协议头 */
		tp.buff = make([]byte, RTMQ_HEAD_SIZE)

		if _, err := io.ReadFull(c.conn, tp.buff); nil != err {
			return
		}

		/* 转换字节序 */
		header := rtmq_head_ntoh(tp)

		/* 读取承载数据 */
		p := &RtmqPacket{}

		p.buff = make([]byte, RTMQ_HEAD_SIZE+header.length)

		copy(p.buff[:RTMQ_HEAD_SIZE], tp.buff)

		if _, err := io.ReadFull(c.conn, p.buff[RTMQ_HEAD_SIZE:]); nil != err {
			return
		}

		/* 放入接收队列 */
		c.recv_chan <- p
	}
}

/* 发送协程的处理流程 */
func (c *RtmqProxyConn) write_routine() {
	c.svr.waitGroup.Add(1)
	defer func() {
		recover()
		c.Close()
		c.svr.waitGroup.Done()
	}()

	for {
		select {
		case <-c.svr.exit_chan:
			return

		case <-c.close_chan:
			return

		case p := <-c.mesg_chan: /* 系统消息发送队列 */
			if _, err := c.conn.Write([]byte(p.buff)); nil != err {
				return
			}

		case p := <-c.send_chan: /* 普通消息发送队列 */
			if _, err := c.conn.Write([]byte(p.buff)); nil != err {
				return
			}

		case <-time.After(1 * time.Second): /* 保活消息 */
			c.keepalive()
			continue

		}
	}
}

/* 发送保活消息 */
func (c *RtmqProxyConn) keepalive() {
	svr := c.svr
	conf := svr.conf

	head := &RtmqHeader{}

	head.cmd = RTMQ_CMD_KPALIVE_REQ
	head.nid = conf.NodeId
	head.flag = RTMQ_SYS_DATA
	head.length = 0
	head.chksum = RTMQ_CHKSUM_VAL

	p := &RtmqPacket{}
	p.buff = make([]byte, RTMQ_HEAD_SIZE)

	rtmq_head_hton(head, p)

	c.mesg_chan <- p
}

/* 链路鉴权请求 */
type RtmqAuthReq struct {
	usr    [RTMQ_USR_MAX_LEN]byte /* 用户名 */
	passwd [RTMQ_PWD_MAX_LEN]byte /* 登录密码 */
}

/* 发送鉴权消息 */
func (c *RtmqProxyConn) auth() {
	svr := c.svr
	conf := svr.conf

	/* > 设置头部数据 */
	head := &RtmqHeader{}

	head.cmd = RTMQ_CMD_LINK_AUTH_REQ
	head.nid = conf.NodeId
	head.flag = RTMQ_SYS_DATA
	head.length = uint32(binary.Size(RtmqAuthReq{}))
	head.chksum = RTMQ_CHKSUM_VAL

	/* > 申请内存空间 */
	p := &RtmqPacket{}
	p.buff = make([]byte, RTMQ_HEAD_SIZE+head.length)

	rtmq_head_hton(head, p)

	fmt.Printf("auth:%s passwd:%s", conf.Usr, conf.Passwd)

	copy(p.buff[RTMQ_HEAD_SIZE:], []byte(conf.Usr))
	copy(p.buff[RTMQ_HEAD_SIZE+RTMQ_USR_MAX_LEN:], []byte(conf.Passwd))

	c.mesg_chan <- p
}

/* 工作协程的处理流程 */
func (c *RtmqProxyConn) handle_routine() {
	c.svr.waitGroup.Add(1)
	defer func() {
		recover()
		c.Close()
		c.svr.waitGroup.Done()
	}()

	for {
		select {
		case <-c.svr.exit_chan:
			return

		case <-c.close_chan:
			return

		case p := <-c.recv_chan:
			c.svr.OnMessage(c, p)
		}
	}
}

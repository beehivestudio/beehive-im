package rtmq

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var (
	RTMQ_HEAD_SIZE uint32 = 17
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
	conn       *net.TCPConn     // the raw connection
	extra      interface{}      // to save extra data
	closeOnce  sync.Once        // close the conn, once, per instance
	is_close   int32            // close flag
	send_chan  chan *RtmqPacket // packet send chanel
	recv_chan  chan *RtmqPacket // packeet receive chanel
	close_chan chan struct{}    // close chanel
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	OnDial() (*net.TCPConn, error)
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*RtmqProxyConn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*RtmqProxyConn, []byte) bool

	// OnClose is called when the connection closed
	OnClose(*RtmqProxyConn)
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

func (p *RtmqPacket) get_flag() uint8 {
	return uint8(binary.BigEndian.Uint32(p.buff[8:9]))
}

func (p *RtmqPacket) get_len() uint32 {
	return binary.BigEndian.Uint32(p.buff[9:13])
}

func (p *RtmqPacket) get_chksum() uint32 {
	return binary.BigEndian.Uint32(p.buff[13:17])
}

/* "主机->网络"字节序 */
func rtmq_head_hton(header *RtmqHeader, p *RtmqPacket) {
	binary.BigEndian.PutUint32(p.buff[0:4], header.cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.buff[4:8], header.nid)      /* NID */
	binary.BigEndian.PutUint32(p.buff[9:13], header.length)  /* LENGTH */
	binary.BigEndian.PutUint32(p.buff[13:17], header.chksum) /* CHKSUM */
}

/* 创建连接对象 */
func rtmq_proxy_conn_creat(conn *net.TCPConn, s *RtmqProxyServer) *RtmqProxyConn {
	return &RtmqProxyConn{
		svr:        s,
		conn:       conn,
		close_chan: make(chan struct{}),
		send_chan:  s.send_chan,
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
		c.svr.callback.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *RtmqProxyConn) IsClosed() bool {
	return atomic.LoadInt32(&c.is_close) == 1
}

/* 各协程启动一个 */
func (c *RtmqProxyConn) Do() {
	if !c.svr.callback.OnConnect(c) {
		return
	}

	go c.handleLoop()
	go c.readLoop()
	go c.writeLoop()
}

/* 启动多个处理协程 */
func (c *RtmqProxyConn) DoPool(num uint32) {
	if !c.svr.callback.OnConnect(c) {
		return
	}
	var i uint32
	for i = 0; i < num; i++ {
		go c.handleLoop()
	}
	go c.readLoop()
	go c.writeLoop()
}

/* 接收协程的处理流程 */
func (c *RtmqProxyConn) readLoop() {
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
func (c *RtmqProxyConn) writeLoop() {
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

		case p := <-c.send_chan:
			if _, err := c.conn.Write([]byte(p.buff)); nil != err {
				return
			}
		}
	}
}

/* 工作协程的处理流程 */
func (c *RtmqProxyConn) handleLoop() {
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
			if !c.svr.callback.OnMessage(c, p.buff) {
				// return
			}
			// go c.svr.callback.OnMessage(c, p)
		}
	}
}

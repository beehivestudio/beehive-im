package rtmq

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var RTMQ_HEAD_SIZE = binary.Size(RtmqHeader)

/* 错误类型 */
var (
	TCP_ERR_CONN_CLOSING   = errors.New("Use of closed network connection")
	TCP_ERR_WRITE_BLOCKING = errors.New("Write packet was blocking")
	TCP_ERR_READ_BLOCKING  = errors.New("Read packet was blocking")
)

/* TCP连接对象 */
type RtmqProxyConn struct {
	s          *RtmqProxyServer
	conn       *net.TCPConn  // the raw connection
	extra      interface{}   // to save extra data
	closeOnce  sync.Once     // close the conn, once, per instance
	is_close   int32         // close flag
	send_chan  *chan Packet  // packet send chanel
	recv_chan  *chan Packet  // packeet receive chanel
	close_chan chan struct{} // close chanel
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	OnDial() (*net.TCPConn, error)
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*RtmqProxyConn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*RtmqProxyConn, Packet) bool

	// OnClose is called when the connection closed
	OnClose(*RtmqProxyConn)
}

/* "网络->主机"字节序 */
func rtmq_proxy_head_ntoh(header *RtmqHeader) {
	header.cmd = binary.BigEndian.Uint32(header.cmd)       /* CMD */
	header.nid = binary.BigEndian.Uint32(header.nid)       /* NID */
	header.length = binary.BigEndian.Uint32(header.length) /* LENGTH */
	header.chksum = binary.BigEndian.Uint32(header.chksum) /* CHKSUM */
}

/* "主机->网络"字节序 */
func rtmq_head_hton(header *RtmqHeader) {
	binary.BigEndian.PutUint32(header.cmd, header.cmd)       /* CMD */
	binary.BigEndian.PutUint32(header.nid, header.nid)       /* NID */
	binary.BigEndian.PutUint32(header.length, header.length) /* LENGTH */
	binary.BigEndian.PutUint32(header.chksum, header.chksum) /* CHKSUM */
}

/* 接收协程 */
func rtmq_proxy_recv_routine(pxy *RtmqProxyCntx, idx int) {
	recv_chan := pxy.recv_chan[idx]

	for {
		conn := pxy.conn[idx]

		/* 读取RTMQ协议头 */
		h := make([]byte, RTMQ_HEAD_SIZE)

		if _, err := io.ReadFull(conn, h); nil != err {
		}

		/* 转换字节序 */
		header := h.(*RtmqHeader)

		rtmq_proxy_head_ntoh(header)

		/* 读取承载数据 */
		buff := make([]byte, RTMQ_HEAD_SIZE+header.length)

		copy(buff[0:RTMQ_HEAD_SIZE], header)

		if _, err := io.ReadFull(buff[RTMQ_HEAD_SIZE:], length); nil != err {
		}

		/* 放入接收队列 */
		recv_chan <- buff
	}
}

/* 创建连接对象 */
func rtmq_proxy_conn_creat(conn *net.TCPConn, s *RtmqProxyServer) *RtmqProxyConn {
	return &RtmqProxyConn{
		svr:        s,
		conn:       conn,
		close_chan: make(chan struct{}),
		send_chan:  &s.send_chan,
		recv_chan:  &s.recv_chan,
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

		/* 读取RTMQ协议头 */
		h := make([]byte, RTMQ_HEAD_SIZE)

		if _, err := io.ReadFull(conn, h); nil != err {
			return
		}

		/* 转换字节序 */
		header := h.(*RtmqHeader)

		rtmq_proxy_head_ntoh(header)

		/* 读取承载数据 */
		buff := make([]byte, RTMQ_HEAD_SIZE+header.length)

		copy(buff[0:RTMQ_HEAD_SIZE], header)

		if _, err := io.ReadFull(buff[RTMQ_HEAD_SIZE:], length); nil != err {
			return
		}

		/* 放入接收队列 */
		c.recv_chan <- buff
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
			if _, err := c.conn.Write([]byte(p)); nil != err {
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
			if !c.svr.callback.OnMessage(c, p) {
				// return
			}
			// go c.svr.callback.OnMessage(c, p)
		}
	}
}

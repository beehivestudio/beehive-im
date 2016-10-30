package rtmq

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/logs"
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

const (
	RTMQ_SSVR_NUM    = 10
	RTMQ_CHKSUM_VAL  = 0x1FE23DC4
	RTMQ_USR_MAX_LEN = 32
	RTMQ_PWD_MAX_LEN = 16
	RTMQ_SYS_DATA    = 0 /* 系统数据 */
	RTMQ_USR_DATA    = 1 /* 业务数据 */
)

/* 系统数据类型 */
const (
	RTMQ_CMD_UNKNOWN       = 0      /* 未知命令 */
	RTMQ_CMD_LINK_AUTH_REQ = 0x0001 /* 链路鉴权请求 */
	RTMQ_CMD_LINK_AUTH_RSP = 0x0002 /* 链路鉴权应答 */
	RTMQ_CMD_KPALIVE_REQ   = 0x0003 /* 链路保活请求 */
	RTMQ_CMD_KPALIVE_RSP   = 0x0004 /* 链路保活应答 */
	RTMQ_CMD_SUB_ONE_REQ   = 0x0005 /* 订阅请求: 将消息只发送给一个用户 */
	RTMQ_CMD_SUB_ONE_RSP   = 0x0006 /* 订阅应答 */
	RTMQ_CMD_SUB_ALL_REQ   = 0x0007 /* 订阅请求: 将消息发送给所有用户 */
	RTMQ_CMD_SUB_ALL_RSP   = 0x0008 /* 订阅应答 */
	RTMQ_CMD_ADD_SCK       = 0x0009 /* 接收客户端数据-请求 */
	RTMQ_CMD_DIST_REQ      = 0x000A /* 分发任务请求 */
	RTMQ_CMD_PROC_REQ      = 0x000B /* 处理客户端数据-请求 */
	RTMQ_CMD_SEND          = 0x000C /* 发送数据-请求 */
	RTMQ_CMD_SEND_ALL      = 0x000D /* 发送所有数据-请求 */
	/* 查询命令 */
	RTMQ_CMD_QUERY_CONF_REQ      = 0x1001 /* 查询配置信息-请求 */
	RTMQ_CMD_QUERY_CONF_REP      = 0x1002 /* 查询配置信息-应答 */
	RTMQ_CMD_QUERY_RECV_STAT_REQ = 0x1003 /* 查询接收状态-请求 */
	RTMQ_CMD_QUERY_RECV_STAT_REP = 0x1004 /* 查询接收状态-应答 */
	RTMQ_CMD_QUERY_PROC_STAT_REQ = 0x1005 /* 查询处理状态-请求 */
	RTMQ_CMD_QUERY_PROC_STAT_REP = 0x1006 /* 查询处理状态-应答 */
)

/* 配置信息 */
type RtmqProxyConf struct {
	NodeId      uint32 /* 结点ID */
	Usr         string /* 用户名 */
	Passwd      string /* 登录密码 */
	RemoteAddr  string /* 对端IP地址 */
	WorkerNum   uint32 /* 工作协程数 */
	SendChanLen uint32 /* 发送队列长度 */
	RecvChanLen uint32 /* 接收队列长度 */
}

type RtmqPacket struct {
	buff []byte /* 接收数据 */
}

/* 协议头 */
type RtmqHeader struct {
	cmd    uint32 /* 消息类型 */
	nid    uint32 /* 结点ID */
	flag   uint32 /* 消息标识(0:系统消息 1:业务消息) */
	length uint32 /* 报体长度 */
	chksum uint32 /* 校验值(固定为0x1FE23DE4) */
}

type RtmqRegCb func(cmd uint32, orig uint32, data []byte, length uint32, param interface{})

/* 回调注册项 */
type RtmqRegItem struct {
	cmd   uint32      /* 命令类型 */
	proc  RtmqRegCb   /* 回调函数 */
	param interface{} /* 附加参数 */
}

type RtmqProxyServer struct {
	ctx       *RtmqProxyCntx   /* 全局对象 */
	conf      *RtmqProxyConf   /* 配置数据 */
	log       *logs.BeeLogger  /* 日志对象 */
	send_chan chan *RtmqPacket /* 发送队列 */
	recv_chan chan *RtmqPacket /* 接收队列 */
	exit_chan chan struct{}    /* 通知所有协程退出 */
	waitGroup *sync.WaitGroup  /* 用于等待所有协程 */
}

/* 上下文信息 */
type RtmqProxyCntx struct {
	conf   *RtmqProxyConf                  /* 配置数据 */
	log    *logs.BeeLogger                 /* 日志对象 */
	reg    map[uint32]*RtmqRegItem         /* 回调注册 */
	server [RTMQ_SSVR_NUM]*RtmqProxyServer /* 服务对象 */
}

/* 连接远端服务 */
func (s *RtmqProxyServer) OnDial() (*net.TCPConn, error) {
	conf := s.conf

	addr, err := net.ResolveTCPAddr("tcp4", conf.RemoteAddr)
	if nil != err {
		s.log.Error("Resolve tcp addr failed! addr:%s errmsg:%s", conf.RemoteAddr, err.Error())
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if nil != err {
		s.log.Error("Dial tcp addr failed! addr:%s errmsg:%s", conf.RemoteAddr, err.Error())
		return nil, err
	}

	return conn, nil
}

/* 连接远端服务 */
func (s *RtmqProxyServer) OnConnect(c *RtmqProxyConn) bool {
	return true
}

func (s *RtmqProxyServer) OnMessage(c *RtmqProxyConn, p *RtmqPacket) bool {
	ctx := s.ctx
	header := rtmq_head_ntoh(p)

	/* 获取CMD对应的注册项 */
	item, ok := ctx.reg[header.cmd]
	if !ok {
		ctx.log.Error("Drop unknown data! cmd:%d", header.cmd)
		return false
	}

	/* 调用注册处理函数 */
	item.proc(header.cmd, header.nid, p.buff[RTMQ_HEAD_SIZE:], header.length, item.param)

	return true
}

func (s *RtmqProxyServer) OnClose(c *RtmqProxyConn) {
	s.log.Error("Connection is close! ip:%s", c.GetRawConn().RemoteAddr())
}

/* 初始化PROXY服务 */
func RtmqProxyInit(conf *RtmqProxyConf, log *logs.BeeLogger) *RtmqProxyCntx {
	ctx := &RtmqProxyCntx{}

	ctx.log = log
	ctx.conf = conf
	for idx := 0; idx < RTMQ_SSVR_NUM; idx += 1 {
		ctx.server[idx] = rtmq_proxy_server_init(ctx, conf)
		if nil == ctx.server[idx] {
			return nil
		}
		go ctx.server[idx].StartConnector(3)
	}

	return ctx
}

/* 回调注册函数 */
func (this *RtmqProxyCntx) Register(cmd uint32, proc RtmqRegCb, param interface{}) {
	item := &RtmqRegItem{}

	item.cmd = cmd
	item.proc = proc
	item.param = param

	this.reg[cmd] = item
}

/* 初始化PROXY服务对象 */
func rtmq_proxy_server_init(ctx *RtmqProxyCntx, conf *RtmqProxyConf) *RtmqProxyServer {
	return &RtmqProxyServer{
		ctx:       ctx,
		conf:      conf,
		log:       ctx.log,
		exit_chan: make(chan struct{}),
		send_chan: make(chan *RtmqPacket, conf.SendChanLen),
		recv_chan: make(chan *RtmqPacket, conf.RecvChanLen),
		waitGroup: &sync.WaitGroup{},
	}
}

/* 启动连接服务 */
func (s *RtmqProxyServer) StartConnector(timeout time.Duration) {
	s.waitGroup.Add(1)
	defer func() {
		s.waitGroup.Done()
	}()

	for {
		/* > 建立TCP连接 */
		conn, err := s.OnDial()
		if nil != err {
			select {
			case <-s.exit_chan:
				return
			case <-time.After(time.Second * timeout):
				continue
			}
		}

		/* > 创建TC连接对象 */
		c := rtmq_proxy_conn_creat(conn, s)
		if 0 == s.conf.WorkerNum {
			go c.Do()
		} else {
			go c.DoPool(s.conf.WorkerNum)
		}

		/* > 等待异常信号 */
		select {
		case <-s.exit_chan:
			return
		case <-c.close_chan:
			time.Sleep(time.Second * timeout)
		}
	}
}

/* 停止服务 */
func (s *RtmqProxyServer) Stop() {
	close(s.exit_chan)
	s.waitGroup.Wait()
}

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
	c.subscribe()

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
	c.subscribe()

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

/* 设置鉴权请求 */
func rtmq_set_auth_req(conf *RtmqProxyConf, p *RtmqPacket) {
	copy(p.buff[RTMQ_HEAD_SIZE:], []byte(conf.Usr))
	copy(p.buff[RTMQ_HEAD_SIZE+RTMQ_USR_MAX_LEN:], []byte(conf.Passwd))
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
	rtmq_set_auth_req(conf, p)

	c.mesg_chan <- p
}

/* 订阅请求 */
type RtmqSubReq struct {
	cmd uint32 /* 被订阅的消息 */
}

/* 设置订阅请求 */
func rtmq_set_sub_req(req *RtmqSubReq, p *RtmqPacket) {
	binary.BigEndian.PutUint32(p.buff[RTMQ_HEAD_SIZE:4], req.cmd) /* CMD */
}

/* 发送订阅消息 */
func (c *RtmqProxyConn) subscribe() {
	svr := c.svr
	ctx := svr.ctx
	conf := svr.conf

	for cmd, _ := range ctx.reg {
		/* > 设置头部数据 */
		head := &RtmqHeader{}

		head.cmd = RTMQ_CMD_SUB_ONE_REQ
		head.nid = conf.NodeId
		head.flag = RTMQ_SYS_DATA
		head.length = uint32(binary.Size(RtmqSubReq{}))
		head.chksum = RTMQ_CHKSUM_VAL

		req := &RtmqSubReq{}
		req.cmd = cmd

		/* > 申请内存空间 */
		p := &RtmqPacket{}
		p.buff = make([]byte, RTMQ_HEAD_SIZE+head.length)

		rtmq_head_hton(head, p)
		rtmq_set_sub_req(req, p)

		c.mesg_chan <- p
	}
}
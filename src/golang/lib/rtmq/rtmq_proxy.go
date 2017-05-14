package rtmq

import (
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/logs"
)

var (
	RTMQ_HEAD_SIZE uint32 = uint32(binary.Size(RtmqHeader{})) /* RTMQ协议头长度 */
)

/* 错误类型 */
var (
	TCP_ERR_CONN_CLOSING   = errors.New("Use of closed network connection")
	TCP_ERR_WRITE_BLOCKING = errors.New("Write packet was blocking")
	TCP_ERR_READ_BLOCKING  = errors.New("Read packet was blocking")
)

/* 常量定义 */
const (
	RTMQ_SSVR_NUM    = 10         /* 服务个数 */
	RTMQ_CHKSUM_VAL  = 0x1FE23DC4 /* 校验值 */
	RTMQ_USR_MAX_LEN = 32         /* 用户名长度 */
	RTMQ_PWD_MAX_LEN = 16         /* 登录密码长度 */
	RTMQ_SYS_DATA    = 0          /* 系统数据 */
	RTMQ_USR_DATA    = 1          /* 业务数据 */
)

/* 保活状态 */
const (
	RTMQ_KPALIVE_STAT_UNKNOWN = 0 /* 未知 */
	RTMQ_KPALIVE_STAT_SENT    = 1 /* 已发送 */
	RTMQ_KPALIVE_STAT_SUCC    = 2 /* 成功 */
	RTMQ_KPALIVE_STAT_FAIL    = 3 /* 失败 */
)

/* 命令类型 */
const (
	RTMQ_CMD_UNKNOWN             = 0      /* 未知命令 */
	RTMQ_CMD_LINK_AUTH_REQ       = 0x0001 /* 链路鉴权请求 */
	RTMQ_CMD_LINK_AUTH_ACK       = 0x0002 /* 链路鉴权应答 */
	RTMQ_CMD_KPALIVE_REQ         = 0x0003 /* 链路保活请求 */
	RTMQ_CMD_KPALIVE_ACK         = 0x0004 /* 链路保活应答 */
	RTMQ_CMD_SUB_ONE_REQ         = 0x0005 /* 订阅请求: 将消息只发送给一个用户 */
	RTMQ_CMD_SUB_ONE_ACK         = 0x0006 /* 订阅应答 */
	RTMQ_CMD_SUB_ALL_REQ         = 0x0007 /* 订阅请求: 将消息发送给所有用户 */
	RTMQ_CMD_SUB_ALL_ACK         = 0x0008 /* 订阅应答 */
	RTMQ_CMD_ADD_SCK             = 0x0009 /* 接收客户端数据-请求 */
	RTMQ_CMD_DIST_REQ            = 0x000A /* 分发任务请求 */
	RTMQ_CMD_PROC_REQ            = 0x000B /* 处理客户端数据-请求 */
	RTMQ_CMD_SEND                = 0x000C /* 发送数据-请求 */
	RTMQ_CMD_SEND_ALL            = 0x000D /* 发送所有数据-请求 */
	RTMQ_CMD_QUERY_CONF_REQ      = 0x1001 /* "查询"配置信息-请求 */
	RTMQ_CMD_QUERY_CONF_ACK      = 0x1002 /* "查询"配置信息-应答 */
	RTMQ_CMD_QUERY_RECV_STAT_REQ = 0x1003 /* "查询"接收状态-请求 */
	RTMQ_CMD_QUERY_RECV_STAT_ACK = 0x1004 /* "查询"接收状态-应答 */
	RTMQ_CMD_QUERY_PROC_STAT_REQ = 0x1005 /* "查询"处理状态-请求 */
	RTMQ_CMD_QUERY_PROC_STAT_ACK = 0x1006 /* "查询"处理状态-应答 */
)

/* 配置信息 */
type ProxyConf struct {
	NodeId      uint32 /* 结点ID */
	Usr         string /* 用户名 */
	Passwd      string /* 登录密码 */
	RemoteAddr  string /* 对端IP地址 */
	WorkerNum   uint32 /* 工作协程数 */
	SendChanLen uint32 /* 发送队列长度 */
	RecvChanLen uint32 /* 接收队列长度 */
}
type RtmqPacket struct {
	head []byte /* 头部数据 */
	body []byte /* 报体数据 */
}
type RtmqRecvPacket struct {
	head []byte /* 头部数据 */
	body []byte /* 报体数据 */
}

/* 协议头 */
type RtmqHeader struct {
	cmd    uint32 /* 消息类型 */
	nid    uint32 /* 结点ID */
	flag   uint32 /* 消息标识(0:系统消息 1:业务消息) */
	length uint32 /* 报体长度 */
	chksum uint32 /* 校验值(固定为0x1FE23DE4) */
}

type RtmqRegCb func(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int

/* 回调注册项 */
type RtmqRegItem struct {
	cmd   uint32      /* 命令类型 */
	proc  RtmqRegCb   /* 回调函数 */
	param interface{} /* 附加参数 */
}

/* TCP连接对象 */
type ProxyConn struct {
	svr           *ProxyServer
	conn          *net.TCPConn         /* 原始TCP连接 */
	extra         interface{}          /* 扩展数据 */
	is_close      int32                /* 连接是否关闭 */
	send_chan     chan *RtmqPacket     /* 普通消息发送队列 */
	mesg_chan     chan *RtmqPacket     /* 系统消息发送队列 */
	recv_chan     chan *RtmqRecvPacket /* 普通消息接收队列 */
	close_chan    chan struct{}        /* 关闭通道 */
	close_once    sync.Once            /* 连接只允许被关闭一次 */
	is_auth       bool                 /* 鉴权是否成功 */
	kpalive_time  int64                /* 发送保活的时间 */
	kpalive_stat  int32                /* 保活状态 */
	kpalive_times int32                /* 保活尝试次数 */
}

/* 代理服务 */
type ProxyServer struct {
	ctx       *Proxy               /* 全局对象 */
	conf      *ProxyConf           /* 配置数据 */
	log       *logs.BeeLogger      /* 日志对象 */
	send_chan chan *RtmqPacket     /* 发送队列 */
	recv_chan chan *RtmqRecvPacket /* 接收队列 */
	exit_chan chan struct{}        /* 通知所有协程退出 */
	waitGroup *sync.WaitGroup      /* 用于等待所有协程 */
}

/* 上下文信息 */
type Proxy struct {
	conf   *ProxyConf                  /* 配置数据 */
	log    *logs.BeeLogger             /* 日志对象 */
	reg    map[uint32]*RtmqRegItem     /* 回调注册 */
	server [RTMQ_SSVR_NUM]*ProxyServer /* 服务对象 */
}

/* 获取日志对象 */
func (pxy *Proxy) GetLog() *logs.BeeLogger {
	return pxy.log
}

/******************************************************************************
 **函数名称: OnDial
 **功    能: 连接远端服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     conn: 连接对象
 **     err: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 20:56:41 #
 ******************************************************************************/
func (svr *ProxyServer) OnDial() (conn *net.TCPConn, err error) {
	conf := svr.conf

	addr, err := net.ResolveTCPAddr("tcp4", conf.RemoteAddr)
	if nil != err {
		svr.log.Error("Resolve tcp addr failed! errmsg:%s", err.Error())
		return nil, err
	}

	conn, err = net.DialTCP("tcp", nil, addr)
	if nil != err {
		svr.log.Error("Dial tcp addr failed! errmsg:%s", err.Error())
		return nil, err
	}

	return conn, nil
}

/******************************************************************************
 **函数名称: OnConnect
 **功    能: 连接远端服务
 **输入参数:
 **     c: 连接对象
 **输出参数: NONE
 **返    回: true:成功 false:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 20:56:41 #
 ******************************************************************************/
func (svr *ProxyServer) OnConnect(c *ProxyConn) bool {
	return true
}

/******************************************************************************
 **函数名称: OnMessage
 **功    能: 消息处理
 **输入参数:
 **     c: 连接对象
 **输出参数: NONE
 **返    回: true:成功 false:失败
 **实现描述:
 **     1. 如果是内部消息, 则调用mesg_handler()进行处理
 **     2. 如果是扩展消息, 则查找对应回调proc()进行处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:06:03 #
 ******************************************************************************/
func (svr *ProxyServer) OnMessage(c *ProxyConn, p *RtmqRecvPacket) bool {
	ctx := svr.ctx
	header := rtmq_head_ntoh(p)

	/* 内部消息处理 */
	if RTMQ_SYS_DATA == header.flag {
		return c.mesg_handler(header.cmd, p)
	}

	/* 获取CMD对应的注册项 */
	item, ok := ctx.reg[header.cmd]
	if !ok {
		item, ok = ctx.reg[0] /* 0:表示默认处理 */
		if !ok {
			ctx.log.Error("Drop unknown data! cmd:%d", header.cmd)
			return false
		}
	}

	/* 调用注册处理函数 */
	item.proc(header.cmd, header.nid, p.body[:], header.length, item.param)

	return true
}

/******************************************************************************
 **函数名称: OnClose
 **功    能: 连接被关闭
 **输入参数:
 **     c: 连接对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:06:03 #
 ******************************************************************************/
func (svr *ProxyServer) OnClose(c *ProxyConn) {
	svr.log.Error("Connection was closed! ip:%svr", c.GetRawConn().RemoteAddr())
}

/******************************************************************************
 **函数名称: ProxyInit
 **功    能: 初始化PROXY服务
 **输入参数:
 **     conf: 配置数据
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 上下文对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:09:33 #
 ******************************************************************************/
func ProxyInit(conf *ProxyConf, log *logs.BeeLogger) *Proxy {
	ctx := &Proxy{}

	ctx.log = log
	ctx.conf = conf
	for idx := 0; idx < RTMQ_SSVR_NUM; idx += 1 {
		ctx.server[idx] = ctx.server_new()
		if nil == ctx.server[idx] {
			log.Error("Init rtmq proxy failed!")
			return nil
		}
	}

	ctx.reg = make(map[uint32]*RtmqRegItem, 0)

	return ctx
}

/******************************************************************************
 **函数名称: Register
 **功    能: 回调注册函数
 **输入参数:
 **     cmd: 消息类型
 **     proc: 消息处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: true:成功 false:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:17:46 #
 ******************************************************************************/
func (ctx *Proxy) Register(cmd uint32, proc RtmqRegCb, param interface{}) bool {
	item := &RtmqRegItem{}

	if _, ok := ctx.reg[cmd]; ok {
		return false
	}

	item.cmd = cmd
	item.proc = proc
	item.param = param

	ctx.reg[cmd] = item
	return true
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动PROXY服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:24:34 #
 ******************************************************************************/
func (ctx *Proxy) Launch() {
	for idx := 0; idx < RTMQ_SSVR_NUM; idx += 1 {
		go ctx.server[idx].StartConnector(3)
	}
}

/******************************************************************************
 **函数名称: AsyncSend
 **功    能: 发送数据
 **输入参数:
 **     cmd: 数据类型
 **     data: 数据内容
 **     length: 数据长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将数据放入发送队列
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 09:36:10 #
 ******************************************************************************/
func (ctx *Proxy) AsyncSend(cmd uint32, data []byte, length uint32) int {
	/* > 设置协议头 */
	head := &RtmqHeader{}

	head.cmd = cmd
	head.nid = ctx.conf.NodeId
	head.length = length
	head.flag = RTMQ_USR_DATA
	head.chksum = RTMQ_CHKSUM_VAL

	/* > 字节序转换 */
	p := &RtmqPacket{}
	p.head = make([]byte, RTMQ_HEAD_SIZE)
	p.body = make([]byte, length)

	rtmq_head_hton(head, p)
	copy(p.body, data)

	/* > 放入发送队列 */
	idx := rand.Intn(RTMQ_SSVR_NUM)

	select {
	case ctx.server[idx].send_chan <- p:
		ctx.log.Debug("Send data success! cmd:0x%04x len:%d", cmd, length)
		return 0
	case <-time.After(3 * time.Second): /* 超时则丢弃 */
		return -1
	}

	return 0
}

/******************************************************************************
 **函数名称: server_new
 **功    能: 新建PROXY服务对象
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 服务对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:17:46 #
 ******************************************************************************/
func (ctx *Proxy) server_new() *ProxyServer {
	conf := ctx.conf
	return &ProxyServer{
		ctx:       ctx,
		conf:      conf,
		log:       ctx.log,
		exit_chan: make(chan struct{}),
		send_chan: make(chan *RtmqPacket, conf.SendChanLen),
		recv_chan: make(chan *RtmqRecvPacket, conf.RecvChanLen),
		waitGroup: &sync.WaitGroup{},
	}
}

/******************************************************************************
 **函数名称: StartConnector
 **功    能: 启动连接服务
 **输入参数:
 **     timeout: 超时等待时间
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:24:34 #
 ******************************************************************************/
func (svr *ProxyServer) StartConnector(timeout time.Duration) {
	svr.waitGroup.Add(1)
	defer func() {
		svr.waitGroup.Done()
	}()

	for {
		/* > 建立TCP连接 */
		conn, err := svr.OnDial()
		if nil != err {
			svr.log.Error("Dial failed! errmsg:%s", err.Error())
			select {
			case <-svr.exit_chan:
				svr.log.Error("Recv exit signal! errmsg:%s", err.Error())
				return
			case <-time.After(time.Second * timeout):
				svr.log.Error("Dial timeout! errmsg:%s", err.Error())
				continue
			}
			continue
		}

		/* > 创建连接对象 */
		c := svr.conn_new(conn)

		go c.Start(svr.conf.WorkerNum)

		/* > 等待异常信号 */
		select {
		case <-svr.exit_chan:
			return
		case <-c.close_chan:
			time.Sleep(time.Second * timeout)
		}
	}
}

/******************************************************************************
 **函数名称: Stop
 **功    能: 停止服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:47:58 #
 ******************************************************************************/
func (svr *ProxyServer) Stop() {
	close(svr.exit_chan)
	svr.waitGroup.Wait()
}

/* "网络->主机"字节序 */
func rtmq_head_ntoh(p *RtmqRecvPacket) *RtmqHeader {
	head := &RtmqHeader{}

	head.cmd = binary.BigEndian.Uint32(p.head[0:4])      /* CMD */
	head.nid = binary.BigEndian.Uint32(p.head[4:8])      /* NID */
	head.flag = binary.BigEndian.Uint32(p.head[8:12])    /* FLAG */
	head.length = binary.BigEndian.Uint32(p.head[12:16]) /* LENGTH */
	head.chksum = binary.BigEndian.Uint32(p.head[16:20]) /* CHKSUM */

	return head
}

/* "主机->网络"字节序 */
func rtmq_head_hton(header *RtmqHeader, p *RtmqPacket) {
	binary.BigEndian.PutUint32(p.head[0:4], header.cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.head[4:8], header.nid)      /* NID */
	binary.BigEndian.PutUint32(p.head[8:12], header.flag)    /* NID */
	binary.BigEndian.PutUint32(p.head[12:16], header.length) /* LENGTH */
	binary.BigEndian.PutUint32(p.head[16:20], header.chksum) /* CHKSUM */
}

/******************************************************************************
 **函数名称: conn_new
 **功    能: 创建连接对象
 **输入参数:
 **     conn: TCP连接
 **输出参数: NONE
 **返    回: 连接对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 21:50:50 #
 ******************************************************************************/
func (svr *ProxyServer) conn_new(conn *net.TCPConn) *ProxyConn {
	return &ProxyConn{
		svr:           svr,
		conn:          conn,
		close_chan:    make(chan struct{}),
		send_chan:     svr.send_chan,
		mesg_chan:     make(chan *RtmqPacket, 1000),
		recv_chan:     svr.recv_chan,
		kpalive_time:  time.Now().Unix(),
		kpalive_stat:  RTMQ_KPALIVE_STAT_UNKNOWN,
		kpalive_times: 0,
	}
}

/* 获取TCP连接对象 */
func (c *ProxyConn) GetRawConn() *net.TCPConn {
	return c.conn
}

/* 关闭连接 */
func (c *ProxyConn) Close() {
	c.close_once.Do(func() {
		atomic.StoreInt32(&c.is_close, 1)
		close(c.mesg_chan)
		close(c.close_chan)
		c.conn.Close()
		c.svr.OnClose(c)
	})
}

/* 判断连接是否关闭 */
func (c *ProxyConn) IsClosed() bool {
	return (1 == atomic.LoadInt32(&c.is_close))
}

/******************************************************************************
 **函数名称: Start
 **功    能: 启动协程
 **输入参数:
 **     num: 工作协程数
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:07:19 #
 ******************************************************************************/
func (c *ProxyConn) Start(num uint32) {
	var i uint32

	if !c.svr.OnConnect(c) {
		return
	}

	c.auth()      // 鉴权
	c.subscribe() // 订阅

	if 0 == num {
		num = 1
	}

	for i = 0; i < num; i++ {
		go c.handle_routine()
	}
	go c.recv_routine()
	c.send_routine()
}

/******************************************************************************
 **函数名称: recv_routine
 **功    能: 接收协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 将发送队列中的数据发送出去
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:10:36 #
 ******************************************************************************/
func (c *ProxyConn) recv_routine() {
	log := c.svr.log
	c.svr.waitGroup.Add(1)
	defer func() {
		if err := recover(); nil != err {
			log.Error("Recv routine crashed! errmsg:%s", err)
		}
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

		p := &RtmqRecvPacket{}

		/* 读取RTMQ协议头 */
		p.head = make([]byte, RTMQ_HEAD_SIZE)

		if _, err := io.ReadFull(c.conn, p.head); nil != err {
			log.Error("Recv head failed! errmsg:%s", err)
			return
		}

		/* 转换字节序 */
		header := rtmq_head_ntoh(p)

		/* 读取承载数据 */
		p.body = make([]byte, header.length)

		if _, err := io.ReadFull(c.conn, p.body[:]); nil != err {
			log.Error("Recv body failed! errmsg:%s", err)
			return
		}

		c.kpalive_times = 0
		c.kpalive_stat = RTMQ_KPALIVE_STAT_SUCC

		/* 放入接收队列 */
		c.recv_chan <- p
	}
}

/******************************************************************************
 **函数名称: send_routine
 **功    能: 发送协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 将发送队列中的数据发送出去
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:10:36 #
 ******************************************************************************/
func (c *ProxyConn) send_routine() {
	log := c.svr.log
	c.svr.waitGroup.Add(1)
	defer func() {
		if err := recover(); nil != err {
			log.Error("Recv routine crashed! errmsg:%s", err)
		}
		c.Close()
		c.svr.waitGroup.Done()
	}()

	for {
		select {
		case <-c.svr.exit_chan:
			return

		case <-c.close_chan:
			return

		case p, _ := <-c.mesg_chan: /* 系统消息发送队列 */
			if _, err := c.conn.Write([]byte(p.head)); nil != err {
				return
			}
			if _, err := c.conn.Write([]byte(p.body)); nil != err {
				return
			}

		case p, _ := <-c.send_chan: /* 普通消息发送队列 */
			if _, err := c.conn.Write([]byte(p.head)); nil != err {
				return
			}
			if _, err := c.conn.Write([]byte(p.body)); nil != err {
				return
			}

		case <-time.After(30 * time.Second): /* 保活消息 */
			if c.kpalive_times > 3 &&
				RTMQ_KPALIVE_STAT_SENT == c.kpalive_stat {
				return
			}
			c.keepalive()
			continue
		}
	}
}

/******************************************************************************
 **函数名称: handle_routine
 **功    能: 工作协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从接收队列中获取数据并处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:10:36 #
 ******************************************************************************/
func (c *ProxyConn) handle_routine() {
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

		case p := <-c.recv_chan: /* 业务消息 */
			c.svr.OnMessage(c, p)
		}
	}
}

/******************************************************************************
 **函数名称: keepalive
 **功    能: 发送保活消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:10:36 #
 ******************************************************************************/
func (c *ProxyConn) keepalive() {
	svr := c.svr
	conf := svr.conf

	/* > 设置协议头 */
	head := &RtmqHeader{}

	head.cmd = RTMQ_CMD_KPALIVE_REQ
	head.nid = conf.NodeId
	head.flag = RTMQ_SYS_DATA
	head.length = 0
	head.chksum = RTMQ_CHKSUM_VAL

	/* > 申请内存空间 */
	p := &RtmqPacket{}
	p.head = make([]byte, RTMQ_HEAD_SIZE)

	rtmq_head_hton(head, p)

	c.kpalive_time = time.Now().Unix()
	c.kpalive_times += 1
	c.mesg_chan <- p
	c.kpalive_stat = RTMQ_KPALIVE_STAT_SENT
}

/* 链路鉴权请求 */
type RtmqAuthReq struct {
	usr    [RTMQ_USR_MAX_LEN]byte /* 用户名 */
	passwd [RTMQ_PWD_MAX_LEN]byte /* 登录密码 */
}

/* 链路鉴权应答 */
type RtmqAuthRsp struct {
	is_succ uint32 /* 鉴权是否成功(0:失败 1:成功) */
}

/* 设置鉴权请求 */
func rtmq_set_auth_req(conf *ProxyConf, p *RtmqPacket) {
	off := 0
	copy(p.body[:RTMQ_USR_MAX_LEN], []byte(conf.Usr))
	off += RTMQ_USR_MAX_LEN
	copy(p.body[off:off+RTMQ_PWD_MAX_LEN], []byte(conf.Passwd))
}

/******************************************************************************
 **函数名称: auth
 **功    能: 发送鉴权消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项: 系统内部消息放在mesg_chan队列中
 **作    者: # Qifeng.zou # 2016.10.30 22:07:19 #
 ******************************************************************************/
func (c *ProxyConn) auth() {
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
	p.head = make([]byte, RTMQ_HEAD_SIZE)
	p.body = make([]byte, head.length)

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
	binary.BigEndian.PutUint32(p.body[0:4], req.cmd) /* CMD */
}

/******************************************************************************
 **函数名称: subscribe
 **功    能: 发送订阅消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 根据ctx.reg映射表发送订阅请求
 **注意事项: 系统内部消息放在mesg_chan队列中
 **作    者: # Qifeng.zou # 2016.10.30 22:03:59 #
 ******************************************************************************/
func (c *ProxyConn) subscribe() {
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

		/* > 设置订阅数据 */
		req := &RtmqSubReq{}
		req.cmd = cmd

		/* > 申请内存空间 */
		p := &RtmqPacket{}
		p.head = make([]byte, RTMQ_HEAD_SIZE)
		p.body = make([]byte, head.length)

		rtmq_head_hton(head, p)
		rtmq_set_sub_req(req, p)

		c.mesg_chan <- p
	}
}

/******************************************************************************
 **函数名称: mesg_handler
 **功    能: 系统消息处理
 **输入参数:
 **     cmd: 消息类型
 **     p: 接收的数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 20:38:12 #
 ******************************************************************************/
func (c *ProxyConn) mesg_handler(cmd uint32, p *RtmqRecvPacket) bool {
	switch cmd {
	case RTMQ_CMD_LINK_AUTH_ACK:
		return c.auth_ack_handler(p)
	case RTMQ_CMD_KPALIVE_ACK:
		return c.keepalive_ack_handler(p)
	}
	return true
}

/******************************************************************************
 **函数名称: auth_ack_handler
 **功    能: 鉴权应答处理
 **输入参数:
 **     p: 数据包对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 20:38:12 #
 ******************************************************************************/
func (c *ProxyConn) auth_ack_handler(p *RtmqRecvPacket) bool {
	log := c.svr.log
	conf := c.svr.conf

	is_succ := binary.BigEndian.Uint32(p.body[:4])
	if 0 == is_succ {
		c.is_auth = false
		log.Error("Auth failed! usr:%s passwd:%s", conf.Usr, conf.Passwd)
		return false
	}

	log.Debug("Auth success! usr:%s passwd:%s", conf.Usr, conf.Passwd)

	c.is_auth = true
	return true
}

/******************************************************************************
 **函数名称: keepalive_ack_handler
 **功    能: 保活应答处理
 **输入参数:
 **     p: 数据包对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 20:38:12 #
 ******************************************************************************/
func (c *ProxyConn) keepalive_ack_handler(p *RtmqRecvPacket) bool {
	log := c.svr.log

	c.kpalive_times = 0
	c.kpalive_stat = RTMQ_KPALIVE_STAT_SUCC

	log.Debug("Keepalive success!")

	return true
}

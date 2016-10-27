package rtmq

import (
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	RTMQ_SSVR_NUM   = 10
	RTMQ_CHKSUM_VAL = 0x1FE23DC4
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
	flag   uint8  /* 消息标识(0:系统消息 1:业务消息) */
	length uint32 /* 报体长度 */
	chksum uint32 /* 校验值(固定为0x1FE23DE4) */
}

type RtmqRegCb func(cmd uint32, orig uint32, data []byte, length uint32, param *interface{})

/* 回调注册项 */
type RtmqRegItem struct {
	Cmd   uint32       /* 命令类型 */
	Proc  RtmqRegCb    /* 回调函数 */
	Param *interface{} /* 附加参数 */
}

type RtmqProxyServer struct {
	conf      *RtmqProxyConf   /* 配置数据 */
	send_chan chan *RtmqPacket /* 发送队列 */
	recv_chan chan *RtmqPacket /* 接收队列 */
	callback  ConnCallback     /* 连接回调 */
	exit_chan chan struct{}    /* 通知所有协程退出 */
	waitGroup *sync.WaitGroup  /* 用于等待所有协程 */
}

/* 上下文信息 */
type RtmqProxyCntx struct {
	conf   *RtmqProxyConf                  /* 配置数据 */
	reg    map[uint32]*RtmqRegItem         /* 回调注册 */
	server [RTMQ_SSVR_NUM]*RtmqProxyServer /* 服务对象 */
}

/* 初始化PROXY服务 */
func RtmqProxyInit(conf *RtmqProxyConf, log *logs.BeeLogger) *RtmqProxyCntx {
	var callback ConnCallback

	pxy := &RtmqProxyCntx{}

	pxy.conf = conf
	for idx := 0; idx < RTMQ_SSVR_NUM; idx += 1 {
		pxy.server[idx] = rtmq_proxy_server_init(conf, callback)
		if nil == pxy.server[idx] {
			return nil
		}
		go pxy.server[idx].StartConnector(3)
	}
	go rtmq_proxy_keepalive_routine(pxy) /* 保活协程 */

	return pxy
}

/* 发送保活消息 */
func rtmq_proxy_send_keepalive(send_chan chan *RtmqPacket) {
	req := &RtmqHeader{}

	req.cmd = RTMQ_CMD_KPALIVE_REQ
	req.flag = 0
	req.length = 0
	req.chksum = RTMQ_CHKSUM_VAL

	p := &RtmqPacket{}
	p.buff = make([]byte, RTMQ_HEAD_SIZE)

	rtmq_head_hton(req, p)

	send_chan <- p
}

/* 保活协程 */
func rtmq_proxy_keepalive_routine(pxy *RtmqProxyCntx) {
	for idx := 0; idx < RTMQ_SSVR_NUM; idx += 1 {
		server := pxy.server[idx]
		rtmq_proxy_send_keepalive(server.send_chan)
		time.Sleep(30)
	}
}

/* 初始化PROXY服务对象 */
func rtmq_proxy_server_init(conf *RtmqProxyConf, callback ConnCallback) *RtmqProxyServer {
	return &RtmqProxyServer{
		conf:      conf,
		callback:  callback,
		exit_chan: make(chan struct{}),
		send_chan: make(chan *RtmqPacket, 20000),
		recv_chan: make(chan *RtmqPacket, 20000),
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
		conn, err := s.callback.OnDial()
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

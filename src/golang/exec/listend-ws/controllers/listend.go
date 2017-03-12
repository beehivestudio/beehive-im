package controllers

import (
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"

	"beehive-im/src/golang/lib/chat_tab"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/lws"
	"beehive-im/src/golang/lib/rtmq"

	"beehive-im/src/golang/exec/listend-ws/controllers/conf"
)

const (
	LSND_SID2CID_LEN = 999
)

/* 连接状态定义 */
const (
	CONN_STATUS_READY  = 1 // 预备状态: 已建立好网络连接
	CONN_STATUS_CHECK  = 2 // 登录校验: 正在进行上线校验
	CONN_STATUS_LOGON  = 3 // 登录成功: 上线校验成功
	CONN_STATUS_KICK   = 4 // 连接被踢: 已加入到被踢列表
	CONN_STATUS_LOGOUT = 5 // 退出登录: 收到退出指令
	CONN_STATUS_CLOSE  = 6 // 连接关闭: 连接已经关闭 等待数据释放
)

/* 上行消息处理回调类型 */
type MesgCallBack func(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int

type MesgCallBackItem struct {
	cmd   uint32       /* 消息ID */
	proc  MesgCallBack /* 处理回调 */
	param interface{}  /* 附加参数 */
}

type MesgCallBackTab struct {
	sync.RWMutex                              /* 读写锁 */
	list         map[uint32]*MesgCallBackItem /* 消息处理回调(上行) */
}

/* SID->CID映射管理 */
type Sid2CidTab struct {
	tab [LSND_SID2CID_LEN]Sid2CidList
}

type Sid2CidList struct {
	sync.RWMutex                   /* 读写锁 */
	list         map[uint64]uint64 /* SID->CID映射 */
}

/* KICK信息 */
type LsndKickItem struct {
	cid uint64 /* 连接ID */
	ttl int64  /* 生命时间 */
}

/* LISTEND上下文 */
type LsndCntx struct {
	conf      *conf.LsndConf      /* 配置信息 */
	log       *logs.BeeLogger     /* 日志对象 */
	frwder    *rtmq.RtmqProxyCntx /* 代理对象 */
	callback  MesgCallBackTab     /* 处理回调 */
	chat      *chat_tab.ChatTab   /* 聊天关系组织表 */
	sid2cid   Sid2CidTab          /* SID->CID映射 */
	lws       *lws.LwsCntx        /* LWS环境 */
	protocol  *lws.Protocol       /* LWS.PROTOCOL */
	kick_list chan *LsndKickItem  /* 被踢列表 */
}

/* 会话扩展数据 */
type LsndSessionExtra struct {
	sid          uint64 /* 会话ID */
	sync.RWMutex        /* 读写锁 */
	cid          uint64 /* 连接ID */
	status       int    /* 连接状态(CONN_STATUS_READY...) */
}

/******************************************************************************
 **函数名称: LsndInit
 **功    能: 初始化对象
 **输入参数:
 **     conf: 配置信息
 **输出参数: NONE
 **返    回:
 **     ctx: 上下文
 **     err: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.08 22:42:49 #
 ******************************************************************************/
func LsndInit(conf *conf.LsndConf) (ctx *LsndCntx, err error) {
	ctx = &LsndCntx{
		conf: conf,
	}

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "websocket.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		ctx.log.Error("Initialize rtmq proxy failed!")
		return nil, errors.New("Initialize rtmq proxy failed!")
	}

	/* > 初始化其他结构 */
	ctx.callback.list = make(map[uint32]*MesgCallBackItem) /* 处理回调 */
	ctx.kick_list = make(chan *LsndKickItem, 100000)       /* 被踢列表 */
	ctx.sid_to_cid_init()                                  /* SID->CID映射 */

	/* > 聊天关系组织表 */
	ctx.chat = chat_tab.Init()

	/* > 初始化LWS模块 */
	lws_conf := &lws.Conf{
		Ip:       conf.WebSocket.Ip,
		Port:     conf.WebSocket.Port,
		Max:      conf.WebSocket.Max,
		Timeout:  conf.WebSocket.Timeout,
		SendqMax: conf.WebSocket.SendqMax,
	}

	ctx.lws = lws.Init(lws_conf, ctx.log)
	if nil == ctx.lws {
		ctx.log.Error("Initialize lws failed!")
		return nil, errors.New("Initialize lws failed!")
	}

	/* > 初始化LWS协议 */
	ctx.protocol = &lws.Protocol{
		Callback:          LsndLwsCallBack,             /* 处理回调 */
		PerPacketHeadSize: uint32(comm.MESG_HEAD_SIZE), /* 每个包的报头长度 */
		Param:             ctx,                         /* 附加参数 */
	}

	return ctx, nil
}

/******************************************************************************
 **函数名称: Register
 **功    能: 注册处理回调
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 注册回调函数
 **注意事项: 请在调用Launch()前完成此函数调用
 **作    者: # Qifeng.zou # 2017.02.09 23:10:38 #
******************************************************************************/
func (ctx *LsndCntx) Register() {
	ctx.UplinkRegister()    // 上行消息注册回调
	ctx.DownlinkRegister()  // 下行消息注册回调
	ctx.lws.Register("/im") // LWS路径注册回调
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动LISTEND服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *LsndCntx) Launch() {
	go ctx.Task()
	ctx.frwder.Launch()
	go ctx.lws.Launch(ctx.protocol)
}

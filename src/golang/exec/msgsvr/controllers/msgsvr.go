package controllers

import (
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/mongo"
	"beehive-im/src/golang/lib/rdb"
	"beehive-im/src/golang/lib/rtmq"

	"beehive-im/src/golang/exec/msgsvr/controllers/conf"
)

/* RID->NID映射表 */
type RidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* RID->NID映射表 */
}

/* 聊天室映射表 */
type RoomMap struct {
	node RidToNidMap /* RID->NID映射表 */
}

/* GID->NID映射表 */
type GidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* GID->NID映射表 */
}

/* 群组映射表 */
type GroupMap struct {
	node GidToNidMap /* GID->NID映射表 */
}

/* 私聊消息 */
type MesgChatItem struct {
	head *comm.MesgHeader /* 头部信息 */
	req  *mesg.MesgChat   /* 请求内容 */
	raw  []byte           /* 原始消息 */
}

/* 群组消息 */
type MesgGroupItem struct {
	head *comm.MesgHeader    /* 头部信息 */
	req  *mesg.MesgGroupChat /* 请求内容 */
	raw  []byte              /* 原始消息 */
}

/* 聊天室消息 */
type MesgRoomItem struct {
	head *comm.MesgHeader   /* 头部信息 */
	req  *mesg.MesgRoomChat /* 请求内容 */
	raw  []byte             /* 原始消息 */
}

/* MSGSVR上下文 */
type MsgSvrCntx struct {
	conf            *conf.MsgSvrConf    /* 配置信息 */
	log             *logs.BeeLogger     /* 日志对象 */
	frwder          *rtmq.Proxy         /* 代理对象 */
	redis           *redis.Pool         /* REDIS连接池 */
	mongo           *mongo.Pool         /* MONGO连接池 */
	room            RoomMap             /* 聊天室映射 */
	group           GroupMap            /* 群组映射 */
	room_mesg_chan  chan *MesgRoomItem  /* 聊天室消息存储队列 */
	group_mesg_chan chan *MesgGroupItem /* 组聊消息存储队列 */
	chat_chan       chan *MesgChatItem  /* 私聊消息存储队列 */
}

/******************************************************************************
 **函数名称: MsgSvrInit
 **功    能: 初始化对象
 **输入参数:
 **     conf: 配置信息
 **输出参数: NONE
 **返    回:
 **     ctx: 上下文
 **     err: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func MsgSvrInit(conf *conf.MsgSvrConf) (ctx *MsgSvrCntx, err error) {
	ctx = &MsgSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "msgsvr.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = rdb.CreatePool(conf.Redis.Addr, conf.Redis.Passwd, 512)
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s passwd:%s",
			conf.Redis.Addr, conf.Redis.Passwd)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > MONGO连接池 */
	ctx.mongo, err = mongo.CreatePool(conf.Mongo.Addr, conf.Mongo.Passwd)
	if nil != err {
		ctx.log.Error("Connect to mongo failed! addr:%s errmsg:%s",
			conf.Mongo.Addr, err.Error())
		return nil, errors.New("Connect to mongo failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		ctx.log.Error("Init rtmq proxy failed! addr:%s", conf.Frwder.RemoteAddr)
		return nil, err
	}

	/* > 初始化存储队列 */
	ctx.room_mesg_chan = make(chan *MesgRoomItem, 100000)
	ctx.group_mesg_chan = make(chan *MesgGroupItem, 100000)
	ctx.chat_chan = make(chan *MesgChatItem, 100000)

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
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) Register() {
	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_SYNC, MsgSvrSyncHandler, ctx)
	//ctx.frwder.Register(comm.CMD_P2P, MsgSvrP2pHandler, ctx)

	//ctx.frwder.Register(comm.CMD_P2P_ACK, MsgSvrP2pAckHandler, ctx)

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_CHAT, MsgSvrChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_CHAT_ACK, MsgSvrChatAckHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CHAT, MsgSvrGroupChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_CHAT_ACK, MsgSvrGroupChatAckHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, MsgSvrRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, MsgSvrRoomChatAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC, MsgSvrRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_ACK, MsgSvrRoomBcAckHandler, ctx)

	/* > 推送消息 */
	ctx.frwder.Register(comm.CMD_BC, MsgSvrBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_BC_ACK, MsgSvrBcAckHandler, ctx)
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动OLSVR服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) Launch() {
	go ctx.task()
	go ctx.update()
	ctx.frwder.Launch()
}

package controllers

import (
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/rtmq"
)

/* RID->NID映射表 */
type MsgSvrRidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* RID->NID映射表 */
}

/* GID->NID映射表 */
type MsgSvrGidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* GID->NID映射表 */
}

/* 私聊消息 */
type mesg_private_item struct {
	head *comm.MesgHeader  /* 头部信息 */
	req  *mesg.MesgPrvtMsg /* 请求内容 */
	raw  []byte            /* 原始消息 */
}

/* 群组消息 */
type mesg_group_item struct {
	head *comm.MesgHeader   /* 头部信息 */
	req  *mesg.MesgGroupMsg /* 请求内容 */
	raw  []byte             /* 原始消息 */
}

/* 聊天室消息 */
type mesg_room_item struct {
	head *comm.MesgHeader  /* 头部信息 */
	req  *mesg.MesgRoomMsg /* 请求内容 */
	raw  []byte            /* 原始消息 */
}

/* MSGSVR上下文 */
type MsgSvrCntx struct {
	conf           *MsgSvrConf         /* 配置信息 */
	log            *logs.BeeLogger     /* 日志对象 */
	frwder         *rtmq.RtmqProxyCntx /* 代理对象 */
	redis          *redis.Pool         /* REDIS连接池 */
	rid_to_nid_map MsgSvrRidToNidMap   /* RID->NID映射表 */
	gid_to_nid_map MsgSvrGidToNidMap   /* GID->NID映射表 */

	room_mesg_chan    chan *mesg_room_item    /* 聊天室消息存储队列 */
	group_mesg_chan   chan *mesg_group_item   /* 组聊消息存储队列 */
	private_mesg_chan chan *mesg_private_item /* 私聊消息存储队列 */
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
func MsgSvrInit(conf *MsgSvrConf) (ctx *MsgSvrCntx, err error) {
	ctx = &MsgSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "msgsvr.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.Redis.Addr)
			if nil != err {
				panic(err.Error())
				return nil, err
			}
			if 0 != len(conf.Redis.Passwd) {
				if _, err := c.Do("AUTH", conf.Redis.Passwd); nil != err {
					c.Close()
					panic(err.Error())
					return nil, err
				}
			}
			return c, err
		},
	}
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s passwd:%s",
			conf.Redis.Addr, conf.Redis.Passwd)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	/* > 初始化存储队列 */
	ctx.room_mesg_chan = make(chan *mesg_room_item, 100000)
	ctx.group_mesg_chan = make(chan *mesg_group_item, 100000)
	ctx.private_mesg_chan = make(chan *mesg_private_item, 100000)

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
	ctx.frwder.Register(comm.CMD_GROUP_MSG, MsgSvrGroupMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_MSG_ACK, MsgSvrGroupMsgAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_PRVT_MSG, MsgSvrPrivateMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_PRVT_MSG_ACK, MsgSvrPrvtMsgAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_BC_MSG, MsgSvrBcMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_BC_MSG_ACK, MsgSvrBcMsgAckHandler, ctx)

	//ctx.frwder.Register(comm.CMD_P2P_MSG, MsgSvrP2pMsgHandler, ctx)
	//ctx.frwder.Register(comm.CMD_P2P_MSG_ACK, MsgSvrP2pMsgAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_MSG, MsgSvrRoomMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_MSG_ACK, MsgSvrRoomMsgAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC_MSG, MsgSvrRoomBcMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_MSG_ACK, MsgSvrRoomBcMsgAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_SYNC_MSG, MsgSvrSyncMsgHandler, ctx)
	//ctx.frwder.Register(comm.CMD_SYNC_MSG_ACK, MsgSvrSyncMsgAckHandler, ctx)
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

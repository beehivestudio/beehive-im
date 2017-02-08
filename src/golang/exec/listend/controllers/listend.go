package controllers

import (
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/lws"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/rtmq"
)

/* RID->NID映射表 */
type LsndRidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* RID->NID映射表 */
}

/* GID->NID映射表 */
type LsndGidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* GID->NID映射表 */
}

/* 私聊消息 */
type mesg_chat_item struct {
	head *comm.MesgHeader /* 头部信息 */
	req  *mesg.MesgChat   /* 请求内容 */
	raw  []byte           /* 原始消息 */
}

/* 群组消息 */
type mesg_group_item struct {
	head *comm.MesgHeader    /* 头部信息 */
	req  *mesg.MesgGroupChat /* 请求内容 */
	raw  []byte              /* 原始消息 */
}

/* 聊天室消息 */
type mesg_room_item struct {
	head *comm.MesgHeader   /* 头部信息 */
	req  *mesg.MesgRoomChat /* 请求内容 */
	raw  []byte             /* 原始消息 */
}

/* MSGSVR上下文 */
type LsndCntx struct {
	conf           *LsndConf           /* 配置信息 */
	log            *logs.BeeLogger     /* 日志对象 */
	frwder         *rtmq.RtmqProxyCntx /* 代理对象 */
	redis          *redis.Pool         /* REDIS连接池 */
	rid_to_nid_map LsndRidToNidMap     /* RID->NID映射表 */
	gid_to_nid_map LsndGidToNidMap     /* GID->NID映射表 */

	room_mesg_chan  chan *mesg_room_item  /* 聊天室消息存储队列 */
	group_mesg_chan chan *mesg_group_item /* 组聊消息存储队列 */
	chat_chan       chan *mesg_chat_item  /* 私聊消息存储队列 */
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
func LsndInit(conf *LsndConf) (ctx *LsndCntx, err error) {
	ctx = &LsndCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "websocket.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.frwder, ctx.log)
	if nil == ctx.frwder {
		ctx.log.Error("Initialize rtmq proxy failed!")
		return nil, errors.New("Initialize rtmq proxy failed!")
	}

	/* > 初始化侦听模块 */
	ctx.lws = lws.Init(conf.lws, ctx.log)
	if nil == ctx.lws {
		ctx.log.Error("Initialize lws failed!")
		return nil, errors.New("Initialize lws failed!")
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
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *LsndCntx) Register() {
	ctx.conn_register()   /* 上行消息 */
	ctx.upconn_register() /* 下行消息 */
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
func (ctx *LsndCntx) Launch() {
	go ctx.task()
	ctx.lws.Launch()
	ctx.frwder.Launch()
}

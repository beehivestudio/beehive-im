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

/* 上行消息处理回调类型 */
type MesgCallback func(conn *LsndConnObj, cmd uint32, data interface{}, length uint32, param interface{}) int

type MesgCallbackObj struct {
	sync.RWLock
	callback map[uint32]MesgCallback /* 消息处理回调(上行) */
}

/* LISTEND上下文 */
type LsndCntx struct {
	conf     *LsndConf           /* 配置信息 */
	log      *logs.BeeLogger     /* 日志对象 */
	frwder   *rtmq.RtmqProxyCntx /* 代理对象 */
	callback MesgCallbackObj     /* 处理回调 */
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

	/* > 初始化LWS模块 */
	addr := fmt.Sprintf("%s:%d", conf.Access.Ip, conf.Access.Port)
	lws_conf := &lws.Conf{
		Ip:       conf.Websocket.Ip,
		Port:     conf.Websocket.Port,
		Max:      conf.Websocket.Max,
		Timeout:  conf.Websocket.Timeout,
		SendqMax: conf.Websocket.SendqMax,
	}

	ctx.lws = lws.Init(lws_conf, ctx.log)
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
 **作    者: # Qifeng.zou # 2017.02.09 23:10:38 #
 ******************************************************************************/
func (ctx *LsndCntx) Register() {
	////////////////////////////////////////////////////////////////////////////
	// WEBSOCKET注册回调(上行)
	/* > 通用消息 */
	ctx.callback.Register(comm.CMD_ONLINE_REQ, LsndOnlineReqHandler, ctx)
	ctx.callback.Register(comm.CMD_OFFLINE_REQ, LsndOfflineReqHandler, ctx)
	ctx.callback.Register(comm.CMD_SYNC, LsndOfflineReqHandler, ctx)

	/* > 私聊消息 */
	ctx.callback.Register(comm.CMD_CHAT, LsndChatHandler, ctx)
	ctx.callback.Register(comm.CMD_CHAT_ACK, LsndChatAckHandler, ctx)

	/* > 群聊消息 */
	ctx.callback.Register(comm.CMD_GROUP_CHAT, LsndGroupChatHandler, ctx)
	ctx.callback.Register(comm.CMD_GROUP_CHAT_ACK, LsndGroupChatAckHandler, ctx)

	/* > 聊天室消息 */
	ctx.callback.Register(comm.CMD_ROOM_CHAT, LsndRoomChatHandler, ctx)
	ctx.callback.Register(comm.CMD_ROOM_CHAT_ACK, LsndRoomChatAckHandler, ctx)

	ctx.callback.Register(comm.CMD_ROOM_BC, LsndRoomBcHandler, ctx)
	ctx.callback.Register(comm.CMD_ROOM_BC_ACK, LsndRoomBcAckHandler, ctx)

	/* > 推送消息 */
	ctx.callback.Register(comm.CMD_BC, LsndBcHandler, ctx)
	ctx.callback.Register(comm.CMD_BC_ACK, LsndBcAckHandler, ctx)

	////////////////////////////////////////////////////////////////////////////
	// FRWDER注册回调(下行)
	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndFrwderOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_OFFLINE_ACK, LsndFrwderP2pMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_SYNC_ACK, LsndSFrwderyncAckHandler, ctx)

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_CHAT, LsndFrwderChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_CHAT_ACK, LsndFrwderChatAckHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CHAT, LsndFrwderGroupChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_CHAT_ACK, LsndFrwderGroupChatAckHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndFrwderRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, LsndFrwderRoomChatAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndFrwderRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_ACK, LsndFrwderRoomBcAckHandler, ctx)

	/* > 推送消息 */
	ctx.frwder.Register(comm.CMD_BC, LsndFrwderBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_BC_ACK, LsndFrwderBcAckHandler, ctx)

	////////////////////////////////////////////////////////////////////////////
	/* > WS注册 */
	ctx.lws.Reigster("/im") /* WS注册路径 */
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

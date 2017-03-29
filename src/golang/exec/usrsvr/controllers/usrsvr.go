package controllers

import (
	"errors"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rtmq"

	"beehive-im/src/golang/exec/usrsvr/controllers/conf"
)

/* 侦听层列表 */
type UsrSvrLsndList struct {
	sync.RWMutex                                  /* 读写锁 */
	list         map[string](map[string][]string) /* 侦听层列表:map[TCP/WS](map[国家/地区](map([运营商名称][]IP列表))) */
}

type UsrSvrLsndNetWork struct {
	sync.RWMutex                         /* 读写锁 */
	types        map[int]*UsrSvrLsndList /* 侦听层类型:map[网络类型]UsrSvrLsndList */
}

/* 用户中心上下文 */
type UsrSvrCntx struct {
	conf    *conf.UsrSvrConf  /* 配置信息 */
	log     *logs.BeeLogger   /* 日志对象 */
	ipdict  *comm.IpDict      /* IP字典 */
	frwder  *rtmq.Proxy       /* 代理对象 */
	redis   *redis.Pool       /* REDIS连接池 */
	listend UsrSvrLsndNetWork /* 侦听层类型 */
}

var g_usrsvr_cntx *UsrSvrCntx /* 全局对象 */

/* 获取全局对象 */
func GetUsrSvrCtx() *UsrSvrCntx {
	return g_usrsvr_cntx
}

/* 设置全局对象 */
func SetUsrSvrCtx(ctx *UsrSvrCntx) {
	g_usrsvr_cntx = ctx
}

/******************************************************************************
 **函数名称: UsrSvrInit
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
func UsrSvrInit(conf *conf.UsrSvrConf) (ctx *UsrSvrCntx, err error) {
	ctx = &UsrSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "usrsvr.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > 加载IP字典 */
	ctx.ipdict, err = comm.LoadIpDict("../conf/ipdict.txt")
	if nil != err {
		return nil, err
	}

	/* > 创建侦听层列表 */
	ctx.listend.types = make(map[int]*UsrSvrLsndList)

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
		ctx.log.Error("Create redis pool failed! addr:%s", conf.Redis.Addr)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	SetUsrSvrCtx(ctx)

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
func (ctx *UsrSvrCntx) Register() {
	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE, UsrSvrOnlineHandler, ctx)
	ctx.frwder.Register(comm.CMD_OFFLINE, UsrSvrOfflineHandler, ctx)
	ctx.frwder.Register(comm.CMD_PING, UsrSvrPingHandler, ctx)
	ctx.frwder.Register(comm.CMD_SUB, UsrSvrSubHandler, ctx)
	ctx.frwder.Register(comm.CMD_UNSUB, UsrSvrUnsubHandler, ctx)
	ctx.frwder.Register(comm.CMD_ALLOC_SEQ, UsrSvrAllocSeqHandler, ctx)

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_BLACKLIST_ADD, UsrSvrBlacklistAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_BLACKLIST_DEL, UsrSvrBlacklistDelHandler, ctx)
	ctx.frwder.Register(comm.CMD_GAG_ADD, UsrSvrGagAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_GAG_DEL, UsrSvrGagDelHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CREAT, UsrSvrGroupCreatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_DISMISS, UsrSvrGroupDismissHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_JOIN, UsrSvrGroupJoinHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_QUIT, UsrSvrGroupQuitHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_INVITE, UsrSvrGroupInviteHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_KICK, UsrSvrGroupKickHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_GAG_ADD, UsrSvrGroupGagAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_GAG_DEL, UsrSvrGroupGagDelHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_BL_ADD, UsrSvrGroupBlacklistAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_BL_DEL, UsrSvrGroupBlacklistDelHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_MGR_ADD, UsrSvrGroupMgrAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_MGR_DEL, UsrSvrGroupMgrDelHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_USR_LIST, UsrSvrGroupUsrListHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CREAT, UsrSvrRoomCreatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_DISMISS, UsrSvrRoomDismissHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_JOIN, UsrSvrRoomJoinHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_QUIT, UsrSvrRoomQuitHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_KICK, UsrSvrRoomKickHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_USR_NUM, UsrSvrRoomUsrNumHandler, ctx)
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
func (ctx *UsrSvrCntx) Launch() {
	ctx.frwder.Launch()

	go ctx.start_task()
}

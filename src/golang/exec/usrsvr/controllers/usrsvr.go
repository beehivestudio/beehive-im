package controllers

import (
	"database/sql"
	"errors"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"

	"beehive-im/src/golang/lib/cache"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/dbase"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/mesg/seqsvr"
	"beehive-im/src/golang/lib/rtmq"
	"beehive-im/src/golang/lib/thrift_pool"

	"beehive-im/src/golang/exec/usrsvr/controllers/conf"
)

/* 侦听层列表 */
type UsrSvrLsndList struct {
	sync.RWMutex                                  /* 读写锁 */
	list         map[string](map[uint32][]string) /* 侦听层列表:map[TCP/WS](map[国家/地区](map([运营商ID][]IP列表))) */
}

type UsrSvrLsndNetWork struct {
	sync.RWMutex                         /* 读写锁 */
	types        map[int]*UsrSvrLsndList /* 侦听层类型:map[网络类型]UsrSvrLsndList */
}

/* 侦听层列表 */
type UsrSvrThriftClient struct {
}

/* 用户中心上下文 */
type UsrSvrCntx struct {
	conf        *conf.UsrSvrConf  /* 配置信息 */
	log         *logs.BeeLogger   /* 日志对象 */
	ipdict      *comm.IpDict      /* IP字典 */
	frwder      *rtmq.Proxy       /* 代理对象 */
	redis       *redis.Pool       /* REDIS连接池 */
	userdb      *sql.DB           /* USERDB数据库 */
	seqsvr_pool *thrift_pool.Pool /* SEQSVR连接池 */
	listend     UsrSvrLsndNetWork /* 侦听层类型 */
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

	///* > 创建侦听层列表 */
	ctx.listend.types = make(map[int]*UsrSvrLsndList)

	/* > REDIS连接池 */
	ctx.redis = cache.CreateRedisPool(conf.Redis.Addr, conf.Redis.Passwd, 2048)
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s", conf.Redis.Addr)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > MYSQL连接池 */
	auth := dbase.MySqlAuthStr(conf.UserDb.Usr, conf.UserDb.Passwd, conf.UserDb.Addr, conf.UserDb.Dbname)

	ctx.userdb, err = sql.Open("mysql", auth)
	if nil != err {
		ctx.log.Error("Connect mysql [%s] failed! errmsg:%s!", auth, err.Error())
		return nil, err
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	/* > SEQSVR连接池 */
	ctx.seqsvr_pool = ctx.seqsvr_pool_init(ctx.conf.Seqsvr.Addr)
	if nil == ctx.seqsvr_pool {
		ctx.log.Error("Connect seqsvr failed! errmsg:%s!", err.Error())
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

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_FRIEND_ADD, UsrSvrFriendAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_FRIEND_DEL, UsrSvrFriendDelHandler, ctx)
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
	ctx.frwder.Register(comm.CMD_ROOM_LSN_STAT, UsrSvrRoomLsnStatHandler, ctx)
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动USRSVR服务
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

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: seqsvr_pool_init
 **功    能: 创建SEQSVR连接池
 **输入参数:
 **     ctx: 全局对象
 **     addr: 服务端IP+PORT
 **输出参数: NONE
 **返    回: SEQSVR连接池
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.08 23:15:35 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) seqsvr_pool_init(addr string) *thrift_pool.Pool {
	pool := &thrift_pool.Pool{
		Dial: func() (interface{}, error) {
			socket, err := thrift.NewTSocket(addr)
			if nil != err {
				ctx.log.Error("Resolve address [%s] failed! errmsg:%s", addr, err.Error())
				return nil, err
			}

			transport := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
			protocol := thrift.NewTBinaryProtocolFactoryDefault()

			trans := transport.GetTransport(socket)
			client := seqsvr.NewSeqSvrThriftClientFactory(trans, protocol)
			if err := client.Transport.Open(); nil != err {
				ctx.log.Error("Opening socket [%s] failed! errmsg:%s", addr, err.Error())
				return nil, err
			}

			return client, nil
		},
		Close: func(client interface{}) error {
			client.(*seqsvr.SeqSvrThriftClient).Transport.Close()
			return nil
		},
		MaxIdle:     2048,
		MaxActive:   2048,
		IdleTimeout: 1800,
	}

	return pool
}

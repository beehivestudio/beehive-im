package controllers

import (
	"errors"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/mesg/seqsvr"
	"beehive-im/src/golang/lib/mongo"
	"beehive-im/src/golang/lib/rtmq"
	"beehive-im/src/golang/lib/thrift_pool"

	"beehive-im/src/golang/exec/chatroom/controllers/conf"
	"beehive-im/src/golang/exec/chatroom/models"
)

/* 侦听层字典 */
type ChatRoomLsndDictItem struct {
	sync.RWMutex                                  /* 读写锁 */
	list         map[string](map[uint32][]string) /* 侦听层列表:map[TCP/WS](map[国家/地区](map([运营商ID][]IP列表))) */
}

type ChatRoomLsndDict struct {
	sync.RWMutex                               /* 读写锁 */
	types        map[int]*ChatRoomLsndDictItem /* 侦听层类型:map[网络类型]ChatRoomLsndDictItem */
}

type ChatRoomLsndList struct {
	sync.RWMutex          /* 读写锁 */
	nodes        []uint32 /* 具体信息:[]结点ID */
}

type ChatRoomLsndData struct {
	dict ChatRoomLsndDict /* 侦听层映射 */
	list ChatRoomLsndList /* 侦听层列表 */
}

/* RID->NID映射表 */
type RidToNidMap struct {
	sync.RWMutex                     /* 读写锁 */
	m            map[uint64][]uint32 /* RID->NID映射表 */
}

/* 聊天室映射表 */
type RoomMap struct {
	node RidToNidMap /* RID->NID映射表 */
}

/* 聊天室消息 */
type MesgRoomItem struct {
	head *comm.MesgHeader   /* 头部信息 */
	req  *mesg.MesgRoomChat /* 请求内容 */
	raw  []byte             /* 原始消息 */
}

/* 用户中心上下文 */
type ChatRoomCntx struct {
	conf           *conf.ChatRoomConf  /* 配置信息 */
	log            *logs.BeeLogger     /* 日志对象 */
	frwder         *rtmq.Proxy         /* 代理对象 */
	cache          models.RoomCacheObj /* 缓存对象 */
	mongo          *mongo.Pool         /* MONGO连接池 */
	userdb         models.RoomDbObj    /* USERDB数据库 */
	seqsvr_pool    *thrift_pool.Pool   /* SEQSVR连接池 */
	listend        ChatRoomLsndData    /* 侦听层数据 */
	room           RoomMap             /* 聊天室映射 */
	room_mesg_chan chan *MesgRoomItem  /* 聊天室消息存储队列 */
}

var g_chatroom_cntx *ChatRoomCntx /* 全局对象 */

/* 获取全局对象 */
func GetRoomSvrCntx() *ChatRoomCntx {
	return g_chatroom_cntx
}

/* 设置全局对象 */
func SetRoomSvrCntx(ctx *ChatRoomCntx) {
	g_chatroom_cntx = ctx
}

/******************************************************************************
 **函数名称: ChatRoomInit
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
func ChatRoomInit(conf *conf.ChatRoomConf) (ctx *ChatRoomCntx, err error) {
	ctx = &ChatRoomCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "chatroom.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > 创建侦听层列表 */
	ctx.listend.dict.types = make(map[int]*ChatRoomLsndDictItem)

	/* > 初始化缓存 */
	err = ctx.cache.Init(conf.Redis.Addr, conf.Redis.Passwd)
	if nil != err {
		ctx.log.Error("Create redis pool failed! addr:%s", conf.Redis.Addr)
		return nil, err
	}

	/* > MONGO连接池 */
	ctx.mongo, err = mongo.CreatePool(conf.Mongo.Addr,
		conf.Mongo.Usr, conf.Mongo.Passwd,
		conf.Mongo.DbName, "maxPoolSize=1000", 30*time.Second)
	if nil != err {
		ctx.log.Error("Connect to mongo failed! addr:%s errmsg:%s",
			conf.Mongo.Addr, err.Error())
		return nil, err
	}

	/* > MYSQL连接池 */
	err = ctx.userdb.Init(conf.UserDb.Usr,
		conf.UserDb.Passwd, conf.UserDb.Addr, conf.UserDb.Dbname)
	if nil != err {
		ctx.log.Error("Connect to mysql failed! addr:%s errmsg:%s",
			conf.UserDb.Addr, err.Error())
		return nil, err
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		ctx.log.Error("Init rtmq proxy failed! addr:%s", conf.Frwder.RemoteAddr)
		return nil, err
	}

	/* > SEQSVR连接池 */
	ctx.seqsvr_pool = ctx.initSeqSvrPool(ctx.conf.Seqsvr.Addr)
	if nil == ctx.seqsvr_pool {
		ctx.log.Error("Connect seqsvr failed! errmsg:%s!", err.Error())
		return nil, err
	}

	/* > 消息队列 */
	ctx.room_mesg_chan = make(chan *MesgRoomItem, 100000)

	SetRoomSvrCntx(ctx)

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
func (ctx *ChatRoomCntx) Register() {
	ctx.frwder.Register(comm.CMD_ONLINE, ChatRoomOnlineHandler, ctx)
	ctx.frwder.Register(comm.CMD_OFFLINE, ChatRoomOfflineHandler, ctx)

	ctx.frwder.Register(comm.CMD_PING, ChatRoomPingHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_CREAT, ChatRoomCreatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_DISMISS, ChatRoomDismissHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_JOIN, ChatRoomJoinHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_QUIT, ChatRoomQuitHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_CHAT, ChatRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, ChatRoomChatAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC, ChatRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_ACK, ChatRoomBcAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_KICK, ChatRoomKickHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_LSN_STAT, ChatRoomLsnStatHandler, ctx)
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
func (ctx *ChatRoomCntx) Launch() {
	ctx.frwder.Launch()

	go ctx.task()
	go ctx.update()
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: initSeqSvrPool
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
func (ctx *ChatRoomCntx) initSeqSvrPool(addr string) *thrift_pool.Pool {
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

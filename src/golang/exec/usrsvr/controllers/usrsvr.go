package controllers

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rtmq"
)

/* 侦听层列表 */
type usrsvr_lsn_list struct {
	sync.RWMutex                                  /* 读写锁 */
	list         map[string](map[string][]string) /* 侦听层列表:map[国家/地区]map([运营商名称][]IP列表) */
}

/* 用户中心上下文 */
type UsrSvrCntx struct {
	conf    *UsrSvrConf         /* 配置信息 */
	log     *logs.BeeLogger     /* 日志对象 */
	ipdict  *comm.IpDict        /* IP字典 */
	frwder  *rtmq.RtmqProxyCntx /* 代理对象 */
	redis   *redis.Pool         /* REDIS连接池 */
	lsnlist usrsvr_lsn_list     /* 侦听层列表 */
}

var g_usrsvr *UsrSvrCntx /* 全局对象 */

/* 获取全局对象 */
func GetUsrSvrCtx() *UsrSvrCntx {
	return g_usrsvr
}

/* 设置全局对象 */
func SetUsrSvrCtx(ctx *UsrSvrCntx) {
	g_usrsvr = ctx
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
func UsrSvrInit(conf *UsrSvrConf) (ctx *UsrSvrCntx, err error) {
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
	ctx.frwder = rtmq.ProxyInit(&conf.frwder, ctx.log)
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
	ctx.frwder.Register(comm.CMD_BAN_ADD, UsrSvrBanAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_BAN_DEL, UsrSvrBanDelHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CREAT, UsrSvrGroupCreatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_DISMISS, UsrSvrGroupDismissHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_JOIN, UsrSvrGroupJoinHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_QUIT, UsrSvrGroupQuitHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_INVITE, UsrSvrGroupInviteHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_KICK, UsrSvrGroupKickHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_BAN_ADD, UsrSvrGroupBanAddHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_BAN_DEL, UsrSvrGroupBanDelHandler, ctx)
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

////////////////////////////////////////////////////////////////////////////////
/* 在线中心配置 */
type UsrSvrConf struct {
	NodeId   uint32             // 结点ID
	Port     int16              // HTTP侦听端口
	WorkPath string             // 工作路径(自动获取)
	AppPath  string             // 程序路径(自动获取)
	ConfPath string             // 配置路径(自动获取)
	Redis    UsrSvrRedisConf    // REDIS配置
	Mysql    UsrSvrMysqlConf    // MYSQL配置
	Mongo    UsrSvrMongoConf    // MONGO配置
	Cipher   string             // 私密密钥
	Log      log.LogConf        // 日志配置
	frwder   rtmq.RtmqProxyConf // RTMQ配置
}

/* 日志配置 */
type UsrSvrLogConf struct {
	Name  xml.Name `xml:"LOG"`        // 结点名
	Level string   `xml:"LEVEL,attr"` // 日志级别
	Path  string   `xml:"PATH,attr"`  // 日志路径
}

/* REDIS配置 */
type UsrSvrRedisConf struct {
	Name   xml.Name `xml:"REDIS"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* MYSQL配置 */
type UsrSvrMysqlConf struct {
	Name   xml.Name `xml:"MYSQL"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* MONGO配置 */
type UsrSvrMongoConf struct {
	Name   xml.Name `xml:"MONGO"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* 鉴权配置 */
type UsrSvrRtmqAuthConf struct {
	Name   xml.Name `xml:"AUTH"`        // 结点名
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* RTMQ代理配置 */
type UsrSvrRtmqProxyConf struct {
	Name        xml.Name           `xml:"FRWDER"`        // 结点名
	Auth        UsrSvrRtmqAuthConf `xml:"AUTH"`          // 鉴权信息
	RemoteAddr  string             `xml:"ADDR,attr"`     // 对端IP(IP+PROT)
	WorkerNum   uint32             `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32             `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32             `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 在线中心XML配置 */
type UsrSvrConfXmlData struct {
	Name   xml.Name            `xml:"USRSVR"`  // 根结点名
	Id     uint32              `xml:"ID,attr"` // 结点ID
	Port   int16               `xml:"PORT"`    // HTTP侦听端口
	Redis  UsrSvrRedisConf     `xml:"REDIS"`   // Redis配置
	Mysql  UsrSvrMysqlConf     `xml:"MYSQL"`   // Mysql配置
	Mongo  UsrSvrMongoConf     `xml:"MONGO"`   // Mongo配置
	Cipher string              `xml:"CIPHER"`  // 私密密钥
	Log    UsrSvrLogConf       `xml:"LOG"`     // 日志配置
	Frwder UsrSvrRtmqProxyConf `xml:"FRWDER"`  // RTMQ PROXY配置
}

/******************************************************************************
 **函数名称: LoadConf
 **功    能: 加载配置信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func (conf *UsrSvrConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "usrsvr.xml")

	return conf.conf_parse()
}

/******************************************************************************
 **函数名称: conf_parse
 **功    能: 解析配置信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 错误描述
 **实现描述: 加载配置并提取有效信息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func (conf *UsrSvrConf) conf_parse() (err error) {
	/* > 加载配置文件 */
	file, err := os.Open(conf.ConfPath)
	if nil != err {
		return err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if nil != err {
		return err
	}

	node := UsrSvrConfXmlData{}

	err = xml.Unmarshal(data, &node)
	if nil != err {
		return err
	}

	/* > 解析配置文件 */
	/* 结点ID */
	conf.NodeId = node.Id
	if 0 == conf.NodeId {
		return errors.New("Get node id failed!")
	}

	/* HTTP侦听端口(PORT) */
	conf.Port = node.Port
	if 0 == conf.Port {
		return errors.New("Get listen port failed!")
	}

	/* Redis配置 */
	conf.Redis.Addr = node.Redis.Addr
	if 0 == len(conf.Redis.Addr) {
		return errors.New("Get redis addr failed!")
	}

	conf.Redis.Usr = node.Redis.Usr
	if 0 == len(conf.Redis.Usr) {
		return errors.New("Get user name of redis failed!")
	}

	conf.Redis.Passwd = node.Redis.Passwd
	if 0 == len(conf.Redis.Passwd) {
		return errors.New("Get password of redis failed!")
	}

	/* MYSQL配置 */
	conf.Mysql.Addr = node.Mysql.Addr
	if 0 == len(conf.Mysql.Addr) {
		return errors.New("Get mysql addr failed!")
	}

	conf.Mysql.Usr = node.Mysql.Usr
	if 0 == len(conf.Mysql.Usr) {
		return errors.New("Get user name of mysql failed!")
	}

	conf.Mysql.Passwd = node.Mysql.Passwd
	if 0 == len(conf.Mysql.Passwd) {
		return errors.New("Get password of mysql failed!")
	}

	/* MONGO配置 */
	conf.Mongo.Addr = node.Mongo.Addr
	if 0 == len(conf.Mongo.Addr) {
		return errors.New("Get mongo addr failed!")
	}

	conf.Mongo.Usr = node.Mongo.Usr
	if 0 == len(conf.Mongo.Usr) {
		return errors.New("Get user name of mongo failed!")
	}

	conf.Mongo.Passwd = node.Mongo.Passwd
	if 0 == len(conf.Mongo.Passwd) {
		return errors.New("Get password of mongo failed!")
	}

	/* > 私密密钥 */
	conf.Cipher = node.Cipher
	if 0 == len(conf.Cipher) {
		return errors.New("Get chiper failed!")
	}

	/* 日志配置 */
	conf.Log.Level = log.GetLevel(node.Log.Level)

	conf.Log.Path = node.Log.Path
	if 0 == len(conf.Log.Path) {
		return errors.New("Get log path failed!")
	}

	/* FRWDER配置 */
	conf.frwder.NodeId = conf.NodeId

	/* 鉴权信息 */
	conf.frwder.Usr = node.Frwder.Auth.Usr
	conf.frwder.Passwd = node.Frwder.Auth.Passwd
	if 0 == len(conf.frwder.Usr) || 0 == len(conf.frwder.Passwd) {
		return errors.New("Get auth conf failed!")
	}

	/* 转发层(IP+PROT) */
	conf.frwder.RemoteAddr = node.Frwder.RemoteAddr
	if 0 == len(conf.frwder.RemoteAddr) {
		return errors.New("Get frwder addr failed!")
	}

	/* 发送队列长度 */
	conf.frwder.SendChanLen = node.Frwder.SendChanLen
	if 0 == conf.frwder.SendChanLen {
		return errors.New("Get send channel length failed!")
	}

	/* 接收队列长度 */
	conf.frwder.RecvChanLen = node.Frwder.RecvChanLen
	if 0 == conf.frwder.RecvChanLen {
		return errors.New("Get recv channel length failed!")
	}

	/* 协程数 */
	conf.frwder.WorkerNum = node.Frwder.WorkerNum
	if 0 == conf.frwder.WorkerNum {
		return errors.New("Get worker number failed!")
	}

	return nil
}

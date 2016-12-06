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

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/log"
	"chat/src/golang/lib/rtmq"
)

/* HTTPSVR配置 */
type HttpSvrConf struct {
	Port      int                // 端口号
	NodeId    uint32             // 结点ID
	WorkPath  string             // 工作路径(自动获取)
	AppPath   string             // 程序路径(自动获取)
	ConfPath  string             // 配置路径(自动获取)
	RedisAddr string             // Redis地址(IP+PORT)
	Cipher    string             // 私密密钥
	Log       log.LogConf        // 日志配置
	frwder    rtmq.RtmqProxyConf // RTMQ配置
}

/* 侦听层列表 */
type httpsvr_lsn_list struct {
	sync.RWMutex                                  /* 读写锁 */
	list         map[string](map[string][]string) /* 侦听层列表:map[国家/地区]运营商名称 */
}

/* HTTPSVR上下文 */
type HttpSvrCntx struct {
	conf    *HttpSvrConf        /* 配置信息 */
	log     *logs.BeeLogger     /* 日志对象 */
	ipdict  *comm.IpDict        /* IP字典 */
	lsnlist httpsvr_lsn_list    /* 侦听层列表:map[国家/地区]运营商名称 */
	frwder  *rtmq.RtmqProxyCntx /* 代理对象 */
	redis   *redis.Pool         /* REDIS连接池 */
}

var g_httpsvr *HttpSvrCntx /* 全局对象 */

/* 获取全局对象 */
func GetHttpCtx() *HttpSvrCntx {
	return g_httpsvr
}

/* 设置全局对象 */
func SetHttpCtx(ctx *HttpSvrCntx) {
	g_httpsvr = ctx
}

/******************************************************************************
 **函数名称: HttpSvrInit
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
func HttpSvrInit(conf *HttpSvrConf) (ctx *HttpSvrCntx, err error) {
	ctx = &HttpSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "httpsvr.log")
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
			c, err := redis.Dial("tcp", conf.RedisAddr)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s", conf.RedisAddr)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	SetHttpCtx(ctx)

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
 **作    者: # Qifeng.zou # 2016.11.20 00:29:41 #
 ******************************************************************************/
func (ctx *HttpSvrCntx) Register() {
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动HTTPSVR服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.20 00:27:03 #
 ******************************************************************************/
func (ctx *HttpSvrCntx) Launch() {
	ctx.frwder.Launch()

	go ctx.start_task()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 侦听配置 */
type HttpSvrConfListenXmlData struct {
	Name xml.Name `xml:"LISTEN"`  // 结点名
	Ip   string   `xml:"IP,attr"` // 格式:"网卡IP:端口号"
}

/* 日志配置 */
type HttpSvrConfLogXmlData struct {
	Name  xml.Name `xml:"LOG"`        // 结点名
	Level string   `xml:"LEVEL,attr"` // 日志级别
	Path  string   `xml:"PATH,attr"`  // 日志路径
}

/* 鉴权配置 */
type HttpSvrConfRtmqAuthXmlData struct {
	Name   xml.Name `xml:"AUTH"`        // 结点名
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* RTMQ代理配置 */
type HttpSvrConfRtmqProxyXmlData struct {
	Name        xml.Name                   `xml:"FRWDER"`        // 结点名
	Auth        HttpSvrConfRtmqAuthXmlData `xml:"AUTH"`          // 鉴权信息
	RemoteAddr  string                     `xml:"REMOTE-ADDR"`   // 对端IP(IP+PROT)
	WorkerNum   uint32                     `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32                     `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32                     `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 在线中心XML配置 */
type HttpSvrConfXmlData struct {
	Name      xml.Name                    `xml:"HTTPSVR"`    // 根结点名
	Id        uint32                      `xml:"ID,attr"`    // 结点ID
	Port      int                         `xml:"PORT"`       // 侦听端口
	RedisAddr string                      `xml:"REDIS-ADDR"` // Redis地址(IP+PORT)
	Cipher    string                      `xml:"CIPHER"`     // 私密密钥
	Log       HttpSvrConfLogXmlData       `xml:"LOG"`        // 日志配置
	Frwder    HttpSvrConfRtmqProxyXmlData `xml:"FRWDER"`     // RTMQ PROXY配置
}

/******************************************************************************
 **函数名称: LoadConf
 **功    能: 加载配置信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func (conf *HttpSvrConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "httpsvr.xml")

	return conf.conf_parse()
}

/******************************************************************************
 **函数名称: conf_parse
 **功    能: 解析配置信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 日志对象
 **实现描述: 加载配置并提取有效信息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func (conf *HttpSvrConf) conf_parse() (err error) {
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

	node := HttpSvrConfXmlData{}

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

	/* 侦听端口(PORT) */
	conf.Port = node.Port
	if 0 == conf.Port {
		return errors.New("Get listen port failed!")
	}

	/* Redis地址(IP+PORT) */
	conf.RedisAddr = node.RedisAddr
	if 0 == len(conf.RedisAddr) {
		return errors.New("Get redis addr failed!")
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

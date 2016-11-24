package controllers

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/log"
	"chat/src/golang/lib/rtmq"
)

/******************************************************************************
 **函数名称: alloc_sid
 **功    能: 申请会话SID
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     sid: 会话SID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 10:48:40 #
 ******************************************************************************/
func (ctx *HttpSvrCntx) alloc_sid() (sid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	for {
		sid, err := redis.Uint64(rds.Do("INCRBY", comm.CHAT_KEY_SID_INCR, 1))
		if nil != err {
			return 0, err
		} else if 0 == sid {
			continue
		}
		return sid, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 在线中心配置 */
type HttpSvrConf struct {
	Port      int32              // 端口号
	NodeId    uint32             // 结点ID
	Listen    string             // 侦听配置
	WorkPath  string             // 工作路径(自动获取)
	AppPath   string             // 程序路径(自动获取)
	ConfPath  string             // 配置路径(自动获取)
	RedisAddr string             // Redis地址(IP+PORT)
	Cipher    string             // 私密密钥
	Log       log.LogConf        // 日志配置
	frwder    rtmq.RtmqProxyConf // RTMQ配置
}

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
	Listen    HttpSvrConfListenXmlData    `xml:"LISTEN"`     // 侦听(网卡IP+端口)
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

	/* 侦听配置(IP:PORT) */
	conf.Listen = node.Listen.Ip
	if 0 == len(conf.Listen) {
		return errors.New("Get listen failed!")
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

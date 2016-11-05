package ctrl

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"chat/src/golang/lib/log"
	"chat/src/golang/lib/rtmq"
)

/* 在线中心配置 */
type UsrSvrConf struct {
	NodeId    uint32             // 结点ID
	WorkPath  string             // 工作路径(自动获取)
	AppPath   string             // 程序路径(自动获取)
	ConfPath  string             // 配置路径(自动获取)
	RedisAddr string             // Redis地址(IP+PORT)
	SecretKey string             // 私密密钥
	Log       log.LogConf        // 日志配置
	proxy     rtmq.RtmqProxyConf // RTMQ配置
}

/* 日志配置 */
type UsrSvrConfLogXmlData struct {
	Name  xml.Name `xml:"LOG"`        // 结点名
	Level string   `xml:"LEVEL,attr"` // 日志级别
	Path  string   `xml:"PATH,attr"`  // 日志路径
}

/* 鉴权配置 */
type UsrSvrConfRtmqAuthXmlData struct {
	Name   xml.Name `xml:"AUTH"`        // 结点名
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* RTMQ代理配置 */
type UsrSvrConfRtmqProxyXmlData struct {
	Name        xml.Name                  `xml:"RTMQ-PROXY"`    // 结点名
	Auth        UsrSvrConfRtmqAuthXmlData `xml:"AUTH"`          // 鉴权信息
	RemoteAddr  string                    `xml:"REMOTE-ADDR"`   // 对端IP(IP+PROT)
	WorkerNum   uint32                    `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32                    `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32                    `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 在线中心XML配置 */
type UsrSvrConfXmlData struct {
	Name      xml.Name                   `xml:"USRSVR"`     // 根结点名
	Id        uint32                     `xml:"ID,attr"`    // 结点ID
	RedisAddr string                     `xml:"REDIS-ADDR"` // Redis地址(IP+PORT)
	SecretKey string                     `xml:"SECRET-KEY"` // 私密密钥
	Log       UsrSvrConfLogXmlData       `xml:"LOG"`        // 日志配置
	RtmqProxy UsrSvrConfRtmqProxyXmlData `xml:"RTMQ-PROXY"` // RTMQ PROXY配置
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
 **     err: 日志对象
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

	/* Redis地址(IP+PORT) */
	conf.RedisAddr = node.RedisAddr
	if 0 == len(conf.RedisAddr) {
		return errors.New("Get redis addr failed!")
	}

	/* > 私密密钥 */
	conf.SecretKey = node.SecretKey
	if 0 == len(conf.SecretKey) {
		return errors.New("Get secret key failed!")
	}

	/* 日志配置 */
	conf.Log.Level = log.GetLevel(node.Log.Level)

	conf.Log.Path = node.Log.Path
	if 0 == len(conf.Log.Path) {
		return errors.New("Get log path failed!")
	}

	/* RTMQ-PROXY配置 */
	conf.proxy.NodeId = conf.NodeId

	/* 鉴权信息 */
	conf.proxy.Usr = node.RtmqProxy.Auth.Usr
	conf.proxy.Passwd = node.RtmqProxy.Auth.Passwd
	if 0 == len(conf.proxy.Usr) || 0 == len(conf.proxy.Passwd) {
		return errors.New("Get auth conf failed!")
	}

	/* 转发层(IP+PROT) */
	conf.proxy.RemoteAddr = node.RtmqProxy.RemoteAddr
	if 0 == len(conf.proxy.RemoteAddr) {
		return errors.New("Get frwder addr failed!")
	}

	/* 发送队列长度 */
	conf.proxy.SendChanLen = node.RtmqProxy.SendChanLen
	if 0 == conf.proxy.SendChanLen {
		return errors.New("Get send channel length failed!")
	}

	/* 接收队列长度 */
	conf.proxy.RecvChanLen = node.RtmqProxy.RecvChanLen
	if 0 == conf.proxy.RecvChanLen {
		return errors.New("Get recv channel length failed!")
	}

	/* 协程数 */
	conf.proxy.WorkerNum = node.RtmqProxy.WorkerNum
	if 0 == conf.proxy.WorkerNum {
		return errors.New("Get worker number failed!")
	}

	return nil
}

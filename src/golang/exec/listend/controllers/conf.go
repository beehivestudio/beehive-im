package controllers

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/lws"
	"beehive-im/src/golang/lib/rtmq"
)

/* 侦听层配置 */
type LsndConf struct {
	NodeId    uint32                  // 结点ID
	WorkPath  string                  // 工作路径(自动获取)
	AppPath   string                  // 程序路径(自动获取)
	ConfPath  string                  // 配置路径(自动获取)
	Log       log.LogConf             // 日志配置
	Operator  LsndConfOperatorXmlData // 运营商信息
	Websocket lws.Conf                // WEBSOCKET配置
	frwder    rtmq.RtmqProxyConf      // RTMQ配置
}

/* 运营商配置 */
type LsndConfOperatorXmlData struct {
	Name   xml.Name `xml:"OPERATOR"`    // 结点名
	Nation string   `xml:"NATION,attr"` // 所属国家
	Name   string   `xml:"NAME,attr"`   // 运营商名称
}

/* 分发队列配置 */
type LsndConfDistqXmlData struct {
	Name xml.Name `xml:"OPERATOR"`  // 结点名
	Num  uint32   `xml:"NUM,attr"`  // 队列个数
	Max  uint32   `xml:"MAX,attr"`  // 队列长度
	Size uint32   `xml:"SIZE,attr"` // 队列SIZE
}

/* 日志配置 */
type LsndConfLogXmlData struct {
	Name  xml.Name `xml:"LOG"`        // 结点名
	Level string   `xml:"LEVEL,attr"` // 日志级别
	Path  string   `xml:"PATH,attr"`  // 日志目录
}

/* 鉴权配置 */
type LsndConfRtmqAuthXmlData struct {
	Name   xml.Name `xml:"AUTH"`        // 结点名
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* WEBSOCKET-CONNECTIONS配置 */
type LsndConfWsConncetionsXmlData struct {
	Name    xml.Name `xml:"CONNECTIONS"` // 结点名
	Max     uint32   `xml:"MAX"`         // 最大连接数
	Timeout uint32   `xml:"TIMEOUT"`     // 连接超时时间
}

/* WEBSOCKET-SENDQ配置 */
type LsndConfWsSendqXmlData struct {
	Name xml.Name `xml:"SENDQ"` // 结点名
	Max  uint32   `xml:"MAX"`   // 队列长度
	Size uint32   `xml:"SIZE"`  // 队列SIZE
}

/* WEBSOCKET代理配置 */
type LsndConfWebsocketXmlData struct {
	Name        xml.Name                     `xml:"WEBSOCKET"`   // 结点名
	Ip          string                       `xml:"IP,attr"`     // 对端IP
	Port        uint32                       `xml:"PORT,attr"`   // 对端PORT
	Connections LsndConfWsConncetionsXmlData `xml:"CONNECTIONS"` // 连接配置
	Sendq       LsndConfWsSendqXmlData       `xml:"SENDQ"`       // 发送队列配置
}

/* FRWDER代理配置 */
type LsndConfRtmqProxyXmlData struct {
	Name        xml.Name                `xml:"FRWDER"`        // 结点名
	RemoteAddr  string                  `xml:"ADDR,attr"`     // 对端IP(IP+PROT)
	Auth        LsndConfRtmqAuthXmlData `xml:"AUTH"`          // 鉴权信息
	WorkerNum   uint32                  `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32                  `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32                  `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 侦听层XML配置 */
type LsndConfXmlData struct {
	Name      xml.Name                 `xml:"LISTEND"`   // 根结点名
	Id        uint32                   `xml:"ID,attr"`   // 结点ID
	Log       LsndConfLogXmlData       `xml:"LOG"`       // 日志配置
	Operator  LsndConfOperatorXmlData  `xml:"OPERATOR"`  // 运营商信息
	Websocket LsndConfWebsocketXmlData `xml:"WEBSOCKET"` // WEBSOCKET配置
	Frwder    LsndConfRtmqProxyXmlData `xml:"FRWDER"`    // RTMQ PROXY配置
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
func (conf *LsndConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "listend.xml")

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
func (conf *LsndConf) conf_parse() (err error) {
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

	node := LsndConfXmlData{}

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

	/* > 日志配置 */
	conf.Log.Level = log.GetLevel(node.Log.Level) // 日志级别
	conf.Log.Path = node.Log.Path                 // 日志路径
	if !len(conf.Log.Path) {
		return errors.New("Get log path failed!")
	}

	/* > 运营商配置 */
	conf.Operator.Nation = node.Operator.Nation // 所属国家
	conf.Operator.Name = node.Operator.Name     // 运营商名称
	if !len(conf.Operator.Nation) || !len(conf.Operator.Name) {
		return errors.New("Get operator information failed!")
	}

	/* > 侦听配置 */
	conf.Websocket.Ip = node.Websocket.Ip                       // IP地址
	conf.Websocket.Port = node.Websocket.Port                   // 端口号
	conf.Websocket.Max = node.Websocket.Connections.Max         // 最大连接限制
	conf.Websocket.Timeout = node.Websocket.Connections.Timeout // 连接超时时间
	conf.Websocket.SendqMax = node.Websocket.Sendq.Max          // 发送队列长度

	/* > FRWDER配置 */
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

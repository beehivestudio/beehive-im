package conf

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"

	"beehive-im/src/golang/lib/log"
)

/* 运营商配置 */
type LsndConfOperatorXmlData struct {
	Id     uint32 `xml:"ID,attr"`     // 运营商ID
	Nation string `xml:"NATION,attr"` // 所属国家
}

/* 分发队列配置 */
type LsndConfDistqXmlData struct {
	Num  uint32 `xml:"NUM,attr"`  // 队列个数
	Max  uint32 `xml:"MAX,attr"`  // 队列长度
	Size uint32 `xml:"SIZE,attr"` // 队列SIZE
}

/* 日志配置 */
type LsndConfLogXmlData struct {
	Level string `xml:"LEVEL,attr"` // 日志级别
	Path  string `xml:"PATH,attr"`  // 日志目录
}

/* 鉴权配置 */
type LsndConfRtmqAuthXmlData struct {
	Usr    string `xml:"USR,attr"`    // 用户名
	Passwd string `xml:"PASSWD,attr"` // 登录密码
}

/* WEBSOCKET-CONNECTIONS配置 */
type LsndConfWsConncetionsXmlData struct {
	Max     uint32 `xml:"MAX"`     // 最大连接数
	Timeout uint32 `xml:"TIMEOUT"` // 连接超时时间
}

/* WEBSOCKET-SENDQ配置 */
type LsndConfWsSendqXmlData struct {
	Max  uint32 `xml:"MAX"`  // 队列长度
	Size uint32 `xml:"SIZE"` // 队列SIZE
}

/* WEBSOCKET代理配置 */
type LsndConfWebsocketXmlData struct {
	Ip          string                       `xml:"IP,attr"`     // 对端IP
	Port        uint32                       `xml:"PORT,attr"`   // 对端PORT
	Connections LsndConfWsConncetionsXmlData `xml:"CONNECTIONS"` // 连接配置
	Sendq       LsndConfWsSendqXmlData       `xml:"SENDQ"`       // 发送队列配置
}

/* FRWDER代理配置 */
type LsndConfRtmqProxyXmlData struct {
	RemoteAddr  string                  `xml:"ADDR,attr"`     // 对端IP(IP+PROT)
	Auth        LsndConfRtmqAuthXmlData `xml:"AUTH"`          // 鉴权信息
	WorkerNum   uint32                  `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32                  `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32                  `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 侦听层XML配置 */
type LsndConfXmlData struct {
	Id        uint32                   `xml:"ID,attr"`   // 结点ID
	Gid       uint32                   `xml:"GID,attr"`  // 分组ID
	Log       LsndConfLogXmlData       `xml:"LOG"`       // 日志配置
	Operator  LsndConfOperatorXmlData  `xml:"OPERATOR"`  // 运营商信息
	WebSocket LsndConfWebsocketXmlData `xml:"WEBSOCKET"` // WEBSOCKET配置
	Frwder    LsndConfRtmqProxyXmlData `xml:"FRWDER"`    // RTMQ PROXY配置
}

/******************************************************************************
 **函数名称: parse
 **功    能: 解析配置信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 错误描述
 **实现描述: 加载配置并提取有效信息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func (conf *LsndConf) parse() (err error) {
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
	conf.Id = node.Id
	if 0 == conf.Id {
		return errors.New("Get node id failed!")
	}

	/* 分组ID */
	conf.Gid = node.Gid
	if 0 == conf.Gid {
		return errors.New("Get gid failed!")
	}

	/* > 日志配置 */
	conf.Log.Level = log.GetLevel(node.Log.Level) // 日志级别
	conf.Log.Path = node.Log.Path                 // 日志路径
	if 0 == len(conf.Log.Path) {
		return errors.New("Get log path failed!")
	}

	/* > 运营商配置 */
	conf.Operator.Id = node.Operator.Id         // 运营商ID
	conf.Operator.Nation = node.Operator.Nation // 所属国家
	if 0 == len(conf.Operator.Nation) {
		return errors.New("Get operator information failed!")
	}

	/* > 侦听配置 */
	conf.WebSocket.Ip = node.WebSocket.Ip                       // IP地址
	conf.WebSocket.Port = node.WebSocket.Port                   // 端口号
	conf.WebSocket.Max = node.WebSocket.Connections.Max         // 最大连接限制
	conf.WebSocket.Timeout = node.WebSocket.Connections.Timeout // 连接超时时间
	conf.WebSocket.SendqMax = node.WebSocket.Sendq.Max          // 发送队列长度

	/* > FRWDER配置 */
	conf.Frwder.Id = conf.Id
	conf.Frwder.Gid = conf.Gid

	/* 鉴权信息 */
	conf.Frwder.Usr = node.Frwder.Auth.Usr
	conf.Frwder.Passwd = node.Frwder.Auth.Passwd
	if 0 == len(conf.Frwder.Usr) || 0 == len(conf.Frwder.Passwd) {
		return errors.New("Get auth conf failed!")
	}

	/* 转发层(IP+PROT) */
	conf.Frwder.RemoteAddr = node.Frwder.RemoteAddr
	if 0 == len(conf.Frwder.RemoteAddr) {
		return errors.New("Get Frwder addr failed!")
	}

	/* 发送队列长度 */
	conf.Frwder.SendChanLen = node.Frwder.SendChanLen
	if 0 == conf.Frwder.SendChanLen {
		return errors.New("Get send channel length failed!")
	}

	/* 接收队列长度 */
	conf.Frwder.RecvChanLen = node.Frwder.RecvChanLen
	if 0 == conf.Frwder.RecvChanLen {
		return errors.New("Get recv channel length failed!")
	}

	/* 协程数 */
	conf.Frwder.WorkerNum = node.Frwder.WorkerNum
	if 0 == conf.Frwder.WorkerNum {
		return errors.New("Get worker number failed!")
	}

	return nil
}

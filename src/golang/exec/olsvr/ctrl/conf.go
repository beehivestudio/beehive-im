package ctrl

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"chat/src/golang/lib/rtmq"
)

/* 在线中心配置 */
type OlSvrConf struct {
	NodeId     uint32             // 结点ID
	WorkPath   string             // 工作路径(自动获取)
	AppPath    string             // 程序路径(自动获取)
	ConfPath   string             // 配置路径(自动获取)
	RedisAddr  string             // Redis地址(IP+PORT)
	LogPath    string             // 日志路径
	rtmq_proxy rtmq.RtmqProxyConf // RTMQ配置
}

/* 在线中心XML配置 */
type OlSvrXmlNode struct {
	Name      xml.Name              `xml:OLSVR`     // 根结点名
	RedisAddr string                `xml:RedisAddr` // Redis地址(IP+PORT)
	LogPath   string                `xml:LogPath`   // 日志路径
	RtmqProxy OlSvrRtmqProxyXmlNode `xml:RtmqProxy` // RTMQ PROXY配置
}

type OlSvrRtmqProxyXmlNode struct {
	Name        xml.Name `xml:RtmqProxy`   // 结点名
	RemoteAddr  string   `xml:RemoteAddr`  // 对端IP(IP+PROT)
	WorkerNum   uint32   `xml:WorkerNum`   // 协程数
	SendChanLen uint32   `xml:SendChanLen` // 发送队列长度
	RecvChanLen uint32   `xml:RecvChanLen` // 接收队列长度
}

/* 加载配置信息 */
func (conf *OlSvrConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "olsvr.xml")

	return conf.conf_parse()
}

/* 解析配置信息 */
func (conf *OlSvrConf) conf_parse() (err error) {
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

	node := OlSvrXmlNode{}

	err = xml.Unmarshal(data, &node)
	if nil != err {
		return err
	}

	/* > 解析配置文件 */
	/* Redis地址(IP+PORT) */
	conf.RedisAddr = node.RedisAddr
	if 0 == len(conf.RedisAddr) {
		return errors.New("Get redis addr failed!")
	}

	/* 日志路径 */
	conf.LogPath = node.LogPath
	if 0 == len(conf.LogPath) {
		return errors.New("Get log path failed!")
	}

	/* RTMQ-PROXY配置 */
	/* 转发层(IP+PROT) */
	conf.rtmq_proxy.RemoteAddr = node.RtmqProxy.RemoteAddr
	if 0 == len(conf.rtmq_proxy.RemoteAddr) {
		return errors.New("Get frwder addr failed!")
	}

	/* 发送队列长度 */
	conf.rtmq_proxy.SendChanLen = node.RtmqProxy.SendChanLen
	if 0 == conf.rtmq_proxy.SendChanLen {
		return errors.New("Get send channel length failed!")
	}

	/* 接收队列长度 */
	conf.rtmq_proxy.RecvChanLen = node.RtmqProxy.RecvChanLen
	if 0 == conf.rtmq_proxy.RecvChanLen {
		return errors.New("Get recv channel length failed!")
	}

	/* 协程数 */
	conf.rtmq_proxy.WorkerNum = node.RtmqProxy.WorkerNum
	if 0 == conf.rtmq_proxy.WorkerNum {
		return errors.New("Get worker number failed!")
	}

	return nil
}

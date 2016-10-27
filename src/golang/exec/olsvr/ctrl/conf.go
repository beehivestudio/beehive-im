package ctrl

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"chat/src/golang/lib/rtmq"
)

type OlSvrConf struct {
	NodeId      uint32              // 结点ID
	WorkPath    string              // 工作路径(自动获取)
	AppPath     string              // 程序路径(自动获取)
	ConfPath    string              // 配置路径(自动获取)
	FrwderAddr  string              // 转发层(IP+PROT)
	SendChanLen uint32              // 发送队列长度
	RecvChanLen uint32              // 接收队列长度
	WorkerNum   uint16              // 协程数
	RedisAddr   string              // Redis地址(IP+PORT)
	LogPath     string              // 日志路径
	rtmq_proxy  *rtmq.RtmqProxyConf // RTMQ配置
}

type OlSvrXmlNode struct {
	NodeName    xml.Name `xml:OlSvr`       // 根结点名
	FrwderAddr  string   `xml:FrwderAddr`  // 转发层(IP+PROT)
	SendChanLen uint32   `xml:SendChanLen` // 发送队列长度
	RecvChanLen uint32   `xml:RecvChanLen` // 接收队列长度
	WorkerNum   uint16   `xml:WorkerNum`   // 协程数
	RedisAddr   string   `xml:RedisAddr`   // Redis地址(IP+PORT)
	LogPath     string   `xml:LogPath`     // 日志路径
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

	v := OlSvrXmlNode{}

	err = xml.Unmarshal(data, &v)
	if nil != err {
		return err
	}

	/* > 解析配置文件 */
	/* 转发层(IP+PROT) */
	conf.FrwderAddr = v.FrwderAddr
	if 0 == len(conf.FrwderAddr) {
		return errors.New("Get frwder addr failed!")
	}

	/* 发送队列长度 */
	conf.SendChanLen = v.SendChanLen
	if 0 == conf.SendChanLen {
		return errors.New("Get send channel length failed!")
	}

	/* 接收队列长度 */
	conf.RecvChanLen = v.RecvChanLen
	if 0 == conf.RecvChanLen {
		return errors.New("Get recv channel length failed!")
	}

	/* 协程数 */
	conf.WorkerNum = v.WorkerNum
	if 0 == conf.WorkerNum {
		return errors.New("Get worker number failed!")
	}

	/* Redis地址(IP+PORT) */
	conf.RedisAddr = v.RedisAddr
	if 0 == len(conf.RedisAddr) {
		return errors.New("Get redis addr failed!")
	}

	/* 日志路径 */
	conf.LogPath = v.LogPath
	if 0 == len(conf.LogPath) {
		return errors.New("Get log path failed!")
	}

	return nil
}

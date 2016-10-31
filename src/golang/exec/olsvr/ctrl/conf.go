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
type OlsvrConf struct {
	NodeId    uint32             // 结点ID
	WorkPath  string             // 工作路径(自动获取)
	AppPath   string             // 程序路径(自动获取)
	ConfPath  string             // 配置路径(自动获取)
	RedisAddr string             // Redis地址(IP+PORT)
	LogPath   string             // 日志路径
	proxy     rtmq.RtmqProxyConf // RTMQ配置
}

/* 鉴权信息 */
type OlsvrConfRtmqAuthXmlData struct {
	Name   xml.Name `xml:Auth`   // 结点名
	Usr    string   `xml:Usr`    // 用户名
	Passwd string   `xml:Passwd` // 登录密码
}

/* RTMQ代理配置 */
type OlsvrConfRtmqProxyXmlData struct {
	Name        xml.Name                 `xml:RtmqProxy`   // 结点名
	Auth        OlsvrConfRtmqAuthXmlData `xml:Auth`        // 鉴权信息
	RemoteAddr  string                   `xml:RemoteAddr`  // 对端IP(IP+PROT)
	WorkerNum   uint32                   `xml:WorkerNum`   // 协程数
	SendChanLen uint32                   `xml:SendChanLen` // 发送队列长度
	RecvChanLen uint32                   `xml:RecvChanLen` // 接收队列长度
}

/* 在线中心XML配置 */
type OlsvrConfXmlData struct {
	Name      xml.Name                  `xml:OLSVR`     // 根结点名
	NodeId    uint32                    `xml:NodeId`    // 结点ID
	RedisAddr string                    `xml:RedisAddr` // Redis地址(IP+PORT)
	LogPath   string                    `xml:LogPath`   // 日志路径
	RtmqProxy OlsvrConfRtmqProxyXmlData `xml:RtmqProxy` // RTMQ PROXY配置
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
func (conf *OlsvrConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "olsvr.xml")

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
func (conf *OlsvrConf) conf_parse() (err error) {
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

	node := OlsvrConfXmlData{}

	err = xml.Unmarshal(data, &node)
	if nil != err {
		return err
	}

	/* > 解析配置文件 */
	/* 结点ID */
	conf.NodeId = node.NodeId
	if 0 == conf.NodeId {
		return errors.New("Get node id failed!")
	}

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

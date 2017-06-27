package conf

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"os"

	"beehive-im/src/golang/lib/log"
)

/* 日志配置 */
type MonConfLogXmlData struct {
	Name  xml.Name `xml:"LOG"`        // 结点名
	Level string   `xml:"LEVEL,attr"` // 日志级别
	Path  string   `xml:"PATH,attr"`  // 日志路径
}

/* REDIS配置 */
type MonRedisConf struct {
	Name   xml.Name `xml:"REDIS"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* MYSQL配置 */
type MonMysqlConf struct {
	Name   xml.Name `xml:"MYSQL"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* MONGO配置 */
type MonMongoConf struct {
	Name   xml.Name `xml:"MONGO"`       // 结点名
	Addr   string   `xml:"ADDR,attr"`   // 地址(IP+端口)
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* 鉴权配置 */
type MonConfRtmqAuthXmlData struct {
	Name   xml.Name `xml:"AUTH"`        // 结点名
	Usr    string   `xml:"USR,attr"`    // 用户名
	Passwd string   `xml:"PASSWD,attr"` // 登录密码
}

/* RTMQ代理配置 */
type MonConfRtmqProxyXmlData struct {
	Name        xml.Name               `xml:"FRWDER"`        // 结点名
	RemoteAddr  string                 `xml:"ADDR,attr"`     // 对端IP(IP+PROT)
	Auth        MonConfRtmqAuthXmlData `xml:"AUTH"`          // 鉴权信息
	WorkerNum   uint32                 `xml:"WORKER-NUM"`    // 协程数
	SendChanLen uint32                 `xml:"SEND-CHAN-LEN"` // 发送队列长度
	RecvChanLen uint32                 `xml:"RECV-CHAN-LEN"` // 接收队列长度
}

/* 在线中心XML配置 */
type MonConfXmlData struct {
	Name   xml.Name                `xml:"MONITOR"`  // 根结点名
	Id     uint32                  `xml:"ID,attr"`  // 结点ID
	Gid    uint32                  `xml:"GID,attr"` // 分组ID
	Redis  MonRedisConf            `xml:"REDIS"`    // Redis地址(IP+PORT)
	Mysql  MonMysqlConf            `xml:"MYSQL"`    // MYSQL地址(IP+PORT)
	Mongo  MonMongoConf            `xml:"MONGO"`    // MONGO地址(IP+PORT)
	Log    MonConfLogXmlData       `xml:"LOG"`      // 日志配置
	Frwder MonConfRtmqProxyXmlData `xml:"FRWDER"`   // RTMQ PROXY配置
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
func (conf *MonConf) parse() (err error) {
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

	node := MonConfXmlData{}

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

	/* 日志配置 */
	conf.Log.Level = log.GetLevel(node.Log.Level)

	conf.Log.Path = node.Log.Path
	if 0 == len(conf.Log.Path) {
		return errors.New("Get log path failed!")
	}

	/* FRWDER配置 */
	conf.Frwder.Id = conf.Id
	conf.Frwder.Gid = conf.Gid

	/* 转发层(IP+PROT) */
	conf.Frwder.RemoteAddr = node.Frwder.RemoteAddr
	if 0 == len(conf.Frwder.RemoteAddr) {
		return errors.New("Get Frwder addr failed!")
	}

	/* 鉴权信息 */
	conf.Frwder.Usr = node.Frwder.Auth.Usr
	conf.Frwder.Passwd = node.Frwder.Auth.Passwd
	if 0 == len(conf.Frwder.Usr) || 0 == len(conf.Frwder.Passwd) {
		return errors.New("Get auth conf failed!")
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

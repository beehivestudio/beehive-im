package conf

import (
	"os"
	"path/filepath"

	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rtmq"
)

/* 在线中心配置 */
type MonConf struct {
	NodeId   uint32             // 结点ID
	WorkPath string             // 工作路径(自动获取)
	AppPath  string             // 程序路径(自动获取)
	ConfPath string             // 配置路径(自动获取)
	Redis    MonRedisConf       // Redis配置
	Mysql    MonMysqlConf       // Mysql配置
	Mongo    MonMongoConf       // Mongo配置
	Log      log.LogConf        // 日志配置
	Frwder   rtmq.RtmqProxyConf // RTMQ配置
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
func (conf *MonConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "../conf", "monitor.xml")

	return conf.conf_parse()
}

/* 获取结点ID */
func (conf *MonConf) GetNid() uint32 {
	return conf.NodeId
}

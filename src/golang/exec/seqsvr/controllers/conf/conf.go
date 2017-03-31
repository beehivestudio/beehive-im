package conf

import (
	"os"
	"path/filepath"

	"beehive-im/src/golang/lib/log"
)

/* 在线中心配置 */
type SeqSvrConf struct {
	NodeId   uint32          // 结点ID
	WorkPath string          // 工作路径(自动获取)
	AppPath  string          // 程序路径(自动获取)
	ConfPath string          // 配置路径(自动获取)
	Port     uint16          // 侦听端口
	Redis    SeqSvrRedisConf // REDIS配置
	Mysql    SeqSvrMysqlConf // MYSQL配置
	Mongo    SeqSvrMongoConf // MONGO配置
	Cipher   string          // 私密密钥
	Log      log.Conf        // 日志配置
}

/******************************************************************************
 **函数名称: Load
 **功    能: 加载配置信息
 **输入参数:
 **     path: 配置路径
 **输出参数: NONE
 **返    回:
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func Load(path string) (conf *SeqSvrConf, err error) {
	conf = &SeqSvrConf{}

	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = path

	err = conf.parse()
	if nil != err {
		return nil, err
	}
	return conf, err
}

/* 获取结点ID */
func (conf *SeqSvrConf) GetNid() uint32 {
	return conf.NodeId
}

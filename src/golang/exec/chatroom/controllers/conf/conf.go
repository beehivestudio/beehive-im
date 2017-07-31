package conf

import (
	"os"
	"path/filepath"

	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rtmq"
)

/* 在线中心配置 */
type ChatRoomConf struct {
	Id       uint32             // 结点ID
	Gid      uint32             // 分组ID
	Port     int16              // HTTP侦听端口
	WorkPath string             // 工作路径(自动获取)
	AppPath  string             // 程序路径(自动获取)
	ConfPath string             // 配置路径(自动获取)
	Seqsvr   ChatRoomSeqsvrConf // SEQSVR配置
	Redis    ChatRoomRedisConf  // REDIS配置
	UserDb   ChatRoomMysqlConf  // USERDB配置(MYSQL)
	Mongo    ChatRoomMongoConf  // MONGO配置
	Cipher   string             // 私密密钥
	Log      log.Conf           // 日志配置
	Frwder   rtmq.ProxyConf     // RTMQ配置
}

/******************************************************************************
 **函数名称: Load
 **功    能: 加载配置信息
 **输入参数:
 **     path: 配置路径
 **输出参数: NONE
 **返    回:
 **     conf: 配置信息
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:35:28 #
 ******************************************************************************/
func Load(path string) (conf *ChatRoomConf, err error) {
	conf = &ChatRoomConf{}

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
func (conf *ChatRoomConf) GetNid() uint32 {
	return conf.Id
}

/* 获取分组ID */
func (conf *ChatRoomConf) GetGid() uint32 {
	return conf.Gid
}

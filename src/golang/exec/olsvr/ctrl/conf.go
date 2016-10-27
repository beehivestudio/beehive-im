package ctrl

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"chat/src/golang/lib/rtmq"
)

type OlSvrConf struct {
	NodeId       uint32              // 结点ID
	WorkPath     string              // 工作路径(自动获取)
	AppPath      string              // 程序路径(自动获取)
	ConfPath     string              // 配置路径(自动获取)
	FrwderAddr   string              // 转发层(IP+PROT)
	SendQueueLen uint32              // 发送队列长度
	RecvQueueLen uint32              // 接收队列长度
	WorkerNum    uint16              // 协程数
	RedisAddr    string              // Redis地址(IP+PORT)
	LogPath      string              // 日志路径
	SecretKey    string              // 密钥
	rtmq_proxy   *rtmq.RtmqProxyConf // RTMQ配置
}

/* 加载配置信息 */
func (conf *OlSvrConf) LoadConf() (err error) {
	conf.WorkPath, _ = os.Getwd()
	conf.WorkPath, _ = filepath.Abs(conf.WorkPath)
	conf.AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	conf.ConfPath = filepath.Join(conf.AppPath, "conf", "olsvr.conf")

	return conf.conf_parse()
}

/* 解析配置信息 */
func (conf *OlSvrConf) conf_parse() (err error) {
	var ok bool
	var key map[string]interface{}

	/* > 加载配置文件 */
	file, err := os.Open(conf.ConfPath)
	if nil != err {
		return err
	}

	defer file.Close()

	data := make([]byte, 10240)

	n, err := file.Read(data)
	if nil != err {
		return err
	}

	err = json.Unmarshal(data[:n], &key)
	if nil != err {
		return err
	}

	/* > 解析配置文件 */
	/* 转发层(IP+PROT) */
	if conf.FrwderAddr, ok = key["FrwderAddr"].(string); !ok {
		return errors.New("Get frwder addr failed!")
	}

	/* 发送队列长度 */
	if digit, ok := key["SendQueueLen"].(float64); !ok {
		conf.SendQueueLen = uint32(digit)
		return errors.New("Get send queue length failed!")
	}

	/* 接收队列长度 */
	if digit, ok := key["RecvQueueLen"].(float64); !ok {
		conf.RecvQueueLen = uint32(digit)
		return errors.New("Get recv queue length failed!")
	}

	/* 协程数 */
	if digit, ok := key["WorkerNum"].(float64); !ok {
		conf.WorkerNum = uint16(digit)
		return errors.New("Get worker number failed!")
	}

	/* Redis地址(IP+PORT) */
	if conf.RedisAddr, ok = key["RedisAddr"].(string); !ok {
		return errors.New("Get redis addr failed!")
	}

	/* 日志路径 */
	if conf.LogPath, ok = key["LogPath"].(string); !ok {
		return errors.New("Get log path failed!")
	}

	/* 密钥 */
	if conf.SecretKey, ok = key["SecretKey"].(string); !ok {
		return errors.New("Get frwder addr failed!")
	}

	return nil
}

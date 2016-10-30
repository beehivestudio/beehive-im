package main

import (
	"chat/src/golang/lib/rtmq"
)

/* ONLINE注册回调 */
func chat_cmd_online_handler(cmd uint32, data []byte, length uint32, param interface{}) {
	pxy, ok := param.(rtmq.RtmqProxyCntx)
	if !ok {
		pxy.log.Error("Get rtmq proxy context failed!")
		return false
	}
	pxy.Send(CHAT_CMD_ONLINE_ACK, data, len(data))

	return true
}

/* OFFLINE注册回调 */
func chat_cmd_offline_handler(cmd uint32, data []byte, length uint32, param interface{}) {
	pxy, ok := param.(rtmq.RtmqProxyCntx)
	if !ok {
		pxy.log.Error("Get rtmq proxy context failed!")
		return false
	}

	pxy.Send(CHAT_CMD_OFFLINE_ACK, data, len(data))

	return true
}

/* 初始化日志 */
func log_init() (log *logs.BeeLogger, err error) {
	log = logs.NewLogger(20000)

	err = os.Mkdir("../log", 0755)
	if nil != err && false == os.IsExist(err) {
		log.Emergency(err.Error())
		return log, err
	}

	log.SetLogger("file", fmt.Sprintf(`{"filename":"./log/demo.log"}`))
	log.SetLevel(logs.LevelDebug)
	return log, nil
}

func main() {
	/* > 加载配置 */
	conf := &RtmqProxyConf{}

	conf.NodeId = 1
	conf.Usr = "qifeng"
	conf.Passwd = "111111"
	conf.RemoteAddr = "127.0.0.1:28889"
	conf.SendChanLen = 10000
	conf.RecvChanLen = 10000
	conf.WorkerNum = 1

	/* > 初始化日志 */
	log, err := log_init()
	if nil != err {
		fmt.Printf("Log initialize failed!\n")
		return
	}

	/* > 初始化RTMQ-PROXY对象 */
	pxy, err := rtmq.ProxyInit(conf, log)
	if nil != err {
		log.Error("Rtmq proxy init failed!")
		return
	}

	/* > 注册回调函数 */
	pxy.Register(CHAT_CMD_ONLINE, chat_cmd_online_handler, pxy)
	pxy.Register(CHAT_CMD_OFFLINE, chat_cmd_offline_handler, pxy)

	/* > 启动PROXY服务 */
	pxy.Launch()
}

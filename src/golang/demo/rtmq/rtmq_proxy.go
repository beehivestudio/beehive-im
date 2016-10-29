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

func main() {
	/* > 加载配置 */
	conf := &RtmqProxyConf{}

	/* > 初始化RTMQ-PROXY对象 */
	pxy, err := rtmq.RtmqProxyInit(conf, log)
	if nil != err {
		log.Error("Rtmq proxy init failed!")
		return
	}

	/* > 注册回调函数 */
	pxy.Register(CHAT_CMD_ONLINE, chat_cmd_online_handler, pxy)
	pxy.Register(CHAT_CMD_OFFLINE, chat_cmd_offline_handler, pxy)
}

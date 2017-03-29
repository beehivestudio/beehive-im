package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/astaxie/beego/logs"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/rtmq"
)

/* ONLINE注册回调 */
func chat_cmd_online_handler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	pxy, ok := param.(*rtmq.Proxy)
	if !ok {
		fmt.Printf("Get rtmq proxy context failed!")
		return -1
	}

	pxy.AsyncSend(uint32(comm.CMD_ONLINE_ACK), data, uint32(len(data)))

	return 0
}

/* OFFLINE注册回调 */
func chat_cmd_offline_handler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	pxy, ok := param.(*rtmq.Proxy)
	if !ok {
		fmt.Printf("Get rtmq proxy context failed!")
		return -1
	}

	pxy.AsyncSend(uint32(comm.CMD_OFFLINE_ACK), data, uint32(len(data)))

	return 0
}

/* OFFLINE注册回调 */
func chat_cmd_2_handler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	pxy, ok := param.(*rtmq.Proxy)
	if !ok {
		fmt.Printf("Get rtmq proxy context failed!")
		return -1
	}

	log := pxy.GetLog()

	log.Debug("Recv command 2! data:%s", string(data))
	fmt.Printf("Recv command 2! data:%s\n", string(data))

	return 0
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
	conf := &rtmq.ProxyConf{}

	conf.NodeId = 1
	conf.Usr = "qifeng"
	conf.Passwd = "111111"
	conf.RemoteAddr = "127.0.0.1:10000"
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
	pxy := rtmq.ProxyInit(conf, log)
	if nil == pxy {
		log.Error("Rtmq proxy init failed!")
		return
	}

	/* > 注册回调函数 */
	pxy.Register(comm.CMD_ONLINE, chat_cmd_online_handler, pxy)
	pxy.Register(comm.CMD_OFFLINE, chat_cmd_offline_handler, pxy)
	pxy.Register(2, chat_cmd_2_handler, pxy)

	/* > 启动PROXY服务 */
	pxy.Launch()

	for {
		idx := rand.Intn(100000)
		data := fmt.Sprintf("This is just a test! idx:%d", idx)
		pxy.AsyncSend(uint32(1), []byte(data), uint32(len(data)))
		time.Sleep(1 * time.Second)
	}
}

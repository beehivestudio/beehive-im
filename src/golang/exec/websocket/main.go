package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"beehive-im/src/golang/exec/websocket/controllers"
	"beehive-im/src/golang/exec/websocket/controllers/conf"
)

/* 输入参数 */
type InputParam struct {
	conf *string /* 配置路径 */
}

/* 提取参数 */
func parse_param() *InputParam {
	param := &InputParam{}

	/* 配置文件 */
	param.conf = flag.String("c", "../conf/websocket.xml", "Configuration path")

	flag.Parse()

	return param
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cpuf, err := os.Create("cpu_profile")
	if nil != err {
		return
	}
	pprof.StartCPUProfile(cpuf)
	defer pprof.StopCPUProfile()

	param := parse_param()

	/* > 加载LSND-WS配置 */
	conf, err := conf.Load(*param.conf)
	if nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化LSND-WS环境 */
	ctx, err := controllers.LsndInit(conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 注册回调函数 */
	ctx.Register()

	/* > 启动侦听服务 */
	ctx.Launch()

	/* > 捕捉中断信号 */
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch

	return
}

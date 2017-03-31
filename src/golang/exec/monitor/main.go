package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"beehive-im/src/golang/exec/monitor/controllers"
	"beehive-im/src/golang/exec/monitor/controllers/conf"
)

/* 输入参数 */
type InputParam struct {
	conf *string /* 配置路径 */
}

/* 提取参数 */
func parse_param() *InputParam {
	param := &InputParam{}

	/* 配置文件 */
	param.conf = flag.String("c", "../conf/monitor.xml", "Configuration path")

	flag.Parse()

	return param
}

func main() {
	var conf conf.MonConf

	runtime.GOMAXPROCS(runtime.NumCPU())

	param := parse_param()

	/* > 加载监控配置 */
	if err := conf.Load(*param.conf); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化环境 */
	ctx, err := controllers.MonInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 注册回调函数 */
	ctx.Register()

	/* > 启动监控服务 */
	ctx.Launch()

	/* > 捕捉中断信号 */
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch

	return
}

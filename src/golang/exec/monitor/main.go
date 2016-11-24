package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"chat/src/golang/exec/monitor/controllers"
)

func main() {
	var conf controllers.MonConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载OLS配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化OLSVR环境 */
	ctx, err := controllers.MonInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 启动OLSVR服务 */
	ctx.Launch()

	/* > 捕捉中断信号 */
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch

	return
}

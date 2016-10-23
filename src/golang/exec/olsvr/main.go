package main

import (
	"flag"
	"fmt"
	"runtime"

	"chat/src/golang/exec/olsvr/ctrl"
)

func main() {
	var conf ctrl.OlSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载OLS配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化OLS环境 */
	ctx, err := ctrl.OlSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 启动OLS服务 */
	ctx.OlSvrLaunch()

	return
}

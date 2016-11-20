package controllers

import (
	"flag"
)

func InitApp() {
	var conf HttpSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载OLS配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化HTTPSVR环境 */
	ctx, err := HttpSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 启动HTTPSVR服务 */
	ctx.Launch()

	return
}

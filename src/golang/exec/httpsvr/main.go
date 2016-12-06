package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"chat/src/golang/exec/httpsvr/controllers"
	"chat/src/golang/exec/httpsvr/routers"
)

/* 设置配置信息 */
func set_conf(conf *controllers.HttpSvrConf) {
	beego.BConfig.AppName = "beehive-im"
	beego.BConfig.Listen.EnableHTTP = true
	beego.BConfig.Listen.HTTPAddr = ""
	beego.BConfig.Listen.HTTPPort = conf.Port
	beego.BConfig.RouterCaseSensitive = true
}

func main() {
	var conf controllers.HttpSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载HTTPSVR配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化HTTPSVR环境 */
	ctx, err := controllers.HttpSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 启动HTTPSVR服务 */
	ctx.Launch()

	routers.Router()

	set_conf(&conf)

	beego.Run()
}

package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"chat/src/golang/exec/httpsvr/controllers"
	"chat/src/golang/exec/httpsvr/routers"
)

/* 设置BEEGO配置 */
func init_bconf(conf *controllers.HttpSvrConf) {
	beego.BConfig.AppName = "beehive-im"
	beego.BConfig.Listen.EnableHTTP = true
	beego.BConfig.Listen.HTTPAddr = ""
	beego.BConfig.Listen.HTTPPort = conf.Port
	beego.BConfig.RouterCaseSensitive = true
}

/* 初始化 */
func _init() *controllers.HttpSvrCntx {
	var conf controllers.HttpSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载HTTPSVR配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return nil
	}

	init_bconf(&conf)

	/* > 初始化HTTPSVR环境 */
	ctx, err := controllers.HttpSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return nil
	}

	return ctx
}

/* 注册回调 */
func register(ctx *controllers.HttpSvrCntx) {
	ctx.Register()
	routers.Router()
}

/* 启动服务 */
func launch(ctx *controllers.HttpSvrCntx) {
	ctx.Launch()
	beego.Run()
}

/* 主函数 */
func main() {
	/* > 初始化 */
	ctx := _init()
	if nil == ctx {
		fmt.Printf("Initialize context failed!\n")
		return
	}

	/* > 注册回调 */
	register(ctx)

	/* > 启动服务 */
	launch(ctx)
}

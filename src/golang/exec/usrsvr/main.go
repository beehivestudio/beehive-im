package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"beehive-im/src/golang/exec/usrsvr/controllers"
	"beehive-im/src/golang/exec/usrsvr/controllers/conf"
	"beehive-im/src/golang/exec/usrsvr/routers"
)

/* 设置BEEGO配置 */
func set_bconf(conf *conf.UsrSvrConf) {
	beego.BConfig.AppName = "beehive-im"
	beego.BConfig.Listen.EnableHTTP = true
	beego.BConfig.Listen.HTTPAddr = ""
	beego.BConfig.Listen.HTTPPort = int(conf.Port)
	beego.BConfig.RouterCaseSensitive = true
}

/* 初始化 */
func _init() *controllers.UsrSvrCntx {
	var conf conf.UsrSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载HTTPSVR配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return nil
	}

	set_bconf(&conf)

	/* > 初始化HTTPSVR环境 */
	ctx, err := controllers.UsrSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return nil
	}

	return ctx
}

/* 注册回调 */
func register(ctx *controllers.UsrSvrCntx) {
	ctx.Register()
	routers.Router()
}

/* 启动服务 */
func launch(ctx *controllers.UsrSvrCntx) {
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

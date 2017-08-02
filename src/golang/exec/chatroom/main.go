package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"beehive-im/src/golang/exec/chatroom/controllers"
	"beehive-im/src/golang/exec/chatroom/controllers/conf"
	"beehive-im/src/golang/exec/chatroom/routers"
)

/* 输入参数 */
type InputParam struct {
	conf *string /* 配置路径 */
}

/* 提取参数 */
func parse_param() *InputParam {
	param := &InputParam{}

	/* 配置文件 */
	param.conf = flag.String("c", "../conf/chatroom.xml", "Configuration path")

	flag.Parse()

	return param
}

/* 设置BEEGO配置 */
func beego_config(conf *conf.ChatRoomConf) {
	beego.BConfig.AppName = "beehive-im"
	beego.BConfig.Listen.EnableHTTP = true
	beego.BConfig.Listen.HTTPAddr = ""
	beego.BConfig.Listen.HTTPPort = int(conf.Port)
	beego.BConfig.RouterCaseSensitive = true
	beego.BConfig.Log.FileLineNum = true
}

/* 初始化 */
func _init() *controllers.ChatRoomCntx {
	runtime.GOMAXPROCS(runtime.NumCPU())

	param := parse_param()

	/* > 加载CHATROOM配置 */
	conf, err := conf.Load(*param.conf)
	if nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return nil
	}

	beego_config(conf)

	/* > 初始化CHATROOM环境 */
	ctx, err := controllers.ChatRoomInit(conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return nil
	}

	return ctx
}

/* 注册回调 */
func register(ctx *controllers.ChatRoomCntx) {
	ctx.Register()
	routers.Router()
}

/* 启动服务 */
func launch(ctx *controllers.ChatRoomCntx) {
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

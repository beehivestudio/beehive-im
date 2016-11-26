package routers

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"chat/src/golang/exec/httpsvr/controllers"
)

/* > 设置路由回调 */
//  beego.Router("/api/list",&RestController{},"*:ListFood")
//  beego.Router("/api/create",&RestController{},"post:CreateFood")
//  beego.Router("/api/update",&RestController{},"put:UpdateFood")
//  beego.Router("/api/delete",&RestController{},"delete:DeleteFood")
func router() {
	beego.Router("/chat/register", &controllers.HttpSvrRegisterCtrl{}, "get:Register")
	beego.Router("/chat/iplist", &controllers.HttpSvrIpListCtrl{}, "get:IpList")
}

func init() {
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

	router()
}

package routers

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/astaxie/beego"

	"chat/src/golang/exec/httpsvr/ctrl"
)

/* > 设置路由回调 */
func router() {
	beego.Router("/chat/register", &ctrl.HttpSvrRegister{})
}

func init() {
	var conf ctrl.HttpSvrConf

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* > 加载HTTPSVR配置 */
	if err := conf.LoadConf(); nil != err {
		fmt.Printf("Load configuration failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 初始化HTTPSVR环境 */
	ctx, err := ctrl.HttpSvrInit(&conf)
	if nil != err {
		fmt.Printf("Initialize context failed! errmsg:%s\n", err.Error())
		return
	}

	/* > 启动HTTPSVR服务 */
	ctx.Launch()

	router()
}

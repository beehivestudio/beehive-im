package routers

import (
	"chat/src/golang/exec/httpsvr/controllers"
	"github.com/astaxie/beego"
)

/* > 设置路由回调 */
//  beego.Router("/api/list",&RestController{},"*:ListFood")
//  beego.Router("/api/create",&RestController{},"post:CreateFood")
//  beego.Router("/api/update",&RestController{},"put:UpdateFood")
//  beego.Router("/api/delete",&RestController{},"delete:DeleteFood")
func Router() {
	beego.Router("/chat/register", &controllers.HttpSvrRegisterCtrl{}, "get:Register")
	beego.Router("/chat/iplist", &controllers.HttpSvrIpListCtrl{}, "get:IpList")
}

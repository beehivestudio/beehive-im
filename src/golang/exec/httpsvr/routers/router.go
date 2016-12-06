package routers

import (
	"github.com/astaxie/beego"

	"beehive-im/src/golang/exec/httpsvr/controllers"
)

/* > 设置路由回调 */
//  beego.Router("/api/list",&RestController{},"*:ListFood")
//  beego.Router("/api/create",&RestController{},"post:CreateFood")
//  beego.Router("/api/update",&RestController{},"put:UpdateFood")
//  beego.Router("/api/delete",&RestController{},"delete:DeleteFood")
func Router() {
	beego.Router("/im/register", &controllers.HttpSvrRegisterCtrl{}, "get:Register")
	beego.Router("/im/iplist", &controllers.HttpSvrIpListCtrl{}, "get:IpList")
}

package routers

import (
	"github.com/astaxie/beego"

	"beehive-im/src/golang/exec/usrsvr/controllers"
)

/* > 设置路由回调 */
//  beego.Router("/api/list",&RestController{},"*:ListFood")
//  beego.Router("/api/create",&RestController{},"post:CreateFood")
//  beego.Router("/api/update",&RestController{},"put:UpdateFood")
//  beego.Router("/api/delete",&RestController{},"delete:DeleteFood")
func Router() {
	beego.Router("/im/register", &controllers.UsrSvrRegisterCtrl{}, "get:Register")
	beego.Router("/im/iplist", &controllers.UsrSvrIplistCtrl{}, "get:Iplist")

	beego.Router("/im/push", &controllers.UsrSvrPushCtrl{}, "post:Push")
	beego.Router("/im/query", &controllers.UsrSvrQueryCtrl{}, "get:Query")
	beego.Router("/im/config", &controllers.UsrSvrConfigCtrl{}, "get:Config")

	//beego.Router("/im/group/query", &controllers.UsrSvrGroupQueryCtrl{}, "get:Query")
	beego.Router("/im/group/config", &controllers.UsrSvrGroupConfigCtrl{}, "get:Config")

	//beego.Router("/im/room/query", &controllers.UsrSvrRoomQueryCtrl{}, "get:Query")
	beego.Router("/im/room/config", &controllers.UsrSvrRoomConfigCtrl{}, "get:Config")
}

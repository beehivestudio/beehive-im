package routers

import (
	"github.com/astaxie/beego"

	"beehive-im/src/golang/exec/chatroom/controllers"
)

/* > 设置路由回调 */
//  beego.Router("/api/list",&RestController{},"*:ListFood")
//  beego.Router("/api/create",&RestController{},"post:CreateFood")
//  beego.Router("/api/update",&RestController{},"put:UpdateFood")
//  beego.Router("/api/delete",&RestController{},"delete:DeleteFood")
func Router() {
	beego.Router("/room/push", &controllers.ChatRoomPushCtrl{}, "post:Push")

	beego.Router("/room/query", &controllers.ChatRoomQueryCtrl{}, "get:Query")
	beego.Router("/room/config", &controllers.ChatRoomConfigCtrl{}, "get:Config")
}

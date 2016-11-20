package routers

import (
	"github.com/astaxie/beego"

	"chat/src/golang/exec/httpsvr/ctrl"
)

func init() {
	beego.Router("/chat/register", &ctrl.HttpSvrRegister{})
}

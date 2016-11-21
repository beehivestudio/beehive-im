package ctrl

import (
	"github.com/astaxie/beego"
)

type ViewController struct {
	beego.Controller
}

func (this *ViewController) Prepare() {
}

type BaseController struct {
	beego.Controller
}

func (this *BaseController) Prepare() {
}

func (c *BaseController) Get() {
}

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
}

package controllers

import (
	"github.com/astaxie/beego"

	"beehive-im/src/golang/lib/comm"
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

/* 异常应答 */
func (this *BaseController) Error(code int, errmsg string) {
	var resp comm.HttpResp

	resp.Code = code
	resp.ErrMsg = errmsg

	this.Data["json"] = &resp
	this.ServeJSON()
}

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["WebSite"] = "beehivestudio.com"
	c.Data["Email"] = "Qifeng.zou.job@hotmail.com"
	c.Data["Name"] = "Qifeng.zou"
}

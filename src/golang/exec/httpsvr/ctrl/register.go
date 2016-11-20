package ctrl

import (
	_ "github.com/astaxie/beego"
)

/* 注册处理 */
type HttpSvrRegister struct {
	BaseController
}

func (this *HttpSvrRegister) Get() {
	ctx := ctrl.GetHttpCtx()

	/* > 提取注册参数 */
	uid := this.GetInt("uid")
	nation := this.GetInt("nation")
	city := this.GetInt("city")
	town := this.GetInt("town")
	if 0 == uid || 0 == nation {
		ctx.log.Error("Register param invalid! uid:%d nation:%d", uid, nation)
		return
	}

	/* > 申请会话ID */
	sid, err := ctx.alloc_sid()
	if nil != err {
		ctx.log.Error("Alloc sid failed! errmsg:%s", err.Error())
		return
	}

	return
}

type HttpSvrRegisterRsp struct {
	Uid    string `json:"uid,omitempty"` // 用户ID
	Sid    uint64 `json:"sid"`           // 会话ID
	Nation uint64 `json:"nation"`        // 国家编号(国)
	City   uint64 `json:"city"`          // 城市编号(市)
	Town   uint64 `json:"town"`          // 城镇编号(县)
}

/******************************************************************************
 **函数名称: Register
 **功    能: 注册处理
 **输入参数:
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.20 12:23:20 #
 ******************************************************************************/
func (this *HttpSvrRegister) Register() {
	return
}

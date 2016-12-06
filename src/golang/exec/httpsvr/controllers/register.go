package controllers

import (
	"errors"

	_ "github.com/astaxie/beego"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
)

/* 注册处理 */
type HttpSvrRegisterCtrl struct {
	BaseController
}

/* 注册参数 */
type HttpSvrRegisterParam struct {
	uid    uint64 // 用户ID
	nation uint64 // 国际ID
	city   uint64 // 地市ID
	town   uint64 // 县城ID
}

func (this *HttpSvrRegisterCtrl) Register() {
	ctx := GetHttpCtx()

	/* > 提取参数 */
	param, err := this.parse_param(ctx)
	if nil != err {
		ctx.log.Error("Parse register failed! uid:%d nation:%d city:%d town:%d",
			param.uid, param.nation, param.city, param.town)
		this.response_fail(param, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return
	}

	ctx.log.Debug("Register param list. uid:%d nation:%d city:%d town:%d",
		param.uid, param.nation, param.city, param.town)

	this.handler(param)

	return
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 解析参数
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     param: 注册参数
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 10:30:09 #
 ******************************************************************************/
func (this *HttpSvrRegisterCtrl) parse_param(ctx *HttpSvrCntx) (*HttpSvrRegisterParam, error) {
	var param HttpSvrRegisterParam

	/* > 提取注册参数 */
	id, _ := this.GetInt64("uid")
	param.uid = uint64(id)

	id, _ = this.GetInt64("nation")
	param.nation = uint64(id)

	id, _ = this.GetInt64("city")
	param.city = uint64(id)

	id, _ = this.GetInt64("town")
	param.town = uint64(id)

	/* > 校验参数合法性 */
	if 0 == param.uid || 0 == param.nation {
		ctx.log.Error("Register param invalid! uid:%d nation:%d", param.uid, param.nation)
		return &param, errors.New("Register param invalid!")
	}

	return &param, nil
}

/* 注册应答 */
type HttpSvrRegisterRsp struct {
	Uid    uint64 `json:"uid"`    // 用户ID
	Sid    uint64 `json:"sid"`    // 会话ID
	Nation uint64 `json:"nation"` // 国家ID(国)
	City   uint64 `json:"city"`   // 城市ID(市)
	Town   uint64 `json:"town"`   // 城镇ID(县)
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

/******************************************************************************
 **函数名称: handler
 **功    能: 注册处理
 **输入参数:
 **     param: 注册参数
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.24 17:34:27 #
 ******************************************************************************/
func (this *HttpSvrRegisterCtrl) handler(param *HttpSvrRegisterParam) {
	ctx := GetHttpCtx()

	/* > 申请会话ID */
	sid, err := im.AllocSid(ctx.redis)
	if nil != err {
		ctx.log.Error("Alloc sid failed! errmsg:%s", err.Error())
		this.response_fail(param, comm.ERR_SYS_RPC, err.Error())
		return
	}

	ctx.log.Debug("Alloc sid success! uid:%d sid:%d", param.uid, sid)

	this.response_success(param, sid)

	return
}

/******************************************************************************
 **函数名称: response_fail
 **功    能: 应答错误信息
 **输入参数:
 **     param: 注册参数
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.24 19:13:29 #
 ******************************************************************************/
func (this *HttpSvrRegisterCtrl) response_fail(param *HttpSvrRegisterParam, code int, errmsg string) {
	var resp HttpSvrRegisterRsp

	resp.Uid = param.uid
	resp.Sid = 0
	resp.Nation = param.nation
	resp.City = param.city
	resp.Town = param.town
	resp.Code = code
	resp.ErrMsg = errmsg

	this.Data["json"] = &resp
	this.ServeJSON()
}

/******************************************************************************
 **函数名称: response_success
 **功    能: 应答处理成功
 **输入参数:
 **     param: 注册参数
 **     sid: 会话SID
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.24 19:13:22 #
 ******************************************************************************/
func (this *HttpSvrRegisterCtrl) response_success(param *HttpSvrRegisterParam, sid uint64) {
	var resp HttpSvrRegisterRsp

	resp.Uid = param.uid
	resp.Sid = sid
	resp.Nation = param.nation
	resp.City = param.city
	resp.Town = param.town
	resp.Code = 0
	resp.ErrMsg = "OK"

	this.Data["json"] = &resp
	this.ServeJSON()
}

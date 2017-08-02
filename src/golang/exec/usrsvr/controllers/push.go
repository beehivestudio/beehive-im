package controllers

import (
	"errors"
	"fmt"

	_ "github.com/garyburd/redigo/redis"
	_ "github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	_ "beehive-im/src/golang/lib/mesg"
)

/* 推送接口 */
type UsrSvrPushCtrl struct {
	BaseController
}

func (this *UsrSvrPushCtrl) Push() {
	ctx := GetUsrSvrCtx()

	dim := this.GetString("dim")
	switch dim {
	case "sid": // SID推送
		this.SidPush(ctx)
		return
	case "uid": // UID推送
		this.UidPush(ctx)
		return
	case "appid": // APP推送
		this.AppPush(ctx)
		return
	case "group": // GROUP推送
		this.GroupPush(ctx)
		return
	case "broadcast": // 全员推送
		this.Broadcast(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this dimension:%s.", dim)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
}

////////////////////////////////////////////////////////////////////////////////
// SID推送

func (this *UsrSvrPushCtrl) SidPush(ctx *UsrSvrCntx) {
	return
}

////////////////////////////////////////////////////////////////////////////////
// UID推送

/* UID推送请求 */
type UidPushReq struct {
	ctrl *UsrSvrPushCtrl
}

/* UID推送参数 */
type UidPushParam struct {
	uid    uint64 // 用户UID
	expire uint32 // 超时时间
}

/* UID推送应答 */
type UidPushRsp struct {
	Uid    uint64 `json:"uid"`    // 用户UID
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

func (this *UsrSvrPushCtrl) UidPush(ctx *UsrSvrCntx) {
	req := &UidPushReq{ctrl: this}

	/* > 提取广播参数 */
	param, err := req.parse_param(ctx)
	if nil != err {
		ctx.log.Error("Parse uid push param failed! uid:%d", param.uid)
		this.Error(comm.ERR_SVR_PARSE_PARAM, err.Error())
		return
	}

	/* > UID推送处理 */
	code, err := req.push_handler(ctx, param)
	if nil != err {
		this.Error(code, err.Error())
		return
	}

	req.push_success(param)

	return

}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 解析参数
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回:
 **     param: 注册参数
 **     err: 错误描述
 **实现描述: 从url参数中抽取
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 22:05:58 #
 ******************************************************************************/
func (req *UidPushReq) parse_param(
	ctx *UsrSvrCntx) (param *UidPushParam, err error) {
	this := req.ctrl
	param = &UidPushParam{}

	/* > 提取注册参数 */
	uid, _ := this.GetInt64("uid")
	param.uid = uint64(uid)

	expire, _ := this.GetInt32("expire")
	param.expire = uint32(expire)

	/* > 校验参数合法性 */
	if 0 == param.uid {
		ctx.log.Error("Paramter is invalid. uid:%d", param.uid)
		return param, errors.New("Paramter is invalid!")
	}

	return param, nil
}

/******************************************************************************
 **函数名称: push_handler
 **功    能: UID推送处理
 **输入参数:
 **     param: 请求参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 按照协议返回http应答
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 22:19:04 #
 ******************************************************************************/
func (req *UidPushReq) push_handler(ctx *UsrSvrCntx, param *UidPushParam) (code int, err error) {
	return 0, nil
}

/******************************************************************************
 **函数名称: push_success
 **功    能: UID推送成功
 **输入参数:
 **     param: 请求参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 按照协议返回http应答
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 22:19:04 #
 ******************************************************************************/
func (req *UidPushReq) push_success(param *UidPushParam) {
	this := req.ctrl
	rsp := &UidPushRsp{
		Uid:    param.uid,
		Code:   0,
		ErrMsg: "OK",
	}

	this.Data["json"] = rsp
	this.ServeJSON()
}

////////////////////////////////////////////////////////////////////////////////
// APP推送

func (this *UsrSvrPushCtrl) AppPush(ctx *UsrSvrCntx) {
	return
}

////////////////////////////////////////////////////////////////////////////////
// GROUP推送

/* GROUP推送参数 */
type GroupPushParam struct {
	gid    uint64 // 群组ID
	expire uint32 // 超时时间
}

/* GROUP推送应答 */
type GroupPushRsp struct {
	Gid    uint64 `json:"gid"`    // 群组ID
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

func (this *UsrSvrPushCtrl) GroupPush(ctx *UsrSvrCntx) {
	return
}

////////////////////////////////////////////////////////////////////////////////
// 全员推送

func (this *UsrSvrPushCtrl) Broadcast(ctx *UsrSvrCntx) {
	return
}

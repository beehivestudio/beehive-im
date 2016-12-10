package controllers

import (
	"fmt"

	"beehive-im/src/golang/lib/comm"
)

/* 系统配置 */
type UsrSvrConfigCtrl struct {
	BaseController
}

func (this *UsrSvrConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	opt := this.GetString("opt")
	switch opt {
	case "user-statis-add": // 添加在线人数统计
		return
	case "user-statis-del": // 删除在线人数统计
		return
	case "user-statis-list": // 在线人数统计列表
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", opt))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 群组配置 */
type UsrSvrGroupConfigCtrl struct {
	BaseController
}

func (this *UsrSvrGroupConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	opt := this.GetString("opt")
	switch opt {
	case "blacklist-add": // 将某人加入群组黑名单
	case "blacklist-del": // 将某人移除群组黑名单
	case "ban-add": // 禁言
	case "ban-del": // 解除禁言
	case "close": // 关闭群组
	case "capacity": // 设置群组容量
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", opt))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 聊天室配置 */
type UsrSvrRoomConfigCtrl struct {
	BaseController
}

func (this *UsrSvrRoomConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	opt := this.GetString("opt")
	switch opt {
	case "blacklist-add": // 将某人加入聊天室黑名单
	case "blacklist-del": // 将某人移除聊天室黑名单
	case "ban-add": // 禁言
	case "ban-del": // 解除禁言
	case "close": // 关闭聊天室
	case "capacity": // 设置聊天室各组容量
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", opt))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

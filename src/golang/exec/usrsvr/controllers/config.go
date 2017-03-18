package controllers

import (
	"fmt"
	"strconv"
	"time"

	"beehive-im/src/golang/lib/comm"
)

/* 系统配置 */
type UsrSvrConfigCtrl struct {
	BaseController
}

func (this *UsrSvrConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "user-statis-add": // 添加在线人数统计
		return
	case "user-statis-del": // 删除在线人数统计
		return
	case "user-statis-list": // 在线人数统计列表
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 群组配置 */
type UsrSvrGroupConfigCtrl struct {
	BaseController
}

func (this *UsrSvrGroupConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "blacklist": // 群组黑名单操作
	case "ban": // 群组禁言操作
	case "close": // 关闭群组
	case "capacity": // 设置群组容量
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 聊天室配置 */
type UsrSvrRoomConfigCtrl struct {
	BaseController
}

func (this *UsrSvrRoomConfigCtrl) Config() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "blacklist": // 聊天室黑名单操作
		this.blacklist(ctx)
	case "gag": // 聊天室禁言操作
		this.gag(ctx)
	case "close": // 关闭聊天室
		this.close(ctx)
	case "capacity": // 聊天室分组容量
		this.capacity(ctx)
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室黑名单操作接口

/******************************************************************************
 **函数名称: blacklist
 **功    能: 黑名单操作
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:33:48 #
 ******************************************************************************/
func (this *UsrSvrRoomConfigCtrl) blacklist(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "add": // 添加聊天室黑名单
		this.blacklist_add(ctx)
	case "del": // 移除聊天室黑名单
		this.blacklist_del(ctx)
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type BlackListAddParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 加入黑名单请求 */
type BlackListAddReq struct {
	ctrl *UsrSvrRoomConfigCtrl
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:47:44 #
 ******************************************************************************/
func (req *BlackListAddReq) parse_param() *BlackListAddParam {
	this := req.ctrl
	param := &BlackListAddParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter rid is invalid!")
		return nil
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter uid is invalid!")
		return nil
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param
}

/******************************************************************************
 **函数名称: blacklist_add
 **功    能: 加入黑名单
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.加入聊天室黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:49:16 #
 ******************************************************************************/
func (this *UsrSvrRoomConfigCtrl) blacklist_add(ctx *UsrSvrCntx) {
	req := &BlackListAddReq{ctrl: this}

	param := req.parse_param()
	if nil == param {
		ctx.log.Error("Parse blacklist add action paramater failed!")
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户加入黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 请求参数 */
type BlackListDelParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type BlackListDelReq struct {
	ctrl *UsrSvrRoomConfigCtrl // 空间对象
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:47:44 #
 ******************************************************************************/
func (req *BlackListDelReq) parse_param() *BlackListDelParam {
	this := req.ctrl
	param := &BlackListDelParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter rid is invalid!")
		return nil
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter uid is invalid!")
		return nil
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param
}

/******************************************************************************
 **函数名称: blacklist_del
 **功    能: 移除黑名单
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除聊天室黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 10:11:20 #
 ******************************************************************************/
func (this *UsrSvrRoomConfigCtrl) blacklist_del(ctx *UsrSvrCntx) {
	req := &BlackListDelReq{ctrl: this}

	param := req.parse_param()
	if nil == param {
		ctx.log.Error("Parse blacklist del action paramater failed!")
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户移除黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室禁言操作接口

func (this *UsrSvrRoomConfigCtrl) gag(ctx *UsrSvrCntx) {
}

func (this *UsrSvrRoomConfigCtrl) close(ctx *UsrSvrCntx) {
}

func (this *UsrSvrRoomConfigCtrl) capacity(ctx *UsrSvrCntx) {
}

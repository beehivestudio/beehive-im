package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"

	"beehive-im/src/golang/exec/chatroom/models"
)

/* 聊天室配置 */
type ChatRoomConfigCtrl struct {
	BaseController
}

func (this *ChatRoomConfigCtrl) Config() {
	ctx := GetRoomSvrCntx()

	option := this.GetString("option")
	switch option {
	case "blacklist": // 聊天室黑名单操作
		this.Blacklist(ctx)
		return
	case "gag": // 聊天室禁言操作
		this.Gag(ctx)
		return
	case "room": // 聊天室操作
		this.Room(ctx)
		return
	case "capacity": // 聊天室分组容量
		this.Capacity(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this option:%s.", option)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
	return
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室黑名单操作接口

/******************************************************************************
 **函数名称: Blacklist
 **功    能: 黑名单操作
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:33:48 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) Blacklist(ctx *ChatRoomCntx) {
	action := this.GetString("action")
	switch action {
	case "add": // 添加聊天室黑名单
		this.blacklist_add(ctx)
		return
	case "del": // 移除聊天室黑名单
		this.blacklist_del(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this action:%s.", action)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)

	return
}

/* 参数列表 */
type RoomBlacklistAddParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 加入黑名单请求 */
type RoomBlacklistAddReq struct {
	ctrl *ChatRoomConfigCtrl
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
func (req *RoomBlacklistAddReq) parse_param() (*RoomBlacklistAddParam, error) {
	this := req.ctrl
	param := &RoomBlacklistAddParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param, nil
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
func (this *ChatRoomConfigCtrl) blacklist_add(ctx *ChatRoomCntx) {
	req := &RoomBlacklistAddReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Add blacklist failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户加入黑名单 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("SADD", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type RoomBlackListDelParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type RoomBlackListDelReq struct {
	ctrl *ChatRoomConfigCtrl // 空间对象
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
func (req *RoomBlackListDelReq) parse_param() (*RoomBlackListDelParam, error) {
	this := req.ctrl
	param := &RoomBlackListDelParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param, nil
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
func (this *ChatRoomConfigCtrl) blacklist_del(ctx *ChatRoomCntx) {
	req := &RoomBlackListDelReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Del blacklist failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户移除黑名单 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("SREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室禁言操作接口

/******************************************************************************
 **函数名称: Gag
 **功    能: 禁言操作
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 11:23:31 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) Gag(ctx *ChatRoomCntx) {
	action := this.GetString("action")
	switch action {
	case "add": // 添加禁言
		this.gag_add(ctx)
		return
	case "del": // 移除禁言
		this.gag_del(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this action:%s.", action)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
	return
}

/* 参数列表 */
type RoomGagAddParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 加入禁言请求 */
type RoomGagAddReq struct {
	ctrl *ChatRoomConfigCtrl
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
func (req *RoomGagAddReq) parse_param() (*RoomGagAddParam, error) {
	this := req.ctrl
	param := &RoomGagAddParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param, nil
}

/******************************************************************************
 **函数名称: gag_add
 **功    能: 添加禁言
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.加入聊天室禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 11:27:21 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) gag_add(ctx *ChatRoomCntx) {
	req := &RoomGagAddReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Add gag failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户加入禁言 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_GAG_SET, param.rid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type RoomGagDelParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type RoomGagDelReq struct {
	ctrl *ChatRoomConfigCtrl // 空间对象
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
func (req *RoomGagDelReq) parse_param() (*RoomGagDelParam, error) {
	this := req.ctrl
	param := &RoomGagDelParam{}

	rid_str := this.GetString("rid")
	uid_str := this.GetString("uid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.rid = uint64(rid)
	param.uid = uint64(uid)

	return param, nil
}

/******************************************************************************
 **函数名称: gag_del
 **功    能: 移除禁言
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 11:29:04 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) gag_del(ctx *ChatRoomCntx) {
	req := &RoomGagDelReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Del gag failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 用户移除黑名单 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_GAG_SET, param.rid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: Room
 **功    能: 聊天室其他操作
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 直接解散聊天室
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:48:21 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) Room(ctx *ChatRoomCntx) {
	action := this.GetString("action")
	switch action {
	case "open": // 打开聊天室
		this.room_open(ctx)
		return
	case "close": // 关闭聊天室
		this.room_close(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this action:%s.", action)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
	return
}

/* 参数列表 */
type RoomOpenParam struct {
	rid uint64 // 聊天室ID
}

/* 打开请求 */
type RoomOpenReq struct {
	ctrl *ChatRoomConfigCtrl
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 08:06:28 #
 ******************************************************************************/
func (req *RoomOpenReq) parse_param() (*RoomOpenParam, error) {
	this := req.ctrl
	param := &RoomOpenParam{}

	rid_str := this.GetString("rid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	param.rid = uint64(rid)

	return param, nil
}

/******************************************************************************
 **函数名称: room_open
 **功    能: 开启聊天室
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.修改聊天室属性
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 08:07:31 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) room_open(ctx *ChatRoomCntx) {
	req := &RoomOpenReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Open room failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 用户加入禁言 */
	key := fmt.Sprintf(models.ROOM_KEY_RID_ATTR, param.rid)

	_, err = rds.Do("HSET", key, "STATUS", models.ROOM_STAT_OPEN)
	if nil != err {
		/* > 回复处理应答 */
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 参数列表 */
type RoomCloseParam struct {
	rid uint64 // 聊天室ID
}

/* 打开请求 */
type RoomCloseReq struct {
	ctrl *ChatRoomConfigCtrl
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 08:06:28 #
 ******************************************************************************/
func (req *RoomCloseReq) parse_param() (*RoomCloseParam, error) {
	this := req.ctrl
	param := &RoomCloseParam{}

	rid_str := this.GetString("rid")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	param.rid = uint64(rid)

	return param, nil
}

/******************************************************************************
 **函数名称: room_close
 **功    能: 关闭聊天室
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.修改聊天室属性
 **注意事项: TODO: 关闭聊天室后, 需要给所有侦听层广播解散聊天室的指令.
 **作    者: # Qifeng.zou # 2017.03.19 08:07:31 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) room_close(ctx *ChatRoomCntx) {
	req := &RoomCloseReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Close room failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 修改聊天室属性 */
	key := fmt.Sprintf(models.ROOM_KEY_RID_ATTR, param.rid)

	_, err = rds.Do("HSET", key, "STATUS", models.ROOM_STAT_CLOSE)
	if nil != err {
		/* > 回复处理应答 */
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: Capacity
 **功    能: 设置聊天室分组容量
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:48:21 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) Capacity(ctx *ChatRoomCntx) {
	action := this.GetString("action")
	switch action {
	case "set": // 设置聊天室分组容量
		this.capacity_set(ctx)
		return
	case "get": // 获取聊天室分组容量
		this.capacity_get(ctx)
		return
	}
}

/* 请求参数 */
type RoomCapSetParam struct {
	rid      uint64 // 聊天室ID
	capacity int    // 分组容量
}

/* 请求对象 */
type RoomCapSetReq struct {
	ctrl *ChatRoomConfigCtrl // 空间对象
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:51:55 #
 ******************************************************************************/
func (req *RoomCapSetReq) parse_param() (*RoomCapSetParam, error) {
	this := req.ctrl
	param := &RoomCapSetParam{}

	rid_str := this.GetString("rid")
	cap_str := this.GetString("cap")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	capacity, _ := strconv.ParseInt(cap_str, 10, 32)
	if 0 == capacity {
		return nil, errors.New("Paramter [cap] is invalid!")
	}

	param.rid = uint64(rid)
	param.capacity = int(capacity)

	return param, nil
}

/******************************************************************************
 **函数名称: capacity_set
 **功    能: 设置聊天室分组人数
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 00:00:39 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) capacity_set(ctx *ChatRoomCntx) {
	req := &RoomCapSetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Set room capacity failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 存储聊天室分组容量 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_GROUP_CAP_ZSET)

	pl.Send("ZADD", key, param.capacity, param.rid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type RoomCapGetParam struct {
	rid      uint64 // 聊天室ID
	capacity int    // 分组容量
}

/* 请求对象 */
type RoomCapGetReq struct {
	ctrl *ChatRoomConfigCtrl // 空间对象
}

/* 请求应答 */
type RoomCapGetRsp struct {
	Rid    uint64 `json:"rid"`    // ROOM ID
	Cap    int    `json:"cap"`    // 分组容量
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:51:55 #
 ******************************************************************************/
func (req *RoomCapGetReq) parse_param() (*RoomCapGetParam, error) {
	this := req.ctrl
	param := &RoomCapGetParam{}

	rid_str := this.GetString("rid")
	cap_str := this.GetString("cap")

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	capacity, _ := strconv.ParseInt(cap_str, 10, 32)
	if 0 == capacity {
		return nil, errors.New("Paramter [cap] is invalid!")
	}

	param.rid = uint64(rid)
	param.capacity = int(capacity)

	return param, nil
}

/******************************************************************************
 **函数名称: capacity_get
 **功    能: 获取聊天室分组人数
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 00:00:39 #
 ******************************************************************************/
func (this *ChatRoomConfigCtrl) capacity_get(ctx *ChatRoomCntx) {
	req := &RoomCapGetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Get room capacity failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 存储聊天室分组容量 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_GROUP_CAP_ZSET)

	capacity, err := redis.Int(rds.Do("ZSCORE", key, param.rid))
	if nil != err {
		ctx.log.Error("Get room capacity failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	rsp := &RoomCapGetRsp{
		Rid:    param.rid,
		Cap:    capacity,
		Code:   comm.OK,
		ErrMsg: "Ok",
	}

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

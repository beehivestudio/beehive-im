package controllers

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/chat"
	"beehive-im/src/golang/lib/comm"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// 系统配置

/* 系统配置 */
type UsrSvrConfigCtrl struct {
	BaseController
}

func (this *UsrSvrConfigCtrl) Config() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "user-statis": // 添加在线人数统计
		this.UserStatis(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
// 用户统计配置

/******************************************************************************
 **函数名称: UserStatis
 **功    能: 用户统计配置
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 22:58:33 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) UserStatis(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "add":
		this.user_statis_add(ctx)
		return
	case "del":
		this.user_statis_del(ctx)
		return
	case "get":
		this.user_statis_get(ctx)
		return
	case "list":
		this.user_statis_list(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type UserStatisAddParam struct {
	prec int // 统计精度
	num  int // 记录总数
}

/* 添加请求 */
type UserStatisAddReq struct {
	ctrl *UsrSvrConfigCtrl
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
func (req *UserStatisAddReq) parse_param() (*UserStatisAddParam, error) {
	this := req.ctrl
	param := &UserStatisAddParam{}

	param.prec, _ = this.GetInt("prec")
	if 0 == param.prec {
		return nil, errors.New("Paramter [prec] is invalid!")
	}

	param.num, _ = this.GetInt("num")
	if 0 == param.num {
		return nil, errors.New("Paramter [prec] is invalid!")
	}

	return param, nil
}

/******************************************************************************
 **函数名称: user_statis_add
 **功    能: 添加用户统计的统计精度
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.添加统计精度
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 23:03:04 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) user_statis_add(ctx *UsrSvrCntx) {
	req := &UserStatisAddReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Add user statis failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 添加统计精度 */
	key := fmt.Sprintf(comm.IM_KEY_PREC_NUM_ZSET)

	pl.Send("ZADD", key, param.num, param.prec)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 参数列表 */
type UserStatisDelParam struct {
	prec int // 统计精度
}

/* 添加请求 */
type UserStatisDelReq struct {
	ctrl *UsrSvrConfigCtrl
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
func (req *UserStatisDelReq) parse_param() (*UserStatisDelParam, error) {
	this := req.ctrl
	param := &UserStatisDelParam{}

	param.prec, _ = this.GetInt("prec")
	if 0 == param.prec {
		return nil, errors.New("Paramter [prec] is invalid!")
	}

	return param, nil
}

/******************************************************************************
 **函数名称: user_statis_del
 **功    能: 删除用户统计的统计精度
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.删除统计精度
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.19 23:03:04 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) user_statis_del(ctx *UsrSvrCntx) {
	req := &UserStatisDelReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Del user statistic failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 添加统计精度 */
	key := fmt.Sprintf(comm.IM_KEY_PREC_NUM_ZSET)

	pl.Send("ZREM", key, param.prec)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 参数列表 */
type UserStatisGetParam struct {
	prec int // 统计精度
	num  int // 记录条数
}

/* 添加请求 */
type UserStatisGetReq struct {
	ctrl *UsrSvrConfigCtrl
}

/* 应答结果 */
type UserStatisGetRsp struct {
	Prec   int            `json:"prec"`   // 统计精度
	Num    int            `json:"num"`    // 统计条数
	List   UserStatisList `json:"list"`   // 统计结果
	Code   int            `json:"code"`   // 返回码
	ErrMsg string         `json:"errmsg"` // 错误描述
}

type UserStatisList []UserStatisItem

/* 统计结果 */
type UserStatisItem struct {
	Idx     int    `json:"idx"`      // 序列号
	Time    int64  `json:"time"`     // 统计时间
	TimeStr string `json:"time-str"` // 统计时间
	Num     int    `json:"num"`      // 在线人数
}

func (list UserStatisList) Len() int           { return len(list) }
func (list UserStatisList) Less(i, j int) bool { return list[i].Time > list[j].Time }
func (list UserStatisList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.20 17:07:21 #
 ******************************************************************************/
func (req *UserStatisGetReq) parse_param() (*UserStatisGetParam, error) {
	this := req.ctrl
	param := &UserStatisGetParam{}

	param.prec, _ = this.GetInt("prec")
	if 0 == param.prec {
		return nil, errors.New("Paramter [prec] is invalid!")
	}

	param.num, _ = this.GetInt("num")
	if 0 == param.num {
		return nil, errors.New("Paramter [num] is invalid!")
	}

	return param, nil
}

/******************************************************************************
 **函数名称: query
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.20 17:07:21 #
 ******************************************************************************/
func (req *UserStatisGetReq) query(ctx *UsrSvrCntx, prec int, num int) (UserStatisList, error) {
	var list UserStatisList

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取统计结果 */
	key := fmt.Sprintf(comm.IM_KEY_PREC_USR_NUM, prec)

	data, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		return list, err
	}

	data_len := len(data)
	for idx := 0; idx < data_len; idx += 2 {
		old_tm, _ := strconv.ParseInt(data[idx], 10, 64)
		user_num, _ := strconv.ParseInt(data[idx+1], 10, 32)

		tm := time.Unix(old_tm, 0)

		item := UserStatisItem{
			Time:    old_tm,
			TimeStr: tm.Format("2006-01-02 15:04:05"),
			Num:     int(user_num),
		}

		list = append(list, item)
	}

	sort.Sort(list) /* 排序处理 */

	/* > 整理统计结果 */
	list = list[0:num]

	num = len(list)
	for idx := 0; idx < num; idx += 1 {
		list[idx].Idx = idx
	}

	return list, nil
}

/******************************************************************************
 **函数名称: user_statis_get
 **功    能: 获取某精度用户统计数据
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.获取统计信息 3.返回统计信息
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.20 17:08:17 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) user_statis_get(ctx *UsrSvrCntx) {
	req := &UserStatisGetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	/* > 获取统计结果 */
	list, err := req.query(ctx, param.prec, param.num)
	if nil != err {
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	rsp := &UserStatisGetRsp{}

	rsp.Prec = param.prec
	rsp.Num = len(list)
	rsp.List = list
	rsp.Code = 0
	rsp.ErrMsg = "Ok"

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

func (this *UsrSvrConfigCtrl) user_statis_list(ctx *UsrSvrCntx) {
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// 群组配置

/* 群组配置 */
type UsrSvrGroupConfigCtrl struct {
	BaseController
}

func (this *UsrSvrGroupConfigCtrl) Config() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "blacklist": // 群组黑名单操作
		return
	case "ban": // 群组禁言操作
		return
	case "close": // 关闭群组
		return
	case "capacity": // 设置群组容量
		return
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

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
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
func (this *UsrSvrRoomConfigCtrl) Blacklist(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "add": // 添加聊天室黑名单
		this.blacklist_add(ctx)
		return
	case "del": // 移除聊天室黑名单
		this.blacklist_del(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type RoomBlacklistAddParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 加入黑名单请求 */
type RoomBlacklistAddReq struct {
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
func (this *UsrSvrRoomConfigCtrl) blacklist_add(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 请求参数 */
type RoomBlackListDelParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type RoomBlackListDelReq struct {
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
func (this *UsrSvrRoomConfigCtrl) blacklist_del(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, param.rid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

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
func (this *UsrSvrRoomConfigCtrl) Gag(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "add": // 添加禁言
		this.gag_add(ctx)
		return
	case "del": // 移除禁言
		this.gag_del(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type RoomGagAddParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 加入禁言请求 */
type RoomGagAddReq struct {
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
func (this *UsrSvrRoomConfigCtrl) gag_add(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_GAG_SET, param.rid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 请求参数 */
type RoomGagDelParam struct {
	rid uint64 // 聊天室ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type RoomGagDelReq struct {
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
func (this *UsrSvrRoomConfigCtrl) gag_del(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_GAG_SET, param.rid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

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
func (this *UsrSvrRoomConfigCtrl) Room(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "open": // 打开聊天室
		this.room_open(ctx)
		return
	case "close": // 关闭聊天室
		this.room_close(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type RoomOpenParam struct {
	rid uint64 // 聊天室ID
}

/* 打开请求 */
type RoomOpenReq struct {
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
func (this *UsrSvrRoomConfigCtrl) room_open(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_RID_ATTR, param.rid)

	_, err = rds.Do("HSET", key, "STATUS", chat.ROOM_STAT_OPEN)
	if nil != err {
		/* > 回复处理应答 */
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 参数列表 */
type RoomCloseParam struct {
	rid uint64 // 聊天室ID
}

/* 打开请求 */
type RoomCloseReq struct {
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
func (this *UsrSvrRoomConfigCtrl) room_close(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_RID_ATTR, param.rid)

	_, err = rds.Do("HSET", key, "STATUS", chat.ROOM_STAT_CLOSE)
	if nil != err {
		/* > 回复处理应答 */
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

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
func (this *UsrSvrRoomConfigCtrl) Capacity(ctx *UsrSvrCntx) {
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
func (this *UsrSvrRoomConfigCtrl) capacity_set(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_GROUP_CAP_ZSET)

	pl.Send("ZADD", key, param.capacity, param.rid)

	/* > 回复处理应答 */
	this.Error(comm.ERR_SUCC, "Ok")

	return
}

/* 请求参数 */
type RoomCapGetParam struct {
	rid      uint64 // 聊天室ID
	capacity int    // 分组容量
}

/* 请求对象 */
type RoomCapGetReq struct {
	ctrl *UsrSvrRoomConfigCtrl // 空间对象
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
func (this *UsrSvrRoomConfigCtrl) capacity_get(ctx *UsrSvrCntx) {
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
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_GROUP_CAP_ZSET)

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
		Code:   comm.ERR_SUCC,
		ErrMsg: "Ok",
	}

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

package controllers

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

/* 群组配置 */
type UsrSvrGroupConfigCtrl struct {
	BaseController
}

func (this *UsrSvrGroupConfigCtrl) Config() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "blacklist": // 群组黑名单
		this.Blacklist(ctx)
		return
	case "gag": // 群组禁言
		this.Gag(ctx)
		return
	case "switch": // 群组开关
		this.Switch(ctx)
		return
	case "capacity": // 群组容量
		this.Capacity(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
// 群组黑名单操作接口

/******************************************************************************
 **函数名称: Blacklist
 **功    能: 黑名单操作
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.23 23:27:49 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) Blacklist(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "get": // 获取群组黑名单
		this.blacklist_get(ctx)
		return
	case "add": // 添加群组黑名单
		this.blacklist_add(ctx)
		return
	case "del": // 移除群组黑名单
		this.blacklist_del(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 参数列表 */
type GroupBlacklistAddParam struct {
	gid uint64 // 群组ID
	uid uint64 // 用户UID
}

/* 加入黑名单请求 */
type GroupBlacklistAddReq struct {
	ctrl *UsrSvrGroupConfigCtrl
}

/******************************************************************************
 **函数名称: parse_param
 **功    能: 参数解析
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 参数信息
 **实现描述: 从url请求中抽取参数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.23 23:29:12 #
 ******************************************************************************/
func (req *GroupBlacklistAddReq) parse_param() (*GroupBlacklistAddParam, error) {
	this := req.ctrl
	param := &GroupBlacklistAddParam{}

	gid_str := this.GetString("gid")
	uid_str := this.GetString("uid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.gid = uint64(gid)
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
 **实现描述: 1.抽取请求参数 2.加入群组黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 09:49:16 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) blacklist_add(ctx *UsrSvrCntx) {
	req := &GroupBlacklistAddReq{ctrl: this}

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
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_BLACKLIST_SET, param.gid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type GroupBlacklistDelParam struct {
	gid uint64 // 群组ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type GroupBlacklistDelReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
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
func (req *GroupBlacklistDelReq) parse_param() (*GroupBlacklistDelParam, error) {
	this := req.ctrl
	param := &GroupBlacklistDelParam{}

	gid_str := this.GetString("gid")
	uid_str := this.GetString("uid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	param.gid = uint64(gid)
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
 **实现描述: 1.抽取请求参数 2.移除群组黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 10:11:20 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) blacklist_del(ctx *UsrSvrCntx) {
	req := &GroupBlacklistDelReq{ctrl: this}

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
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_BLACKLIST_SET, param.gid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type GroupBlacklistGetParam struct {
	gid uint64 // 群组ID
}

/* 请求对象 */
type GroupBlacklistGetReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
}

/* 应答结果 */
type GroupBlacklistGetRsp struct {
	Len    int            `json:"len"`    // 列表长度
	List   GroupBlacklist `json:"list"`   // 黑名单列表
	Code   int            `json:"code"`   // 错误码
	ErrMsg string         `json:"errmsg"` // 错误描述
}

type GroupBlacklist []GroupBlacklistItem

/* 应用列表 */
type GroupBlacklistItem struct {
	Idx int    `json:"idx"` // 索引ID
	Uid uint64 `json:"uid"` // 用户ID
}

func (list GroupBlacklist) Len() int           { return len(list) }
func (list GroupBlacklist) Less(i, j int) bool { return list[i].Uid < list[j].Uid }
func (list GroupBlacklist) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

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
func (req *GroupBlacklistGetReq) parse_param() (*GroupBlacklistGetParam, error) {
	this := req.ctrl
	param := &GroupBlacklistGetParam{}

	gid_str := this.GetString("gid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	param.gid = uint64(gid)

	return param, nil
}

/******************************************************************************
 **函数名称: blacklist_get
 **功    能: 获取黑名单
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.获取群组黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 10:11:20 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) blacklist_get(ctx *UsrSvrCntx) {
	rsp := &GroupBlacklistGetRsp{}

	req := &GroupBlacklistGetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Get blacklist failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取用户黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_BLACKLIST_SET, param.gid)

	list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, 0, "+inf"))
	if nil != err {
		ctx.log.Error("Get blacklist failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	num := len(list)
	for idx := 0; idx < num; idx += 1 {
		uid, _ := strconv.ParseInt(list[idx], 10, 64)
		if 0 == uid {
			continue
		}

		item := &GroupBlacklistItem{
			Idx: idx,
			Uid: uint64(uid),
		}

		rsp.List = append(rsp.List, *item)
	}

	sort.Sort(rsp.List) /* 按uid排序 */

	rsp.Len = num
	for idx := 0; idx < num; idx += 1 {
		rsp.List[idx].Idx = idx
	}

	/* > 回复处理应答 */
	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

////////////////////////////////////////////////////////////////////////////////
// 群组禁言操作接口

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
func (this *UsrSvrGroupConfigCtrl) Gag(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "get": // 查找禁言
		this.gag_get(ctx)
		return
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

/* 请求参数 */
type GroupGagGetParam struct {
	gid uint64 // 群组ID
}

/* 请求对象 */
type GroupGagGetReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
}

/* 应答结果 */
type GroupGagGetRsp struct {
	Len    int          `json:"len"`    // 列表长度
	List   GroupGagList `json:"list"`   // 禁言列表
	Code   int          `json:"code"`   // 错误码
	ErrMsg string       `json:"errmsg"` // 错误描述
}

type GroupGagList []GroupGagItem

/* 应用列表 */
type GroupGagItem struct {
	Idx int    `json:"idx"` // 索引ID
	Uid uint64 `json:"uid"` // 用户ID
}

func (list GroupGagList) Len() int           { return len(list) }
func (list GroupGagList) Less(i, j int) bool { return list[i].Uid < list[j].Uid }
func (list GroupGagList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

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
func (req *GroupGagGetReq) parse_param() (*GroupGagGetParam, error) {
	this := req.ctrl
	param := &GroupGagGetParam{}

	gid_str := this.GetString("gid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	param.gid = uint64(gid)

	return param, nil
}

/******************************************************************************
 **函数名称: gag_get
 **功    能: 获取禁言列表
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.获取群组黑名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 10:11:20 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) gag_get(ctx *UsrSvrCntx) {
	rsp := &GroupGagGetRsp{}

	req := &GroupGagGetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Get gag list failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取禁言用户列表 */
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_GAG_SET, param.gid)

	list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, 0, "+inf"))
	if nil != err {
		ctx.log.Error("Get gag list failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	num := len(list)
	for idx := 0; idx < num; idx += 1 {
		uid, _ := strconv.ParseInt(list[idx], 10, 64)
		if 0 == uid {
			continue
		}

		item := &GroupGagItem{
			Idx: idx,
			Uid: uint64(uid),
		}

		rsp.List = append(rsp.List, *item)
	}

	sort.Sort(rsp.List) /* 按uid排序 */

	rsp.Len = num
	for idx := 0; idx < num; idx += 1 {
		rsp.List[idx].Idx = idx
	}

	/* > 回复处理应答 */
	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

/* 参数列表 */
type GroupGagAddParam struct {
	gid uint64 // 群组ID
	uid uint64 // 用户UID
}

/* 加入禁言请求 */
type GroupGagAddReq struct {
	ctrl *UsrSvrGroupConfigCtrl
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
func (req *GroupGagAddReq) parse_param() (*GroupGagAddParam, error) {
	this := req.ctrl
	param := &GroupGagAddParam{}

	gid_str := this.GetString("gid")
	uid_str := this.GetString("uid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [rid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.gid = uint64(gid)
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
 **实现描述: 1.抽取请求参数 2.加入群组禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 11:27:21 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) gag_add(ctx *UsrSvrCntx) {
	req := &GroupGagAddReq{ctrl: this}

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
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_GAG_SET, param.gid)

	pl.Send("ZADD", key, time.Now().Unix(), param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type GroupGagDelParam struct {
	gid uint64 // 群组ID
	uid uint64 // 用户UID
}

/* 请求对象 */
type GroupGagDelReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
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
func (req *GroupGagDelReq) parse_param() (*GroupGagDelParam, error) {
	this := req.ctrl
	param := &GroupGagDelParam{}

	gid_str := this.GetString("gid")
	uid_str := this.GetString("uid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		return nil, errors.New("Paramter [uid] is invalid!")
	}

	param.gid = uint64(gid)
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
func (this *UsrSvrGroupConfigCtrl) gag_del(ctx *UsrSvrCntx) {
	req := &GroupGagDelReq{ctrl: this}

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
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_USR_GAG_SET, param.gid)

	pl.Send("ZREM", key, param.uid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////
// 群组开关操作

/******************************************************************************
 **函数名称: Switch
 **功    能: 群组开关
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:48:21 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) Switch(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "on": // 打开群组
		this.switch_on(ctx)
		return
	case "off": // 关闭群组
		this.switch_off(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
	return
}

/* 请求参数 */
type GroupSwitchParam struct {
	gid uint64 // 群组ID
}

/* 请求对象 */
type GroupSwitchReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
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
func (req *GroupSwitchReq) parse_param() (*GroupSwitchParam, error) {
	this := req.ctrl
	param := &GroupSwitchParam{}

	gid_str := this.GetString("gid")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	param.gid = uint64(gid)

	return param, nil
}

/******************************************************************************
 **函数名称: switch_on
 **功    能: 打开群组
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.修改群组状态
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.26 20:38:14 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) switch_on(ctx *UsrSvrCntx) {
	req := &GroupSwitchReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Parse switch failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 关闭群组 */
	key := fmt.Sprintf(comm.CHAT_KEY_GID_ATTR, param.gid)

	rds.Do("HSET", key, comm.CHAT_GID_ATTR_SWITCH, 1)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/******************************************************************************
 **函数名称: switch_off
 **功    能: 关闭群组
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.修改群组状态
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.26 20:47:39 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) switch_off(ctx *UsrSvrCntx) {
	req := &GroupSwitchReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Parse switch failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 关闭群组 */
	key := fmt.Sprintf(comm.CHAT_KEY_GID_ATTR, param.gid)

	rds.Do("HSET", key, "SWITCH", 0)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

////////////////////////////////////////////////////////////////////////////////
// 群组容量操作

/******************************************************************************
 **函数名称: Capacity
 **功    能: 设置群组分组容量
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 23:48:21 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) Capacity(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "set": // 设置群组容量
		this.capacity_set(ctx)
		return
	case "get": // 设置群组容量
		this.capacity_get(ctx)
		return
	}
}

/* 请求参数 */
type GroupCapSetParam struct {
	gid      uint64 // 群组ID
	capacity int    // 群组容量
}

/* 请求对象 */
type GroupCapSetReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
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
func (req *GroupCapSetReq) parse_param() (*GroupCapSetParam, error) {
	this := req.ctrl
	param := &GroupCapSetParam{}

	gid_str := this.GetString("gid")
	cap_str := this.GetString("cap")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	capacity, _ := strconv.ParseInt(cap_str, 10, 32)
	if 0 == capacity {
		return nil, errors.New("Paramter [cap] is invalid!")
	}

	param.gid = uint64(gid)
	param.capacity = int(capacity)

	return param, nil
}

/******************************************************************************
 **函数名称: capacity_set
 **功    能: 设置群组人数
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 00:00:39 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) capacity_set(ctx *UsrSvrCntx) {
	req := &GroupCapSetReq{ctrl: this}

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

	/* > 存储群组分组容量 */
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_CAP_ZSET)

	pl.Send("ZADD", key, param.capacity, param.gid)

	/* > 回复处理应答 */
	this.Error(comm.OK, "Ok")

	return
}

/* 请求参数 */
type GroupCapGetParam struct {
	gid      uint64 // 群组ID
	capacity int    // 分组容量
}

/* 请求对象 */
type GroupCapGetReq struct {
	ctrl *UsrSvrGroupConfigCtrl // 空间对象
}

/* 请求应答 */
type GroupCapGetRsp struct {
	Gid    uint64 `json:"gid"`    // 群组ID
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
func (req *GroupCapGetReq) parse_param() (*GroupCapGetParam, error) {
	this := req.ctrl
	param := &GroupCapGetParam{}

	gid_str := this.GetString("gid")
	cap_str := this.GetString("cap")

	gid, _ := strconv.ParseInt(gid_str, 10, 64)
	if 0 == gid {
		return nil, errors.New("Paramter [gid] is invalid!")
	}

	capacity, _ := strconv.ParseInt(cap_str, 10, 32)
	if 0 == capacity {
		return nil, errors.New("Paramter [cap] is invalid!")
	}

	param.gid = uint64(gid)
	param.capacity = int(capacity)

	return param, nil
}

/******************************************************************************
 **函数名称: capacity_get
 **功    能: 获取群组人数
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 1.抽取请求参数 2.移除禁言名单
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.18 00:00:39 #
 ******************************************************************************/
func (this *UsrSvrGroupConfigCtrl) capacity_get(ctx *UsrSvrCntx) {
	req := &GroupCapGetReq{ctrl: this}

	param, err := req.parse_param()
	if nil != err {
		ctx.log.Error("Get group capacity failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SVR_INVALID_PARAM, err.Error())
		return
	}

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 存储群组容量 */
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_CAP_ZSET)

	capacity, err := redis.Int(rds.Do("ZSCORE", key, param.gid))
	if nil != err {
		ctx.log.Error("Get group capacity failed! errmsg:%s", err.Error())
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	rsp := &GroupCapGetRsp{
		Gid:    param.gid,
		Cap:    capacity,
		Code:   comm.OK,
		ErrMsg: "Ok",
	}

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

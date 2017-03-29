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

/* 系统配置 */
type UsrSvrConfigCtrl struct {
	BaseController
}

func (this *UsrSvrConfigCtrl) Config() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "frwder": // 转发层操作
		this.Frwder(ctx)
		return
	case "listend": // 侦听层操作
		this.Listend(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

// 转发层操作

/******************************************************************************
 **函数名称: Frwder
 **功    能: 转发层操作
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.22 22:23:26 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) Frwder(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "list": // 转发层列表
		this.ListFrwder(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
}

/* 应答结果 */
type FrwderListRsp struct {
	Len    int        `json:"len"`    // 列表长度
	List   FrwderList `json:"list"`   // 分组列表
	Code   int        `json:"code"`   // 错误码
	ErrMsg string     `json:"errmsg"` // 错误描述
}

type FrwderList []FrwderListItem

/* 应用列表 */
type FrwderListItem struct {
	Idx     int    `json:"idx"`      // 索引IDX
	Nid     uint32 `json:"nid"`      // 分组列表
	IpAddr  string `json:"ipaddr"`   // IP
	FwdPort uint16 `json:"fwd-port"` // FORWARD-PORT
	BcPort  uint16 `json:"bc-port"`  // BACKEND-PORT
	Status  uint32 `json:"status"`   // 当前状态
}

func (list FrwderList) Len() int           { return len(list) }
func (list FrwderList) Less(i, j int) bool { return list[i].Nid < list[j].Nid }
func (list FrwderList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

/******************************************************************************
 **函数名称: ListFrwder
 **功    能: 查询转发层列表
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.22 21:13:52 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) ListFrwder(ctx *UsrSvrCntx) {
	rsp := &FrwderListRsp{}

	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 获取转发层列表 */
	nodes, err := redis.Strings(rds.Do(
		"ZRANGEBYSCORE", comm.IM_KEY_FRWD_NID_ZSET, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get frwder list failed! errmsg:%s", err.Error())
		return
	}

	num := len(nodes)
	for idx := 0; idx < num; idx += 2 {
		var item FrwderListItem

		nid, _ := strconv.ParseInt(nodes[idx], 10, 64)

		key := fmt.Sprintf(comm.IM_KEY_FRWD_ATTR, nid)

		vals, err := redis.Strings(rds.Do("HMGET", key,
			comm.IM_FRWD_ATTR_ADDR, comm.IM_FRWD_ATTR_FWD_PORT, comm.IM_FRWD_ATTR_BC_PORT))
		if nil != err {
			continue
		}

		ttl, _ := strconv.ParseInt(nodes[idx+1], 10, 64)
		fwd_port, _ := strconv.ParseInt(vals[1], 10, 32)
		bc_port, _ := strconv.ParseInt(vals[2], 10, 32)

		item.Nid = uint32(nid)
		item.IpAddr = vals[0]
		if ttl < ctm {
			item.Status = uint32(comm.PROC_STATUS_EXIT)
		} else {
			item.Status = uint32(comm.PROC_STATUS_EXEC)
		}
		item.FwdPort = uint16(fwd_port)
		item.BcPort = uint16(bc_port)

		rsp.List = append(rsp.List, item)
	}

	sort.Sort(rsp.List)
	rsp.Len = len(rsp.List)
	rsp.Code = 0
	rsp.ErrMsg = "Ok"
	for idx := 0; idx < rsp.Len; idx += 1 {
		item := &rsp.List[idx]

		item.Idx = idx + 1
	}

	this.Data["json"] = rsp
	this.ServeJSON()
	return
}

// 侦听层操作

/******************************************************************************
 **函数名称: Listend
 **功    能: 侦听层操作
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 根据action调用对应的处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.22 22:23:26 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) Listend(ctx *UsrSvrCntx) {
	action := this.GetString("action")
	switch action {
	case "list": // 侦听层列表
		this.ListListend(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this action:%s.", action))
}

/* 应答结果 */
type ListendListRsp struct {
	Len    int         `json:"len"`    // 列表长度
	List   ListendList `json:"list"`   // 分组列表
	Code   int         `json:"code"`   // 错误码
	ErrMsg string      `json:"errmsg"` // 错误描述
}

type ListendList []ListendListItem

/* 应用列表 */
type ListendListItem struct {
	Idx    int    `json:"idx"`    // 索引IDX
	Nid    uint32 `json:"nid"`    // 分组列表
	Type   uint32 `json:"type"`   // 侦听层类型(0:未知 1:TCP 2:WS)
	IpAddr string `json:"ipaddr"` // IP+PORT
	Status uint32 `json:"status"` // 当前状态
	Total  uint32 `json:"total"`  // 在线人数
}

func (list ListendList) Len() int           { return len(list) }
func (list ListendList) Less(i, j int) bool { return list[i].Nid < list[j].Nid }
func (list ListendList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

/******************************************************************************
 **函数名称: ListListend
 **功    能: 侦听层操作
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.22 21:13:52 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) ListListend(ctx *UsrSvrCntx) {
	rsp := &ListendListRsp{}

	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 获取侦听层列表 */
	nodes, err := redis.Strings(rds.Do(
		"ZRANGEBYSCORE", comm.IM_KEY_LSND_NID_ZSET, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get listend list failed! errmsg:%s", err.Error())
		return
	}

	num := len(nodes)
	for idx := 0; idx < num; idx += 2 {
		var item ListendListItem

		nid, _ := strconv.ParseInt(nodes[idx], 10, 64)

		key := fmt.Sprintf(comm.IM_KEY_LSND_ATTR, nid)

		vals, err := redis.Strings(rds.Do("HMGET", key,
			comm.IM_LSND_ATTR_TYPE, comm.IM_LSND_ATTR_ADDR,
			comm.IM_LSND_ATTR_STATUS, comm.IM_LSND_ATTR_CONNECTION))
		if nil != err {
			continue
		}

		ttl, _ := strconv.ParseInt(nodes[idx+1], 10, 64)

		typ, _ := strconv.ParseInt(vals[0], 10, 32)
		item.Type = uint32(typ)
		item.Nid = uint32(nid)
		item.IpAddr = vals[1]
		status, _ := strconv.ParseInt(vals[2], 10, 32)
		if ttl < ctm {
			item.Status = uint32(comm.PROC_STATUS_EXIT)
		} else {
			item.Status = uint32(status)
		}
		total, _ := strconv.ParseInt(vals[3], 10, 32)
		item.Total = uint32(total)

		rsp.List = append(rsp.List, item)
	}

	sort.Sort(rsp.List)
	rsp.Len = len(rsp.List)
	rsp.Code = 0
	rsp.ErrMsg = "Ok"
	for idx := 0; idx < rsp.Len; idx += 1 {
		item := &rsp.List[idx]

		item.Idx = idx + 1
	}

	this.Data["json"] = rsp
	this.ServeJSON()
	return
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
	this.Error(comm.OK, "Ok")

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
	this.Error(comm.OK, "Ok")

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
	key := fmt.Sprintf(comm.IM_KEY_PREC_USR_MAX_NUM, prec)

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

/* 添加请求 */
type UserStatisListReq struct {
	ctrl *UsrSvrConfigCtrl
}

/* 应答结果 */
type UserStatisListRsp struct {
	Len    int                `json:"len"`    // 统计条数
	List   UserStatisPrecList `json:"list"`   // 统计结果
	Code   int                `json:"code"`   // 返回码
	ErrMsg string             `json:"errmsg"` // 错误描述
}

type UserStatisPrecList []UserStatisPrecItem

/* 统计结果 */
type UserStatisPrecItem struct {
	Idx  int `json:"idx"`  // 序列号
	Prec int `json:"prec"` // 统计精度
	Num  int `json:"num"`  // 最大记录数
}

func (list UserStatisPrecList) Len() int           { return len(list) }
func (list UserStatisPrecList) Less(i, j int) bool { return list[i].Prec < list[j].Prec }
func (list UserStatisPrecList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

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
func (req *UserStatisListReq) query(ctx *UsrSvrCntx) (UserStatisPrecList, error) {
	var list UserStatisPrecList

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取统计精度 */
	key := fmt.Sprintf(comm.IM_KEY_PREC_NUM_ZSET)

	data, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		return list, err
	}

	data_len := len(data)
	for idx := 0; idx < data_len; idx += 2 {
		prec, _ := strconv.Atoi(data[idx])
		num, _ := strconv.Atoi(data[idx+1])

		item := UserStatisPrecItem{
			Prec: int(prec),
			Num:  int(num),
		}

		list = append(list, item)
	}

	sort.Sort(list) /* 排序处理 */

	/* > 整理统计精度 */
	num := len(list)
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
 **实现描述: 1.抽取请求参数 2.获取统计精度 3.返回统计精度
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.20 18:51:38 #
 ******************************************************************************/
func (this *UsrSvrConfigCtrl) user_statis_list(ctx *UsrSvrCntx) {
	req := &UserStatisListReq{ctrl: this}

	/* > 获取统计结果 */
	list, err := req.query(ctx)
	if nil != err {
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	/* > 回复处理应答 */
	rsp := &UserStatisListRsp{}

	rsp.Len = len(list)
	rsp.List = list
	rsp.Code = 0
	rsp.ErrMsg = "Ok"

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

package controllers

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

type UsrSvrQueryCtrl struct {
	BaseController
}

func (this *UsrSvrQueryCtrl) Query() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "sid-list":
		this.SidList(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s", option))
	return
}

/* 请求对象 */
type SidListGetReq struct {
	ctrl *UsrSvrQueryCtrl // 空间对象
}

/* 应答结果 */
type SidListGetRsp struct {
	Uid    uint64  `json:"uid"`    // 用户ID
	Len    int     `json:"len"`    // 列表长度
	List   SidList `json:"list"`   // SID列表
	Code   int     `json:"code"`   // 错误码
	ErrMsg string  `json:"errmsg"` // 错误描述
}

type SidList []SidListItem

/* 应用列表 */
type SidListItem struct {
	Idx int    `json:"idx"` // 索引ID
	Sid uint64 `json:"sid"` // 会话ID
}

func (list SidList) Len() int           { return len(list) }
func (list SidList) Less(i, j int) bool { return list[i].Sid < list[j].Sid }
func (list SidList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

/******************************************************************************
 **函数名称: SidList
 **功    能: 获取SID列表
 **输入参数:
 **     ctx: 上下文
 **输出参数:
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.26 21:37:08 #
 ******************************************************************************/
func (this *UsrSvrQueryCtrl) SidList(ctx *UsrSvrCntx) {
	rsp := &SidListGetRsp{}

	uid_str := this.GetString("uid")
	if "" == uid_str {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter [uid] is invalied!")
		return
	}

	uid, _ := strconv.ParseInt(uid_str, 10, 64)
	if 0 == uid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Paramter [uid] is invalied!")
		return
	}

	rsp.Uid = uint64(uid)

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取会话SID列表 */
	key := fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, rsp.Uid)

	list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		this.Error(comm.ERR_SYS_SYSTEM, err.Error())
		return
	}

	num := len(list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(list[idx], 10, 64)
		if 0 == sid {
			continue
		}

		item := &SidListItem{
			Idx: idx,
			Sid: uint64(sid),
		}

		rsp.List = append(rsp.List, *item)
	}

	sort.Sort(rsp.List)

	for idx := 0; idx < num; idx += 1 {
		rsp.List[idx].Idx = idx
	}

	/* 回复应答 */
	rsp.Len = num
	rsp.Code = 0
	rsp.ErrMsg = "Ok"

	this.Data["json"] = rsp
	this.ServeJSON()

	return
}

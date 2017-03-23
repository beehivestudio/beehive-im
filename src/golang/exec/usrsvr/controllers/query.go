package controllers

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

type UsrSvrQueryCtrl struct {
	BaseController
}

func (this *UsrSvrQueryCtrl) Query() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	}
	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s", option))
}

type UsrSvrRoomQueryCtrl struct {
	BaseController
}

func (this *UsrSvrRoomQueryCtrl) Query() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "top-list":
		this.TopList(ctx)
		return
	}
	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s", option))
}

/* 应答结果 */
type RoomTopListRsp struct {
	Len    int         `json:"len"`    // 列表长度
	List   RoomTopList `json:"list"`   // 分组列表
	Code   int         `json:"code"`   // 错误码
	ErrMsg string      `json:"errmsg"` // 错误描述
}

type RoomTopList []RoomTopItem

/* 应用列表 */
type RoomTopItem struct {
	Idx   int    `json:"idx"`   // 索引IDX
	Rid   uint64 `json:"rid"`   // 聊天室ID
	Total uint32 `json:"total"` // 在线人数
}

func (list RoomTopList) Len() int           { return len(list) }
func (list RoomTopList) Less(i, j int) bool { return list[i].Total < list[j].Total }
func (list RoomTopList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

func (this *UsrSvrRoomQueryCtrl) TopList(ctx *UsrSvrCntx) {
	rsp := &RoomTopListRsp{}

	rds := ctx.redis.Get()
	defer rds.Close()

	num, err := this.GetInt("num")
	if nil != err {
		num = 10 // 默认抽取排名前10的聊天室
	}

	/* > 获取聊天室列表 */
	off := 0
	ctm := time.Now().Unix()
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, ctm, "+inf", "LIMIT", comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get room list failed! errmsg:%s", err.Error())
			break
		}

		count := len(rid_list)
		for idx := 0; idx < count; idx += 1 {
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			if 0 == rid {
				continue
			}

			key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, uint64(rid))

			total, err := redis.Int(rds.Do("ZCARD", key))
			if nil != err {
				ctx.log.Error("Get room user number failed! rid:%d", uint64(rid))
				continue
			}

			item := &RoomTopItem{
				Idx:   idx,
				Rid:   uint64(rid),
				Total: uint32(total),
			}

			rsp.List = append(rsp.List, *item)
		}

		if count < comm.CHAT_BAT_NUM {
			break
		}

		off += count
	}

	/* > 进行排序处理 */
	sort.Sort(rsp.List)

	if len(rsp.List) > num {
		rsp.List = rsp.List[0 : num-1]
	}

	num = len(rsp.List)
	for idx := 0; idx < num; idx += 1 {
		rsp.List[idx].Idx = idx
	}

	this.Data["json"] = rsp
	this.ServeJSON()
	return
}

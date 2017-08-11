package controllers

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"

	"beehive-im/src/golang/exec/chatroom/models"
)

// 聊天室信息查询
type ChatRoomQueryCtrl struct {
	BaseController
}

func (this *ChatRoomQueryCtrl) Query() {
	ctx := GetRoomSvrCntx()

	option := this.GetString("option")
	switch option {
	case "top-list":
		this.TopList(ctx)
		return
	case "group-list":
		this.GroupList(ctx)
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

/******************************************************************************
 **函数名称: TopList
 **功    能: 聊天室TOP-LIST排行
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 获取聊天室和对应人数, 并按人数从"多->少"排行
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.23 22:51:05 #
 ******************************************************************************/
func (this *ChatRoomQueryCtrl) TopList(ctx *ChatRoomCntx) {
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
			models.ROOM_KEY_RID_ZSET, ctm, "+inf", "LIMIT", comm.CHAT_BAT_NUM))
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

			key := fmt.Sprintf(models.ROOM_KEY_RID_TO_SID_ZSET, uint64(rid))

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

/* 应答结果 */
type RoomGroupListRsp struct {
	Rid    uint64        `json:"rid"`    // 聊天室ID
	Len    int           `json:"len"`    // 列表长度
	List   RoomGroupList `json:"list"`   // 分组列表
	Code   int           `json:"code"`   // 错误码
	ErrMsg string        `json:"errmsg"` // 错误描述
}

type RoomGroupList []RoomGroupItem

/* 应用列表 */
type RoomGroupItem struct {
	Idx   int    `json:"idx"`   // 索引IDX
	Gid   uint32 `json:"gid"`   // 分组ID
	Total uint32 `json:"total"` // 在线人数
}

func (list RoomGroupList) Len() int           { return len(list) }
func (list RoomGroupList) Less(i, j int) bool { return list[i].Gid < list[j].Gid }
func (list RoomGroupList) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }

/******************************************************************************
 **函数名称: GroupList
 **功    能: 聊天室分组信息
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 获取各组和对应人数, 并按分组序列号排序
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.23 22:51:05 #
 ******************************************************************************/
func (this *ChatRoomQueryCtrl) GroupList(ctx *ChatRoomCntx) {
	rsp := &RoomGroupListRsp{}

	rds := ctx.redis.Get()
	defer rds.Close()

	rid_str := this.GetString("rid")
	if "" == rid_str {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Rid is invalid!")
		return
	}

	rid, _ := strconv.ParseInt(rid_str, 10, 64)
	if 0 == rid {
		this.Error(comm.ERR_SVR_INVALID_PARAM, "Rid is invalid!")
		return
	}

	/* 获取聊天室分组列表 */
	off := 0
	key := fmt.Sprintf(models.ROOM_KEY_RID_GID_TO_NUM_ZSET, rid)
	for {
		gid_num_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			key, "-inf", "+inf", "WITHSCORES", "LIMIT", comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get room list failed! errmsg:%s", err.Error())
			break
		}

		num := len(gid_num_list) / 2
		for idx := 0; idx < num; idx += 1 {
			gid, _ := strconv.ParseInt(gid_num_list[2*idx], 10, 64)
			total, _ := strconv.ParseInt(gid_num_list[2*idx+1], 10, 32)

			item := &RoomGroupItem{
				Idx:   idx,
				Gid:   uint32(gid),
				Total: uint32(total),
			}

			rsp.List = append(rsp.List, *item)
		}

		if num < comm.CHAT_BAT_NUM {
			break
		}

		off += num
	}

	/* > 进行排序处理 */
	sort.Sort(rsp.List)

	num := len(rsp.List)
	for idx := 0; idx < num; idx += 1 {
		rsp.List[idx].Idx = idx
	}

	this.Data["json"] = rsp
	this.ServeJSON()
	return
}

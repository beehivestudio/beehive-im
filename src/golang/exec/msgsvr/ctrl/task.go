package ctrl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: update
 **功    能: 启动update服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 01:00:10 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) update() {
	for {
		ctx.update_rid_to_nid_map()

		time.Sleep(30 * time.Second)
	}
}

/******************************************************************************
 **函数名称: update_rid_to_nid_map
 **功    能: 更新RID->NID映射表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 00:21:54 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) update_rid_to_nid_map() {
	var items map[uint64]MsgSvrRidToNidItem

	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	ctm := time.Now().Unix()
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, ctm, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get rid zset failed! errmsg:%s", err.Error())
			return
		}

		rid_num := len(rid_list)
		for idx := 0; idx < rid_num; idx += 1 {
			rid_int, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			rid := uint64(rid_int)
			key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, rid)
			nid_list, err := redis.Ints(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
			if nil != err {
				ctx.log.Error("Get nid list by rid failed! errmsg:%s", err.Error())
				return
			} else if 0 == len(nid_list) {
				break
			}

			var item MsgSvrRidToNidItem

			item.nid_list = nid_list
			items[rid] = item
		}

		if rid_num < comm.CHAT_BAT_NUM {
			break
		}
		off += rid_num
	}

	ctx.rid_to_nid_map.Lock()
	ctx.rid_to_nid_map.items = items
	ctx.rid_to_nid_map.Unlock()
}

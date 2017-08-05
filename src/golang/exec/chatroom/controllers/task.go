package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"

	"beehive-im/src/golang/exec/chatroom/models"
)

/******************************************************************************
 **函数名称: task
 **功    能: 开启定时任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 08:20:07 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) task() {
	go ctx.task_room_mesg_chan_pop()
	go ctx.task_room_mesg_queue_clean()

	/* 每1秒执行一次任务 */
	go func() {
		for {
			ctx.listend_dict_update()                             // 更新侦听层字典
			ctx.listend_list_update()                             // 更新侦听层列表
			models.RoomSendUsrNum(ctx.log, ctx.frwder, ctx.redis) // 下发聊天室人数

			time.Sleep(time.Second)
		}
	}()

	/* 每30秒执行一次任务 */
	go func() {
		for {
			ctm := time.Now().Unix()
			ctx.clean_rid_zset(ctm) // 定时清理超时聊天室

			time.Sleep(30 * time.Second)
		}
	}()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: listend_dict_update
 **功    能: 更新侦听层字典
 **输入参数: NONE
 **输出参数: NONE
 **返    回: NONE
 **实现描述:
 **     1. 获取所有侦听层网络类型.
 **     2. 根据侦听层网络类型获取侦听层的列表.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:11:08 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) listend_dict_update() {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 获取"网络类型"列表 */
	types, err := redis.Ints(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_LSND_TYPE_ZSET, ctm, "+inf"))
	if nil != err {
		/* 清理所有数据 */
		ctx.listend.dict.Lock()
		defer ctx.listend.dict.Unlock()
		ctx.listend.dict.types = make(map[int]*ChatRoomLsndDictItem)
		ctx.log.Error("Get listend type list failed! errmsg:%s", err.Error())
		return
	}

	/* > 清理所有数据 */
	ctx.listend.dict.Lock()
	defer ctx.listend.dict.Unlock()
	ctx.listend.dict.types = make(map[int]*ChatRoomLsndDictItem)

	/* > 重新设置数据 */
	num := len(types)
	for idx := 0; idx < num; idx += 1 {
		typ := types[idx]

		list := ctx.listend_dict_fetch(typ)
		if nil == list {
			ctx.log.Error("Get listen list failed! type:%d", typ)
			delete(ctx.listend.dict.types, typ)
			continue
		}

		ctx.listend.dict.types[typ] = list
	}
}

/******************************************************************************
 **函数名称: listend_dict_fetch
 **功    能: 获取侦听层字典
 **输入参数:
 **     typ: 网络类型(0:Unkonwn 1:TCP 2:WS)
 **输出参数: NONE
 **返    回: 侦听层IP列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:09:55 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) listend_dict_fetch(typ int) *ChatRoomLsndDictItem {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	lsnd := &ChatRoomLsndDictItem{
		list: make(map[string](map[uint32][]string)),
	}

	/* > 获取"国家/地区"列表 */
	key := fmt.Sprintf(comm.IM_KEY_LSND_NATION_ZSET, typ)

	nations, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
	if nil != err {
		ctx.log.Error("Get nation list failed! errmsg:%s", err.Error())
		return nil
	}

	nation_num := len(nations)
	for m := 0; m < nation_num; m += 1 {
		//ctx.log.Debug("Nation:%s", nations[m])
		/* > 获取"国家/地区"对应的"运营商"列表 */
		key := fmt.Sprintf(comm.IM_KEY_LSND_OP_ZSET, typ, nations[m])

		operators, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
		if nil != err {
			ctx.log.Error("Get operator list by nation failed! errmsg:%s", err.Error())
			return nil
		}

		operator_set := make(map[uint32][]string, 0)

		operator_num := len(operators)
		for n := 0; n < operator_num; n += 1 {
			opid, _ := strconv.ParseInt(operators[n], 10, 32)

			//ctx.log.Debug("    Operator:%d", uint32(opid))
			/* > 获取"运营商"对应的"IP+PORT"列表 */
			key := fmt.Sprintf(comm.IM_KEY_LSND_IP_ZSET, typ, nations[m], uint32(opid))

			iplist, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
			if nil != err {
				ctx.log.Error("Get operator list by nation failed! errmsg:%s", err.Error())
				return nil
			}

			iplist_num := len(iplist)
			for k := 0; k < iplist_num; k += 1 {
				//ctx.log.Debug("    iplist:%s", iplist[k])
				operator_set[uint32(opid)] = append(operator_set[uint32(opid)], iplist[k])
			}
		}

		lsnd.list[nations[m]] = operator_set
	}

	return lsnd
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: get_lsn_id_list
 **功    能: 更新侦听层ID列表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 侦听层ID列表
 **实现描述: 获取有效的侦听层ID列表
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:48:23 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) listend_list_update() {
	var list []uint32

	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 获取侦听层列表 */
	nodes, err := redis.Strings(rds.Do(
		"ZRANGEBYSCORE", comm.IM_KEY_LSND_NID_ZSET, ctm, "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get listend list failed! errmsg:%s", err.Error())
		ctx.listend.list.Lock()
		defer ctx.listend.list.Unlock()
		ctx.listend.list.nodes = make([]uint32, 0)
		return
	}

	num := len(nodes)
	for idx := 0; idx < num; idx += 2 {
		nid, _ := strconv.ParseInt(nodes[idx], 10, 32)
		list = append(list, uint32(nid))
	}

	ctx.listend.list.Lock()
	defer ctx.listend.list.Unlock()

	ctx.listend.list.nodes = list

	return
}

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
func (ctx *ChatRoomCntx) update() {
	for {
		ctx.update_rid_to_nid_map()

		time.Sleep(5 * time.Second)
	}
}

/******************************************************************************
 **函数名称: update_rid_to_nid_map
 **功    能: 更新聊天室RID->NID映射表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项: 为减小锁的粒度, 获取rid->nid的映射时, 无需加锁.
 **作    者: # Qifeng.zou # 2016.11.05 00:21:54 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) update_rid_to_nid_map() {
	m, err := models.RoomGetRidToNidMap(ctx.redis)
	if nil != err {
		ctx.log.Error("Get rid to nid map failed! errmsg:%s", err.Error())
		return
	}

	ctx.room.node.Lock()
	ctx.room.node.m = m
	ctx.room.node.Unlock()
}

/******************************************************************************
 **函数名称: clean_by_rid
 **功    能: 通过RID清理资源
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 12:08:43 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) clean_by_rid(rid uint64) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(models.ROOM_KEY_RID_GID_TO_NUM_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(models.ROOM_KEY_RID_NID_TO_NUM_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_NID_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_UID_SID_ZSET, rid)
	ctx.clean_uid_by_rid(rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_SID_ZSET, rid)
	ctx.clean_sid_by_rid(rid)
	pl.Send("DEL", key)

	pl.Send("ZREM", models.ROOM_KEY_RID_ZSET, rid)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: clean_sid_by_rid
 **功    能: 通过RID清理SID数据
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:16:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) clean_sid_by_rid(rid uint64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取RID->SID集合 */
	off := 0
	for {
		key := fmt.Sprintf(models.ROOM_KEY_RID_TO_SID_ZSET, rid)
		sid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			key, "-inf", "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		sid_len := len(sid_list)
		for idx := 0; idx < sid_len; idx += 1 {
			/* > 逐一清理SID->RID记录 */
			sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)
			key = fmt.Sprintf(models.ROOM_KEY_SID_TO_RID_ZSET, sid)
			pl.Send("ZREM", key, rid)
		}

		if sid_len < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
		pl.Do("")
	}
}

/******************************************************************************
 **函数名称: clean_uid_by_rid
 **功    能: 通过RID清理UID数据
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:16:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) clean_uid_by_rid(rid uint64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取RID->UID集合 */
	off := 0
	for {
		key := fmt.Sprintf(models.ROOM_KEY_RID_TO_UID_SID_ZSET, rid)
		uid_sid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			key, "-inf", "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		uid_sid_num := len(uid_sid_list)
		for idx := 0; idx < uid_sid_num; idx += 1 {
			vals := strings.Split(uid_sid_list[idx], ":")
			if 2 != len(vals) {
				ctx.log.Error("Format of [uid:sid] is invalid! str:%s", uid_sid_list[idx])
				continue
			}

			uid_str := vals[0] // 用户ID
			sid_str := vals[1] // 会话ID

			ctx.log.Debug("uid:%d sid:%d", uid_str, sid_str)
		}

		if uid_sid_num < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
		pl.Do("")
	}
}

/******************************************************************************
 **函数名称: clean_rid_zset
 **功    能: 清理聊天室RID资源
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 12:08:43 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) clean_rid_zset(ctm int64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			models.ROOM_KEY_RID_ZSET, "-inf", ctm,
			"LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		rid_num := len(rid_list)
		for idx := 0; idx < rid_num; idx += 1 {
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			ctx.clean_by_rid(uint64(rid))
		}

		if rid_num < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
	}
}

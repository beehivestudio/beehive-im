package ctrl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: timer_clean
 **功    能: 定时清理操作
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 12:08:43 #
 ******************************************************************************/
func (ctx *TaskerCntx) timer_clean() {
	for {
		ctm := time.Now().Unix()

		ctx.clean_sid_zset(ctm)
		ctx.clean_rid_zset(ctm)
		ctx.clean_uid_zset(ctm)

		time.Sleep(30)
	}
	return
}

/******************************************************************************
 **函数名称: clean_by_sid
 **功    能: 通过SID清理资源
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 12:08:43 #
 ******************************************************************************/
func (ctx *TaskerCntx) clean_by_sid(sid uint64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID对应的数据 */
	key := fmt.Sprintf(comm.CHAT_KEY_SID_ATTR, sid)

	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		ctx.log.Error("Get sid attr failed! errmsg:%s", err.Error())
		return
	}

	id, _ := strconv.ParseInt(vals[0], 10, 64)
	uid := uint64(id)
	id, _ = strconv.ParseInt(vals[1], 10, 64)
	nid := uint32(id)

	ctx.log.Debug("Delete sid [%d] data! uid:%d nid:%d", sid, uid, nid)

	/* > 删除SID对应的数据 */
	num, err := redis.Int(rds.Do("DEL", key))
	if nil != err {
		ctx.log.Error("Delete sid attr failed! errmsg:%s", err.Error())
		return
	} else if 0 == num {
		ctx.log.Error("Sid [%d] was deleted!", sid)
		pl.Send("ZREM", comm.CHAT_KEY_SID_ZSET, sid)
		return
	}

	/* > 清理相关资源 */
	key = fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)
	rid_gid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get rid list by sid failed! sid:%d", sid)
		return
	}

	rid_num := len(rid_gid_list)
	for idx := 0; idx < rid_num; idx += 2 {
		rid, _ := strconv.ParseInt(rid_gid_list[idx], 10, 64)
		gid, _ := strconv.ParseInt(rid_gid_list[idx+1], 10, 64)

		/* 更新统计计数 */
		key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, rid)
		pl.Send("ZINCRBY", key, -1, gid)
		key = fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, rid)
		pl.Send("ZINCRBY", key, -1, nid)
	}

	key = fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, sid)
	pl.Send("ZREM", key, sid)

	pl.Send("ZREM", comm.CHAT_KEY_SID_ZSET, sid)
}

/******************************************************************************
 **函数名称: clean_sid_zset
 **功    能: 清理会话SID资源
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 12:08:43 #
 ******************************************************************************/
func (ctx *TaskerCntx) clean_sid_zset(ctm int64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		sid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_SID_ZSET, "-inf", ctm,
			"LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		sid_num := len(sid_list)
		for idx := 0; idx < sid_num; idx += 1 {
			sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)
			ctx.clean_by_sid(uint64(sid))
		}

		if sid_num < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
	}
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
func (ctx *TaskerCntx) clean_sid_by_rid(rid uint64) {
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
		key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, rid)
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
			key = fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)
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
func (ctx *TaskerCntx) clean_uid_by_rid(rid uint64) {
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
		key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_ZSET, rid)
		uid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			key, "-inf", "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		uid_len := len(uid_list)
		for idx := 0; idx < uid_len; idx += 1 {
			/* > 逐一清理UID->RID记录 */
			uid, _ := strconv.ParseInt(uid_list[idx], 10, 64)
			key = fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, uid)
			pl.Send("HDEL", key, rid)
		}

		if uid_len < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
		pl.Do("")
	}
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
func (ctx *TaskerCntx) clean_by_rid(rid uint64) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_ZSET, rid)
	ctx.clean_uid_by_rid(rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, rid)
	ctx.clean_sid_by_rid(rid)
	pl.Send("DEL", key)

	pl.Send("ZREM", comm.CHAT_KEY_RID_ZSET, rid)
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
func (ctx *TaskerCntx) clean_rid_zset(ctm int64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, "-inf", ctm,
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

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: clean_by_uid
 **功    能: 通过UID清理资源
 **输入参数:
 **     uid: 用户UID
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:36:41 #
 ******************************************************************************/
func (ctx *TaskerCntx) clean_by_uid(uid uint64) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID对应的数据 */
	key := fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, uid)
	pl.Send("DEL", key)

	pl.Send("ZREM", comm.CHAT_KEY_UID_ZSET, uid)
}

/******************************************************************************
 **函数名称: clean_uid_zset
 **功    能: 清理UID资源
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:32:11 #
 ******************************************************************************/
func (ctx *TaskerCntx) clean_uid_zset(ctm int64) {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		uid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_UID_ZSET, "-inf", ctm, "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		uid_num := len(uid_list)
		for idx := 0; idx < uid_num; idx += 1 {
			uid, _ := strconv.ParseInt(uid_list[idx], 10, 64)
			ctx.clean_by_uid(uint64(uid))
		}

		if uid_num < comm.CHAT_BAT_NUM {
			break
		}
		off += comm.CHAT_BAT_NUM
	}
}

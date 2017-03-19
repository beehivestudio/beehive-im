package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
)

/******************************************************************************
 **函数名称: timer_clean
 **功    能: 定时清理操作
 **输入参数: NONE
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

		time.Sleep(30 * time.Second)
	}
	return
}

/******************************************************************************
 **函数名称: timer_update
 **功    能: 定时更新操作
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:46:18 #
 ******************************************************************************/
func (ctx *TaskerCntx) timer_update() {
	for {
		ctx.update_prec_statis()

		time.Sleep(30 * time.Second)
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

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
			comm.IM_KEY_SID_ZSET, "-inf", ctm,
			"LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get sid list failed! errmsg:%s", err.Error())
			return
		}

		sid_num := len(sid_list)
		for idx := 0; idx < sid_num; idx += 1 {
			sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)
			im.CleanSidData(ctx.redis, uint64(sid))
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
		key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, rid)
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

			/* > 逐一清理UID->RID记录 */
			uid, _ := strconv.ParseInt(uid_str, 10, 64)
			key = fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, uid)
			pl.Send("HDEL", key, rid)
		}

		if uid_sid_num < comm.CHAT_BAT_NUM {
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

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, rid)
	pl.Send("DEL", key)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, rid)
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

	pl.Send("ZREM", comm.IM_KEY_UID_ZSET, uid)
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
			comm.IM_KEY_UID_ZSET, "-inf", ctm, "LIMIT", off, comm.CHAT_BAT_NUM))
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

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: update_prec_statis
 **功    能: 更新各精度用户数
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:49:10 #
 ******************************************************************************/
func (ctx *TaskerCntx) update_prec_statis() {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	defer ctx.clean_prec_statis()

	/* > 获取当前并发数 */
	sid_num, err := redis.Int64(rds.Do("ZCARD", comm.IM_KEY_SID_ZSET))
	if nil != err {
		ctx.log.Error("Get sid num failed! errmsg:%s", err.Error())
		return
	}

	/* > 遍历统计精度列表 */
	prec_rnum_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_PREC_NUM_ZSET, 0, "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get prec list failed! errmsg:%s", err.Error())
		return
	}

	ctm := uint64(time.Now().Unix())

	prec_num := len(prec_rnum_list)
	for idx := 0; idx < prec_num; idx += 2 {
		prec, _ := strconv.ParseInt(prec_rnum_list[idx], 10, 64)
		rnum, _ := strconv.ParseInt(prec_rnum_list[idx+1], 10, 64)
		if 0 == prec || 0 == rnum {
			continue
		}

		seg := (ctm / uint64(prec)) * uint64(prec)

		/* > 更新最大峰值 */
		key := fmt.Sprintf(comm.IM_KEY_PREC_USR_MAX_NUM, prec)
		has, err := redis.Bool(rds.Do("HEXISTS", key, seg))
		if nil != err {
			ctx.log.Error("Exec hexists failed! errmsg:%s", err.Error())
			break
		} else if false == has {
			pl.Send("HSET", key, seg, sid_num)
		}

		max, err := redis.Int64(rds.Do("HGET", key, seg))
		if nil != err {
			ctx.log.Error("Get max num failed! errmsg:%s", err.Error())
			continue
		} else if max <= sid_num {
			pl.Send("HSET", key, seg, sid_num)
		}

		/* > 更新最低峰值 */
		key = fmt.Sprintf(comm.IM_KEY_PREC_USR_MIN_NUM, prec)
		has, err = redis.Bool(rds.Do("HEXISTS", key, seg))
		if nil != err {
			ctx.log.Error("Exec hexists failed! errmsg:%s", err.Error())
			break
		} else if false == has {
			pl.Send("HSET", key, seg, sid_num)
		}

		min, err := redis.Int64(rds.Do("HGET", key, seg))
		if nil != err {
			ctx.log.Error("Get min num failed! errmsg:%s", err.Error())
			continue
		} else if min > sid_num {
			pl.Send("HSET", key, seg, sid_num)
		}
	}
}

/******************************************************************************
 **函数名称: clean_prec_statis
 **功    能: 删除各精度用户数
 **输入参数:
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 17:49:10 #
 ******************************************************************************/
func (ctx *TaskerCntx) clean_prec_statis() {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ctm := time.Now().Unix()

	/* > 遍历统计精度列表 */
	prec_rnum_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_PREC_NUM_ZSET, 0, "+inf", "WITHSCORES"))
	if nil != err {
		ctx.log.Error("Get prec list failed! errmsg:%s", err.Error())
		return
	}

	prec_num := len(prec_rnum_list)
	for idx := 0; idx < prec_num; idx += 2 {
		prec, _ := strconv.ParseInt(prec_rnum_list[idx], 10, 64)
		rnum, _ := strconv.ParseInt(prec_rnum_list[idx+1], 10, 64)
		if 0 == prec || 0 == rnum {
			continue
		}

		seg := (ctm / prec) * prec

		/* > 清理最大峰值 */
		key := fmt.Sprintf(comm.IM_KEY_PREC_USR_MAX_NUM, prec)
		time_list, err := redis.Strings(rds.Do("HKEYS", key))
		if nil == err {
			time_num := len(time_list)
			for idx := 0; idx < time_num; idx += 1 {
				tm, _ := strconv.ParseInt(time_list[idx], 10, 64)
				intval_num := (seg - tm) / prec
				if intval_num > rnum {
					pl.Send("HDEL", key, tm)
				}
			}
		}

		/* > 清理最低峰值 */
		key = fmt.Sprintf(comm.IM_KEY_PREC_USR_MIN_NUM, prec)
		time_list, err = redis.Strings(rds.Do("HKEYS", key))
		if nil == err {
			time_num := len(time_list)
			for idx := 0; idx < time_num; idx += 1 {
				tm, _ := strconv.ParseInt(time_list[idx], 10, 64)
				intval_num := (seg - tm) / prec
				if intval_num > rnum {
					pl.Send("HDEL", key, tm)
				}
			}
		}
	}
}

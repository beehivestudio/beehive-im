package chat

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: RoomGetRidToNidSet
 **功    能: 通过聊天室RID获取对应的帧听层NID列表
 **输入参数:
 **     pool: REDIS连接池
 **     gid: 群ID
 **输出参数: NONE
 **返    回: 帧听层ID列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.07 23:03:03 #
 ******************************************************************************/
func RoomGetRidToNidSet(pool *redis.Pool, rid uint64) (list []uint32, err error) {
	rds := pool.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, rid)
	nid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
	if nil != err {
		return nil, err
	}

	num := len(nid_list)
	list = make([]uint32, 0)
	for idx := 0; idx < num; idx += 1 {
		nid, _ := strconv.ParseInt(nid_list[idx], 10, 32)
		list = append(list, uint32(nid))
	}

	return list, nil
}

/******************************************************************************
 **函数名称: RoomGetRidToNidMap
 **功    能: 通过聊天室RID获取对应的帧听层NID映射表
 **输入参数:
 **     pool: REDIS连接池
 **输出参数: NONE
 **返    回: 帧听层ID列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.07 23:03:03 #
 ******************************************************************************/
func RoomGetRidToNidMap(pool *redis.Pool) (m map[uint64][]uint32, err error) {
	rds := pool.Get()
	defer rds.Close()

	ctm := time.Now().Unix()
	m = make(map[uint64][]uint32)

	off := 0
	key := fmt.Sprintf(comm.CHAT_KEY_RID_ZSET)
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			key, ctm, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			return nil, err
		}

		num := len(rid_list)
		for idx := 0; idx < num; idx += 1 {
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)

			m[uint64(rid)], err = RoomGetRidToNidSet(pool, uint64(rid))
			if nil != err {
				return nil, err
			}
		}

		if num < comm.CHAT_BAT_NUM {
			break
		}
		off += num
	}

	return m, nil
}

/******************************************************************************
 **函数名称: RoomCleanBySid
 **功    能: 清理指定会话的聊天室数据
 **输入参数:
 **     pool: REDIS连接池
 **     uid: 用户UID
 **     nid: 侦听层ID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.10 23:00:26 #
 ******************************************************************************/
func RoomCleanBySid(pool *redis.Pool, uid uint64, nid uint32, sid uint64) error {
	rds := pool.Get()
	defer rds.Close()

	pl := pool.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID -> RID列表 */
	key := fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)

	rid_gid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		return err
	}

	rid_num := len(rid_gid_list)
	for idx := 0; idx < rid_num; idx += 2 {
		rid, _ := strconv.ParseInt(rid_gid_list[idx], 10, 64)
		gid, _ := strconv.ParseInt(rid_gid_list[idx+1], 10, 64)

		/* 清理各种数据 */
		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, rid)
		pl.Send("ZREM", key, sid)

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, rid)
		member := fmt.Sprintf(comm.UID_SID_STR, uid, sid)
		pl.Send("ZREM", key, member)

		/* 更新统计计数 */
		key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, rid)
		pl.Send("ZINCRBY", key, -1, gid)

		key = fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, rid)
		pl.Send("ZINCRBY", key, -1, nid)
	}

	/* 清理各种数据 */
	key = fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)
	pl.Send("DEL", key)

	return nil
}

/******************************************************************************
 **函数名称: RoomUpdateBySid
 **功    能: 更新指定会话的聊天室数据
 **输入参数:
 **     pool: REDIS连接池
 **     uid: 用户UID
 **     nid: 侦听层ID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.11 08:50:23 #
 ******************************************************************************/
func RoomUpdateBySid(pool *redis.Pool, uid uint64, nid uint32, sid uint64) error {
	rds := pool.Get()
	defer rds.Close()

	pl := pool.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID -> RID列表 */
	key := fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, sid)

	rid_gid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		return err
	}

	ttl := time.Now().Unix() + comm.CHAT_SID_TTL

	rid_num := len(rid_gid_list)
	for idx := 0; idx < rid_num; idx += 2 {
		rid, _ := strconv.ParseInt(rid_gid_list[idx], 10, 64)
		//gid, _ := strconv.ParseInt(rid_gid_list[idx+1], 10, 64)

		/* 清理各种数据 */
		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, rid)
		pl.Send("ZADD", key, ttl, sid)

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, rid)
		member := fmt.Sprintf(comm.UID_SID_STR, uid, sid)
		pl.Send("ZADD", key, ttl, member)
	}

	return nil
}

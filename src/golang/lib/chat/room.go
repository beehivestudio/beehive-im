package chat

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/comm"
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
	list = make([]uint32)
	for idx := 0; idx < num; idx += 1 {
		nid, _ := strconv.ParseInt(nid_list[idx], 10, 32)
		list.append(nid)
	}

	return list, nil
}

/******************************************************************************
 **函数名称: RoomGetRidToNidMap
 **功    能: 通过聊天室RID获取对应的帧听层NID映射表
 **输入参数:
 **     pool: REDIS连接池
 **     gid: 群ID
 **输出参数: NONE
 **返    回: 帧听层ID列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.07 23:03:03 #
 ******************************************************************************/
func RoomGetRidToNidMap(pool *redis.Pool, rid uint64) (m map[uint64][]uint32, err error) {
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

		num := len(nid_list)
		for idx := 0; idx < num; idx += 1 {
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)

			m[rid], err = RoomGetRidToNidSet(pool, rid)
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

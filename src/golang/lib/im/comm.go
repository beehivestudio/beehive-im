package im

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: AllocSid
 **功    能: 申请会话SID
 **输入参数:
 **     pool: Redis连接池
 **输出参数: NONE
 **返    回:
 **     sid: 会话SID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 10:48:40 #
 ******************************************************************************/
func AllocSid(pool *redis.Pool) (sid uint64, err error) {
	rds := pool.Get()
	defer rds.Close()

	for {
		sid, err := redis.Uint64(rds.Do("INCRBY", comm.IM_KEY_SID_INCR, 1))
		if nil != err {
			return 0, err
		} else if 0 == sid {
			continue
		}
		return sid, nil
	}

	return 0, errors.New("Alloc sid failed!")
}

/* 会话属性 */
type SidAttr struct {
	Sid uint64 // 会话SID
	Uid uint64 // 用户ID
	Nid uint32 // 侦听层ID
}

/******************************************************************************
 **函数名称: GetSidAttr
 **功    能: 获取会话属性
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 会话属性
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.09 08:35:54 #
 ******************************************************************************/
func GetSidAttr(pool *redis.Pool, sid uint64) *SidAttr {
	var attr SidAttr

	rds := pool.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)
	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		return nil
	}

	attr.Sid = sid
	uid_int, _ := strconv.ParseInt(vals[0], 10, 64)
	attr.Uid = uint64(uid_int)
	nid_int, _ := strconv.ParseInt(vals[1], 10, 64)
	attr.Nid = uint32(nid_int)

	return &attr
}

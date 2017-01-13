package im

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/chat"
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
func GetSidAttr(pool *redis.Pool, sid uint64) (attr *SidAttr, err error) {
	rds := pool.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)

	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		return nil, err
	}

	attr = &SidAttr{}

	attr.Sid = sid
	uid_int, _ := strconv.ParseInt(vals[0], 10, 64)
	attr.Uid = uint64(uid_int)
	nid_int, _ := strconv.ParseInt(vals[1], 10, 64)
	attr.Nid = uint32(nid_int)

	return attr, nil
}

/******************************************************************************
 **函数名称: CleanSidData
 **功    能: 清理会话数据
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.09 08:35:54 #
 ******************************************************************************/
func CleanSidData(pool *redis.Pool, sid uint64) error {
	rds := pool.Get()
	defer rds.Close()

	pl := pool.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID对应的数据 */
	attr, err := GetSidAttr(pool, sid)
	if nil != err {
		return err
	}

	/* > 删除SID对应的数据 */
	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)

	num, err := redis.Int(rds.Do("DEL", key))
	if nil != err {
		return err
	} else if 0 == num {
		pl.Send("ZREM", comm.IM_KEY_SID_ZSET, sid)
		return nil
	}

	/* > 清理相关资源 */
	chat.RoomCleanBySid(pool, attr.Uid, attr.Nid, sid)

	pl.Send("ZREM", comm.IM_KEY_SID_ZSET, sid)

	return nil
}

/******************************************************************************
 **函数名称: UpdateSidData
 **功    能: 更新会话数据
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述: 更新会话属性以及对应的聊天室信息.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.11 23:34:31 #
 ******************************************************************************/
func UpdateSidData(pool *redis.Pool, nid uint32, sid uint64) (code uint32, err error) {
	pl := pool.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* 获取会话属性 */
	attr, err := GetSidAttr(pool, sid)
	if nil != err {
		return comm.ERR_SYS_SYSTEM, err
	} else if 0 == attr.Uid {
		return comm.ERR_SVR_DATA_COLLISION, errors.New("Get sid attribute failed!")
	} else if nid != attr.Nid {
		return comm.ERR_SVR_DATA_COLLISION, errors.New("Node of session is collision!")
	}

	/* 更新会话属性 */
	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", comm.IM_KEY_SID_ZSET, ttl, sid)
	pl.Send("ZADD", comm.IM_KEY_UID_ZSET, ttl, attr.Uid)

	/* > 更新聊天室信息 */
	chat.RoomUpdateBySid(pool, attr.Uid, attr.Nid, sid)

	return 0, nil
}

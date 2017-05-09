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

/* 会话属性 */
type SidAttr struct {
	sid uint64 // 会话SID
	cid uint64 // 连接CID
	uid uint64 // 用户ID
	nid uint32 // 侦听层ID
}

func (attr *SidAttr) GetSid() uint64 { return attr.sid }
func (attr *SidAttr) GetCid() uint64 { return attr.cid }
func (attr *SidAttr) GetUid() uint64 { return attr.uid }
func (attr *SidAttr) GetNid() uint32 { return attr.nid }

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

	/* 获取会话属性 */
	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)

	vals, err := redis.Strings(rds.Do("HMGET", key, "CID", "UID", "NID"))
	if nil != err {
		return nil, err
	}

	cid, _ := strconv.ParseInt(vals[0], 10, 64)
	uid, _ := strconv.ParseInt(vals[1], 10, 64)
	nid, _ := strconv.ParseInt(vals[2], 10, 64)

	attr = &SidAttr{
		sid: sid,
		cid: uint64(cid),
		uid: uint64(uid),
		nid: uint32(nid),
	}
	return attr, nil
}

/******************************************************************************
 **函数名称: CleanSessionData
 **功    能: 清理会话数据
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **     cid: 连接CID
 **     nid: 节点ID
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项: 当会话属性中的cid和nid与参数不一致时, 不进行清理操作.
 **作    者: # Qifeng.zou # 2017.05.10 06:32:05 #
 ******************************************************************************/
func CleanSessionData(pool *redis.Pool, sid uint64, cid uint64, nid uint32) error {
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
	} else if attr.GetCid() != cid || attr.GetNid() != nid {
		return nil /* 数据不一致, 不进行清理操作 */
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
	chat.RoomCleanBySid(pool, attr.uid, attr.nid, sid)

	pl.Send("ZREM", comm.IM_KEY_SID_ZSET, sid)

	return nil
}

/******************************************************************************
 **函数名称: CleanSessionDataBySid
 **功    能: 通过SID清理会话数据
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.09 08:35:54 #
 ******************************************************************************/
func CleanSessionDataBySid(pool *redis.Pool, sid uint64) error {
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
	chat.RoomCleanBySid(pool, attr.uid, attr.nid, sid)

	pl.Send("ZREM", comm.IM_KEY_SID_ZSET, sid)

	return nil
}

/******************************************************************************
 **函数名称: UpdateSidData
 **功    能: 更新会话数据
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **     cid: 连接CID
 **     nid: 节点ID
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述: 更新会话属性以及对应的聊天室信息.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.11 23:34:31 #
 ******************************************************************************/
func UpdateSessionData(pool *redis.Pool, sid uint64, cid uint64, nid uint32) (code uint32, err error) {
	pl := pool.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* 获取会话属性 */
	attr, err := GetSidAttr(pool, sid)
	if nil != err {
		return comm.ERR_SYS_SYSTEM, err
	} else if 0 == attr.uid {
		return comm.ERR_SVR_DATA_COLLISION, errors.New("Get sid attribute failed!")
	} else if nid != attr.nid {
		return comm.ERR_SVR_DATA_COLLISION, errors.New("Nid is collision!")
	} else if cid != attr.cid {
		return comm.ERR_SVR_DATA_COLLISION, errors.New("Cid is collision!")
	}

	/* 更新会话属性 */
	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", comm.IM_KEY_SID_ZSET, ttl, sid)
	pl.Send("ZADD", comm.IM_KEY_UID_ZSET, ttl, attr.uid)

	/* > 更新聊天室信息 */
	chat.RoomUpdateBySid(pool, attr.uid, attr.nid, sid)

	return 0, nil
}

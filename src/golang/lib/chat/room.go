package chat

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/rtmq"
)

/* 聊天室角色 */
const (
	ROOM_ROLE_OWNER   = 1 // 聊天室-所有者
	ROOM_ROLE_MANAGER = 2 // 聊天室-管理员
)

/* 聊天室状态 */
const (
	ROOM_STAT_OPEN  = 1 // 聊天室-开启
	ROOM_STAT_CLOSE = 0 // 聊天室-关闭
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
 **返    回: rid->nid映射表
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
		member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, uid, sid)
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
		member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, uid, sid)
		pl.Send("ZADD", key, ttl, member)
	}

	return nil
}

/******************************************************************************
 **函数名称: IsRoomOwner
 **功    能: 用户是否是聊天室的所有者
 **输入参数:
 **     pool: REDIS连接池
 **     rid: 聊天室ID
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: true:是 false:不是
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 07:35:22 #
 ******************************************************************************/
func IsRoomOwner(pool *redis.Pool, rid uint64, uid uint64) bool {
	rds := pool.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_ROLE_TAB, rid)

	role, err := redis.Int(rds.Do("HGET", key, uid))
	if nil != err {
		return false
	} else if ROOM_ROLE_OWNER == role {
		return true
	}

	return true
}

/******************************************************************************
 **函数名称: IsRoomManager
 **功    能: 用户是否是聊天室的管理员
 **输入参数:
 **     pool: REDIS连接池
 **     rid: 聊天室ID
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: true:是 false:不是
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 08:15:03 #
 ******************************************************************************/
func IsRoomManager(pool *redis.Pool, rid uint64, uid uint64) bool {
	rds := pool.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_ROLE_TAB, rid)

	role, err := redis.Int(rds.Do("HGET", key, uid))
	if nil != err {
		return false
	} else if ROOM_ROLE_OWNER == role || ROOM_ROLE_MANAGER == role {
		return true
	}

	return true
}

/******************************************************************************
 **函数名称: RoomSendUsrNum
 **功    能: 下发聊天室人数
 **输入参数:
 **     pool: REDIS连接池
 **输出参数: NONE
 **返    回: 错误码+错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 11:02:22 #
 ******************************************************************************/
func RoomSendUsrNum(frwder *rtmq.Proxy, pool *redis.Pool) error {
	rds := pool.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* 获取侦听层ID列表 */
	lsn_nid_list, err := redis.Ints(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_LSND_NID_ZSET, ctm, "+inf"))
	if nil != err {
		return err
	}

	lsn_num := len(lsn_nid_list)

	off := 0
	for {
		/* 获取聊天室列表 */
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, ctm, "+inf", off, comm.CHAT_BAT_NUM))
		if nil != err {
			return err
		}

		rid_num := len(rid_list)

		/* 遍历下发聊天室人数 */
		for m := 0; m < rid_num; m += 1 {
			rid, _ := strconv.ParseInt(rid_list[m], 10, 64)

			/* 获取聊天室人数 */
			key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, uint64(rid))
			usr_num, err := redis.Int(rds.Do("ZCARD", key))
			if nil != err {
				continue
			}

			/* 下发聊天室人数 */
			for n := 0; n < lsn_num; n += 1 {
				send_room_usr_num(frwder, pool,
					uint64(rid), uint32(lsn_nid_list[n]), uint32(usr_num))
			}
		}

		if off < comm.CHAT_BAT_NUM {
			break
		}

		off += rid_num
	}

	return nil
}

/******************************************************************************
 **函数名称: send_room_usr_num
 **功    能: 发送ROOM-USR-NUM信息
 **输入参数:
 **     head: 协议头
 **     req: 请求数据
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **协议格式:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 11:15:03 #
 ******************************************************************************/
func send_room_usr_num(frwder *rtmq.Proxy,
	pool *redis.Pool, rid uint64, nid uint32, num uint32) int {
	rds := pool.Get()
	defer rds.Close()

	/* > 设置协议体 */
	rsp := &mesg.MesgRoomUsrNum{
		Rid: proto.Uint64(rid),
		Num: proto.Uint32(num),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head := &comm.MesgHeader{}

	head.Cmd = comm.CMD_ROOM_USR_NUM
	head.Sid = rid // 会话ID改为聊天室ID
	head.Nid = uint32(nid)
	head.Length = uint32(length)
	head.ChkSum = comm.MSG_CHKSUM_VAL

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	frwder.AsyncSend(comm.CMD_ROOM_USR_NUM, p.Buff, uint32(len(p.Buff)))

	return 0
}

package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/rdb"
	"beehive-im/src/golang/lib/rtmq"
)

type RoomCacheObj struct {
	redis *redis.Pool
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化
 **输入参数:
 **     addr: 地址
 **     pwd: 密码
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.09.16 09:55:43 #
 ******************************************************************************/
func (c *RoomCacheObj) Init(addr string, pwd string) error {
	/* > REDIS连接池 */
	c.redis = rdb.CreatePool(addr, pwd, 2048)
	if nil == c.redis {
		return errors.New("Create redis pool failed!")
	}
	return nil
}

/******************************************************************************
 **函数名称: Get
 **功    能: 获取连接对象
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 连接对象(Redis)
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.09.16 09:56:28 #
 ******************************************************************************/
func (c *RoomCacheObj) Get() redis.Conn {
	return c.redis.Get()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: getRidToNidSetFromRds
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
func (c *RoomCacheObj) getRidToNidSetFromRds(rid uint64) (list []uint32, err error) {
	rds := c.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	key := fmt.Sprintf(ROOM_KEY_RID_TO_NID_ZSET, rid)
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
 **函数名称: GetRidToNidMapFromRds
 **功    能: 通过聊天室RID获取对应的帧听层NID映射表
 **输入参数:
 **     pool: REDIS连接池
 **输出参数: NONE
 **返    回: rid->nid映射表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.07 23:03:03 #
 ******************************************************************************/
func (c *RoomCacheObj) GetRidToNidMapFromRds() (m map[uint64][]uint32, err error) {
	rds := c.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()
	m = make(map[uint64][]uint32)

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			ROOM_KEY_RID_ZSET, ctm, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			return nil, err
		}

		num := len(rid_list)
		for idx := 0; idx < num; idx += 1 {
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)

			m[uint64(rid)], err = c.getRidToNidSetFromRds(uint64(rid))
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
 **注意事项: 不进行各组&各侦听层的人数统计, 而以侦听层的上报数据为准进行统计.
 **作    者: # Qifeng.zou # 2017.01.10 23:00:26 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomCleanBySid(uid uint64, nid uint32, sid uint64) error {
	rds := c.redis.Get()
	defer rds.Close()

	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID -> RID列表 */
	key := fmt.Sprintf(ROOM_KEY_SID_TO_RID_ZSET, sid)

	rid_gid_list, err := redis.Strings(rds.Do(
		"ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES"))
	if nil != err {
		return err
	}

	rid_num := len(rid_gid_list)
	for idx := 0; idx < rid_num; idx += 2 {
		rid, _ := strconv.ParseInt(rid_gid_list[idx], 10, 64)
		gid, _ := strconv.ParseInt(rid_gid_list[idx+1], 10, 64)

		/* 清理各种数据 */
		key = fmt.Sprintf(ROOM_KEY_RID_TO_SID_ZSET, rid)
		pl.Send("ZREM", key, sid)

		key = fmt.Sprintf(ROOM_KEY_RID_TO_UID_SID_ZSET, rid)
		member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, uid, sid)
		pl.Send("ZREM", key, member)

		/* 更新统计计数 */
		key = fmt.Sprintf(ROOM_KEY_RID_GID_TO_NUM_ZSET, rid)
		pl.Send("ZINCRBY", key, -1, gid)
	}

	/* 清理各种数据 */
	key = fmt.Sprintf(ROOM_KEY_SID_TO_RID_ZSET, sid)
	pl.Send("DEL", key)

	return nil
}

/******************************************************************************
 **函数名称: RoomUpdateBySid
 **功    能: 更新指定会话的聊天室数据
 **输入参数:
 **     uid: 用户UID
 **     nid: 侦听层ID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.11 08:50:23 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomUpdateBySid(uid uint64, nid uint32, sid uint64) error {
	rds := c.redis.Get()
	defer rds.Close()

	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID -> RID列表 */
	key := fmt.Sprintf(ROOM_KEY_SID_TO_RID_ZSET, sid)

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
		key = fmt.Sprintf(ROOM_KEY_RID_TO_SID_ZSET, rid)
		pl.Send("ZADD", key, ttl, sid)

		key = fmt.Sprintf(ROOM_KEY_RID_TO_UID_SID_ZSET, rid)
		member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, uid, sid)
		pl.Send("ZADD", key, ttl, member)
	}

	return nil
}

/******************************************************************************
 **函数名称: IsRoomOwner
 **功    能: 用户是否是聊天室的所有者
 **输入参数:
 **     rid: 聊天室ID
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: true:是 false:不是
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 07:35:22 #
 ******************************************************************************/
func (c *RoomCacheObj) IsRoomOwner(rid uint64, uid uint64) bool {
	rds := c.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(ROOM_KEY_ROOM_ROLE_TAB, rid)

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
 **     rid: 聊天室ID
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: true:是 false:不是
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 08:15:03 #
 ******************************************************************************/
func (c *RoomCacheObj) IsRoomManager(rid uint64, uid uint64) bool {
	rds := c.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(ROOM_KEY_ROOM_ROLE_TAB, rid)

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
 **     frwder: 转发层代理
 **输出参数: NONE
 **返    回: 错误码+错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.13 11:02:22 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomSendUsrNum(
	log *logs.BeeLogger, frwder *rtmq.Proxy) error {
	rds := c.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* 获取侦听层ID列表 */
	lsn_nid_list, err := redis.Ints(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_LSND_NID_ZSET, ctm, "+inf"))
	if nil != err {
		log.Error("Get listend list failed! errmsg:%s", err.Error())
		return err
	}

	lsn_num := len(lsn_nid_list)

	off := 0
	for {
		/* 获取聊天室列表 */
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			ROOM_KEY_RID_ZSET, ctm, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			log.Error("Get room list failed! errmsg:%s", err.Error())
			return err
		}

		rid_num := len(rid_list)

		/* 遍历下发聊天室人数 */
		for m := 0; m < rid_num; m += 1 {
			rid, _ := strconv.ParseInt(rid_list[m], 10, 64)

			/* 获取聊天室人数 */
			key := fmt.Sprintf(ROOM_KEY_RID_TO_UID_SID_ZSET, uint64(rid))
			usr_num, err := redis.Int(rds.Do("ZCARD", key))
			if nil != err {
				log.Error("Get user num of room failed! errmsg:%s", err.Error())
				continue
			}

			/* 下发聊天室人数 */
			for n := 0; n < lsn_num; n += 1 {
				c.room_send_usr_num(frwder,
					uint64(rid), uint32(lsn_nid_list[n]), uint32(usr_num))
			}
		}

		if rid_num < comm.CHAT_BAT_NUM {
			break
		}

		off += rid_num
	}

	return nil
}

/******************************************************************************
 **函数名称: room_send_usr_num
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
func (c *RoomCacheObj) room_send_usr_num(
	frwder *rtmq.Proxy, rid uint64, nid uint32, num uint32) int {
	rds := c.redis.Get()
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

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	frwder.AsyncSend(comm.CMD_ROOM_USR_NUM, p.Buff, uint32(len(p.Buff)))

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 会话属性 */
type RoomSidAttr struct {
	sid uint64 // 会话SID
	cid uint64 // 连接CID
	uid uint64 // 用户ID
	nid uint32 // 侦听层ID
}

func (attr *RoomSidAttr) GetSid() uint64 { return attr.sid }
func (attr *RoomSidAttr) GetCid() uint64 { return attr.cid }
func (attr *RoomSidAttr) GetUid() uint64 { return attr.uid }
func (attr *RoomSidAttr) GetNid() uint32 { return attr.nid }

/******************************************************************************
 **函数名称: RoomGetSidAttr
 **功    能: 获取会话属性
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 会话属性
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.09 08:35:54 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomGetSidAttr(sid uint64) (attr *RoomSidAttr, err error) {
	rds := c.redis.Get()
	defer rds.Close()

	/* 获取会话属性 */
	key := fmt.Sprintf(ROOM_KEY_SID_ATTR, sid)

	vals, err := redis.Strings(rds.Do("HMGET", key, "CID", "UID", "NID"))
	if nil != err {
		return nil, err
	}

	cid, _ := strconv.ParseInt(vals[0], 10, 64)
	uid, _ := strconv.ParseInt(vals[1], 10, 64)
	nid, _ := strconv.ParseInt(vals[2], 10, 64)

	attr = &RoomSidAttr{
		sid: sid,
		cid: uint64(cid),
		uid: uint64(uid),
		nid: uint32(nid),
	}
	return attr, nil
}

/******************************************************************************
 **函数名称: RoomCleanSessionData
 **功    能: 清理聊天室会话数据
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
func (c *RoomCacheObj) RoomCleanSessionData(sid uint64, cid uint64, nid uint32) error {
	rds := c.redis.Get()
	defer rds.Close()

	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID对应的数据 */
	attr, err := c.RoomGetSidAttr(sid)
	if nil != err {
		return err
	} else if attr.GetCid() != cid || attr.GetNid() != nid {
		return errors.New("Data is collision!") /* 数据不一致, 不进行清理操作 */
	}

	/* > 删除SID对应的数据 */
	key := fmt.Sprintf(ROOM_KEY_SID_ATTR, sid)

	num, err := redis.Int(rds.Do("DEL", key))
	if nil != err {
		return err
	} else if 0 == num {
		pl.Send("ZREM", ROOM_KEY_SID_ZSET, sid)
		return nil
	}

	/* > 清理相关资源 */
	c.RoomCleanBySid(attr.uid, attr.nid, sid)

	pl.Send("ZREM", ROOM_KEY_SID_ZSET, sid)

	return nil
}

/******************************************************************************
 **函数名称: RoomCleanSessionDataBySid
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
func (c *RoomCacheObj) RoomCleanSessionDataBySid(sid uint64) error {
	rds := c.redis.Get()
	defer rds.Close()

	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取SID对应的数据 */
	attr, err := c.RoomGetSidAttr(sid)
	if nil != err {
		return err
	}

	/* > 删除SID对应的数据 */
	key := fmt.Sprintf(ROOM_KEY_SID_ATTR, sid)

	num, err := redis.Int(rds.Do("DEL", key))
	if nil != err {
		return err
	} else if 0 == num {
		pl.Send("ZREM", ROOM_KEY_SID_ZSET, sid)
		return nil
	}

	/* > 清理相关资源 */
	c.RoomCleanBySid(attr.uid, attr.nid, sid)

	pl.Send("ZREM", ROOM_KEY_SID_ZSET, sid)

	return nil
}

/******************************************************************************
 **函数名称: RoomUpdateSessionData
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
func (c *RoomCacheObj) RoomUpdateSessionData(
	sid uint64, cid uint64, nid uint32) (code uint32, err error) {
	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* 获取会话属性 */
	attr, err := c.RoomGetSidAttr(sid)
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
	pl.Send("ZADD", ROOM_KEY_SID_ZSET, ttl, sid)
	pl.Send("ZADD", ROOM_KEY_UID_ZSET, ttl, attr.uid)

	/* > 更新聊天室信息 */
	c.RoomUpdateBySid(attr.uid, attr.nid, sid)

	return 0, nil
}

/******************************************************************************
 **函数名称: RoomListBySid
 **功    能: 根据会话sid获取其当前加入的聊天室列表
 **输入参数:
 **     pool: REDIS连接池
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.08.11 20:28:04 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomListBySid(sid uint64) ([]string, error) {
	rds := c.redis.Get()
	defer rds.Close()

	/* > 获取SID -> RID列表 */
	key := fmt.Sprintf(ROOM_KEY_SID_TO_RID_ZSET, sid)

	return redis.Strings(rds.Do("ZRANGEBYSCORE", key, "-inf", "+inf"))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: RoomAdd
 **功    能: 添加聊天室
 **输入参数:
 **     pool: REDIS连接池
 **     rid: 聊天室ID
 **     uid: 用户UID
 **     name: 聊天室名称
 **     desc: 聊天室描述
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.09.16 00:41:59 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomAdd(rid uint64, req *mesg.MesgRoomCreat) error {
	pl := c.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	rds := c.redis.Get()
	defer rds.Close()

	/* > 设置聊天室所有者 */
	key := fmt.Sprintf(ROOM_KEY_ROOM_ROLE_TAB, rid)

	ok, err := redis.Bool(rds.Do("HSETNX", key, req.GetUid(), ROOM_ROLE_OWNER))
	if nil != err {
		return err
	} else if !ok {
		return errors.New("Set room owner failed!")
	}

	/* > 设置聊天室信息 */
	key = fmt.Sprintf(ROOM_KEY_ROOM_INFO_TAB, rid)

	pl.Send("HMSET", key, "NAME", req.GetName(), "DESC", req.GetDesc())

	return nil
}

/******************************************************************************
 **函数名称: RoomCapacity
 **功    能: 添加聊天室
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.09.16 00:41:59 #
 ******************************************************************************/
func (c *RoomCacheObj) RoomCapacity(rid uint64) (capacity int, err error) {
	rds := c.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(ROOM_KEY_ROOM_GROUP_CAP_ZSET)

	capacity, err = redis.Int(rds.Do("ZSCORE", key, rid))
	if nil != err {
		return 0, err
	}

	return capacity, nil
}

/******************************************************************************
 **函数名称: UpdateOnlineData
 **功    能: 更新在线数据
 **输入参数:
 **     uid: 用户UID
 **     sid: 会话SID
 **     nid: 结点ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 错误码
 **实现描述: 更新缓存中的在线数据
 **注意事项:
 **作    者: # Qifeng.zou # 2017.10.12 19:07:46 #
 ******************************************************************************/
func (c *RoomCacheObj) UpdateOnlineData(uid uint64, sid uint64, nid uint32, cid uint64) int {
	var key string

	pl := c.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ttl := time.Now().Unix() + comm.CHAT_SID_TTL

	/* 记录SID集合 */
	pl.Send("ZADD", ROOM_KEY_SID_ZSET, ttl, sid)

	/* 记录UID集合 */
	pl.Send("ZADD", ROOM_KEY_UID_ZSET, ttl, uid)

	/* 记录SID->UID/NID */
	key = fmt.Sprintf(ROOM_KEY_SID_ATTR, sid)
	pl.Send("HMSET", key, "CID", cid, "UID", uid, "NID", nid)

	/* 记录UID->SID集合 */
	key = fmt.Sprintf(ROOM_KEY_UID_TO_SID_SET, uid)
	pl.Send("SADD", key, sid)

	return 0
}

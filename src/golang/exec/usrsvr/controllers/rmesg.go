package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/chat"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg"
)

// 聊天室

////////////////////////////////////////////////////////////////////////////////
/* 创建聊天室 */

/******************************************************************************
 **函数名称: room_creat_parse
 **功    能: 解析ROOM-CREAT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:32:20 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_creat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-creat failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomCreat{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-creat request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_HEAD_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_creat_failed
 **功    能: 发送ROOM-CREAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-CREAT请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:33:42 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_creat_failed(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, code uint32, err error) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomCreatAck{
		Rid:    proto.Uint64(0),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(err.Error()),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_creat_ack
 **功    能: 发送ROOM-CREAT应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:28:40 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_creat_ack(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, rid uint64) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomCreatAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(rid),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: alloc_rid
 **功    能: 申请聊天室ID
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     rid: 聊天室ID
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:24:23 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) alloc_rid() (rid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 申请聊天室ID */
	rid_str, err := redis.String(rds.Do("INCR", comm.CHAT_KEY_RID_INCR))
	if nil != err {
		return 0, err
	}

	rid_int, _ := strconv.ParseInt(rid_str, 10, 64)

	return uint64(rid_int), nil
}

/******************************************************************************
 **函数名称: room_creat_handler
 **功    能: ROOM-CREAT处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-CREAT请求
 **输出参数: NONE
 **返    回:
 **     rid: 聊天室ID
 **     err: 错误描述
 **实现描述:
 **注意事项: 已验证了ROOM-CREAT请求的合法性
 **作    者: # Qifeng.zou # 2017.01.19 22:22:50 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_creat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat) (rid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 分配聊天室ID */
	rid, err = ctx.alloc_rid()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 设置聊天室所有者 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_ROLE_TAB, rid)

	ok, err := redis.Int(rds.Do("HSETNX", key, req.GetUid(), chat.ROOM_ROLE_OWNER))
	if nil != err {
		ctx.log.Error("Set room owner failed! uid:%d errmsg:%s",
			req.GetUid(), err.Error())
		return 0, err
	} else if 0 == ok {
		ctx.log.Error("Set room owner failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 设置聊天室信息 */
	key = fmt.Sprintf(comm.CHAT_KEY_ROOM_INFO_TAB, rid)

	pl.Send("HMSET", key, "NAME", req.GetName(), "DESC", req.GetDesc())

	return rid, nil
}

/******************************************************************************
 **函数名称: UsrSvrRoomCreatHandler
 **功    能: 创建聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required string name = 2;   // M|聊天室名称|字串|
 **        required string desc = 3;   // M|聊天室描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:21:48 #
 ******************************************************************************/
func UsrSvrRoomCreatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv join request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析创建请求 */
	head, req, code, err := ctx.room_creat_parse(data)
	if nil == req {
		ctx.log.Error("Parse room-creat request failed!")
		ctx.room_creat_failed(head, req, code, err)
		return -1
	}

	/* > 创建聊天室处理 */
	rid, err := ctx.room_creat_handler(head, req)
	if nil != err {
		ctx.log.Error("Room creat handler failed!")
		ctx.room_creat_failed(head, req, comm.ERR_SYS_SYSTEM, err)
		return -1
	}

	/* > 发送应答 */
	ctx.room_creat_ack(head, req, rid)

	return 0
}

/* 解散聊天室 */
func UsrSvrRoomDismissHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 加入聊天室 */

/******************************************************************************
 **函数名称: room_join_parse
 **功    能: 解析ROOM-JOIN请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 16:41:17 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_join_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomJoin, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-join failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomJoin{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal join request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_HEAD_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_join_failed
 **功    能: 发送ROOM-JOIN应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-JOIN请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 17:12:36 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_join_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomJoin, code uint32, err error) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomJoinAck{
		Gid:    proto.Uint32(0),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(err.Error()),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_join_ack
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_join_ack(head *comm.MesgHeader, req *mesg.MesgRoomJoin, gid uint32) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomJoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(gid),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: alloc_room_gid
 **功    能: 分配组ID
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 20:08:06 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) alloc_room_gid(rid uint64) (gid uint32, err error) {
	var num int

	rds := ctx.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, rid)

	/* > 优先加入到gid为0的分组 */
	num, err = redis.Int(rds.Do("ZSCORE", key, "0"))
	if uint32(num) < comm.CHAT_ROOM_GROUP_MAX_NUM {
		return 0, nil
	}

	/* > 获取有序GID列表: 以人数从多到少进行排序(加入人数最少的分组) */
	min := 0
	max := comm.CHAT_ROOM_GROUP_MAX_NUM - 1

	gid_lst, err := redis.Ints(rds.Do("ZRANGEBYSCORE", key, min, max, "LIMIT", 0, 1))
	if nil != err {
		ctx.log.Error("Get group list failed! errmsg:%s", err)
		return 0, err
	} else if len(gid_lst) > 0 {
		return uint32(gid_lst[0]), nil
	}

	grp_num, err := redis.Int(rds.Do("ZCARD", key))
	if nil != err {
		ctx.log.Error("Get group num failed! errmsg:%s", err)
		return 0, err
	}

	return uint32(grp_num), nil
}

/******************************************************************************
 **函数名称: room_join_handler
 **功    能: ROOM-JOIN处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-JOIN请求
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项: 已验证了ROOM-JOIN请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 19:51:46 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_join_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomJoin) (gid uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 判断UID是否在黑名单中 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, req.GetRid())
	ok, err := redis.Bool(rds.Do("SISMEMBER", key, req.GetUid()))
	if nil != err {
		ctx.log.Error("Exec command [SISMEMBER] failed! rid:%d uid:%d err:",
			req.GetRid(), req.GetUid(), err.Error())
		return 0, err
	} else if true == ok {
		ctx.log.Error("User is in blacklist! rid:%d uid:%d", req.GetRid(), req.GetUid())
		return 0, errors.New("User is in blacklist!")
	}

GET_GID:
	/* > 判断UID是否登录过 */
	key = fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, req.GetUid())
	ok, err = redis.Bool(rds.Do("HEXISTS", key, req.GetRid()))
	if nil != err {
		ctx.log.Error("Get rid [%d] by uid failed! errmsg:%s", req.GetRid(), err.Error())
		return 0, err
	} else if true == ok { // 已登录过则放在同一个分组中...
		ctx.log.Debug("Uid has been login! rid:%d uid:%d", req.GetRid(), req.GetUid())
		gid_int, err := redis.Int(rds.Do("HGET", key, req.GetRid()))
		if nil != err {
			ctx.log.Error("Get rid [%d] by uid failed! errmsg:%s", req.GetRid(), err.Error())
			return 0, err
		}

		ctx.log.Debug("Uid has been login! rid:%d uid:%d gid:%d", req.GetRid(), req.GetUid(), gid_int)

		/* > 判断SID依然在聊天室列表 */
		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())

		exist, _ := redis.Uint64(rds.Do("ZSCORE", key, head.GetSid()))

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, req.GetRid())
		member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, req.GetUid(), head.GetSid())
		ttl := time.Now().Unix() + comm.CHAT_SID_TTL
		pl.Send("ZADD", key, ttl, member) // 加入RID -> UID集合"${uid}:${sid}"

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
		pl.Send("ZADD", key, ttl, head.GetSid()) // 加入RID -> SID集合

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, req.GetRid())
		pl.Send("ZADD", key, ttl, head.GetNid()) // 加入RID -> NID集合

		key = fmt.Sprintf(comm.CHAT_KEY_SID_TO_RID_ZSET, head.GetSid())
		pl.Send("ZADD", key, uint32(gid_int), req.GetRid()) // 加入SID -> RID集合

		/* > 更新数据库统计 */
		if 0 == exist {
			key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, req.GetRid())
			pl.Send("ZINCRBY", key, 1, uint32(gid_int))

			key = fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, req.GetRid())
			pl.Send("ZINCRBY", key, 1, head.GetNid())
		}

		return uint32(gid_int), nil
	}

	/* > 分配新的分组 */
	gid, err = ctx.alloc_room_gid(req.GetRid())
	if nil != err {
		ctx.log.Error("Alloc gid failed! rid:%d", req.GetRid())
		return 0, err
	}

	/* > 设置UID的RID组GID */
	key = fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, req.GetUid())
	ok, err = redis.Bool(rds.Do("HSETNX", key, req.GetRid(), gid)) /* 防止冲突 */
	if nil != err {
		ctx.log.Error("Get rid [%d] by uid failed!", req.GetRid())
		return 0, err
	} else if !ok {
		goto GET_GID /* 存在冲突 */
	}

	/* > 更新数据库统计 */
	key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZINCRBY", key, 1, gid)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZINCRBY", key, 1, head.GetNid())

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, req.GetRid())
	member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, req.GetUid(), head.GetSid())
	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", key, ttl, member) // 加入RID -> UID集合"${uid}:${sid}"

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetSid()) // 加入RID -> SID集合

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetNid()) // 加入RID -> NID集合

	return gid, nil
}

/******************************************************************************
 **函数名称: UsrSvrRoomJoinHandler
 **功    能: 加入聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func UsrSvrRoomJoinHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv join request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* 1. > 解析ROOM-JOIN请求 */
	head, req, code, err := ctx.room_join_parse(data)
	if nil == req {
		ctx.log.Error("Parse room-join request failed!")
		ctx.room_join_failed(head, req, code, err)
		return -1
	}

	/* 2. > 初始化上线环境 */
	gid, err := ctx.room_join_handler(head, req)
	if nil != err {
		ctx.log.Error("Room join handler failed!")
		ctx.room_join_failed(head, req, comm.ERR_SYS_SYSTEM, err)
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.room_join_ack(head, req, gid)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 退出聊天室 */

/******************************************************************************
 **函数名称: room_quit_isvalid
 **功    能: 判断ROOM-QUIT是否合法
 **输入参数:
 **     req: ROOM-QUIT请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:26:22 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_quit_isvalid(req *mesg.MesgRoomQuit) bool {
	if 0 == req.GetUid() || 0 == req.GetRid() {
		return false
	}
	return true
}

/******************************************************************************
 **函数名称: room_quit_failed
 **功    能: 发送ROOM-QUIT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-QUIT请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **        required uint32 gid = 3;    // M|分组ID|数字|
 **        optional uint32 code = 4; // M|错误码|数字|
 **        optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:20:34 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_quit_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomQuit, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomQuitAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_QUIT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_quit_parse
 **功    能: 解析ROOM-QUIT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:18:29 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_quit_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomQuit, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-quit failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomQuit{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room quit request failed! errmsg:%s", err.Error())
		return nil, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Parse body failed!")
	}

	/* > 校验协议合法性 */
	if !ctx.room_quit_isvalid(req) {
		ctx.log.Error("Room quit request is invalid!")
		return nil, nil, comm.ERR_SVR_CHECK_FAIL, errors.New("Check request failed!")
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_quit_ack
 **功    能: 发送ROOM-QUIT应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_quit_ack(head *comm.MesgHeader, req *mesg.MesgRoomQuit) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomQuitAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_QUIT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_quit_handler
 **功    能: 退出聊天室处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-QUIT请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项: 已验证了ROOM-QUIT请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 21:28:18 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_quit_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomQuit) (code uint32, err error) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZREM", key, head.GetSid()) // 清理RID -> SID集合

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_SID_ZSET, req.GetRid())
	member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, req.GetUid(), head.GetSid())
	pl.Send("ZREM", key, member) // 清理RID -> UID集合"${uid}:${sid}"

	return 0, nil
}

/******************************************************************************
 **函数名称: UsrSvrRoomQuitHandler
 **功    能: 退出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项: 需要对协议头进行字节序转换
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func UsrSvrRoomQuitHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit request!")

	/* 1. > 解析ROOM-QUIT请求 */
	head, req, code, err := ctx.room_quit_parse(data)
	if nil != err {
		ctx.log.Error("Parse room quit request failed!")
		ctx.room_quit_failed(head, req, code, err.Error())
		return -1
	}

	/* 2. > 退出聊天室处理 */
	code, err = ctx.room_quit_handler(head, req)
	if nil != err {
		ctx.log.Error("Hanle room quit request failed!")
		ctx.room_quit_failed(head, req, code, err.Error())
		return -1
	}

	/* 3. > 发送ROOM-QUIT应答 */
	ctx.room_quit_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 踢出聊天室 */

/******************************************************************************
 **函数名称: room_kick_parse
 **功    能: 解析KICK请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:21:37 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_kick_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomKick, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-kick is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomKick{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-kick request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_kick_failed
 **功    能: 发送KICK应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-KICK请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 17:12:36 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_kick_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomKick, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomKickAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_KICK_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_kick_ack
 **功    能: 发送ROOM-KICK应答
 **输入参数:
 **     head: 协议头
 **     req: 请求数据
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:32:20 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_kick_ack(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomKickAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_KICK_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_kick_notify
 **功    能: 发送ROOM-KICK-NTC通知
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
 **作    者: # Qifeng.zou # 2017.01.13 07:46:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_kick_notify(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 设置协议体 */
	ack := &mesg.MesgRoomKickNtc{
		Uid: proto.Uint64(req.GetUid()),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	nid_list, err := redis.Ints(rds.Do("ZRANGEBYSCORE", comm.IM_KEY_LSND_NID_ZSET, "-inf", "+inf"))
	if nil != err {
		ctx.log.Error("Get listen nid failed! errmsg:%s", err.Error())
		return -1
	}

	num := len(nid_list)

	for idx := 0; idx < num; idx += 1 {
		head.Cmd = comm.CMD_ROOM_KICK_NTC
		head.Length = uint32(length)
		head.Sid = req.GetRid() // 会话ID改为聊天室ID
		head.Nid = uint32(nid_list[idx])
		head.ChkSum = comm.MSG_CHKSUM_VAL

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_NTC, p.Buff, uint32(len(p.Buff)))
	}

	return 0
}

/******************************************************************************
 **函数名称: room_kick_handler
 **功    能: ROOM-KICK处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-KICK请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述:
 **注意事项: 已验证了ROOM-KICK请求的合法性
 **作    者: # Qifeng.zou # 2017.01.12 23:34:28 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) room_kick_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomKick) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取会话属性 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get sid attr failed! rid:%d uid:%d errmsg:%s",
			req.GetRid(), req.GetUid(), err.Error())
		return comm.ERR_SYS_SYSTEM, err
	} else if !chat.IsRoomManager(ctx.redis, req.GetRid(), attr.GetUid()) {
		ctx.log.Error("You're not owner! rid:%d uid:%d", req.GetRid(), req.GetUid())
		return comm.ERR_SYS_PERM_DENIED, errors.New("You're not room owner!")
	}

	/* > 用户加入黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, req.GetRid())

	pl.Send("ZADD", key, req.GetUid())

	return 0, nil
}

/******************************************************************************
 **函数名称: UsrSvrRoomKickHandler
 **功    能: 踢出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **        required uint32 code = 3;   // M|错误码|数字|
 **        required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:58:49 #
 ******************************************************************************/
func UsrSvrRoomKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-kick request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析ROOM-KICK请求 */
	head, req, code, err := ctx.room_kick_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-kick request failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 执行ROOM-KICK操作 */
	code, err = ctx.room_kick_handler(head, req)
	if nil != err {
		ctx.log.Error("Room kick handler failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送ROOM-KICK应答 */
	ctx.room_kick_ack(head, req)

	/* > 发送ROOM-KICK-NTC通知 */
	ctx.room_kick_notify(head, req)

	return 0
}

/* 获取聊天室人数 */
func UsrSvrRoomUsrNumHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

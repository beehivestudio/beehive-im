package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"labix.org/v2/mgo"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/chat"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/mesg/seqsvr"
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
func (ctx *ChatRoomCntx) room_creat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-creat failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
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
func (ctx *ChatRoomCntx) room_creat_failed(
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
func (ctx *ChatRoomCntx) room_creat_ack(
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

	ctx.log.Debug("Create room success! rid:%d uid:%d name:%s desc:%s",
		rid, req.GetUid(), req.GetName(), req.GetDesc())

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
 **作    者: # Qifeng.zou # 2017.05.29 08:48:19 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) alloc_rid() (rid uint64, err error) {
	/* > 获取连接对象 */
	conn, err := ctx.seqsvr_pool.Get()
	if nil != err {
		ctx.log.Error("Get seqsvr connection pool failed! errmsg:%s", err.Error())
		return 0, err
	}

	client := conn.(*seqsvr.SeqSvrThriftClient)
	defer ctx.seqsvr_pool.Put(client, false)

	/* > 申请聊天室ID */
	_rid, err := client.AllocRoomId()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	ctx.log.Debug("Alloc rid success! rid:%d", _rid)

	return uint64(_rid), nil
}

/******************************************************************************
 **函数名称: room_add
 **功    能: 添加聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     req: 创建请求
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.29 20:37:40 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_add(rid uint64, req *mesg.MesgRoomCreat) error {
	/* > 准备SQL语句 */
	sql := fmt.Sprintf("INSERT INTO CHAT_ROOM_INFO_%d(rid, name, status, description, create_time, update_time, owner) VALUES(?, ?, ?, ?, ?, ?, ?)", rid%256)

	stmt, err := ctx.userdb.Prepare(sql)
	if nil != err {
		ctx.log.Error("Prepare sql failed! sql:%s errmsg:%s", sql, err.Error())
		return err
	}

	defer stmt.Close()

	/* > 执行SQL语句 */
	_, err = stmt.Exec(rid, req.GetName(), chat.ROOM_STAT_OPEN,
		req.GetDesc(), time.Now().Unix(), time.Now().Unix(), req.GetUid())
	if nil != err {
		ctx.log.Error("Add rid failed! errmsg:%s", err.Error())
		return err
	}

	ctx.log.Debug("Add rid success!")

	return nil
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
func (ctx *ChatRoomCntx) room_creat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat) (rid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	defer func() {
		if err := recover(); nil != err {
			ctx.log.Error("Routine crashed! errmsg:%s", err)
		}
	}()

	/* > 分配聊天室ID */
	rid, err = ctx.alloc_rid()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	err = ctx.room_add(rid, req)
	if nil != err {
		ctx.log.Error("Room add failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 设置聊天室所有者 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_ROLE_TAB, rid)

	ok, err := redis.Bool(rds.Do("HSETNX", key, req.GetUid(), chat.ROOM_ROLE_OWNER))
	if nil != err {
		ctx.log.Error("Set room owner failed! uid:%d rid:%d errmsg:%s",
			req.GetUid(), rid, err.Error())
		return 0, err
	} else if !ok {
		ctx.log.Debug("Set room owner failed! uid:%d rid:%d", req.GetUid(), rid)
		return 0, errors.New("Set room owner failed!")
	}

	/* > 设置聊天室信息 */
	key = fmt.Sprintf(comm.CHAT_KEY_ROOM_INFO_TAB, rid)

	pl.Send("HMSET", key, "NAME", req.GetName(), "DESC", req.GetDesc())

	return rid, nil
}

/******************************************************************************
 **函数名称: ChatRoomCreatHandler
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
func ChatRoomCreatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-creat request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

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
func ChatRoomDismissHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
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
func (ctx *ChatRoomCntx) room_join_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomJoin, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-join failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
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
func (ctx *ChatRoomCntx) room_join_failed(head *comm.MesgHeader,
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
func (ctx *ChatRoomCntx) room_join_ack(head *comm.MesgHeader, req *mesg.MesgRoomJoin, gid uint32) int {
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
 **函数名称: room_join_notify
 **功    能: 发送上线通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:51:09 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_notify(head *comm.MesgHeader, req *mesg.MesgRoomJoin) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomJoinNtf{
		Uid: proto.Uint64(req.GetUid()),
		Rid: proto.Uint64(req.GetRid()),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发上线通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head.Cmd = comm.CMD_ROOM_JOIN_NTF
		head.Length = uint32(length)
		head.Nid = ctx.listend.list.nodes[idx]

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_NTF, p.Buff, uint32(len(p.Buff)))
	}

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
func (ctx *ChatRoomCntx) alloc_room_gid(rid uint64) (gid uint32, err error) {
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
func (ctx *ChatRoomCntx) room_join_handler(
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

	/* > 分配新的分组 */
	gid, err = ctx.alloc_room_gid(req.GetRid())
	if nil != err {
		ctx.log.Error("Alloc gid failed! rid:%d", req.GetRid())
		return 0, err
	}

	/* > 更新数据库统计 */
	key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZINCRBY", key, 1, gid)

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
 **函数名称: ChatRoomJoinHandler
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
func ChatRoomJoinHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
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
	ctx.room_join_notify(head, req)

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
func (ctx *ChatRoomCntx) room_quit_isvalid(req *mesg.MesgRoomQuit) bool {
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
func (ctx *ChatRoomCntx) room_quit_failed(head *comm.MesgHeader,
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
func (ctx *ChatRoomCntx) room_quit_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomQuit, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-quit failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
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
func (ctx *ChatRoomCntx) room_quit_ack(head *comm.MesgHeader, req *mesg.MesgRoomQuit) int {
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
 **函数名称: room_quit_notify
 **功    能: 发送下线通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:51:09 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_notify(head *comm.MesgHeader, req *mesg.MesgRoomQuit) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomQuitNtf{
		Uid: proto.Uint64(req.GetUid()),
		Rid: proto.Uint64(req.GetRid()),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发下线通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head.Cmd = comm.CMD_ROOM_QUIT_NTF
		head.Length = uint32(length)
		head.Nid = ctx.listend.list.nodes[idx]

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_NTF, p.Buff, uint32(len(p.Buff)))
	}

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
func (ctx *ChatRoomCntx) room_quit_handler(
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
 **函数名称: ChatRoomQuitHandler
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
func ChatRoomQuitHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
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
	ctx.room_quit_notify(head, req)

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
func (ctx *ChatRoomCntx) room_kick_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomKick, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-kick is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
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
func (ctx *ChatRoomCntx) room_kick_failed(head *comm.MesgHeader,
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
 **函数名称: room_kick_by_uid
 **功    能: 通过UID下发KICK指令
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
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.19 23:06:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_by_uid(rid uint64, uid uint64) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 获取会话列表 */
	key := fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, uid)

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid list by uid failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	num := len(sid_list)
	if 0 == num {
		return 0, nil
	}

	/* > 生成PB数据 */
	req := &mesg.MesgRoomKick{
		Uid: proto.Uint64(uid),
		Rid: proto.Uint64(rid),
	}

	body, err := proto.Marshal(req)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return comm.ERR_SVR_BODY_INVALID, err
	}

	length := len(body)

	/* > 遍历会话列表 */
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		/* > 获取会话属性 */
		attr, err := im.GetSidAttr(ctx.redis, uint64(sid))
		if nil != err {
			ctx.log.Error("Get sid attr failed! rid:%d uid:%d errmsg:%s",
				rid, uid, err.Error())
			return comm.ERR_SYS_SYSTEM, err
		}

		/* > 下发踢除指令 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head := &comm.MesgHeader{
			Cmd:    comm.CMD_ROOM_KICK,
			Sid:    attr.GetSid(),
			Cid:    attr.GetCid(),
			Nid:    attr.GetNid(),
			Length: uint32(length),
			Seq:    0,
		}

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK, p.Buff, uint32(len(p.Buff)))
	}
	return 0, nil
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
func (ctx *ChatRoomCntx) room_kick_ack(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
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
 **功    能: 发送被踢通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:59:32 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_notify(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomKickNtf{
		Uid: proto.Uint64(req.GetUid()),
		Rid: proto.Uint64(req.GetRid()),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发被踢通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head.Cmd = comm.CMD_ROOM_KICK_NTF
		head.Length = uint32(length)
		head.Nid = ctx.listend.list.nodes[idx]

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_NTF, p.Buff, uint32(len(p.Buff)))
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
func (ctx *ChatRoomCntx) room_kick_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomKick) (code uint32, err error) {
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
		ctx.log.Error("You're not owner! rid:%d kicked-uid:%d attr.uid:%d",
			req.GetRid(), req.GetUid(), attr.GetUid())
		return comm.ERR_SYS_PERM_DENIED, errors.New("You're not room owner!")
	}

	/* > 用户加入黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_USR_BLACKLIST_SET, req.GetRid())

	pl.Send("SADD", key, req.GetUid())

	/* > 提交MONGO存储 */
	data := &chat.RoomBlacklistTabRow{
		Rid:    req.GetRid(),             // 聊天室ID
		Uid:    req.GetUid(),             // 用户ID
		Status: chat.ROOM_USER_STAT_KICK, // 状态(被踢)
		Ctm:    time.Now().Unix(),        // 设置时间
	}

	cb := func(c *mgo.Collection) (err error) {
		c.Insert(data)
		return err
	}

	ctx.mongo.Exec(ctx.conf.Mongo.DbName, chat.ROOM_TAB_BLACKLIST, cb)

	/* > 遍历下发踢除指令 */
	ctx.room_kick_by_uid(req.GetRid(), req.GetUid())

	return 0, nil
}

/******************************************************************************
 **函数名称: ChatRoomKickHandler
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
func ChatRoomKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-kick request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析ROOM-KICK请求 */
	head, req, code, err := ctx.room_kick_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-kick failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 执行ROOM-KICK操作 */
	code, err = ctx.room_kick_handler(head, req)
	if nil != err {
		ctx.log.Error("Room-kick handler failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送ROOM-KICK应答 */
	ctx.room_kick_ack(head, req)
	ctx.room_kick_notify(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 聊天室各侦听统计 */

/******************************************************************************
 **函数名称: room_lsn_stat_parse
 **功    能: 解析ROOM-LSN-STAT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.13 06:51:41 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_lsn_stat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomLsnStat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(0) {
		errmsg := "Header of room-lsn-stat is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomLsnStat{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-lsn-stat request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_lsn_stat_handler
 **功    能: ROOM-LSN-STAT处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-LSN-STAT请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述:
 **注意事项: 已验证了ROOM-LSN-STAT请求的合法性
 **作    者: # Qifeng.zou # 2017.05.13 06:54:54 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_lsn_stat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomLsnStat) (code uint32, err error) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ttl := time.Now().Unix() + 30

	/* > 更新统计数据 */
	key := fmt.Sprintf(comm.CHAT_KEY_RID_NID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZADD", key, req.GetNum(), req.GetNid())

	pl.Send("ZADD", comm.CHAT_KEY_RID_ZSET, ttl, req.GetRid())

	return 0, nil
}

/******************************************************************************
 **函数名称: ChatRoomKickHandler
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
 **        required uint64 rid = 1;     // M|聊天室ID|数字|
 **        required uint32 nid = 2;     // M|结点ID|数字|
 **        required uint32 num = 3;     // M|在线人数|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.13 07:03:24 #
 ******************************************************************************/
func ChatRoomLsnStatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-lsn-stat request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析ROOM-LSN-STAT请求 */
	head, req, code, err := ctx.room_lsn_stat_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-lsn-stat request failed!")
		return -1
	}

	/* > 执行ROOM-LSN-STAT操作 */
	code, err = ctx.room_lsn_stat_handler(head, req)
	if nil != err {
		ctx.log.Error("Room lsn stat handler failed! code:%d errmsg:%s", code, err.Error())
		return -1
	}

	return 0
}

/******************************************************************************
 **函数名称: room_chat_parse
 **功    能: 解析ROOM-CHAT
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
 **作    者: # Qifeng.zou # 2016.11.04 22:29:23 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_chat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomChat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header is invalid!")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomChat{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-msg failed! sid:%d cid:%d nid:%d errmsg:%s",
			head.GetSid(), head.GetCid(), head.GetNid(), err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_chat_failed
 **功    能: 发送ROOM-CHAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: 聊天消息
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:52:14 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_chat_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomChat, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
		ack.Gid = proto.Uint32(req.GetGid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_ROOM_CHAT_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: room_chat_ack
 **功    能: 发送聊天消息应答
 **输入参数:
 **     head: 协议头
 **     req: 聊天室消息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 生成PB格式消息应答 并发送应答.
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_chat_ack(head *comm.MesgHeader, req *mesg.MesgRoomChat) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(req.GetGid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_ROOM_CHAT_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: room_chat_handler
 **功    能: ROOM-CHAT处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-CHAT请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **     1. 将消息存放在聊天室历史消息表中
 **     2. 遍历rid->nid列表, 并转发聊天室消息
 **注意事项: TODO: 增加敏感词过滤功能, 屏蔽政治、低俗、侮辱性词汇.
 **作    者: # Qifeng.zou # 2016.11.04 22:34:55 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_chat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomChat, data []byte) (err error) {
	item := &MesgRoomItem{}

	ctx.log.Debug("rid:%d gid:%d sid:%d cid:%d nid:%d",
		req.GetRid(), req.GetGid(), head.GetSid(), head.GetCid(), head.GetNid())

	/* 1. 放入存储队列 */
	item.head = head
	item.req = req
	item.raw = data

	ctx.room_mesg_chan <- item

	/* 2. 下发聊天室消息 */
	ctx.room.node.RLock()
	defer ctx.room.node.RUnlock()
	nid_list, ok := ctx.room.node.m[req.GetRid()]
	if !ok {
		ctx.log.Error("Get node list failed! rid:%d", req.GetRid())
		return nil
	}

	/* > 遍历rid->nid列表, 并下发聊天室消息 */
	for idx, nid := range nid_list {
		ctx.log.Debug("idx:%d rid:%d nid:%d", idx, req.GetRid(), nid)

		ctx.send_data(comm.CMD_ROOM_CHAT, head.GetSid(), 0, uint32(nid),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}
	return err
}

/******************************************************************************
 **函数名称: ChatRoomChatHandler
 **功    能: 聊天室消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将消息放入历史队列
 **     3. 将消息发送分发到聊天室对应帧听层.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:28:02 #
 ******************************************************************************/
func ChatRoomChatHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-chat message!")

	/* > 解析ROOM-CHAT协议 */
	head, req, code, err := ctx.room_chat_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-msg failed! code:%d errmsg:%s", code, err.Error())
		if nil != head {
			ctx.room_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		}
		return -1
	}

	/* > 进行业务处理 */
	err = ctx.room_chat_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle room message failed!")
		ctx.room_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	return ctx.room_chat_ack(head, req)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomChatAckHandler
 **功    能: 聊天室消息应答
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 暂不处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:01:06 #
 ******************************************************************************/
func ChatRoomChatAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomBcHandler
 **功    能: 聊天室广播消息处理
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将消息放入聊天室广播队列
 **     3. 将消息发送分发到聊天室对应帧听层.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:01:06 #
 ******************************************************************************/
func ChatRoomBcHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomBcAckHandler
 **功    能: 聊天室广播消息处理
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 暂不处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:01:06 #
 ******************************************************************************/
func ChatRoomBcAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task_room_mesg_chan_pop
 **功    能: 聊天室消息的存储任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从聊天室消息队列中取出消息, 并进行存储处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 23:43:03 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) task_room_mesg_chan_pop() {
	for item := range ctx.room_mesg_chan {
		item.storage(ctx)
	}
}

/******************************************************************************
 **函数名称: storage
 **功    能: 聊天室消息的存储处理
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 将消息存入聊天室缓存和数据库
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:05:51 #
 ******************************************************************************/
func (item *MesgRoomItem) storage(ctx *ChatRoomCntx) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 解析PB协议 */
	msg := &mesg.MesgRoomChat{}

	err := proto.Unmarshal(item.raw[comm.MESG_HEAD_SIZE:], msg)
	if nil != err {
		ctx.log.Error("Unmarshal room-chat-mesg failed!")
		return
	}

	/* > 提交REDIS缓存 */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_MESG_QUEUE, item.req.GetRid())
	pl.Send("LPUSH", key, item.raw[comm.MESG_HEAD_SIZE:])

	/* > 提交MONGO存储 */
	data := &chat.RoomChatTabRow{
		Rid:  msg.GetRid(),
		Uid:  msg.GetUid(),
		Ctm:  time.Now().Unix(),
		Data: item.raw,
	}

	cb := func(c *mgo.Collection) (err error) {
		c.Insert(data)
		return err
	}

	ctx.mongo.Exec(ctx.conf.Mongo.DbName, chat.ROOM_TAB_MESG, cb)
}

/******************************************************************************
 **函数名称: task_room_mesg_queue_clean
 **功    能: 清理聊天室缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持聊天室缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:34:18 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) task_room_mesg_queue_clean() {
	for {
		ctx.room_mesg_queue_clean()

		time.Sleep(30 * time.Second)
	}
}

/******************************************************************************
 **函数名称: room_mesg_queue_clean
 **功    能: 清理聊天室缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持聊天室缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:34:18 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_mesg_queue_clean() {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, 0, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get rid list failed! errmsg:%s", err.Error())
			break
		}

		num := len(rid_list)
		for idx := 0; idx < num; idx += 1 {
			/* 保持聊天室缓存消息为最新的100条 */
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			key := fmt.Sprintf(comm.CHAT_KEY_ROOM_MESG_QUEUE, uint64(rid))

			rds.Do("LTRIM", key, 0, 99)
		}

		if num < comm.CHAT_BAT_NUM {
			break
		}
		off += num
	}
}

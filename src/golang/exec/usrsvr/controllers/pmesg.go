package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
/* 添加好友 */

/******************************************************************************
 **函数名称: friend_add_parse
 **功    能: 解析FRIEND-ADD消息
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述: 1.对通用头进行字节序转换 2.解析PB协议体
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 21:57:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_add_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgFriendAdd, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header of friend-add is invalid")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgFriendAdd{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	} else if 0 == req.GetOrig() || 0 == req.GetDest() {
		ctx.log.Error("Paramter isn't right! orig:%d dest:%d", req.GetOrig(), req.GetDest())
		return head, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Paramter isn't right!")
	}

	return head, req, comm.OK, nil
}

/******************************************************************************
 **函数名称: friend_add_failed
 **功    能: 发送FRIEND-ADD应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-ADD请求
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
func (ctx *UsrSvrCntx) friend_add_failed(head *comm.MesgHeader,
	req *mesg.MesgFriendAdd, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgFriendAddAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 发送应答数据 */
	return ctx.send_data(comm.CMD_FRIEND_ADD_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: friend_add_ack
 **功    能: 发送FRIEND-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-ADD请求
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 22:14:27 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_add_ack(head *comm.MesgHeader, req *mesg.MesgFriendAdd) int {
	/* > 设置协议体 */
	ack := &mesg.MesgFriendAddAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 发送应答数据 */
	return ctx.send_data(comm.CMD_FRIEND_ADD_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: friend_add_handler
 **功    能: FRIEND-ADD处理
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-ADD请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误码+错误描述
 **实现描述:
 **     1. 将消息放入UID离线队列
 **     2. 判断接收方是否在线.
 **        > 如果在线, 则直接下发消息
 **        > 如果不在线, 则无需下发消息
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 22:07:00 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_add_handler(head *comm.MesgHeader,
	req *mesg.MesgFriendAdd, data []byte) (code uint32, err error) {
	var key string

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 将消息放入UID离线队列 */

	/* > 发送给"接收方"所有终端.
	        1.如果在线, 则直接下发消息
		    2.如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetDest())

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetDest())
		return comm.ERR_SYS_DB, err
	}

	num := len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		attr, _ := im.GetSidAttr(ctx.redis, uint64(sid))
		if nil == attr {
			continue
		} else if 0 == attr.GetNid() {
			ctx.log.Error("Nid is invalid! uid:%d sid:%d cid:%d nid:%d!",
				req.GetDest(), sid, attr.GetCid(), attr.GetNid())
			continue
		} else if uint64(attr.GetUid()) != req.GetDest() {
			ctx.log.Error("uid:%d sid:%d cid:%d nid:%d!",
				req.GetDest(), sid, attr.GetCid(), attr.GetNid())
			continue
		}

		ctx.log.Debug("uid:%d sid:%d cid:%d nid:%d!",
			req.GetDest(), sid, attr.GetCid(), attr.GetNid())

		ctx.send_data(comm.CMD_FRIEND_ADD,
			uint64(sid), attr.GetCid(), uint32(attr.GetNid()),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: UsrSvrFriendAddHandler
 **功    能: FRIEND-ADD消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 首先将FRIEND-ADD消息放入接收方离线队列.
 **     2. 如果接收方当前在线, 则直接下发FRIEND-ADD消息;
 **        如果接收方当前"不"在线, 则"不"下发FRIEND-ADD消息.
 **     3. 给发送方下发FRIEND-ADD应答消息.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 21:49:36 #
 ******************************************************************************/
func UsrSvrFriendAddHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv friend add request!")

	/* > 解析FRIEND-ADD协议 */
	head, req, code, err := ctx.friend_add_parse(data)
	if nil == head {
		ctx.log.Error("Parse friend add failed! errmsg:%s", err.Error())
		return -1
	} else if nil == req {
		ctx.log.Error("Parse friend add failed! errmsg:%s", err.Error())
		ctx.friend_add_failed(head, req, code, err.Error())
		return -1
	}

	ctx.log.Debug("Uid [%d] send friend-add to uid [%d]!", req.GetOrig(), req.GetDest())

	/* > 进行业务处理 */
	code, err = ctx.friend_add_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle friend add failed! errmsg:%s", err.Error())
		ctx.friend_add_failed(head, req, code, err.Error())
		return -1
	}

	ctx.friend_add_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 删除好友 */

/******************************************************************************
 **函数名称: friend_del_parse
 **功    能: 解析FRIEND-DEL消息
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述: 1.对通用头进行字节序转换 2.解析PB协议体
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 21:57:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_del_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgFriendDel, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header of friend-add is invalid")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgFriendDel{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	} else if 0 == req.GetOrig() || 0 == req.GetDest() {
		ctx.log.Error("Paramter isn't right! orig:%d dest:%d", req.GetOrig(), req.GetDest())
		return head, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Paramter isn't right!")
	}

	return head, req, comm.OK, nil
}

/******************************************************************************
 **函数名称: friend_del_failed
 **功    能: 发送FRIEND-DEL应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-DEL请求
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
 **作    者: # Qifeng.zou # 2017.06.07 22:19:11 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_del_failed(head *comm.MesgHeader,
	req *mesg.MesgFriendDel, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgFriendDelAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 发送应答数据 */
	return ctx.send_data(comm.CMD_FRIEND_DEL_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: friend_del_ack
 **功    能: 发送FRIEND-DEL应答
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-ADD请求
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 22:14:27 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_del_ack(head *comm.MesgHeader, req *mesg.MesgFriendDel) int {
	/* > 设置协议体 */
	ack := &mesg.MesgFriendDelAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 发送应答数据 */
	return ctx.send_data(comm.CMD_FRIEND_DEL_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: friend_del_handler
 **功    能: FRIEND-DEL处理
 **输入参数:
 **     head: 协议头
 **     req: FRIEND-DEL请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误码+错误描述
 **实现描述:
 **     1. 将消息放入UID离线队列
 **     2. 判断接收方是否在线.
 **        > 如果在线, 则直接下发消息
 **        > 如果不在线, 则无需下发消息
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 22:07:00 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) friend_del_handler(head *comm.MesgHeader,
	req *mesg.MesgFriendDel, data []byte) (code uint32, err error) {
	var key string

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 将消息放入UID离线队列 */

	/* > 发送给"接收方"所有终端.
	        1.如果在线, 则直接下发消息
		    2.如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetDest())

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetDest())
		return comm.ERR_SYS_DB, err
	}

	num := len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		attr, _ := im.GetSidAttr(ctx.redis, uint64(sid))
		if nil == attr {
			continue
		} else if 0 == attr.GetNid() {
			ctx.log.Error("Nid is invalid! uid:%d sid:%d cid:%d nid:%d!",
				req.GetDest(), sid, attr.GetCid(), attr.GetNid())
			continue
		} else if uint64(attr.GetUid()) != req.GetDest() {
			ctx.log.Error("uid:%d sid:%d cid:%d nid:%d!",
				req.GetDest(), sid, attr.GetCid(), attr.GetNid())
			continue
		}

		ctx.log.Debug("uid:%d sid:%d cid:%d nid:%d!",
			req.GetDest(), sid, attr.GetCid(), attr.GetNid())

		ctx.send_data(comm.CMD_FRIEND_DEL,
			uint64(sid), attr.GetCid(), uint32(attr.GetNid()),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: UsrSvrFriendDelHandler
 **功    能: FRIEND-DEL消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 首先将FRIEND-DEL消息放入接收方离线队列.
 **     2. 如果接收方当前在线, 则直接下发FRIEND-ADD消息;
 **        如果接收方当前"不"在线, 则"不"下发FRIEND-ADD消息.
 **     3. 给发送方下发FRIEND-DEL应答消息.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.07 21:49:36 #
 ******************************************************************************/
func UsrSvrFriendDelHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv friend del request!")

	/* > 解析FRIEND-DEL协议 */
	head, req, code, err := ctx.friend_del_parse(data)
	if nil == head {
		ctx.log.Error("Parse friend del failed! errmsg:%s", err.Error())
		return -1
	} else if nil == req {
		ctx.log.Error("Parse friend del failed! errmsg:%s", err.Error())
		ctx.friend_del_failed(head, req, code, err.Error())
		return -1
	}

	ctx.log.Debug("Uid [%d] send friend-del to uid [%d]!", req.GetOrig(), req.GetDest())

	/* > 进行业务处理 */
	code, err = ctx.friend_del_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle friend del failed! errmsg:%s", err.Error())
		ctx.friend_del_failed(head, req, code, err.Error())
		return -1
	}

	ctx.friend_del_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 加入黑名单 */

/******************************************************************************
 **函数名称: blacklist_add_parse
 **功    能: 解析BLACKLIST-ADD请求
 **输入参数:
 **     data: 原始数据
 **输出参数: NONE
 **返    回:
 **     head: 协议头
 **     req: 请求内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 10:06:14 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_add_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgBlacklistAdd, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of blacklist-add is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body of blacklist-add failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: blacklist_add_handler
 **功    能: 进行BLACKLIST-ADD处理
 **输入参数:
 **     head: 协议头
 **     req: 请求内容
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 10:30:04 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_add_handler(
	head *comm.MesgHeader, req *mesg.MesgBlacklistAdd) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 加入用户黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_BLACKLIST_ZSET, req.GetOrig())

	_, err = rds.Do("ZADD", key, ctm, req.GetDest())
	if nil != err {
		ctx.log.Error("Add into blacklist failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: blacklist_add_failed
 **功    能: 发送BLACKLIST-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: BLACKLIST-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_add_failed(head *comm.MesgHeader,
	req *mesg.MesgBlacklistAdd, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgOnlineAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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

	head.Cmd = comm.CMD_BLACKLIST_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_BLACKLIST_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: blacklist_add_ack
 **功    能: 发送BLACKLIST-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: BLACKLIST-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 10:40:03 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_add_ack(
	head *comm.MesgHeader, req *mesg.MesgBlacklistAdd) int {
	/* > 设置协议体 */
	ack := &mesg.MesgBlacklistAddAck{
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

	head.Cmd = comm.CMD_BLACKLIST_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_BLACKLIST_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: UsrSvrBlacklistAddHandler
 **功    能: 加入黑名单
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 09:49:10 #
 ******************************************************************************/
func UsrSvrBlacklistAddHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	/* > 解析BLACKLIST-ADD请求 */
	head, req, code, err := ctx.blacklist_add_parse(data)
	if nil != err {
		ctx.log.Error("Parse blacklist-add failed! code:%d errmsg:%s", code, err.Error())
		ctx.blacklist_add_failed(head, req, code, err.Error())
		return -1
	}

	/* > 验证请求合法性 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get attr by sid failed! errmsg:%s", err.Error())
		ctx.blacklist_add_failed(head, req, code, err.Error())
		return -1
	} else if 0 != attr.GetUid() && attr.GetUid() != req.GetOrig() {
		errmsg := "Uid is collision!"
		ctx.log.Error("errmsg:%s", errmsg)
		ctx.blacklist_add_failed(head, req, comm.ERR_SYS_SYSTEM, errmsg)
		return -1
	}

	/* > 进行BLACKLIST-ADD处理 */
	code, err = ctx.blacklist_add_handler(head, req)
	if nil != err {
		ctx.log.Error("Handle blacklist-add failed! code:%d errmsg:%s", code, err.Error())
		ctx.blacklist_add_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送BLACKLIST-ADD应答 */
	ctx.blacklist_add_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 移除黑名单 */

/******************************************************************************
 **函数名称: blacklist_del_parse
 **功    能: 解析BLACKLIST-DEL请求
 **输入参数:
 **     data: 原始数据
 **输出参数: NONE
 **返    回:
 **     head: 协议头
 **     req: 请求内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:03:54 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_del_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgBlacklistDel, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of blacklist-del is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body of blacklist-add failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: blacklist_del_handler
 **功    能: 进行BLACKLIST-DEL处理
 **输入参数:
 **     head: 协议头
 **     req: 请求内容
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:04:31 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_del_handler(
	head *comm.MesgHeader, req *mesg.MesgBlacklistDel) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 移除用户黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_BLACKLIST_ZSET, req.GetOrig())

	_, err = rds.Do("ZREM", key, req.GetDest())
	if nil != err {
		ctx.log.Error("Remove blacklist failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: blacklist_del_failed
 **功    能: 发送BLACKLIST-DEL应答
 **输入参数:
 **     head: 协议头
 **     req: BLACKLIST-DEL请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 11:05:32 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_del_failed(head *comm.MesgHeader,
	req *mesg.MesgBlacklistDel, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgBlacklistDelAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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

	head.Cmd = comm.CMD_BLACKLIST_DEL_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_BLACKLIST_DEL_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: blacklist_del_ack
 **功    能: 发送BLACKLIST-DEL应答
 **输入参数:
 **     head: 协议头
 **     req: BLACKLIST-DEL请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) blacklist_del_ack(
	head *comm.MesgHeader, req *mesg.MesgBlacklistDel) int {
	/* > 设置协议体 */
	ack := &mesg.MesgBlacklistDelAck{
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

	head.Cmd = comm.CMD_BLACKLIST_DEL_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_BLACKLIST_DEL_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: UsrSvrBlacklistDelHandler
 **功    能: 移除黑名单
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:54 #
 ******************************************************************************/
func UsrSvrBlacklistDelHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	/* > 解析BLACKLIST-DEL请求 */
	head, req, code, err := ctx.blacklist_del_parse(data)
	if nil != err {
		ctx.log.Error("Parse blacklist-del failed! code:%d errmsg:%s", code, err.Error())
		ctx.blacklist_del_failed(head, req, code, err.Error())
		return -1
	}

	/* > 验证请求合法性 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get attr by sid failed! errmsg:%s", err.Error())
		ctx.blacklist_del_failed(head, req, code, err.Error())
		return -1
	} else if 0 != attr.GetUid() && attr.GetUid() != req.GetOrig() {
		errmsg := "Uid is collision!"
		ctx.log.Error("errmsg:%s", errmsg)
		ctx.blacklist_del_failed(head, req, comm.ERR_SYS_SYSTEM, errmsg)
		return -1
	}

	/* > 进行BLACKLIST-DEL处理 */
	code, err = ctx.blacklist_del_handler(head, req)
	if nil != err {
		ctx.log.Error("Handle blacklist-del failed! code:%d errmsg:%s", code, err.Error())
		ctx.blacklist_del_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送BLACKLIST-DEL应答 */
	ctx.blacklist_del_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 设置禁言 */

/******************************************************************************
 **函数名称: gag_add_parse
 **功    能: 解析GAG-ADD请求
 **输入参数:
 **     data: 原始数据
 **输出参数: NONE
 **返    回:
 **     head: 协议头
 **     req: 请求内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:03:54 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_add_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgGagAdd, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of gag-add is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body of gag-add failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: gag_add_handler
 **功    能: 进行GAG-ADD处理
 **输入参数:
 **     head: 协议头
 **     req: 请求内容
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:04:31 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_add_handler(
	head *comm.MesgHeader, req *mesg.MesgGagAdd) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 移除用户黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_GAG_ZSET, req.GetOrig())

	_, err = rds.Do("ZADD", key, ctm, req.GetDest())
	if nil != err {
		ctx.log.Error("Add into gag-list failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: gag_add_failed
 **功    能: 发送GAG-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: GAG-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 11:05:32 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_add_failed(head *comm.MesgHeader,
	req *mesg.MesgGagAdd, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgGagAddAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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

	head.Cmd = comm.CMD_GAG_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GAG_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: gag_add_ack
 **功    能: 发送GAG-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: GAG-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_add_ack(
	head *comm.MesgHeader, req *mesg.MesgGagAdd) int {
	/* > 设置协议体 */
	ack := &mesg.MesgGagDelAck{
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

	head.Cmd = comm.CMD_GAG_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GAG_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: UsrSvrGagAddHandler
 **功    能: 设置禁言
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:54 #
 ******************************************************************************/
func UsrSvrGagAddHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	/* > 解析GAG-ADD请求 */
	head, req, code, err := ctx.gag_add_parse(data)
	if nil != err {
		ctx.log.Error("Parse gag-add failed! code:%d errmsg:%s", code, err.Error())
		ctx.gag_add_failed(head, req, code, err.Error())
		return -1
	}

	/* > 验证请求合法性 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get attr by sid failed! errmsg:%s", err.Error())
		ctx.gag_add_failed(head, req, code, err.Error())
		return -1
	} else if 0 != attr.GetUid() && attr.GetUid() != req.GetOrig() {
		errmsg := "Uid is collision!"
		ctx.log.Error("errmsg:%s", errmsg)
		ctx.gag_add_failed(head, req, comm.ERR_SYS_SYSTEM, errmsg)
		return -1
	}

	/* > 进行GAG-ADD处理 */
	code, err = ctx.gag_add_handler(head, req)
	if nil != err {
		ctx.log.Error("Handle gag-add failed! code:%d errmsg:%s", code, err.Error())
		ctx.gag_add_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送GAG-ADD应答 */
	ctx.gag_add_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 解除禁言 */

/******************************************************************************
 **函数名称: gag_del_parse
 **功    能: 解析GAG-ADD请求
 **输入参数:
 **     data: 原始数据
 **输出参数: NONE
 **返    回:
 **     head: 协议头
 **     req: 请求内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:03:54 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_del_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgGagDel, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of gag-del is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body of gag-del failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: gag_del_handler
 **功    能: 进行GAG-ADD处理
 **输入参数:
 **     head: 协议头
 **     req: 请求内容
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:04:31 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_del_handler(
	head *comm.MesgHeader, req *mesg.MesgGagDel) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 移除用户黑名单 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_GAG_ZSET, req.GetOrig())

	_, err = rds.Do("ZREM", key, req.GetDest())
	if nil != err {
		ctx.log.Error("Remove gag failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: gag_del_failed
 **功    能: 发送GAG-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: GAG-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 11:05:32 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_del_failed(head *comm.MesgHeader,
	req *mesg.MesgGagDel, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgGagDelAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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

	head.Cmd = comm.CMD_GAG_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GAG_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: gag_del_ack
 **功    能: 发送GAG-ADD应答
 **输入参数:
 **     head: 协议头
 **     req: GAG-ADD请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint32 code = 1;       // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) gag_del_ack(
	head *comm.MesgHeader, req *mesg.MesgGagDel) int {
	/* > 设置协议体 */
	ack := &mesg.MesgGagDelAck{
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

	head.Cmd = comm.CMD_GAG_ADD_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GAG_ADD_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: UsrSvrGagDelHandler
 **功    能: 解除禁言
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 11:07:54 #
 ******************************************************************************/
func UsrSvrGagDelHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	/* > 解析GAG-ADD请求 */
	head, req, code, err := ctx.gag_del_parse(data)
	if nil != err {
		ctx.log.Error("Parse gag-del failed! code:%d errmsg:%s", code, err.Error())
		ctx.gag_del_failed(head, req, code, err.Error())
		return -1
	}

	/* > 验证请求合法性 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get attr by sid failed! errmsg:%s", err.Error())
		ctx.gag_del_failed(head, req, code, err.Error())
		return -1
	} else if 0 != attr.GetUid() && attr.GetUid() != req.GetOrig() {
		errmsg := "Uid is collision!"
		ctx.log.Error("errmsg:%s", errmsg)
		ctx.gag_del_failed(head, req, comm.ERR_SYS_SYSTEM, errmsg)
		return -1
	}

	/* > 进行GAG-ADD处理 */
	code, err = ctx.gag_del_handler(head, req)
	if nil != err {
		ctx.log.Error("Handle gag-del failed! code:%d errmsg:%s", code, err.Error())
		ctx.gag_del_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送GAG-ADD应答 */
	ctx.gag_del_ack(head, req)

	return 0
}

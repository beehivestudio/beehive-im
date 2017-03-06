package controllers

import (
	_ "encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	_ "time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: DownlinkRegister
 **功    能: 下行消息回调注册
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func (ctx *LsndCntx) DownlinkRegister() {
	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndFrwderOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_OFFLINE_ACK, LsndFrwderOfflineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_SUB_ACK, LsndFrwderSubAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_UNSUB_ACK, LsndFrwderUnsubAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_SYNC_ACK, LsndFrwderCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_ALLOC_SEQ_ACK, LsndFrwderCommHandler, ctx)

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_CHAT, LsndFrwderCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_CHAT_ACK, LsndFrwderCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_FRIEND_ADD_ACK, LsndFrwderCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_FRIEND_DEL_ACK, LsndFrwderCommHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CHAT, LsndFrwderGroupChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_CHAT_ACK, LsndFrwderGroupChatAckHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndFrwderRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, LsndFrwderRoomChatAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndFrwderRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_ACK, LsndFrwderRoomBcAckHandler, ctx)

	/* > 推送消息 */
	ctx.frwder.Register(comm.CMD_BC, LsndFrwderCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_BC_ACK, LsndFrwderCommHandler, ctx)
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndFrwderCommHandler
 **功    能: 通用消息处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func LsndFrwderCommHandler(cmd uint32, data []byte, length uint32, param interface{}) int {
	var req mesg.MesgChat
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv command [%d]!", cmd)

	/* > 验证合法性 */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Mesg head is invalid!")
		return -1
	}

	/* > 下发私聊消息 */
	cid := ctx.find_cid_by_sid(head.GetSid())
	if 0 == cid {
		ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.lws.AsyncSend(cid, data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndBcAckHandler
 **功    能: 广播消息应答处理(待商议)
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:54:37 #
 ******************************************************************************/
func LsndBcAckHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndP2pMsgHandler
 **功    能: 点到点消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断点到点消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将消息放入离线队列
 **     3. 将消息发送给在线的人员.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:56:56 #
 ******************************************************************************/
func LsndP2pMsgHandler(cmd uint32, dest uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndP2pMsgAckHandler
 **功    能: 点到点应答的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将离线消息从离线队列中删除.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:58:12 #
 ******************************************************************************/
func LsndP2pMsgAckHandler(cmd uint32, dest uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: sync_parse
 **功    能: 解析SYNC请求
 **输入参数:
 **     data: 收到数据
 **输出参数: NONE
 **返    回:
 **     head: 协议头
 **     req: 请求数据
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 22:53:36 #
 ******************************************************************************/
func (ctx *LsndCntx) sync_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgSync, code uint32, err error) {
	/* 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Message header of sync is invalid!")
		return nil, nil, comm.ERR_SVR_PARSE_PARAM, errors.New("Header of sync-req invalid!")
	}

	/* 解析协议体 */
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body of sync is invalid! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_PARSE_PARAM, errors.New("Boby of sync-req invalid!")
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: sync_handler
 **功    能: 处理SYNC请求
 **输入参数:
 **     head: 头部数据
 **     req: SYNC请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.15 00:32:53 #
 ******************************************************************************/
func (ctx *LsndCntx) sync_handler(
	head *comm.MesgHeader, req *mesg.MesgSync) (code uint32, err error) {
	rds := ctx.redis.Get()
	rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取离线消息ID列表 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_OFFLINE_ZSET, req.GetUid())

	mesg_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, 0, "+inf"))
	if nil != err {
		ctx.log.Error("Get offline message failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	/* > 遍历离线消息ID列表 */
	num := len(mesg_list)
	for idx := 0; idx < num; idx += 1 {
		vals := strings.Split(mesg_list[idx], ":")
		uid, err := strconv.ParseInt(vals[0], 10, 64)   // 消息发送者的UID
		msgid, err := strconv.ParseInt(vals[1], 10, 64) // 消息发送者的消息ID
		if 0 == msgid || 0 == uid {
			ctx.log.Error("Parse offline message failed! mesg:%s", mesg_list[idx])
			pl.Send("ZREM", mesg_list[idx])
			continue
		}

		/* > 获取离线消息*/
		mesg_key := fmt.Sprintf(comm.CHAT_KEY_USR_SEND_MESG_HTAB, uid)

		data, err := redis.String(rds.Do("HGET", mesg_key, msgid))
		if nil != err {
			ctx.log.Error("Get offline message failed! mesg:%s", mesg_list[idx])
			pl.Send("ZREM", key, mesg_list[idx])
			pl.Send("HDEL", mesg_key, msgid)
			continue
		}

		/* > 判断消息合法性 */
		hhead := comm.MesgHeadNtoh([]byte(data))
		if !hhead.IsValid() {
			ctx.log.Error("Message header is invalid!")
			pl.Send("ZREM", key, mesg_list[idx])
			pl.Send("HDEL", mesg_key, msgid)
			continue
		}

		switch hhead.GetCmd() {
		case comm.CMD_CHAT: // 私聊消息
			msg := &mesg.MesgChat{}

			err = proto.Unmarshal([]byte(data[comm.MESG_HEAD_SIZE:]), msg)
			if nil != err {
				ctx.log.Error("Unmarshal offline message failed! uid:%d msgid:%d errmsg:%s",
					uid, msgid, err.Error())
				pl.Send("ZREM", key, mesg_list[idx])
				pl.Send("HDEL", mesg_key, msgid)
				continue
			}

			/* > 下发离线消息*/
			ctx.send_data(comm.CMD_CHAT, head.GetSid(), head.GetNid(),
				uint64(msgid), []byte(data[comm.MESG_HEAD_SIZE:]), hhead.GetLength())
		}
	}
	return 0, nil
}

/******************************************************************************
 **函数名称: send_err_sync_ack
 **功    能: 发送SYNC应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: SYNC请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint32 code = 2;   // M|错误码|数字|
 **         required string errmsg = 3; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 23:12:29 #
 ******************************************************************************/
func (ctx *LsndCntx) send_err_sync_ack(head *comm.MesgHeader,
	req *mesg.MesgSync, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgSyncAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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

	head.Cmd = comm.CMD_SYNC_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_SYNC_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: send_sync_ack
 **功    能: 发送SYNC消息应答
 **输入参数:
 **     head: 协议头
 **     req: SYNC请求
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 生成PB格式消息应答 并发送应答.
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint32 code = 2;   // M|错误码|数字|
 **         required string errmsg = 3; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 23:08:08 #
 ******************************************************************************/
func (ctx *LsndCntx) send_sync_ack(head *comm.MesgHeader, req *mesg.MesgSync) int {
	/* > 设置协议体 */
	ack := &mesg.MesgSyncAck{
		Uid:    proto.Uint64(req.GetUid()),
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

	head.Cmd = comm.CMD_SYNC_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_SYNC_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: LsndSyncHandler
 **功    能: 同步请求的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 收到同步请求后, 下发离线消息.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 22:49:17 #
 ******************************************************************************/
func LsndSyncHandler(cmd uint32, dest uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv sync-req message!")

	/* > 解析消息同步请求 */
	head, req, code, err := ctx.sync_parse(data)
	if nil != err {
		ctx.log.Error("Parse sync message failed! errmsg:%s", err.Error())
		ctx.send_err_sync_ack(head, req, code, err.Error())
		return -1
	}

	/* > 校验消息同步请求 */
	attr, err := im.GetSidAttr(ctx.redis, head.GetSid())
	if nil != err {
		ctx.log.Error("Get session attr failed! errmsg:%s", err.Error())
		ctx.send_err_sync_ack(head, req, code, err.Error())
		return -1
	} else if req.GetUid() != attr.Uid {
		ctx.log.Error("Sync data failed! uid:%d/%d nid:%d/%d",
			req.GetUid(), attr.Uid, head.GetNid(), attr.Nid)
		ctx.send_err_sync_ack(head, req, comm.ERR_SVR_DATA_COLLISION, "Uid is collision!")
		return -1
	}

	/* > 处理消息同步请求 */
	code, err = ctx.sync_handler(head, req)
	if nil != err {
		ctx.log.Error("Handle sync message failed! errmsg:%s", err.Error())
		ctx.send_err_sync_ack(head, req, code, err.Error())
		return -1
	}

	ctx.send_sync_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

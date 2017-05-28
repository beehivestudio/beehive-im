package controllers

import (
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: UpMesgRegister
 **功    能: 下行消息回调注册
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 为"下行"消息注册处理函数
 **注意事项: "下行"消息指的是从转发层发送过来的消息
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func (ctx *LsndCntx) UpMesgRegister() {
	/* > 未知消息 */
	ctx.frwder.Register(comm.CMD_UNKNOWN, LsndUpMesgCommHandler, ctx)

	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndUpMesgOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_KICK, LsndUpMesgKickHandler, ctx)
	ctx.frwder.Register(comm.CMD_SUB_ACK, LsndUpMesgSubAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_UNSUB_ACK, LsndUpMesgCommHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_JOIN_ACK, LsndUpMesgRoomJoinAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndUpMesgRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, LsndUpMesgCommHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndUpMesgRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_KICK, LsndUpMesgRoomKickHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_USR_NUM, LsndUpMesgRoomUsrNumHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_JOIN_NTC, LsndUpMesgRoomJoinNtcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_QUIT_NTC, LsndUpMesgRoomQuitNtcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_KICK_NTC, LsndUpMesgRoomKickNtcHandler, ctx)

	/* > 内部运维消息 */
	ctx.frwder.Register(comm.CMD_LSND_INFO_ACK, LsndUpMesgLsndInfoAckHandler, ctx)
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgCommHandler
 **功    能: 通用消息处理
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
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func LsndUpMesgCommHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv command [0x%04X]!", cmd)

	/* > 验证合法性 */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Mesg head is invalid!")
		return -1
	}

	/* > 获取会话数据 */
	cid := head.GetCid()
	if 0 == cid {
		cid = ctx.chat.GetCidBySid(head.GetSid())
		if 0 == cid {
			ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
			return -1
		}
	}

	extra := ctx.chat.SessionGetParam(head.GetSid(), cid)
	if nil == extra {
		ctx.log.Error("Didn't find conn data! sid:%d cid:%d", head.GetSid(), cid)
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.log.Debug("Session extra data. sid:%d cid:%d status:%d",
		conn.GetSid(), conn.GetCid(), conn.GetStatus())

	ctx.lws.AsyncSend(conn.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_error_online_ack_handler
 **功    能: ONLINE-ACK的异常处理
 **输入参数:
 **     cid: 连接ID
 **     head: 通用头(主机字节序)
 **     data: 下发数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.13 01:06:41 #
 ******************************************************************************/
func (ctx *LsndCntx) lsnd_error_online_ack_handler(cid uint64, head *comm.MesgHeader, data []byte) int {
	/* > 下发ONLINE-ACK消息 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换(主机 - > 网络) */

	ctx.lws.AsyncSend(cid, data)

	/* > 加入被踢列表 */
	ctx.kick_add(cid)

	return -1
}

/******************************************************************************
 **函数名称: LsndUpMesgOnlineAckHandler
 **功    能: ONLINE-ACK消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.07 23:49:03 #
 ******************************************************************************/
func LsndUpMesgOnlineAckHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv online ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of online-ack is invalid!")
		return -1
	}

	cid := head.GetCid()

	/* > 消息ONLINE-ACK的处理 */
	ack := &mesg.MesgOnlineAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal online-ack failed! errmsg:%s", err.Error())
		return ctx.lsnd_error_online_ack_handler(cid, head, data)
	} else if 0 != ack.GetCode() {
		ctx.log.Error("Online failed! cid:%d sid:%d code:%d errmsg:%s",
			head.GetCid(), ack.GetSid(), ack.GetCode(), ack.GetErrmsg())
		return ctx.lsnd_error_online_ack_handler(cid, head, data)
	}

	/* > 获取&更新会话状态 */
	extra := ctx.chat.SessionGetParam(ack.GetSid(), cid)
	if nil == extra {
		ctx.log.Error("Didn't find conn data! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		return ctx.lsnd_error_online_ack_handler(cid, head, data)
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		return ctx.lsnd_error_online_ack_handler(cid, head, data)
	} else if !conn.SetSeq(ack.GetSeq()) {
		ctx.log.Error("Update conn req failed! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		return ctx.lsnd_error_online_ack_handler(cid, head, data)
	}

	conn.SetStatus(CONN_STATUS_LOGIN) /* 已登录 */

	/* 更新SID->CID映射 */
	_cid := ctx.chat.GetCidBySid(ack.GetSid())
	if cid != _cid {
		ctx.kick_add(_cid)
	}
	ctx.chat.SessionSetCid(ack.GetSid(), cid)

	/* > 下发ONLINE-ACK消息 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换(主机 - > 网络) */

	ctx.lws.AsyncSend(cid, data)

	ctx.log.Debug("Send online ack success! cid:%d/%d sid:%d", conn.GetCid(), cid, ack.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgKickHandler
 **功    能: KICK消息的处理
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
 **作    者: # Qifeng.zou # 2017.04.29 19:39:30 #
 ******************************************************************************/
func LsndUpMesgKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv kick command!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(0) {
		ctx.log.Error("Kick head is invalid! sid:%d cid:%d seq:%d nid:%d",
			head.GetSid(), head.GetCid(), head.GetSeq(), head.GetNid())
		return -1
	}

	/* > 消息KICK的处理 */
	kick := &mesg.MesgKick{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], kick) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal kick failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.log.Debug("Kick command! code:%d errmsg:%s", kick.GetCode(), kick.GetErrmsg())

	/* > 下发KICK消息 */
	cid := ctx.chat.GetCidBySid(head.GetSid())
	if 0 == cid {
		ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
		return -1
	}

	extra := ctx.chat.SessionGetParam(head.GetSid(), cid)
	if nil == extra {
		ctx.log.Error("Didn't find conn data! sid:%d", head.GetSid())
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.lws.AsyncSend(conn.GetCid(), data)

	/* > 执行KICK指令 */
	ctx.kick_add(conn.GetCid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgSubAckHandler
 **功    能: SUB-ACK消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 10:25:58 #
 ******************************************************************************/
func LsndUpMesgSubAckHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv sub ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of sub-ack is invalid!")
		return -1
	}

	/* > 消息SUB-ACK的处理 */
	ack := &mesg.MesgSubAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal sub-ack failed! errmsg:%s", err.Error())
		return -1
	} else if 0 != ack.GetCode() {
		ctx.log.Error("Sub command [0x%04X] failed! sid:%d code:%d errmsg:%s",
			ack.GetCmd(), head.GetSid(), ack.GetCode(), ack.GetErrmsg())
		return 0
	}

	extra := ctx.chat.SessionGetParam(head.GetSid(), head.GetCid())
	if nil == extra {
		ctx.log.Error("Didn't find conn data! sid:%d", head.GetSid())
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! sid:%d", head.GetSid())
		return -1
	}

	/* > 更新订阅列表 */
	ctx.chat.SubAdd(head.GetSid(), head.GetCid(), ack.GetCmd())

	/* > 下发SUB-ACK消息 */
	ctx.lws.AsyncSend(conn.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
// 聊天室相关操作

/* 聊天室待发消息参数 */
type LsndRoomDataParam struct {
	ctx  *LsndCntx // 全局对象
	data []byte    // 待发数据
}

/******************************************************************************
 **函数名称: LsndRoomSendDataCb
 **功    能: 将聊天室各种消息下发给指定客户端
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **     _param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func LsndRoomSendDataCb(sid uint64, cid uint64, _param interface{}) int {
	p, ok := _param.(*LsndRoomDataParam)
	if !ok {
		return -1
	}

	ctx := p.ctx
	data := p.data

	ctx.log.Debug("Send room data! sid:%d cid:%d", sid, cid)

	/* > 下发ROOM各种消息 */
	ctx.lws.AsyncSend(cid, data)

	return 0
}

/******************************************************************************
 **函数名称: LsndRoomSendSubDataCb
 **功    能: 发送消息各客户端(注: 须订阅)
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **     _param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.20 13:35:23 #
 ******************************************************************************/
func LsndRoomSendSubDataCb(sid uint64, cid uint64, _param interface{}) int {
	p, ok := _param.(*LsndRoomDataParam)
	if !ok {
		return -1
	}

	ctx := p.ctx
	data := p.data

	head := comm.MesgHeadNtoh(data)

	ctx.log.Debug("Send room data! sid:%d cid:%d cmd:0x%04X", sid, cid, head.GetCmd())

	/* > 是否订阅 */
	if !ctx.chat.IsSub(sid, cid, head.GetCmd()) {
		ctx.log.Debug("Didn't sub! sid:%d cid:%d cmd:0x%04X", sid, cid, head.GetCmd())
		return 0
	}

	/* > 下发消息 */
	ctx.lws.AsyncSend(cid, data)

	return 0
}

/******************************************************************************
 **函数名称: LsndUpMesgRoomJoinAckHandler
 **功    能: ROOM-JOIN-ACK消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 22:43:55 #
 ******************************************************************************/
func LsndUpMesgRoomJoinAckHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-join-ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-join-ack is invalid!")
		return -1
	}

	/* > 解析ROOM-JOIN-ACK消息 */
	ack := &mesg.MesgRoomJoinAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-join-ack failed! errmsg:%s", err.Error())
		return -1
	} else if 0 != ack.GetCode() {
		ctx.log.Debug("Join room failed. uid:%d rid:%d gid:%d code:%d errmsg:%s",
			ack.GetUid(), ack.GetRid(), ack.GetGid(), ack.GetCode(), ack.GetErrmsg())
		return 0
	}

	ctx.log.Debug("Room join ack. uid:%d rid:%d gid:%d code:%d errmsg:%s",
		ack.GetUid(), ack.GetRid(), ack.GetGid(), ack.GetCode(), ack.GetErrmsg())

	/* > 获取会话数据 */
	cid := ctx.chat.GetCidBySid(head.GetSid())
	if 0 == cid {
		ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
		return -1
	}

	/* > 加入聊天室 */
	ctx.chat.RoomJoin(ack.GetRid(), ack.GetGid(), head.GetSid(), cid)

	extra := ctx.chat.SessionGetParam(head.GetSid(), cid)
	if nil == extra {
		ctx.log.Error("Didn't find conn data! sid:%d", head.GetSid())
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.log.Debug("Session extra data. sid:%d cid:%d status:%d",
		conn.GetSid(), conn.GetCid(), conn.GetStatus())

	/* > 下发ROOM-JOIN-ACK消息 */
	ctx.lws.AsyncSend(conn.GetCid(), data)

	return 0
}

/******************************************************************************
 **函数名称: LsndUpMesgRoomChatHandler
 **功    能: ROOM-CHAT消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func LsndUpMesgRoomChatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room chat message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-chat is invalid!")
		return -1
	}

	/* > 解析ROOM-CHAT消息 */
	req := &mesg.MesgRoomChat{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-chat failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-CHAT消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(req.GetRid(), req.GetGid(), LsndRoomSendDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomBcHandler
 **功    能: ROOM-BC消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func LsndUpMesgRoomBcHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room broadcast message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Room-broadcast head is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomBc{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room broadcast failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.log.Debug("Recv room broadcast! rid:%d", req.GetRid())

	/* > 遍历下发ROOM-CHAT消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(req.GetRid(), 0, LsndRoomSendDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomKickHandler
 **功    能: ROOM-KICK消息的处理
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
 **作    者: # Qifeng.zou # 2017.05.19 22:18:04 #
 ******************************************************************************/
func LsndUpMesgRoomKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room kick!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Room-kick head is invalid!")
		return -1
	}

	/* > 解析ROOM-KICK消息 */
	req := &mesg.MesgRoomKick{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-kick failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.log.Debug("Room-kick command. uid:%d rid:%d", req.GetUid(), req.GetRid())

	/* > 获取会话数据 */
	cid := ctx.chat.GetCidBySid(head.GetSid())
	if 0 == cid {
		ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
		return -1
	}

	extra := ctx.chat.SessionGetParam(head.GetSid(), cid)
	if nil == extra {
		ctx.log.Error("Didn't find conn data! sid:%d", head.GetSid())
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		ctx.log.Error("Convert conn extra failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.log.Debug("Session extra data. sid:%d cid:%d status:%d",
		conn.GetSid(), conn.GetCid(), conn.GetStatus())

	/* > 执行ROOM-KICK操作 */
	ctx.chat.RoomQuit(req.GetRid(), conn.GetSid(), conn.GetCid())

	/* > 下发ROOM-KICK消息 */
	ctx.lws.AsyncSend(conn.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomUsrNumHandler
 **功    能: ROOM-USR-NUM消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 11:25:28 #
 ******************************************************************************/
func LsndUpMesgRoomUsrNumHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room usr number message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-usr-num is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomUsrNum{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room user number failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-USR-NUM消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(req.GetRid(), 0, LsndRoomSendSubDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomJoinNtcHandler
 **功    能: ROOM-JOIN-NTC消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 11:26:07 #
 ******************************************************************************/
func LsndUpMesgRoomJoinNtcHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room join notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-join-ntc is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomJoinNtc{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-join-ntc failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-JOIN-NTC消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(head.GetSid(), 0, LsndRoomSendDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomQuitNtcHandler
 **功    能: ROOM-QUIT-NTC消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 11:28:34 #
 ******************************************************************************/
func LsndUpMesgRoomQuitNtcHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-quit-ntc is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomQuitNtc{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-quit-ntc failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-QUIT-NTC消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(head.GetSid(), 0, LsndRoomSendDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUpMesgRoomKickNtcHandler
 **功    能: ROOM-KICK-NTC消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.08 11:28:34 #
 ******************************************************************************/
func LsndUpMesgRoomKickNtcHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room kick notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of room-kick-ntc is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomKickNtc{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room-kick-ntc failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-KICK-NTC消息 */
	p := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.TravRoomSession(head.GetSid(), 0, LsndRoomSendDataCb, p)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// 运维消息

/******************************************************************************
 **函数名称: LsndUpMesgLsndInfoAckHandler
 **功    能: LSND-INFO-ACK消息的处理
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
 **作    者: # Qifeng.zou # 2017.03.21 18:05:18 #
 ******************************************************************************/
func LsndUpMesgLsndInfoAckHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv lsnd info ack!")

	return 0
}

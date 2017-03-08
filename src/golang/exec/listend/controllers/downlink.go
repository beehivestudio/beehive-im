package controllers

import (
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
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
	/* > 未知消息 */
	ctx.frwder.Register(comm.CMD_UNKNOWN, LsndDownlinkCommHandler, ctx)

	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndDownlinkOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_SUB_ACK, LsndDownlinkSubAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_UNSUB_ACK, LsndDownlinkUnsubAckHandler, ctx)

	/* > 群聊消息 */
	//ctx.frwder.Register(comm.CMD_GROUP_CHAT, LsndDownlinkGroupChatHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndDownlinkRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndDownlinkRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_USR_NUM, LsndDownlinkRoomUsrNumHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_JOIN_NTC, LsndDownlinkRoomJoinNtcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_QUIT_NTC, LsndDownlinkRoomQuitNtcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_KICK_NTC, LsndDownlinkRoomKickNtcHandler, ctx)
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkCommHandler
 **功    能: 通用消息处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func LsndDownlinkCommHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
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

/******************************************************************************
 **函数名称: LsndDownlinkOnlineAckHandler
 **功    能: ONLINE-ACK消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.07 23:49:03 #
 ******************************************************************************/
func LsndDownlinkOnlineAckHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv online ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Header of online-ack is invalid!")
		return -1
	}

	cid := head.GetCid()

	/* > 消息ONLINE-ACK的处理 */
	ack := &mesg.MesgOnlineAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal online-ack failed! errmsg:%s", err.Error())
		ctx.lws.Kick(cid)
		return -1
	} else if 0 != ack.GetCode() {
		ctx.log.Error("Logon failed! cid:%d sid:%d code:%d errmsg:%s",
			head.GetCid(), ack.GetSid(), ack.GetCode(), ack.GetErrmsg())
		ctx.lws.Kick(cid)
		return 0
	}

	head.SetSid(ack.GetSid())

	/* > 获取&更新会话状态 */
	sp := ctx.chat.SessionGetParam(ack.GetSid())
	if nil == sp {
		ctx.log.Error("Didn't find session data! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		ctx.lws.Kick(cid)
		return -1
	}

	session, ok := sp.(*LsndSessionExtra)
	if !ok {
		ctx.log.Error("Convert session extra failed! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		ctx.lws.Kick(cid)
		return -1
	}

	session.SetStatus(CONN_STATUS_LOGON) /* 已登录 */

	/* > 下发ONLINE-ACK消息 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换(主机 - > 网络) */

	ctx.lws.AsyncSend(cid, data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkSubAckHandler
 **功    能: SUB-ACK消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:25:58 #
 ******************************************************************************/
func LsndDownlinkSubAckHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv sub ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
			ack.GetSub(), head.GetSid(), ack.GetCode(), ack.GetErrmsg())
		return 0
	}

	/* > 更新订阅列表 */
	ctx.chat.SubAdd(head.GetSid(), ack.GetSub())

	/* > 下发SUB-ACK消息 */
	sp := ctx.chat.SessionGetParam(head.GetSid())
	if nil == sp {
		ctx.log.Error("Didn't find session data! sid:%d", head.GetSid())
		return -1
	}

	session, ok := sp.(*LsndSessionExtra)
	if !ok {
		ctx.log.Error("Convert session extra failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.lws.AsyncSend(session.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkUnsubAckHandler
 **功    能: UNSUB-ACK消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 取消指定消息的订阅
 **     2. 转发UNSUB-ACK给对应客户端
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:33:30 #
 ******************************************************************************/
func LsndDownlinkUnsubAckHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv unsub ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Header of unsub-ack is invalid!")
		return -1
	}

	/* > 消息UNSUB-ACK的处理 */
	ack := &mesg.MesgUnsubAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal unsub-ack failed! errmsg:%s", err.Error())
		return -1
	} else if 0 != ack.GetCode() {
		ctx.log.Error("Unsub command [0x%04X] failed! sid:%d code:%d errmsg:%s",
			ack.GetSub(), head.GetSid(), ack.GetCode(), ack.GetErrmsg())
		return 0
	}

	/* > 更新订阅列表 */
	ctx.chat.SubDel(head.GetSid(), ack.GetSub())

	/* > 获取会话数据 */
	sp := ctx.chat.SessionGetParam(head.GetSid())
	if nil == sp {
		ctx.log.Error("Didn't find session data! sid:%d", head.GetSid())
		return -1
	}

	session, ok := sp.(*LsndSessionExtra)
	if !ok {
		ctx.log.Error("Convert session extra failed! sid:%d", head.GetSid())
		return -1
	}

	/* > 下发SUB-ACK消息 */
	ctx.lws.AsyncSend(session.GetCid(), data)

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
 **函数名称: lsnd_room_send_data_cb
 **功    能: 将聊天室各种消息下发给指定客户端
 **输入参数:
 **     sid: 会话SID
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func lsnd_room_send_data_cb(sid uint64, param interface{}) int {
	dp, ok := param.(*LsndRoomDataParam)
	if !ok {
		return -1
	}

	ctx := dp.ctx

	/* > 获取会话数据 */
	sp := ctx.chat.SessionGetParam(sid)
	if nil == sp {
		ctx.log.Error("Didn't find session data! sid:%d", sid)
		return -1
	}

	session, ok := sp.(*LsndSessionExtra)
	if !ok {
		ctx.log.Error("Convert session extra failed! sid:%d", sid)
		return -1
	}

	/* > 下发ROOM各种消息 */
	ctx.lws.AsyncSend(session.GetCid(), dp.data)

	return 0
}

/******************************************************************************
 **函数名称: LsndDownlinkRoomChatHandler
 **功    能: ROOM-CHAT消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func LsndDownlinkRoomChatHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room chat message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(req.GetRid(), req.GetGid(), lsnd_room_send_data_cb, dp)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkRoomBcHandler
 **功    能: ROOM-BC消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 10:38:39 #
 ******************************************************************************/
func LsndDownlinkRoomBcHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room chat message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Header of room-chat is invalid!")
		return -1
	}

	/* > 解析ROOM-BC消息 */
	req := &mesg.MesgRoomBc{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room broadcast failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 遍历下发ROOM-CHAT消息 */
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(req.GetRid(), 0, lsnd_room_send_data_cb, dp)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkRoomUsrNumHandler
 **功    能: ROOM-USR-NUM消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 11:25:28 #
 ******************************************************************************/
func LsndDownlinkRoomUsrNumHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room usr number message!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(req.GetRid(), 0, lsnd_room_send_data_cb, dp)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkRoomJoinNtcHandler
 **功    能: ROOM-JOIN-NTC消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 11:26:07 #
 ******************************************************************************/
func LsndDownlinkRoomJoinNtcHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room join notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(head.GetSid(), 0, lsnd_room_send_data_cb, dp)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkRoomQuitNtcHandler
 **功    能: ROOM-QUIT-NTC消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 11:28:34 #
 ******************************************************************************/
func LsndDownlinkRoomQuitNtcHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(head.GetSid(), 0, lsnd_room_send_data_cb, dp)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkRoomKickNtcHandler
 **功    能: ROOM-KICK-NTC消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 11:28:34 #
 ******************************************************************************/
func LsndDownlinkRoomKickNtcHandler(cmd uint32, dest uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room kick notification!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
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
	dp := &LsndRoomDataParam{ctx: ctx, data: data}

	ctx.chat.Trav(head.GetSid(), 0, lsnd_room_send_data_cb, dp)

	return 0
}

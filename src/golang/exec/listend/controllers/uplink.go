package controllers

import (
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: UplinkRegister
 **功    能: 上行消息回调注册
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 为各消息注册处理函数
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:59:58 #
 ******************************************************************************/
func (ctx *LsndCntx) UplinkRegister() {
	/* 未知消息 */
	ctx.callback.Register(comm.CMD_UNKNOWN, LsndUplinkCommHandler, ctx)

	/* 通用消息 */
	ctx.callback.Register(comm.CMD_ONLINE_REQ, LsndOnlineReqHandler, ctx)   /* 上线消息 */
	ctx.callback.Register(comm.CMD_OFFLINE_REQ, LsndOfflineReqHandler, ctx) /* 下线消息 */
	ctx.callback.Register(comm.CMD_PING, LsndPingHandler, ctx)              /* 心跳请求 */
	ctx.callback.Register(comm.CMD_UNSUB_REQ, LsndUnsubReqHandler, ctx)     /* 取消订阅 */

	/* 聊天室消息 */
	ctx.callback.Register(comm.CMD_ROOM_QUIT_REQ, LsndRoomQuitReqHandler, ctx)
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUplinkCommHandler
 **功    能: 消息通用处理 - 直接将消息转发给上游模块
 **输入参数:
 **     session: 连接会话
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 直接将消息转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 22:49:17 #
 ******************************************************************************/
func LsndUplinkCommHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	ctx.log.Debug("Recv cmd [0x%04X] request! sid:%d cid:%d",
		head.GetCmd(), head.GetSid(), session.GetCid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndOnlineReqHandler
 **功    能: ONLINE消息的处理
 **输入参数:
 **     session: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将ONLINE请求转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 23:10:58 #
 ******************************************************************************/
func LsndOnlineReqHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	/* > 验证当前状态 */
	if !session.IsStatus(CONN_STATUS_READY) &&
		!session.IsStatus(CONN_STATUS_CHECK) {
		ctx.log.Error("Drop online request! cid:%d sid:%d status:%d",
			session.GetCid(), session.GetSid(), session.GetStatus())
		return 0
	}

	/* > "网络->主机"字节序 */
	head := comm.MesgHeadNtoh(data)

	session.SetSid(head.GetSid())
	ctx.chat.SessionSetParam(head.GetSid(), session)

	head.SetSid(session.GetCid())
	head.SetNid(ctx.conf.GetNid())

	ctx.log.Debug("Recv online request! cmd:0x%04X sid:%d cid:%d",
		head.GetCmd(), session.GetSid(), session.GetCid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* "主机->网络"字节序 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游模块 */

	/* > 更新连接状态 */
	session.SetStatus(CONN_STATUS_CHECK)
	ctx.add_sid_to_cid(head.GetSid(), session.GetCid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndOfflineReqHandler
 **功    能: OFFLINE消息的处理
 **输入参数:
 **     session: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将OFFLINE请求转发给上游模块, 再将指定连接踢下线.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 21:39:22 #
 ******************************************************************************/
func LsndOfflineReqHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv offline request! sid:%d cid:%d", session.GetSid(), session.GetCid())

	/* > "网络->主机"字节序 */
	head := comm.MesgHeadNtoh(data)

	session.SetSid(head.GetSid())

	head.SetSid(session.GetCid())
	head.SetNid(ctx.conf.GetNid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* "主机->网络"字节序 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游模块 */

	/* > 更新会话状态 */
	session.SetStatus(CONN_STATUS_LOGOUT)

	/* > 将某连接踢下线 */
	ctx.lws.Kick(session.GetCid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndPingHandler
 **功    能: PING处理
 **输入参数:
 **     session: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 转发给上游模块, 并回复PONG应答.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 23:40:55 #
 ******************************************************************************/
func LsndPingHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv ping! sid:%d cid:%d", session.GetSid(), session.GetCid())

	/* > "网络->主机"字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* "主机->网络"字节序 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游模块 */

	/* > 回复PONG应答 */
	head.SetCmd(comm.CMD_PONG)

	data = make([]byte, comm.MESG_HEAD_SIZE)

	p = &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* "主机->网络"字节序 */

	ctx.lws.AsyncSend(session.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUnsubReqHandler
 **功    能: 取消订阅处理
 **输入参数:
 **     session: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 移除订阅消息, 再转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 22:58:12 #
 ******************************************************************************/
func LsndUnsubReqHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv unsub request! sid:%d cid:%d", session.GetSid(), session.GetCid())

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	/* > 移除订阅消息 */
	req := &mesg.MesgUnsubReq{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal unsub request failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.chat.SubDel(head.GetSid(), req.GetSub()) /* 移除订阅消息 */

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游 */

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndRoomQuitReqHandler
 **功    能: 退出聊天室的处理
 **输入参数:
 **     session: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 退出聊天室, 再转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 21:55:26 #
 ******************************************************************************/
func LsndRoomQuitReqHandler(session *LsndSessionExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit request! sid:%d cid:%d", session.GetSid(), session.GetCid())

	/* > 字节序转换(网络->主机) */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	/* > 退出指定聊天室 */
	req := &mesg.MesgRoomQuit{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal room quit request failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.chat.RoomQuit(req.GetRid(), session.GetSid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游 */

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

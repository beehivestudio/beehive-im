package controllers

import (
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: MesgRegister
 **功    能: 上行消息回调注册
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 为"上行"消息注册处理函数
 **注意事项: "上行"消息指的是从客户端发送过来的消息
 **作    者: # Qifeng.zou # 2017.03.06 17:59:58 #
 ******************************************************************************/
func (ctx *LsndCntx) MesgRegister() {
	/* 未知消息 */
	ctx.callback.Register(comm.CMD_UNKNOWN, LsndMesgCommHandler, ctx) /* 默认处理 */

	/* 通用消息 */
	ctx.callback.Register(comm.CMD_ONLINE, LsndMesgOnlineHandler, ctx)   /* 上线消息 */
	ctx.callback.Register(comm.CMD_OFFLINE, LsndMesgOfflineHandler, ctx) /* 下线消息 */
	ctx.callback.Register(comm.CMD_PING, LsndMesgPingHandler, ctx)       /* 心跳请求 */
	ctx.callback.Register(comm.CMD_UNSUB, LsndMesgUnsubHandler, ctx)     /* 取消订阅 */

	/* 聊天室消息 */
	ctx.callback.Register(comm.CMD_ROOM_QUIT, LsndMesgRoomQuitHandler, ctx) /* 退出聊天室 */
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndMesgCommHandler
 **功    能: 消息通用处理 - 直接将消息转发给上游模块
 **输入参数:
 **     conn: 连接会话
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
func LsndMesgCommHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	ctx.log.Debug("Recv cmd [0x%04X] request! sid:%d cid:%d",
		head.GetCmd(), head.GetSid(), conn.GetCid())

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
 **函数名称: LsndMesgOnlineHandler
 **功    能: ONLINE消息的处理
 **输入参数:
 **     conn: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将ONLINE请求转发给上游模块
 **注意事项: 无需验证消息序列号
 **作    者: # Qifeng.zou # 2017.03.04 23:10:58 #
 ******************************************************************************/
func LsndMesgOnlineHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	/* > 验证当前状态 */
	if !conn.IsStatus(CONN_STATUS_READY) &&
		!conn.IsStatus(CONN_STATUS_CHECK) {
		ctx.log.Error("Drop online request! cid:%d sid:%d status:%d",
			conn.GetCid(), conn.GetSid(), conn.GetStatus())
		return 0
	}

	/* > "网络->主机"字节序 */
	head := comm.MesgHeadNtoh(data)

	conn.SetSid(head.GetSid())
	ctx.chat.SessionSetParam(head.GetSid(), conn)

	head.SetSid(conn.GetCid())
	head.SetNid(ctx.conf.GetNid())

	ctx.log.Debug("Recv online request! cmd:0x%04X sid:%d cid:%d",
		head.GetCmd(), conn.GetSid(), conn.GetCid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* "主机->网络"字节序 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游模块 */

	/* > 更新连接状态 */
	conn.SetStatus(CONN_STATUS_CHECK)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndMesgOfflineHandler
 **功    能: OFFLINE消息的处理
 **输入参数:
 **     conn: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 再将指定连接踢下线.
 **注意事项: 执行关闭操作时"再将"OFFLINE请求转发给上游模块
 **作    者: # Qifeng.zou # 2017.03.06 21:39:22 #
 ******************************************************************************/
func LsndMesgOfflineHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv offline request! sid:%d cid:%d", conn.GetSid(), conn.GetCid())

	/* > 更新会话状态 */
	conn.SetStatus(CONN_STATUS_LOGOUT)

	/* > 将某连接踢下线 */
	ctx.lws.Kick(conn.GetCid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndMesgPingHandler
 **功    能: PING处理
 **输入参数:
 **     conn: 会话连接
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 转发给上游模块, 并回复PONG应答.
 **注意事项: 无需验证消息序列号
 **作    者: # Qifeng.zou # 2017.03.04 23:40:55 #
 ******************************************************************************/
func LsndMesgPingHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv ping! sid:%d cid:%d", conn.GetSid(), conn.GetCid())

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

	ctx.lws.AsyncSend(conn.GetCid(), data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndMesgUnsubHandler
 **功    能: 取消订阅处理
 **输入参数:
 **     conn: 会话连接
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
func LsndMesgUnsubHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv unsub request! sid:%d cid:%d", conn.GetSid(), conn.GetCid())

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.conf.GetNid())

	/* > 移除订阅消息 */
	req := &mesg.MesgUnsub{}

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
 **函数名称: LsndMesgRoomQuitHandler
 **功    能: 退出聊天室的处理
 **输入参数:
 **     conn: 会话连接
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
func LsndMesgRoomQuitHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit request! sid:%d cid:%d", conn.GetSid(), conn.GetCid())

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

	ctx.chat.RoomQuit(req.GetRid(), conn.GetSid())

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游 */

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

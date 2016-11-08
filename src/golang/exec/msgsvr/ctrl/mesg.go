package ctrl

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrGroupMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg request!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrGroupMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: prvt_msg_parse
 **功    能: 解析私聊消息
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 13:23:54 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) prvt_msg_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgPrvtMsg) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgPrvtMsg{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal prvt-msg failed! errmsg:%s", err.Error())
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_prvt_msg_ack
 **功    能: 发送ROOM-MSG应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **     errno: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:52:14 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_err_prvt_msg_ack(head *comm.MesgHeader,
	req *mesg.MesgRoomMsg, errno uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgRoomAck{
		ErrNum: proto.Uint32(errno),
		ErrMsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+length)

	head.Cmd = comm.CMD_ROOM_MSG_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.AsyncSend(comm.CMD_ROOM_MSG_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: send_prvt_msg_ack
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_prvt_msg_ack(head *comm.MesgHeader, req *mesg.MesgRoomMsg) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgRoomAck{
		ErrNum: proto.Uint32(0),
		ErrMsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+length)

	head.Cmd = comm.CMD_ROOM_MSG_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.AsyncSend(comm.CMD_ROOM_MSG_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: prvt_msg_handler
 **功    能: ROOM-MSG处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-MSG请求
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:34:55 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) prvt_msg_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomMsg) (err error) {
	ctx.rid_to_nid_map.RLock()
	nid_list, ok := ctx.rid_to_nid_map.m[req.GetRid()]
	if false == ok {
		ctx.rid_to_nid_map.RUnlock()
		return nil
	}
	for nid := range nid_list {
		ctx.log.Debug("rid:%d nid:%d", req.GetRid(), nid)
	}
	ctx.rid_to_nid_map.RUnlock()
	return err
}

/******************************************************************************
 **函数名称: MsgSvrPrvtMsgHandler
 **功    能: 私聊消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 首先将私聊消息放入接收方离线队列.
 **     2. 如果接收方当前在线, 则直接下发私聊消息;
 **        如果接收方当前"不"在线, 则"不"下发私聊消息.
 **     3. 给发送方下发私聊应答消息.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 13:05:26 #
 ******************************************************************************/
func MsgSvrPrvtMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv private msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrPrvtMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrBcMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrBcMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrP2pMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrP2pMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: room_msg_parse
 **功    能: 解析ROOM-MSG
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:29:23 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) room_msg_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomMsg) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgRoomMsg{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-msg failed! errmsg:%s", err.Error())
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_room_msg_ack
 **功    能: 发送ROOM-MSG应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **     errno: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:52:14 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_err_room_msg_ack(head *comm.MesgHeader,
	req *mesg.MesgRoomMsg, errno uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgRoomAck{
		ErrNum: proto.Uint32(errno),
		ErrMsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+length)

	head.Cmd = comm.CMD_ROOM_MSG_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.AsyncSend(comm.CMD_ROOM_MSG_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: send_room_msg_ack
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_room_msg_ack(head *comm.MesgHeader, req *mesg.MesgRoomMsg) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgRoomAck{
		ErrNum: proto.Uint32(0),
		ErrMsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+length)

	head.Cmd = comm.CMD_ROOM_MSG_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.AsyncSend(comm.CMD_ROOM_MSG_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_msg_handler
 **功    能: ROOM-MSG处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-MSG请求
 **输出参数: NONE
 **返    回:
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 22:34:55 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) room_msg_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomMsg) (err error) {
	ctx.rid_to_nid_map.RLock()
	nid_list, ok := ctx.rid_to_nid_map.m[req.GetRid()]
	if false == ok {
		ctx.rid_to_nid_map.RUnlock()
		return nil
	}
	for nid := range nid_list {
		ctx.log.Debug("rid:%d nid:%d", req.GetRid(), nid)
	}
	ctx.rid_to_nid_map.RUnlock()
	return err
}

/******************************************************************************
 **函数名称: MsgSvrRoomMsgHandler
 **功    能: 聊天室消息的处理
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
 **作    者: # Qifeng.zou # 2016.11.04 22:28:02 #
 ******************************************************************************/
func MsgSvrRoomMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv room message!")

	/* > 解析ROOM-MSG协议 */
	head, req := ctx.room_msg_parse(data)
	if nil == head || nil == req {
		ctx.log.Error("Parse room-msg failed!")
		return -1
	}

	/* > 进行业务处理 */
	err := ctx.room_msg_handler(head, req)
	if nil != err {
		ctx.log.Error("Parse room-msg failed!")
		ctx.send_err_room_msg_ack(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	ctx.send_room_msg_ack(head, req)
	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrRoomMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrRoomBcMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrRoomBcMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrSyncMsgHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func MsgSvrSyncMsgAckHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

package controllers

import (
	_ "fmt"
	_ "strconv"

	_ "github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: group_msg_parse
 **功    能: 解析GROUP-MSG
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
func (ctx *MsgSvrCntx) group_msg_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgGroupMsg) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgGroupMsg{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal group-msg failed! errmsg:%s", err.Error())
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_group_msg_ack
 **功    能: 发送GROUP-MSG应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
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
 **作    者: # Qifeng.zou # 2016.12.17 13:44:00 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_err_group_msg_ack(head *comm.MesgHeader,
	req *mesg.MesgGroupMsg, code uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgGroupAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_GROUP_MSG_ACK, head.GetSid(),
		head.GetNid(), head.GetSerial(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: send_group_msg_ack
 **功    能: 发送上线应答
 **输入参数:
 **     head: 协议头
 **     req: 协议体
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.17 13:44:49 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_group_msg_ack(head *comm.MesgHeader, req *mesg.MesgGroupMsg) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgGroupAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_GROUP_MSG_ACK, head.GetSid(),
		head.GetNid(), head.GetSerial(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: group_msg_handler
 **功    能: GROUP-MSG处理
 **输入参数:
 **     head: 协议头
 **     req: GROUP-MSG请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **     1. 将消息存放在聊天室历史消息表中
 **     2. 遍历rid->nid列表, 并转发聊天室消息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.17 13:48:00 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) group_msg_handler(
	head *comm.MesgHeader, req *mesg.MesgGroupMsg, data []byte) (err error) {
	ctx.gid_to_nid_map.RLock()
	nid_list, ok := ctx.gid_to_nid_map.m[req.GetGid()]
	if false == ok {
		ctx.gid_to_nid_map.RUnlock()
		return nil
	}

	/* > 遍历rid->nid列表, 并下发聊天室消息 */
	for nid := range nid_list {
		ctx.log.Debug("gid:%d nid:%d", req.GetGid(), nid)

		ctx.send_data(comm.CMD_GROUP_MSG, req.GetGid(), uint32(nid),
			head.GetSerial(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}
	ctx.gid_to_nid_map.RUnlock()
	return err
}

/******************************************************************************
 **函数名称: MsgSvrGroupMsgHandler
 **功    能: 群消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断群消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将群消息放入群历史消息
 **     3. 将群消息发送给在线的人员.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 08:45:19 #
 ******************************************************************************/
func MsgSvrGroupMsgHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group message!")

	/* > 解析ROOM-MSG协议 */
	head, req := ctx.group_msg_parse(data)
	if nil == head || nil == req {
		ctx.log.Error("Parse group-msg failed!")
		return -1
	}

	/* > 进行业务处理 */
	err := ctx.group_msg_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Parse group-msg failed!")
		ctx.send_err_group_msg_ack(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	ctx.send_group_msg_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrGroupMsgAckHandler
 **功    能: 群消息应答处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 无需处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:43:01 #
 ******************************************************************************/
func MsgSvrGroupMsgAckHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

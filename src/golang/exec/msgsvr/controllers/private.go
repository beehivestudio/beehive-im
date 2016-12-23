package controllers

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: private_msg_parse
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
func (ctx *MsgSvrCntx) private_msg_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgPrvtMsg) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgPrvtMsg{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal prvt-msg failed! errmsg:%s", err.Error())
		return nil, nil
	} else if 0 == req.GetOrig() || 0 == req.GetDest() {
		ctx.log.Error("Paramter isn't right! orig:%d dest:%d", req.GetOrig(), req.GetDest())
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_prvt_msg_ack
 **功    能: 发送PRVT-MSG应答(异常)
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
 **作    者: # Qifeng.zou # 2016.11.04 22:52:14 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_err_prvt_msg_ack(head *comm.MesgHeader,
	req *mesg.MesgPrvtMsg, code uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgPrvtAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_PRVT_MSG_ACK, head.GetSid(), head.GetNid(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: send_prvt_msg_ack
 **功    能: 发送私聊应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1; // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_prvt_msg_ack(head *comm.MesgHeader, req *mesg.MesgPrvtMsg) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgPrvtAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_PRVT_MSG_ACK, head.GetSid(), head.GetNid(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: private_msg_handler
 **功    能: PRVT-MSG处理
 **输入参数:
 **     head: 协议头
 **     req: PRVT-MSG请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回:
 **实现描述:
 **     1. 将消息放入UID离线队列
 **     3. 判断接收方是否在线.
 **        > 如果在线, 则直接下发消息
 **        > 如果不在线, 则无需下发消息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.18 20:33:18 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) private_msg_handler(
	head *comm.MesgHeader, req *mesg.MesgPrvtMsg, data []byte) (err error) {
	var key string

	rds := ctx.redis.Get()
	defer rds.Close()

	/* 1. 将消息放入离线队列 */
	key = fmt.Sprintf(comm.CHAT_KEY_USR_OFFLINE_QUEUE, req.GetDest())
	num, err := redis.Int(rds.Do("RPUSH", key, data[comm.MESG_HEAD_SIZE:]))
	if nil != err {
		ctx.log.Error("Push data into offline queue failed! uid:%d", req.GetDest())
		return err
	}

	/* 2. 发送给"发送方"的其他终端.
	   > 如果在线, 则直接下发消息
	   > 如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetOrig())

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetOrig())
		return err
	}

	num = len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)
		if uint64(sid) == head.GetSid() {
			continue
		}

		attr := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)
		vals, _ := redis.Strings(rds.Do("HGET", attr, "UID", "NID"))

		uid, _ := strconv.ParseInt(vals[0], 10, 64)
		nid, _ := strconv.ParseInt(vals[1], 10, 32)
		if uint64(uid) != req.GetOrig() || 0 == nid {
			continue
		}

		ctx.send_data(comm.CMD_PRVT_MSG, uint64(sid),
			uint32(nid), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	/* 3. 发送给"接收方"所有终端.
	   > 如果在线, 则直接下发消息
	   > 如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetDest())

	sid_list, err = redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetDest())
		return err
	}

	num = len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		attr := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)
		vals, _ := redis.Strings(rds.Do("HGET", attr, "UID", "NID"))

		uid, _ := strconv.ParseInt(vals[0], 10, 64)
		nid, _ := strconv.ParseInt(vals[1], 10, 32)
		if uint64(uid) != req.GetDest() || 0 == nid {
			continue
		}

		ctx.send_data(comm.CMD_PRVT_MSG, uint64(sid),
			uint32(nid), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	return err
}

/******************************************************************************
 **函数名称: MsgSvrPrivateMsgHandler
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
func MsgSvrPrivateMsgHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv private msg ack!")

	/* > 解析ROOM-MSG协议 */
	head, req := ctx.private_msg_parse(data)
	if nil == head || nil == req {
		ctx.log.Error("Parse private message failed!")
		return -1
	}

	/* > 进行业务处理 */
	err := ctx.private_msg_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Parse private message failed!")
		ctx.send_err_prvt_msg_ack(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	ctx.send_prvt_msg_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrPrvtMsgAckHandler
 **功    能: 私聊消息应答处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 收到私有消息的应答后, 删除离线队列中的对应数据.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:45:08 #
 ******************************************************************************/
func MsgSvrPrvtMsgAckHandler(cmd uint32, orig uint32,
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

package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

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
func (ctx *MsgSvrCntx) room_chat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomChat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Message header of room-chat is invalid!")
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header is invalid!")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomChat{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-msg failed! errmsg:%s", err.Error())
		return nil, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Body is invalid!")
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: send_err_room_chat_ack
 **功    能: 发送ROOM-CHAT应答(异常)
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
func (ctx *MsgSvrCntx) send_err_room_chat_ack(head *comm.MesgHeader,
	req *mesg.MesgRoomChat, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_ROOM_CHAT_ACK, head.GetSid(),
		head.GetNid(), head.GetSerial(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: send_room_chat_ack
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
func (ctx *MsgSvrCntx) send_room_chat_ack(head *comm.MesgHeader, req *mesg.MesgRoomChat) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
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
		head.GetNid(), head.GetSerial(), body, uint32(len(body)))
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
func (ctx *MsgSvrCntx) room_chat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomChat, data []byte) (err error) {
	item := &MesgRoomItem{}

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
		return nil
	}

	/* > 遍历rid->nid列表, 并下发聊天室消息 */
	for nid := range nid_list {
		ctx.log.Debug("rid:%d nid:%d", req.GetRid(), nid)

		ctx.send_data(comm.CMD_ROOM_CHAT, req.GetRid(), uint32(nid),
			head.GetSerial(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}
	return err
}

/******************************************************************************
 **函数名称: MsgSvrRoomChatHandler
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
func MsgSvrRoomChatHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-chat message!")

	/* > 解析ROOM-CHAT协议 */
	head, req, code, err := ctx.room_chat_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-msg failed! code:%d errmsg:%s", code, err.Error())
		return -1
	}

	/* > 进行业务处理 */
	err = ctx.room_chat_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle room message failed!")
		ctx.send_err_room_chat_ack(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	return ctx.send_room_chat_ack(head, req)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrRoomChatAckHandler
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
func MsgSvrRoomChatAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrRoomBcHandler
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
func MsgSvrRoomBcHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrRoomBcAckHandler
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
func MsgSvrRoomBcAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: room_mesg_storage_task
 **功    能: 聊天室消息的存储任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从聊天室消息队列中取出消息, 并进行存储处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 23:43:03 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) room_mesg_storage_task() {
	for item := range ctx.room_mesg_chan {
		ctx.room_mesg_storage_proc(item.head, item.req, item.raw)
	}
}

/******************************************************************************
 **函数名称: room_mesg_storage_proc
 **功    能: 聊天室消息的存储处理
 **输入参数:
 **     head: 消息头
 **     req: 请求内容
 **     raw: 原始数据
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 将消息存入聊天室缓存和数据库
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:05:51 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) room_mesg_storage_proc(
	head *comm.MesgHeader, req *mesg.MesgRoomChat, raw []byte) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_MESG_QUEUE, req.GetRid())
	pl.Send("LPUSH", key, raw[comm.MESG_HEAD_SIZE:])
}

/******************************************************************************
 **函数名称: room_mesg_queue_clean_task
 **功    能: 清理聊天室缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持聊天室缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:34:18 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) room_mesg_queue_clean_task() {
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
func (ctx *MsgSvrCntx) room_mesg_queue_clean() {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_RID_ZSET, 0, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get rid list failed! errmsg:%s", err.Error())
			continue
		}

		num := len(rid_list)
		for idx := 0; idx < num; idx += 1 {
			/* 保持聊天室缓存消息为最新的100条 */
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			key := fmt.Sprintf(comm.CHAT_KEY_ROOM_MESG_QUEUE, uint64(rid))

			rds.Do("LTRIM", key, 0, 99)
		}
	}
}

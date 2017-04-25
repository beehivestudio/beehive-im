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

/******************************************************************************
 **函数名称: chat_parse
 **功    能: 解析私聊消息
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
 **作    者: # Qifeng.zou # 2016.11.05 13:23:54 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) chat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgChat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header of chat is invalid")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgChat{}
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
 **函数名称: chat_failed
 **功    能: 发送CHAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: CHAT请求
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
func (ctx *MsgSvrCntx) chat_failed(head *comm.MesgHeader,
	req *mesg.MesgChat, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgChatAck{
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
	body, err = proto.Marshal(ack)
	return ctx.send_data(comm.CMD_CHAT_ACK, head.GetSid(),
		head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: chat_ack
 **功    能: 发送私聊应答
 **输入参数:
 **     head: 协议头
 **     req: CHAT请求
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
func (ctx *MsgSvrCntx) chat_ack(head *comm.MesgHeader, req *mesg.MesgChat) int {
	/* > 设置协议体 */
	ack := &mesg.MesgChatAck{
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
	return ctx.send_data(comm.CMD_CHAT_ACK, head.GetSid(),
		head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: chat_handler
 **功    能: CHAT处理
 **输入参数:
 **     head: 协议头
 **     req: CHAT请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误码+错误描述
 **实现描述:
 **     1. 将消息放入UID离线队列
 **	    2. 发送给"发送方"的其他终端.
 **        > 如果在线, 则直接下发消息
 **        > 如果不在线, 则无需下发消息
 **     3. 判断接收方是否在线.
 **        > 如果在线, 则直接下发消息
 **        > 如果不在线, 则无需下发消息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.18 20:33:18 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) chat_handler(head *comm.MesgHeader,
	req *mesg.MesgChat, data []byte) (code uint32, err error) {
	var key string

	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 将消息放入离线队列 */
	item := &MesgChatItem{
		head: head,
		req:  req,
		raw:  data,
	}

	ctx.chat_chan <- item

	/* > 发送给"发送方"的其他终端.
	   1.如果在线, 则直接下发消息
	   2.如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetOrig())

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetOrig())
		return comm.ERR_SYS_DB, err
	}

	num := len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)
		if uint64(sid) == head.GetSid() {
			continue
		}

		attr, _ := im.GetSidAttr(ctx.redis, uint64(sid))
		if nil == attr {
			continue
		} else if 0 == attr.GetNid() {
			continue
		} else if uint64(attr.GetUid()) != req.GetOrig() {
			continue
		}

		ctx.send_data(comm.CMD_CHAT, uint64(sid), uint32(attr.GetNid()),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	/* > 发送给"接收方"所有终端.
		   1.如果在线, 则直接下发消息
	       2.如果不在线, 则无需下发消息 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetDest())

	sid_list, err = redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid set by uid [%d] failed!", req.GetDest())
		return comm.ERR_SYS_DB, err
	}

	num = len(sid_list)
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		attr, _ := im.GetSidAttr(ctx.redis, uint64(sid))
		if nil == attr {
			continue
		} else if 0 == attr.GetNid() {
			continue
		} else if uint64(attr.GetUid()) != req.GetOrig() {
			continue
		} else if uint64(attr.GetUid()) != req.GetDest() {
			continue
		}

		ctx.send_data(comm.CMD_CHAT, uint64(sid), uint32(attr.GetNid()),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: MsgSvrChatHandler
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
func MsgSvrChatHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv chat message!")

	/* > 解析CHAT协议 */
	head, req, code, err := ctx.chat_parse(data)
	if nil == head {
		ctx.log.Error("Parse chat failed! errmsg:%s", err.Error())
		return -1
	} else if nil == req {
		ctx.log.Error("Parse chat failed! errmsg:%s", err.Error())
		ctx.chat_failed(head, req, code, err.Error())
		return -1
	}

	/* > 进行业务处理 */
	code, err = ctx.chat_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle chat failed! errmsg:%s", err.Error())
		ctx.chat_failed(head, req, code, err.Error())
		return -1
	}

	ctx.chat_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_ack_parse
 **功    能: 解析私聊应答消息
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
 **作    者: # Qifeng.zou # 2016.12.26 20:38:42 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) chat_ack_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgChatAck, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header of chat is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header of chat-ack invalid!")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgChatAck{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal chat-ack failed! errmsg:%s", err.Error())
		return nil, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: chat_ack_handler
 **功    能: CHAT-ACK处理
 **输入参数:
 **     head: 协议头
 **     req: CHAT-ACK请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误码+错误信息
 **实现描述: 清理离线消息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.26 21:01:12 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) chat_ack_handler(
	head *comm.MesgHeader, req *mesg.MesgChatAck, data []byte) (code uint32, err error) {
	rds := ctx.redis.Get()
	defer func() {
		rds.Do("")
		rds.Close()
	}()

	if 0 != req.GetCode() {
		ctx.log.Error("code:%d errmsg:%s", req.GetCode(), req.GetErrmsg())
		return req.GetCode(), errors.New(req.GetErrmsg())
	}

	/* 清理离线消息 */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_OFFLINE_ZSET, req.GetDest())
	field := fmt.Sprintf(comm.CHAT_FMT_UID_MSGID_STR, req.GetOrig(), head.GetSeq())
	rds.Send("ZREM", key, field)

	key = fmt.Sprintf(comm.CHAT_KEY_USR_SEND_MESG_HTAB, req.GetOrig())
	rds.Send("HDEL", key, head.GetSeq())

	return 0, nil
}

/******************************************************************************
 **函数名称: MsgSvrChatAckHandler
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
func MsgSvrChatAckHandler(cmd uint32, orig uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	/* > 解析CHAT-ACK协议 */
	head, req, code, err := ctx.chat_ack_parse(data)
	if nil != err {
		ctx.log.Error("Parse private message ack failed! code:%d errmsg:%s",
			code, err.Error())
		return -1
	}

	/* > 进行业务处理 */
	code, err = ctx.chat_ack_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle chat ack failed! code:%d errmsg:%s", code, err.Error())
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// 定时任务

/******************************************************************************
 **函数名称: task_chat_chan_pop
 **功    能: 私聊消息的存储任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从私聊消息队列中取出消息, 并进行存储处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 11:03:42 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) task_chat_chan_pop() {
	for item := range ctx.chat_chan {
		item.storage(ctx)
	}
}

/******************************************************************************
 **函数名称: storage
 **功    能: 私聊消息的存储处理
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 将私聊消息存入缓存和数据库
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 11:03:42 #
 ******************************************************************************/
func (item *MesgChatItem) storage(ctx *MsgSvrCntx) {
	var key string

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ctm := time.Now().Unix()

	/* > 加入接收者离线列表 */
	key = fmt.Sprintf(comm.CHAT_KEY_USR_OFFLINE_ZSET, item.req.GetDest())
	member := fmt.Sprintf(comm.CHAT_FMT_UID_MSGID_STR, item.req.GetOrig(), item.head.GetSeq())
	pl.Send("ZADD", key, member, ctm)

	/* > 存储发送者离线消息 */
	key = fmt.Sprintf(comm.CHAT_KEY_USR_SEND_MESG_HTAB, item.req.GetOrig())
	pl.Send("HSETNX", key, item.raw)
}

////////////////////////////////////////////////////////////////////////////////

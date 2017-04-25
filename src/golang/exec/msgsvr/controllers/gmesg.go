package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
// 群组消息的发送采用写扩散的机制

////////////////////////////////////////////////////////////////////////////////
// 群组消息

/******************************************************************************
 **函数名称: group_chat_parse
 **功    能: 解析GROUP-CHAT
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
func (ctx *MsgSvrCntx) group_chat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgGroupChat) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil
	}

	/* > 解析PB协议 */
	req = &mesg.MesgGroupChat{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal group-chat failed! errmsg:%s", err.Error())
		return head, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: group_chat_failed
 **功    能: 发送GROUP-CHAT应答(异常)
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
func (ctx *MsgSvrCntx) group_chat_failed(head *comm.MesgHeader,
	req *mesg.MesgGroupChat, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgGroupChatAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_GROUP_CHAT_ACK, head.GetSid(),
		head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: group_chat_ack
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
func (ctx *MsgSvrCntx) group_chat_ack(head *comm.MesgHeader, req *mesg.MesgGroupChat) int {
	/* > 设置协议体 */
	ack := &mesg.MesgGroupChatAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_GROUP_CHAT_ACK, head.GetSid(),
		head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: group_chat_handler
 **功    能: GROUP-MSG处理
 **输入参数:
 **     head: 协议头
 **     req: GROUP-MSG请求
 **     data: 原始数据
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **     1. 将消息存放在聊天室历史消息表中
 **     2. 遍历rid->nid列表, 并转发群聊消息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.17 13:48:00 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) group_chat_handler(
	head *comm.MesgHeader, req *mesg.MesgGroupChat, data []byte) (err error) {
	/* > 放入存储队列 */
	item := &MesgGroupItem{
		head: head,
		req:  req,
		raw:  data,
	}

	ctx.group_mesg_chan <- item

	/* > 下发群聊消息 */
	ctx.group.node.RLock()
	defer ctx.group.node.RUnlock()

	nid_list, ok := ctx.group.node.m[req.GetGid()]
	if !ok {
		return nil
	}

	/* > 遍历rid->nid列表, 并下发聊天室消息 */
	for nid := range nid_list {
		ctx.log.Debug("gid:%d nid:%d", req.GetGid(), nid)

		ctx.send_data(comm.CMD_GROUP_CHAT, req.GetGid(), uint32(nid),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}
	return err
}

/******************************************************************************
 **函数名称: MsgSvrGroupChatHandler
 **功    能: 群消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断群消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将群消息放入群历史消息, 再放入所有组员的消息队列中.
 **     3. 将群消息发送给在线的人员.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 08:45:19 #
 ******************************************************************************/
func MsgSvrGroupChatHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MsgSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group chat message!")

	/* > 解析ROOM-MSG协议 */
	head, req := ctx.group_chat_parse(data)
	if nil == head {
		ctx.log.Error("Parse header of group chat failed!")
		return -1
	} else if nil == req {
		ctx.log.Error("Parse body of group chat failed!")
		ctx.group_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, "Body is invalid!")
		return -1
	}

	/* > 进行业务处理 */
	err := ctx.group_chat_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Parse group-msg failed!")
		ctx.group_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	ctx.group_chat_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: MsgSvrGroupChatAckHandler
 **功    能: 群消息应答处理
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 无需处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:43:01 #
 ******************************************************************************/
func MsgSvrGroupChatAckHandler(cmd uint32, nid uint32,
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
 **函数名称: task_group_mesg_chan_pop
 **功    能: 群聊消息的存储任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从群聊消息队列中取出消息, 并进行存储处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 23:45:01 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) task_group_mesg_chan_pop() {
	for item := range ctx.group_mesg_chan {
		item.storage(ctx)
	}
}

/******************************************************************************
 **函数名称: storage
 **功    能: 群聊消息的存储处理
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 将消息存入聊天室缓存和数据库
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:05:51 #
 ******************************************************************************/
func (item *MesgGroupItem) storage(ctx *MsgSvrCntx) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_MESG_QUEUE, item.req.GetGid())
	pl.Send("LPUSH", key, item.raw[comm.MESG_HEAD_SIZE:])
}

/******************************************************************************
 **函数名称: task_group_mesg_queue_clean
 **功    能: 清理聊天室缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持群聊缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 08:38:44 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) task_group_mesg_queue_clean() {
	for {
		ctx.group_mesg_queue_clean()

		time.Sleep(30 * time.Second)
	}
}

/******************************************************************************
 **函数名称: group_mesg_queue_clean
 **功    能: 清理群聊缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持群聊缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.29 08:38:38 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) group_mesg_queue_clean() {
	rds := ctx.redis.Get()
	defer rds.Close()

	off := 0
	for {
		gid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			comm.CHAT_KEY_GID_ZSET, 0, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get gid list failed! errmsg:%s", err.Error())
			continue
		}

		num := len(gid_list)
		for idx := 0; idx < num; idx += 1 {
			/* 保持聊天室缓存消息为最新的100条 */
			gid, _ := strconv.ParseInt(gid_list[idx], 10, 64)
			key := fmt.Sprintf(comm.CHAT_KEY_GROUP_MESG_QUEUE, uint64(gid))

			rds.Do("LTRIM", key, 0, 99)
		}
	}
}

package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"labix.org/v2/mgo"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/crypt"
	"beehive-im/src/golang/lib/mesg"
	"beehive-im/src/golang/lib/mesg/seqsvr"

	"beehive-im/src/golang/exec/chatroom/models"
)

// 聊天室

// 通用请求

////////////////////////////////////////////////////////////////////////////////
// 上线请求

type OnlineToken struct {
	uid uint64 /* 用户ID */
	ttl int64  /* TTL */
	sid uint64 /* 会话SID */
}

/******************************************************************************
 **函数名称: online_token_decode
 **功    能: 解码TOKEN
 **输入参数:
 **     token: TOKEN字串
 **输出参数: NONE
 **返    回: TOKEN字段
 **实现描述: 解析token, 并提取有效数据.
 **注意事项:
 **     TOKEN的格式"uid:${uid}:ttl:${ttl}:sid:${sid}:end"
 **     uid: 用户ID
 **     ttl: 该token的最大生命时间
 **     sid: 会话SID
 **作    者: # Qifeng.zou # 2016.11.20 09:28:06 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) online_token_decode(token string) *OnlineToken {
	tk := &OnlineToken{}

	/* > TOKEN解码 */
	cry := crypt.CreateEncodeCtx(ctx.conf.Cipher)
	orig_token := crypt.Decode(cry, token)
	words := strings.Split(orig_token, ":")
	if 7 != len(words) {
		ctx.log.Error("Token format not right! token:%s orig:%s", token, orig_token)
		return nil
	}

	ctx.log.Debug("token:%s orig:%s", token, orig_token)

	/* > 验证TOKEN合法性 */
	uid, _ := strconv.ParseInt(words[1], 10, 64)
	tk.uid = uint64(uid)
	ctx.log.Debug("words[1]:%s uid:%d", words[1], tk.uid)

	ttl, _ := strconv.ParseInt(words[3], 10, 64)
	tk.ttl = int64(ttl)
	ctx.log.Debug("words[3]:%s ttl:%d", words[3], tk.ttl)

	sid, _ := strconv.ParseInt(words[5], 10, 64)
	tk.sid = uint64(sid)
	ctx.log.Debug("words[5]:%s sid:%d sid:%d", words[5], sid, tk.sid)

	return tk
}

/******************************************************************************
 **函数名称: online_req_check
 **功    能: 检验ONLINE请求合法性
 **输入参数:
 **     req: ONLINE请求
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述: 计算TOKEN合法性
 **注意事项:
 **     1.TOKEN的格式"${uid}:${ttl}:${sid}"
 **         uid: 用户ID
 **         ttl: 该token的最大生命时间
 **         sid: 会话SID
 **     2.头部数据(MesgHeader)中的SID此时表示的是客户端的连接CID.
 **作    者: # Qifeng.zou # 2016.11.02 10:20:57 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) online_req_check(req *mesg.MesgOnline) error {
	token := ctx.online_token_decode(req.GetToken())
	if nil == token {
		ctx.log.Error("Decode token failed!")
		return errors.New("Decode token failed!")
	} else if token.ttl < time.Now().Unix() {
		ctx.log.Error("Token is timeout! uid:%d sid:%d ttl:%d", token.uid, token.sid, token.ttl)
		return errors.New("Token is timeout!")
	} else if uint64(token.uid) != req.GetUid() || uint64(token.sid) != req.GetSid() {
		ctx.log.Error("Token is invalid! uid:%d/%d sid:%d/%d ttl:%d",
			token.uid, req.GetUid(), token.sid, req.GetSid(), token.ttl)
		return errors.New("Token is invalid!!")
	}

	return nil
}

/******************************************************************************
 **函数名称: online_parse
 **功    能: 解析上线请求
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
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) online_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgOnline, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of online is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	ctx.log.Debug("Online request header! cmd:0x%04X length:%d cid:%d nid:%d seq:%d head:%d",
		head.GetCmd(), head.GetLength(),
		head.GetSid(), head.GetNid(),
		head.GetSeq(), comm.MESG_HEAD_SIZE)

	/* > 解析PB协议 */
	req = &mesg.MesgOnline{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal body failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Unmarshal body failed!")
	}

	/* > 校验协议合法性 */
	err = ctx.online_req_check(req)
	if nil != err {
		ctx.log.Error("Check online-request failed!")
		return head, req, comm.ERR_SVR_CHECK_FAIL, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: online_handler
 **功    能: 上线处理
 **输入参数:
 **     req: 上线请求
 **输出参数: NONE
 **返    回: 消息序列号+异常信息
 **实现描述:
 **     1. 校验是否SID上线信息是否存在冲突. 如果存在冲突, 则将之前的连接踢下线.
 **     2. 更新数据库信息
 **注意事项:
 **     1. 在上线请求中, head中的sid此时为侦听层cid
 **     2. 在上线请求中, req中的sid此时为会话sid
 **作    者: # Qifeng.zou # 2016.11.01 21:12:36 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) online_handler(head *comm.MesgHeader, req *mesg.MesgOnline) {
	/* 获取会话属性 */
	attr, err := ctx.cache.RoomGetSidAttr(req.GetSid())
	if nil != err {
		ctx.send_kick(req.GetSid(), head.GetCid(), head.GetNid(), comm.ERR_SYS_SYSTEM, err.Error())
		return
	} else if (0 != attr.GetUid() && attr.GetUid() != req.GetUid()) ||
		(0 != attr.GetNid() && (attr.GetNid() != head.GetNid() || attr.GetCid() != head.GetCid())) {
		// 注意：当nid为0时表示会话SID之前并未登录.
		ctx.log.Error("Session's nid is conflict! uid:%d sid:%d nid:[%d/%d] cid:%d",
			attr.GetUid(), req.GetSid(), attr.GetNid(), head.GetNid(), head.GetCid())
		/* 清理会话数据 */
		ctx.cache.RoomCleanSessionData(head.GetSid(), attr.GetCid(), attr.GetNid())
		/* 将老连接踢下线 */
		ctx.send_kick(req.GetSid(), attr.GetCid(), attr.GetNid(), comm.ERR_SVR_DATA_COLLISION, "Session's nid is collision!")
	}

	/* 更新在线数据 */
	ctx.cache.UpdateOnlineData(req.GetUid(), req.GetSid(), head.GetNid(), head.GetCid())
}

/******************************************************************************
 **函数名称: ChatRoomOnlineHandler
 **功    能: 上线请求
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;         // M|用户ID|数字|
 **        required uint64 sid = 2;         // M|会话ID|数字|
 **        required string token = 3;       // M|鉴权TOKEN|字串|
 **        required string app = 4;         // M|APP名|字串|
 **        required string version = 5;     // M|APP版本|字串|
 **        optional uint32 terminal = 6;    // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     }
 **注意事项:
 **     1. 首先需要调用MesgHeadNtoh()对头部数据进行直接序转换.
 **     2. 在上线请求中, head中的sid此时为侦听层cid
 **     3. 在上线请求中, req中的sid此时为会话sid
 **     4. 由于ONLINE请求由USRSVR模块处理, 此处只存储该状态, 无需发送应答.
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func ChatRoomOnlineHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv online request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析上线请求 */
	head, req, code, err := ctx.online_parse(data)
	if nil != err {
		ctx.log.Error("Parse online request failed! code:%d errmsg:%s", code, err.Error())
		return -1
	}

	/* > 初始化上线环境 */
	ctx.online_handler(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// 下线请求

/******************************************************************************
 **函数名称: offline_parse
 **功    能: 解析Offline请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 22:17:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) offline_parse(data []byte) (head *comm.MesgHeader) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil
	}

	return head
}

/******************************************************************************
 **函数名称: offline_handler
 **功    能: Offline处理
 **输入参数:
 **     head: 协议头
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述: 下发下线通知后, 再清理会话数据.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.11 23:23:50 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) offline_handler(head *comm.MesgHeader) error {
	/* > 下发下线通知 */
	rlist, err := ctx.cache.RoomListBySid(head.GetSid())
	if nil != err {
		ctx.log.Error("Get room list by sid failed! errmsg:%s", err.Error())
		return err
	}

	attr, err := ctx.cache.RoomGetSidAttr(head.GetSid())
	if nil != err {
		ctx.log.Error("Get sid attr failed! errmsg:%s", err.Error())
		return err
	}

	for idx, rid_str := range rlist {
		ctx.log.Debug("idx:%d rid:%s sid:%d", idx, rid_str, head.GetSid())

		rid, _ := strconv.ParseInt(rid_str, 10, 64)

		ctx.room_quit_notify(head.GetSid(), attr.GetUid(), uint64(rid))
	}

	/* > 清理会话数据 */
	return ctx.cache.RoomCleanSessionData(
		head.GetSid(), head.GetCid(), head.GetNid())
}

/******************************************************************************
 **函数名称: ChatRoomOfflineHandler
 **功    能: 下线请求
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
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func ChatRoomOfflineHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv offline request!")

	/* 1. > 解析下线请求 */
	head := ctx.offline_parse(data)
	if nil == head {
		ctx.log.Error("Parse offline request failed!")
		return -1
	}

	ctx.log.Debug("Offline data! sid:%d cid:%d nid:%d", head.GetSid(), head.GetCid(), head.GetNid())

	/* 2. > 清理会话数据 */
	err := ctx.offline_handler(head)
	if nil != err {
		ctx.log.Error("Offline handler failed!")
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// PING请求

/******************************************************************************
 **函数名称: ping_parse
 **功    能: 解析PING请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回: 协议头
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:18:29 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) ping_parse(data []byte) (head *comm.MesgHeader) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil
	}

	return head
}

/******************************************************************************
 **函数名称: ping_handler
 **功    能: PING处理
 **输入参数:
 **     head: 协议头
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 更新会话相关的TTL. 如果发现数据异常, 则需要清除该会话的数据, 并将该
 **          会话踢下线.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:53:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) ping_handler(head *comm.MesgHeader) {
	code, err := ctx.cache.RoomUpdateSessionData(
		head.GetSid(), head.GetCid(), head.GetNid())
	if nil != err {
		// 清理会话数据
		ctx.cache.RoomCleanSessionData(head.GetSid(), head.GetCid(), head.GetNid())
		ctx.send_kick(head.GetSid(), head.GetCid(), head.GetNid(), code, err.Error())
	}
}

/******************************************************************************
 **函数名称: ChatRoomPingHandler
 **功    能: 客户端PING
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 由侦听层给各终端回复PONG请求
 **作    者: # Qifeng.zou # 2016.11.03 21:40:30 #
 ******************************************************************************/
func ChatRoomPingHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv ping request!")

	/* > 解析PING请求 */
	head := ctx.ping_parse(data)
	if nil == head {
		ctx.log.Error("Parse ping request failed!")
		return -1
	}

	/* > PING请求处理 */
	ctx.ping_handler(head)

	return 0
}

/******************************************************************************
 **函数名称: send_kick
 **功    能: 发送踢人操作
 **输入参数:
 **     sid: 会话ID
 **     cid: 连接ID
 **     nid: 结点ID
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required int code = 1;          // M|错误码|数字|
 **     required string errmsg = 2;     // M|错误描述|数字|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.16 20:49:02 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) send_kick(sid uint64, cid uint64, nid uint32, code uint32, errmsg string) int {
	var head comm.MesgHeader

	ctx.log.Debug("Send kick command! sid:%d nid:%d", sid, nid)

	/* > 设置协议体 */
	req := &mesg.MesgKick{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(req)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_KICK
	head.Sid = sid
	head.Cid = cid
	head.Nid = nid
	head.Length = uint32(length)

	comm.MesgHeadHton(&head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_KICK, p.Buff, uint32(len(p.Buff)))

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 创建聊天室 */

/******************************************************************************
 **函数名称: room_creat_parse
 **功    能: 解析ROOM-CREAT请求
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
 **作    者: # Qifeng.zou # 2017.01.19 22:32:20 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_creat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-creat failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomCreat{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-creat request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_HEAD_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_creat_failed
 **功    能: 发送ROOM-CREAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-CREAT请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:33:42 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_creat_failed(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, code uint32, err error) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomCreatAck{
		Rid:    proto.Uint64(0),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(err.Error()),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_creat_ack
 **功    能: 发送ROOM-CREAT应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:28:40 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_creat_ack(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat, rid uint64) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomCreatAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(rid),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	ctx.log.Debug("Create room success! rid:%d uid:%d name:%s desc:%s",
		rid, req.GetUid(), req.GetName(), req.GetDesc())

	return 0
}

/******************************************************************************
 **函数名称: alloc_rid
 **功    能: 申请聊天室ID
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     rid: 聊天室ID
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.29 08:48:19 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) alloc_rid() (rid uint64, err error) {
	/* > 获取连接对象 */
	conn, err := ctx.seqsvr_pool.Get()
	if nil != err {
		ctx.log.Error("Get seqsvr connection pool failed! errmsg:%s", err.Error())
		return 0, err
	}

	client := conn.(*seqsvr.SeqSvrThriftClient)
	defer ctx.seqsvr_pool.Put(client, false)

	/* > 申请聊天室ID */
	_rid, err := client.AllocRoomId()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	ctx.log.Debug("Alloc rid success! rid:%d", _rid)

	return uint64(_rid), nil
}

/******************************************************************************
 **函数名称: room_creat_handler
 **功    能: ROOM-CREAT处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-CREAT请求
 **输出参数: NONE
 **返    回:
 **     rid: 聊天室ID
 **     err: 错误描述
 **实现描述:
 **注意事项: 已验证了ROOM-CREAT请求的合法性
 **作    者: # Qifeng.zou # 2017.01.19 22:22:50 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_creat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomCreat) (rid uint64, err error) {
	defer func() {
		if err := recover(); nil != err {
			ctx.log.Error("Routine crashed! errmsg:%s", err)
		}
	}()

	/* > 分配聊天室ID */
	rid, err = ctx.alloc_rid()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 更新数据到MYSQL */
	err = ctx.userdb.RoomAdd(rid, req)
	if nil != err {
		ctx.log.Error("Room add into mysql failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 更新数据到REDIS */
	err = ctx.cache.RoomAdd(rid, req)
	if nil != err {
		ctx.log.Error("Room add into redis failed! errmsg:%s", err.Error())
		return 0, err
	}

	return rid, nil
}

/******************************************************************************
 **函数名称: ChatRoomCreatHandler
 **功    能: 创建聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required string name = 2;   // M|聊天室名称|字串|
 **        required string desc = 3;   // M|聊天室描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:21:48 #
 ******************************************************************************/
func ChatRoomCreatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-creat request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析创建请求 */
	head, req, code, err := ctx.room_creat_parse(data)
	if nil == req {
		ctx.log.Error("Parse room-creat request failed!")
		ctx.room_creat_failed(head, req, code, err)
		return -1
	}

	/* > 创建聊天室处理 */
	rid, err := ctx.room_creat_handler(head, req)
	if nil != err {
		ctx.log.Error("Room creat handler failed!")
		ctx.room_creat_failed(head, req, comm.ERR_SYS_SYSTEM, err)
		return -1
	}

	/* > 发送应答 */
	ctx.room_creat_ack(head, req, rid)

	return 0
}

/* 解散聊天室 */
func ChatRoomDismissHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 加入聊天室 */

/******************************************************************************
 **函数名称: room_join_parse
 **功    能: 解析ROOM-JOIN请求
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
 **作    者: # Qifeng.zou # 2016.11.03 16:41:17 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomJoin, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-join failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomJoin{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal join request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_HEAD_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_join_failed
 **功    能: 发送ROOM-JOIN应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-JOIN请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 17:12:36 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomJoin, code uint32, err error) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomJoinAck{
		Gid:    proto.Uint32(0),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(err.Error()),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_join_ack
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_ack(head *comm.MesgHeader, req *mesg.MesgRoomJoin, gid uint32) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomJoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(gid),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_join_notify
 **功    能: 发送上线通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:51:09 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_notify(head *comm.MesgHeader, req *mesg.MesgRoomJoin) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomJoinNtf{
		Uid: proto.Uint64(req.GetUid()),
		Rid: proto.Uint64(req.GetRid()),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发上线通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head.Cmd = comm.CMD_ROOM_JOIN_NTF
		head.Length = uint32(length)
		head.Nid = ctx.listend.list.nodes[idx]

		ctx.log.Debug("Send room join notification! nid:%d", head.Nid)

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_JOIN_NTF, p.Buff, uint32(len(p.Buff)))
	}

	return 0
}

/******************************************************************************
 **函数名称: alloc_room_gid
 **功    能: 分配组ID
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 20:08:06 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) alloc_room_gid(rid uint64) (gid uint32, err error) {
	var num int

	rds := ctx.cache.Get()
	defer rds.Close()

	key := fmt.Sprintf(models.ROOM_KEY_RID_GID_TO_NUM_ZSET, rid)

	/* > 优先加入到gid为0的分组 */
	num, err = redis.Int(rds.Do("ZSCORE", key, "0"))
	if uint32(num) < comm.CHAT_ROOM_GROUP_MAX_NUM {
		return 0, nil
	}

	/* > 获取有序GID列表: 以人数从多到少进行排序(加入人数最少的分组) */
	min := 0
	max := comm.CHAT_ROOM_GROUP_MAX_NUM - 1

	gid_lst, err := redis.Ints(rds.Do("ZRANGEBYSCORE", key, min, max, "LIMIT", 0, 1))
	if nil != err {
		ctx.log.Error("Get group list failed! errmsg:%s", err)
		return 0, err
	} else if len(gid_lst) > 0 {
		return uint32(gid_lst[0]), nil
	}

	grp_num, err := redis.Int(rds.Do("ZCARD", key))
	if nil != err {
		ctx.log.Error("Get group num failed! errmsg:%s", err)
		return 0, err
	}

	return uint32(grp_num), nil
}

/******************************************************************************
 **函数名称: room_join_handler
 **功    能: ROOM-JOIN处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-JOIN请求
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项: 已验证了ROOM-JOIN请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 19:51:46 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_join_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomJoin) (gid uint32, err error) {
	rds := ctx.cache.Get()
	defer rds.Close()

	pl := ctx.cache.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 判断UID是否在黑名单中 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_BLACKLIST_SET, req.GetRid())
	ok, err := redis.Bool(rds.Do("SISMEMBER", key, req.GetUid()))
	if nil != err {
		ctx.log.Error("Exec command [SISMEMBER] failed! rid:%d uid:%d err:",
			req.GetRid(), req.GetUid(), err.Error())
		return 0, err
	} else if true == ok {
		ctx.log.Error("User is in blacklist! rid:%d uid:%d", req.GetRid(), req.GetUid())
		return 0, errors.New("User is in blacklist!")
	}

	/* > 分配新的分组 */
	gid, err = ctx.alloc_room_gid(req.GetRid())
	if nil != err {
		ctx.log.Error("Alloc gid failed! rid:%d", req.GetRid())
		return 0, err
	}

	/* > 更新数据库统计 */
	key = fmt.Sprintf(models.ROOM_KEY_RID_GID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZINCRBY", key, 1, gid)

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_UID_SID_ZSET, req.GetRid())
	member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, req.GetUid(), head.GetSid())
	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", key, ttl, member) // 加入RID -> UID集合"${uid}:${sid}"

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetSid()) // 加入RID -> SID集合

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_NID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetNid()) // 加入RID -> NID集合

	key = fmt.Sprintf(models.ROOM_KEY_SID_TO_RID_ZSET, head.GetSid())
	pl.Send("ZADD", key, ttl, req.GetRid()) /* 记录SID->RID集合 */

	return gid, nil
}

/******************************************************************************
 **函数名称: ChatRoomJoinHandler
 **功    能: 加入聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func ChatRoomJoinHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv join request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* 1. > 解析ROOM-JOIN请求 */
	head, req, code, err := ctx.room_join_parse(data)
	if nil == req {
		ctx.log.Error("Parse room-join request failed!")
		ctx.room_join_failed(head, req, code, err)
		return -1
	}

	/* 2. > 初始化上线环境 */
	gid, err := ctx.room_join_handler(head, req)
	if nil != err {
		ctx.log.Error("Room join handler failed!")
		ctx.room_join_failed(head, req, comm.ERR_SYS_SYSTEM, err)
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.room_join_ack(head, req, gid)
	ctx.room_join_notify(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 退出聊天室 */

/******************************************************************************
 **函数名称: room_quit_isvalid
 **功    能: 判断ROOM-QUIT是否合法
 **输入参数:
 **     req: ROOM-QUIT请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:26:22 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_isvalid(req *mesg.MesgRoomQuit) bool {
	if 0 == req.GetUid() || 0 == req.GetRid() {
		return false
	}
	return true
}

/******************************************************************************
 **函数名称: room_quit_failed
 **功    能: 发送ROOM-QUIT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-QUIT请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **        required uint32 gid = 3;    // M|分组ID|数字|
 **        optional uint32 code = 4; // M|错误码|数字|
 **        optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:20:34 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomQuit, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomQuitAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_QUIT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_quit_parse
 **功    能: 解析ROOM-QUIT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:18:29 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomQuit, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-quit failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomQuit{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room quit request failed! errmsg:%s", err.Error())
		return nil, nil, comm.ERR_SVR_BODY_INVALID, errors.New("Parse body failed!")
	}

	/* > 校验协议合法性 */
	if !ctx.room_quit_isvalid(req) {
		ctx.log.Error("Room quit request is invalid!")
		return nil, nil, comm.ERR_SVR_CHECK_FAIL, errors.New("Check request failed!")
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_quit_ack
 **功    能: 发送ROOM-QUIT应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 code = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_ack(head *comm.MesgHeader, req *mesg.MesgRoomQuit) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomQuitAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_QUIT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_quit_notify
 **功    能: 发送下线通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:51:09 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_notify(sid uint64, uid uint64, rid uint64) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomQuitNtf{
		Uid: proto.Uint64(uid),
		Rid: proto.Uint64(rid),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发下线通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		head := &comm.MesgHeader{
			Cmd:    comm.CMD_ROOM_QUIT_NTF,
			Length: uint32(length),
			Nid:    ctx.listend.list.nodes[idx],
		}

		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_QUIT_NTF, p.Buff, uint32(len(p.Buff)))
	}

	return 0
}

/******************************************************************************
 **函数名称: room_quit_handler
 **功    能: 退出聊天室处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-QUIT请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项: 已验证了ROOM-QUIT请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 21:28:18 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_quit_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomQuit) (code uint32, err error) {
	pl := ctx.cache.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(models.ROOM_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZREM", key, head.GetSid()) // 清理RID -> SID集合

	key = fmt.Sprintf(models.ROOM_KEY_RID_TO_UID_SID_ZSET, req.GetRid())
	member := fmt.Sprintf(comm.CHAT_FMT_UID_SID_STR, req.GetUid(), head.GetSid())
	pl.Send("ZREM", key, member) // 清理RID -> UID集合"${uid}:${sid}"

	return 0, nil
}

/******************************************************************************
 **函数名称: ChatRoomQuitHandler
 **功    能: 退出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项: 需要对协议头进行字节序转换
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func ChatRoomQuitHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room quit request!")

	/* 1. > 解析ROOM-QUIT请求 */
	head, req, code, err := ctx.room_quit_parse(data)
	if nil != err {
		ctx.log.Error("Parse room quit request failed!")
		ctx.room_quit_failed(head, req, code, err.Error())
		return -1
	}

	/* 2. > 退出聊天室处理 */
	code, err = ctx.room_quit_handler(head, req)
	if nil != err {
		ctx.log.Error("Hanle room quit request failed!")
		ctx.room_quit_failed(head, req, code, err.Error())
		return -1
	}

	/* 3. > 发送ROOM-QUIT应答 */
	ctx.room_quit_ack(head, req)
	ctx.room_quit_notify(head.GetSid(), req.GetUid(), req.GetRid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 踢出聊天室 */

/******************************************************************************
 **函数名称: room_kick_parse
 **功    能: 解析KICK请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:21:37 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomKick, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of room-kick is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomKick{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-kick request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_kick_failed
 **功    能: 发送KICK应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: ROOM-KICK请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 17:12:36 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomKick, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomKickAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_KICK_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_kick_by_uid
 **功    能: 通过UID下发KICK指令
 **输入参数:
 **     head: 协议头
 **     req: ROOM-KICK请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.19 23:06:38 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_by_uid(rid uint64, uid uint64) (code uint32, err error) {
	rds := ctx.cache.Get()
	defer rds.Close()

	/* > 获取会话列表 */
	key := fmt.Sprintf(models.ROOM_KEY_UID_TO_SID_SET, uid)

	sid_list, err := redis.Strings(rds.Do("SMEMBERS", key))
	if nil != err {
		ctx.log.Error("Get sid list by uid failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	num := len(sid_list)
	if 0 == num {
		return 0, nil
	}

	/* > 生成PB数据 */
	req := &mesg.MesgRoomKick{
		Uid: proto.Uint64(uid),
		Rid: proto.Uint64(rid),
	}

	body, err := proto.Marshal(req)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return comm.ERR_SVR_BODY_INVALID, err
	}

	length := len(body)

	/* > 遍历会话列表 */
	for idx := 0; idx < num; idx += 1 {
		sid, _ := strconv.ParseInt(sid_list[idx], 10, 64)

		/* > 获取会话属性 */
		attr, err := ctx.cache.RoomGetSidAttr(uint64(sid))
		if nil != err {
			ctx.log.Error("Get sid attr failed! rid:%d uid:%d errmsg:%s",
				rid, uid, err.Error())
			return comm.ERR_SYS_SYSTEM, err
		}

		/* > 下发踢除指令 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head := &comm.MesgHeader{
			Cmd:    comm.CMD_ROOM_KICK,
			Sid:    attr.GetSid(),
			Cid:    attr.GetCid(),
			Nid:    attr.GetNid(),
			Length: uint32(length),
			Seq:    0,
		}

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK, p.Buff, uint32(len(p.Buff)))
	}
	return 0, nil
}

/******************************************************************************
 **函数名称: room_kick_ack
 **功    能: 发送ROOM-KICK应答
 **输入参数:
 **     head: 协议头
 **     req: 请求数据
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 code = 3;   // M|错误码|数字|
 **         required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:32:20 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_ack(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomKickAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ROOM_KICK_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: room_kick_notify
 **功    能: 发送被踢通知
 **输入参数:
 **     head: 请求消息头
 **     req: 请求消息
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 10:59:32 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_notify(head *comm.MesgHeader, req *mesg.MesgRoomKick) int {
	/* > 设置协议体 */
	ntf := &mesg.MesgRoomKickNtf{
		Uid: proto.Uint64(req.GetUid()),
		Rid: proto.Uint64(req.GetRid()),
	}

	/* > 生成PB数据 */
	body, err := proto.Marshal(ntf)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 下发被踢通知 */
	ctx.listend.list.RLock()
	defer ctx.listend.list.RUnlock()

	num := len(ctx.listend.list.nodes)

	for idx := 0; idx < num; idx += 1 {
		/* > 拼接协议包 */
		p := &comm.MesgPacket{}
		p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

		head.Cmd = comm.CMD_ROOM_KICK_NTF
		head.Length = uint32(length)
		head.Nid = ctx.listend.list.nodes[idx]

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_KICK_NTF, p.Buff, uint32(len(p.Buff)))
	}

	return 0
}

/******************************************************************************
 **函数名称: room_kick_handler
 **功    能: ROOM-KICK处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-KICK请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述:
 **注意事项: 已验证了ROOM-KICK请求的合法性
 **作    者: # Qifeng.zou # 2017.01.12 23:34:28 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_kick_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomKick) (code uint32, err error) {
	pl := ctx.cache.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 获取会话属性 */
	attr, err := ctx.cache.RoomGetSidAttr(head.GetSid())
	if nil != err {
		ctx.log.Error("Get sid attr failed! rid:%d uid:%d errmsg:%s",
			req.GetRid(), req.GetUid(), err.Error())
		return comm.ERR_SYS_SYSTEM, err
	} else if !ctx.cache.IsRoomManager(req.GetRid(), attr.GetUid()) {
		ctx.log.Error("You're not owner! rid:%d kicked-uid:%d attr.uid:%d",
			req.GetRid(), req.GetUid(), attr.GetUid())
		return comm.ERR_SYS_PERM_DENIED, errors.New("You're not room owner!")
	}

	/* > 用户加入黑名单 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_USR_BLACKLIST_SET, req.GetRid())

	pl.Send("SADD", key, req.GetUid())

	/* > 提交MONGO存储 */
	blacklist := &models.RoomBlacklistTabRow{
		Rid:    req.GetRid(),               // 聊天室ID
		Uid:    req.GetUid(),               // 用户ID
		Status: models.ROOM_USER_STAT_KICK, // 状态(被踢)
		Ctm:    time.Now().Unix(),          // 设置时间
	}

	cb := func(c *mgo.Collection) (err error) {
		c.Insert(blacklist)
		return err
	}

	ctx.mongo.Exec(ctx.conf.Mongo.DbName, models.ROOM_TAB_BLACKLIST, cb)

	/* > 遍历下发踢除指令 */
	ctx.room_kick_by_uid(req.GetRid(), req.GetUid())

	return 0, nil
}

/******************************************************************************
 **函数名称: ChatRoomKickHandler
 **功    能: 踢出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **        required uint32 code = 3;   // M|错误码|数字|
 **        required string errmsg = 4; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.12 23:58:49 #
 ******************************************************************************/
func ChatRoomKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-kick request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析ROOM-KICK请求 */
	head, req, code, err := ctx.room_kick_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-kick failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 执行ROOM-KICK操作 */
	code, err = ctx.room_kick_handler(head, req)
	if nil != err {
		ctx.log.Error("Room-kick handler failed!")
		ctx.room_kick_failed(head, req, code, err.Error())
		return -1
	}

	/* > 发送ROOM-KICK应答 */
	ctx.room_kick_ack(head, req)
	ctx.room_kick_notify(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 聊天室各侦听统计 */

/******************************************************************************
 **函数名称: room_lsn_stat_parse
 **功    能: 解析ROOM-LSN-STAT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.13 06:51:41 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_lsn_stat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomLsnStat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(0) {
		errmsg := "Header of room-lsn-stat is invalid!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	ctx.log.Debug("Recv room-lsn-stat data! cmd:0x%04X nid:%d",
		head.GetCmd(), head.GetNid())

	/* > 解析PB协议 */
	req = &mesg.MesgRoomLsnStat{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-lsn-stat request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	ctx.log.Debug("nid:%d rid:%d num:%d", head.GetNid(), req.GetRid(), req.GetNum())

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_lsn_stat_handler
 **功    能: ROOM-LSN-STAT处理
 **输入参数:
 **     head: 协议头
 **     req: ROOM-LSN-STAT请求
 **输出参数: NONE
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述:
 **注意事项: 已验证了ROOM-LSN-STAT请求的合法性
 **作    者: # Qifeng.zou # 2017.05.13 06:54:54 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) room_lsn_stat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomLsnStat) (code uint32, err error) {
	pl := ctx.cache.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ttl := time.Now().Unix() + models.ROOM_TTL_SEC

	/* > 更新统计数据 */
	key := fmt.Sprintf(models.ROOM_KEY_RID_NID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZADD", key, req.GetNum(), req.GetNid())

	pl.Send("ZADD", models.ROOM_KEY_RID_ZSET, ttl, req.GetRid())

	return 0, nil
}

/******************************************************************************
 **函数名称: ChatRoomKickHandler
 **功    能: 踢出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 rid = 1;     // M|聊天室ID|数字|
 **        required uint32 nid = 2;     // M|结点ID|数字|
 **        required uint32 num = 3;     // M|在线人数|数字|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.13 07:03:24 #
 ******************************************************************************/
func ChatRoomLsnStatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-lsn-stat request!")

	/* > 解析ROOM-LSN-STAT请求 */
	head, req, code, err := ctx.room_lsn_stat_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-lsn-stat request failed!")
		return -1
	}

	/* > 执行ROOM-LSN-STAT操作 */
	code, err = ctx.room_lsn_stat_handler(head, req)
	if nil != err {
		ctx.log.Error("Room lsn stat handler failed! code:%d errmsg:%s", code, err.Error())
		return -1
	}

	return 0
}

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
func (ctx *ChatRoomCntx) room_chat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgRoomChat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d",
			head.GetCmd(), head.GetNid())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New("Header is invalid!")
	}

	/* > 解析PB协议 */
	req = &mesg.MesgRoomChat{}

	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal room-msg failed! sid:%d cid:%d nid:%d errmsg:%s",
			head.GetSid(), head.GetCid(), head.GetNid(), err.Error())
		return head, nil, comm.ERR_SVR_BODY_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: room_chat_failed
 **功    能: 发送ROOM-CHAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: 聊天消息
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
func (ctx *ChatRoomCntx) room_chat_failed(head *comm.MesgHeader,
	req *mesg.MesgRoomChat, code uint32, errmsg string) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
	}

	if nil != req {
		ack.Uid = proto.Uint64(req.GetUid())
		ack.Rid = proto.Uint64(req.GetRid())
		ack.Gid = proto.Uint32(req.GetGid())
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	return ctx.send_data(comm.CMD_ROOM_CHAT_ACK, head.GetSid(),
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
}

/******************************************************************************
 **函数名称: room_chat_ack
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
func (ctx *ChatRoomCntx) room_chat_ack(head *comm.MesgHeader, req *mesg.MesgRoomChat) int {
	/* > 设置协议体 */
	ack := &mesg.MesgRoomChatAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(req.GetGid()),
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
		head.GetCid(), head.GetNid(), head.GetSeq(), body, uint32(len(body)))
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
func (ctx *ChatRoomCntx) room_chat_handler(
	head *comm.MesgHeader, req *mesg.MesgRoomChat, data []byte) (err error) {
	item := &MesgRoomItem{}

	ctx.log.Debug("rid:%d gid:%d sid:%d cid:%d nid:%d",
		req.GetRid(), req.GetGid(), head.GetSid(), head.GetCid(), head.GetNid())

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
		ctx.log.Error("Get node list failed! rid:%d", req.GetRid())
		return nil
	}

	/* > 遍历rid->nid列表, 并下发聊天室消息 */
	for idx, nid := range nid_list {
		ctx.log.Debug("idx:%d rid:%d nid:%d", idx, req.GetRid(), nid)

		ctx.send_data(comm.CMD_ROOM_CHAT, head.GetSid(), 0, uint32(nid),
			head.GetSeq(), data[comm.MESG_HEAD_SIZE:], head.GetLength())
	}
	return err
}

/******************************************************************************
 **函数名称: ChatRoomChatHandler
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
func ChatRoomChatHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv room-chat message!")

	/* > 解析ROOM-CHAT协议 */
	head, req, code, err := ctx.room_chat_parse(data)
	if nil != err {
		ctx.log.Error("Parse room-msg failed! code:%d errmsg:%s", code, err.Error())
		if nil != head {
			ctx.room_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		}
		return -1
	}

	/* > 进行业务处理 */
	err = ctx.room_chat_handler(head, req, data)
	if nil != err {
		ctx.log.Error("Handle room message failed!")
		ctx.room_chat_failed(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	return ctx.room_chat_ack(head, req)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomChatAckHandler
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
func ChatRoomChatAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomBcHandler
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
func ChatRoomBcHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: ChatRoomBcAckHandler
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
func ChatRoomBcAckHandler(cmd uint32, nid uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*ChatRoomCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task_room_mesg_chan_pop
 **功    能: 聊天室消息的存储任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 从聊天室消息队列中取出消息, 并进行存储处理
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 23:43:03 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) task_room_mesg_chan_pop() {
	for item := range ctx.room_mesg_chan {
		item.storage(ctx)
	}
}

/******************************************************************************
 **函数名称: storage
 **功    能: 聊天室消息的存储处理
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: NONE
 **实现描述: 将消息存入聊天室缓存和数据库
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:05:51 #
 ******************************************************************************/
func (item *MesgRoomItem) storage(ctx *ChatRoomCntx) {
	pl := ctx.cache.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 解析PB协议 */
	msg := &mesg.MesgRoomChat{}

	err := proto.Unmarshal(item.raw[comm.MESG_HEAD_SIZE:], msg)
	if nil != err {
		ctx.log.Error("Unmarshal room-chat-mesg failed!")
		return
	}

	/* > 提交REDIS缓存 */
	key := fmt.Sprintf(models.ROOM_KEY_ROOM_MESG_QUEUE, item.req.GetRid())
	pl.Send("LPUSH", key, item.raw[comm.MESG_HEAD_SIZE:])

	/* > 提交MONGO存储 */
	data := &models.RoomChatTabRow{
		Rid:  msg.GetRid(),
		Uid:  msg.GetUid(),
		Ctm:  time.Now().Unix(),
		Data: item.raw,
	}

	cb := func(c *mgo.Collection) (err error) {
		c.Insert(data)
		return err
	}

	ctx.mongo.Exec(ctx.conf.Mongo.DbName, models.ROOM_TAB_MESG, cb)
}

/******************************************************************************
 **函数名称: task_room_mesg_queue_clean
 **功    能: 清理聊天室缓存消息
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **实现描述: 保持聊天室缓存消息为最新的100条
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 22:34:18 #
 ******************************************************************************/
func (ctx *ChatRoomCntx) task_room_mesg_queue_clean() {
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
func (ctx *ChatRoomCntx) room_mesg_queue_clean() {
	rds := ctx.cache.Get()
	defer rds.Close()

	off := 0
	for {
		rid_list, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
			models.ROOM_KEY_RID_ZSET, 0, "+inf", "LIMIT", off, comm.CHAT_BAT_NUM))
		if nil != err {
			ctx.log.Error("Get rid list failed! errmsg:%s", err.Error())
			break
		}

		num := len(rid_list)
		for idx := 0; idx < num; idx += 1 {
			/* 保持聊天室缓存消息为最新的100条 */
			rid, _ := strconv.ParseInt(rid_list[idx], 10, 64)
			key := fmt.Sprintf(models.ROOM_KEY_ROOM_MESG_QUEUE, uint64(rid))

			rds.Do("LTRIM", key, 0, 99)
		}

		if num < comm.CHAT_BAT_NUM {
			break
		}
		off += num
	}
}

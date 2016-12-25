package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/crypt"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

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
func (ctx *UsrSvrCntx) online_token_decode(token string) *OnlineToken {
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
func (ctx *UsrSvrCntx) online_req_check(req *mesg.MesgOnlineReq) error {
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
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) online_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgOnlineReq, err error) {

	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !comm.MesgHeadIsValid(head) {
		ctx.log.Error("Header of Online request is invalid! cmd:0x%04X flag:%d length:%d chksum:0x%08X sid:%d nid:%d serial:%d head:%d",
			head.GetCmd(), head.GetFlag(), head.GetLength(),
			head.GetChkSum(), head.GetSid(), head.GetNid(),
			head.GetSerial(), comm.MESG_HEAD_SIZE)
		return nil, nil, errors.New("Header of online request is invalied!")
	}

	ctx.log.Debug("Online request header! cmd:0x%04X flag:%d length:%d chksum:0x%08X sid:%d nid:%d serial:%d head:%d",
		head.GetCmd(), head.GetFlag(), head.GetLength(),
		head.GetChkSum(), head.GetSid(), head.GetNid(),
		head.GetSerial(), comm.MESG_HEAD_SIZE)

	/* > 解析PB协议 */
	req = &mesg.MesgOnlineReq{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal online request failed! errmsg:%s", err.Error())
		return head, nil, errors.New("Unmarshal online request failed!")
	}

	/* > 校验协议合法性 */
	err = ctx.online_req_check(req)
	if nil != err {
		return head, nil, err
	}

	return head, req, nil
}

/******************************************************************************
 **函数名称: send_err_online_ack
 **功    能: 发送上线应答
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 ** {
 **     required uint64 uid = 1;        // M|用户ID|数字|
 **     required uint64 sid = 2;        // M|连接ID|数字|内部使用
 **     required string app = 3;        // M|APP名|字串|
 **     required string version = 4;    // M|APP版本|字串|
 **     optional uint32 terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     optional uint32 errnum = 6;     // M|错误码|数字|
 **     optional string errmsg = 7;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_err_online_ack(head *comm.MesgHeader,
	req *mesg.MesgOnlineReq, code uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgOnlineAck{
		Uid:      proto.Uint64(req.GetUid()),
		Sid:      proto.Uint64(0),
		App:      proto.String(req.GetApp()),
		Version:  proto.String(req.GetVersion()),
		Terminal: proto.Uint32(req.GetTerminal()),
		Code:     proto.Uint32(code),
		Errmsg:   proto.String(errmsg),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ONLINE_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ONLINE_ACK, p.Buff, uint32(len(p.Buff)))

	ctx.log.Debug("Send online ack succ!")
	return 0
}

/******************************************************************************
 **函数名称: send_online_ack
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 ** {
 **     required uint64 uid = 1;        // M|用户ID|数字|
 **     required uint64 sid = 2;        // M|会话ID|数字|内部使用
 **     required string app = 3;        // M|APP名|字串|
 **     required string version = 4;    // M|APP版本|字串|
 **     optional uint32 terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     optional uint32 errnum = 6;     // M|错误码|数字|
 **     optional string errmsg = 7;     // M|错误描述|字串|
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_online_ack(head *comm.MesgHeader, req *mesg.MesgOnlineReq) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgOnlineAck{
		Uid:      proto.Uint64(req.GetUid()),
		Sid:      proto.Uint64(head.GetSid()),
		App:      proto.String(req.GetApp()),
		Version:  proto.String(req.GetVersion()),
		Terminal: proto.Uint32(req.GetTerminal()),
		Code:     proto.Uint32(0),
		Errmsg:   proto.String("Ok"),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_ONLINE_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ONLINE_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: online_handler
 **功    能: 上线处理
 **输入参数:
 **     req: 上线请求
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述:
 **     1. 校验是否SID上线信息是否存在冲突. 如果存在冲突, 则将之前的连接踢下线.
 **     2. 更新数据库信息
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 21:12:36 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) online_handler(head *comm.MesgHeader, req *mesg.MesgOnlineReq) (err error) {
	var key string

	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ttl := time.Now().Unix() + comm.CHAT_SID_TTL

	/* 获取SID -> (UID/NID)的映射 */
	key = fmt.Sprintf(comm.IM_KEY_SID_ATTR, head.GetSid())

	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		ctx.log.Error("Get sid attribution failed! errmsg:%s", err)
		return err
	}

	id, _ := strconv.ParseInt(vals[0], 10, 64)
	uid := uint64(id)
	id, _ = strconv.ParseInt(vals[1], 10, 32)
	nid := uint32(id)

	if nid != head.GetNid() {
		ctx.log.Error("Session's nid is conflict! uid:%d sid:%d nid:[%d/%d]",
			uid, head.GetSid(), nid, head.GetNid())
		ctx.send_kick(head.GetSid(), nid, comm.ERR_SVR_DATA_COLLISION, "Session's nid is collision!")
	}

	/* 记录SID集合 */
	pl.Send("ZADD", comm.IM_KEY_SID_ZSET, ttl, head.GetSid())

	/* 记录UID集合 */
	pl.Send("ZADD", comm.IM_KEY_UID_ZSET, ttl, req.GetUid())

	/* 记录SID->UID/NID */
	key = fmt.Sprintf(comm.IM_KEY_SID_ATTR, head.GetSid())
	pl.Send("HMSET", key, "UID", req.GetUid(), "NID", head.GetNid())

	/* 记录UID->SID集合 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, req.GetUid())
	pl.Send("SADD", key, head.GetSid())

	return err
}

/******************************************************************************
 **函数名称: UsrSvrOnlineReqHandler
 **功    能: 上线请求
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
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
 **注意事项: 首先需要调用MesgHeadNtoh()对头部数据进行直接序转换.
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func UsrSvrOnlineReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request! cmd:0x%04X orig:%d length:%d", cmd, orig, length)

	/* 1. > 解析上线请求 */
	head, req, err := ctx.online_parse(data)
	if nil == head {
		ctx.log.Error("Parse online request failed! Header is invalid.")
		return -1
	} else if nil == req {
		ctx.log.Error("Parse online request failed! Request is invalid.")
		ctx.send_err_online_ack(head, req, comm.ERR_SVR_PARSE_PARAM, err.Error())
		return -1
	}

	/* 2. > 初始化上线环境 */
	err = ctx.online_handler(head, req)
	if nil != err {
		ctx.log.Error("Online handler failed!")
		ctx.send_err_online_ack(head, req, comm.ERR_SYS_SYSTEM, err.Error())
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.send_online_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

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
func (ctx *UsrSvrCntx) offline_parse(data []byte) (head *comm.MesgHeader) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	return head
}

/******************************************************************************
 **函数名称: offline_handler
 **功    能: Offline处理
 **输入参数:
 **     head: 协议头
 **     req: 下线请求
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 22:22:02 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) offline_handler(head *comm.MesgHeader) error {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	// 获取SID -> (UID/NID)的映射
	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, head.GetSid())

	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		ctx.log.Error("Get sid attribution failed! errmsg:%s", err)
		return err
	}

	id, _ := strconv.ParseInt(vals[0], 10, 64)
	uid := uint64(id)
	id, _ = strconv.ParseInt(vals[1], 10, 32)
	nid := uint32(id)

	if nid != head.GetNid() {
		ctx.log.Error("Nid isn't right! nid:%d/%d", nid, head.GetNid())
		return errors.New("Node id isn't right!")
	}

	// 删除SID -> (UID/NID)的映射
	num, err := redis.Int(rds.Do("DEL", key))
	if nil != err {
		ctx.log.Error("Delete key failed! errmsg:%s", err)
		return err
	} else if 0 == num {
		ctx.log.Error("Sid [%d] was cleaned!", head.GetSid())
		return nil
	}

	/* 删除SID集合 */
	pl.Send("ZREM", comm.IM_KEY_SID_ZSET, head.GetSid())

	/* 记录UID->SID集合 */
	key = fmt.Sprintf(comm.IM_KEY_UID_TO_SID_SET, uid)
	pl.Send("SREM", key, head.GetSid())

	return nil
}

/******************************************************************************
 **函数名称: UsrSvrOfflineReqHandler
 **功    能: 下线请求
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func UsrSvrOfflineReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv offline request!")

	/* 1. > 解析下线请求 */
	head := ctx.offline_parse(data)
	if nil == head {
		ctx.log.Error("Parse offline request failed!")
		return -1
	}

	/* 2. > 初始化上线环境 */
	err := ctx.offline_handler(head)
	if nil != err {
		ctx.log.Error("Offline handler failed!")
		return -1
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: join_req_isvalid
 **功    能: 判断JOIN是否合法
 **输入参数:
 **     req: JOIN请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述: 计算TOKEN合法性
 **注意事项:
 **     TOKEN的格式"uid:${uid}:rid:${rid}:ttl:${ttl}"
 **     uid: 用户ID
 **     ttl: 该token的最大生命时间
 **作    者: # Qifeng.zou # 2016.11.03 16:41:28 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) join_req_isvalid(req *mesg.MesgJoinReq) bool {
	/* > TOKEN解码 */
	cry := crypt.CreateEncodeCtx(ctx.conf.Cipher)
	token := crypt.Decode(cry, req.GetToken())
	words := strings.Split(token, ":")
	if 4 != len(words) {
		ctx.log.Error("Token format not right! token:%s", token)
		return false
	}

	/* > 验证TOKEN合法性 */
	uid, _ := strconv.ParseInt(words[1], 10, 64)
	rid, _ := strconv.ParseInt(words[3], 10, 64)
	ttl, _ := strconv.ParseInt(words[5], 10, 64)
	if ttl < time.Now().Unix() {
		ctx.log.Error("Token is timeout!")
		return false
	} else if uint64(uid) != req.GetUid() {
		ctx.log.Error("Token is invalid! uid:%d/%d", uid, req.GetUid())
		return false
	} else if uint64(rid) != req.GetRid() {
		ctx.log.Error("Token is invalid! rid:%d/%d", rid, req.GetRid())
		return false
	}

	return true
}

/******************************************************************************
 **函数名称: join_parse
 **功    能: 解析JOIN请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 16:41:17 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) join_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgJoinReq) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgJoinReq{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal join request failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */
	if !ctx.join_req_isvalid(req) {
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_join_ack
 **功    能: 发送JOIN应答(异常)
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
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 17:12:36 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_err_join_ack(head *comm.MesgHeader,
	req *mesg.MesgJoinReq, code uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgJoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(0),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: send_join_ack
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
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_join_ack(head *comm.MesgHeader, req *mesg.MesgJoinReq, gid uint32) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgJoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Gid:    proto.Uint32(gid),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_JOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_JOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: alloc_gid
 **功    能: 分配组ID
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 20:08:06 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) alloc_gid(rid uint64) (gid uint32, err error) {
	var num int

	rds := ctx.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, rid)

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
 **函数名称: join_handler
 **功    能: JOIN处理
 **输入参数:
 **     head: 协议头
 **     req: JOIN请求
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项: 已验证了JION请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 19:51:46 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) join_handler(
	head *comm.MesgHeader, req *mesg.MesgJoinReq) (gid uint32, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

GET_GID:
	/* > 判断UID是否登录过 */
	key := fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, req.GetUid())
	ok, err := redis.Bool(rds.Do("HEXISTS", key, req.GetRid()))
	if nil != err {
		ctx.log.Error("Get rid [%d] by uid failed!", req.GetRid())
		return 0, err
	} else if true == ok {
		gid_int, err := redis.Int(rds.Do("HGET", key, req.GetRid()))
		if nil != err {
			ctx.log.Error("Get rid [%d] by uid failed!", req.GetRid())
			return 0, err
		}

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_ZSET, req.GetRid())
		ttl := time.Now().Unix() + comm.CHAT_SID_TTL
		pl.Send("ZINCRBY", key, ttl, req.GetUid())

		key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
		pl.Send("ZADD", key, ttl, head.GetSid())
		return uint32(gid_int), nil
	}

	/* > 分配新的分组 */
	gid, err = ctx.alloc_gid(req.GetRid())
	if nil != err {
		ctx.log.Error("Alloc gid failed! rid:%d", req.GetRid())
		return 0, err
	}

	/* > 设置UID的RID组GID */
	key = fmt.Sprintf(comm.CHAT_KEY_UID_TO_RID, req.GetUid())
	ok, err = redis.Bool(rds.Do("HSETNX", key, req.GetRid(), gid)) /* 防止冲突 */
	if nil != err {
		ctx.log.Error("Get rid [%d] by uid failed!", req.GetRid())
		return 0, err
	} else if false == ok {
		goto GET_GID /* 存在冲突 */
	}

	/* > 更新数据库统计 */
	key = fmt.Sprintf(comm.CHAT_KEY_RID_GID_TO_NUM_ZSET, req.GetRid())
	pl.Send("ZINCRBY", key, 1, gid)

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_UID_ZSET, req.GetRid())
	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZINCRBY", key, ttl, req.GetUid())

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetSid())

	key = fmt.Sprintf(comm.CHAT_KEY_RID_TO_NID_ZSET, req.GetRid())
	pl.Send("ZADD", key, ttl, head.GetNid())

	return gid, nil
}

/******************************************************************************
 **函数名称: UsrSvrJoinReqHandler
 **功    能: 加入聊天室
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
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
 **        required string token = 3;  // M|鉴权TOKEN|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func UsrSvrJoinReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv join request! cmd:0x%04X orig:%d length:%d", cmd, orig, length)

	/* 1. > 解析JOIN请求 */
	head, req := ctx.join_parse(data)
	if nil == head || nil != req {
		ctx.log.Error("Parse join request failed!")
		ctx.send_err_join_ack(head, req, comm.ERR_SVR_PARSE_PARAM, "Parse join request failed!")
		return -1
	}

	/* 2. > 初始化上线环境 */
	gid, err := ctx.join_handler(head, req)
	if nil != err {
		ctx.log.Error("Online handler failed!")
		ctx.send_err_join_ack(head, req, comm.ERR_SYS_SYSTEM, err.Error())
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.send_join_ack(head, req, gid)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: unjoin_req_isvalid
 **功    能: 判断UNJOIN是否合法
 **输入参数:
 **     req: UNJOIN请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:26:22 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) unjoin_req_isvalid(req *mesg.MesgUnjoinReq) bool {
	if 0 == req.GetUid() || 0 == req.GetRid() {
		return false
	}
	return true
}

/******************************************************************************
 **函数名称: send_err_unjoin_ack
 **功    能: 发送UNJOIN应答(异常)
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
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required uint64 rid = 2;    // M|聊天室ID|数字|
 **        required uint32 gid = 3;    // M|分组ID|数字|
 **        optional uint32 errnum = 4; // M|错误码|数字|
 **        optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:20:34 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_err_unjoin_ack(head *comm.MesgHeader,
	req *mesg.MesgUnjoinReq, code uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgUnjoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_UNJOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_UNJOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: unjoin_parse
 **功    能: 解析UNJOIN请求
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
func (ctx *UsrSvrCntx) unjoin_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgUnjoinReq) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgUnjoinReq{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal join request failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */
	if !ctx.unjoin_req_isvalid(req) {
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_unjoin_ack
 **功    能: 发送UNJOIN应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint64 uid = 1;    // M|用户ID|数字|
 **         required uint64 rid = 2;    // M|聊天室ID|数字|
 **         required uint32 gid = 3;    // M|分组ID|数字|
 **         optional uint32 errnum = 4; // M|错误码|数字|
 **         optional string errmsg = 5; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) send_unjoin_ack(head *comm.MesgHeader, req *mesg.MesgUnjoinReq) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgUnjoinAck{
		Uid:    proto.Uint64(req.GetUid()),
		Rid:    proto.Uint64(req.GetRid()),
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_UNJOIN_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_UNJOIN_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: unjoin_handler
 **功    能: UNJOIN处理
 **输入参数:
 **     head: 协议头
 **     req: UNJOIN请求
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项: 已验证了UNJION请求的合法性
 **作    者: # Qifeng.zou # 2016.11.03 21:28:18 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) unjoin_handler(
	head *comm.MesgHeader, req *mesg.MesgUnjoinReq) (err error) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	key := fmt.Sprintf(comm.CHAT_KEY_RID_TO_SID_ZSET, req.GetRid())
	pl.Send("ZREM", key, head.GetSid())

	return nil
}

/******************************************************************************
 **函数名称: UsrSvrUnjoinReqHandler
 **功    能: 退出聊天室
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
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
func UsrSvrUnjoinReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv unjoin request!")

	/* 1. > 解析UNJOIN请求 */
	head, req := ctx.unjoin_parse(data)
	if nil == head || nil != req {
		ctx.log.Error("Parse join request failed!")
		ctx.send_err_unjoin_ack(head, req, comm.ERR_SVR_PARSE_PARAM, "Parse join request failed!")
		return -1
	}

	/* 2. > 退出聊天室处理 */
	ctx.unjoin_handler(head, req)

	/* 3. > 发送UNJOIN应答 */
	ctx.send_unjoin_ack(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

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
func (ctx *UsrSvrCntx) ping_parse(data []byte) (head *comm.MesgHeader) {
	/* > 字节序转换 */
	return comm.MesgHeadNtoh(data)
}

/******************************************************************************
 **函数名称: ping_handler
 **功    能: PING处理
 **输入参数:
 **     head: 协议头
 **输出参数: NONE
 **返    回: 组ID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:53:38 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) ping_handler(head *comm.MesgHeader) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* 获取SID->UID/NID */
	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, head.GetSid())
	vals, err := redis.Strings(rds.Do("HMGET", key, "UID", "NID"))
	if nil != err {
		ctx.log.Error("Get uid by sid [%d] failed!", head.GetSid())
		return
	}

	uid_int, _ := strconv.ParseInt(vals[0], 10, 64)
	uid := uint64(uid_int)
	nid_int, _ := strconv.ParseInt(vals[1], 10, 64)
	nid := uint32(nid_int)

	if nid != head.GetNid() {
		ctx.log.Error("Node id isn't right! sid:%d nid:%d/%d",
			head.GetSid(), nid, head.GetNid())
		return
	}

	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", comm.IM_KEY_SID_ZSET, ttl, head.GetSid())
	pl.Send("ZADD", comm.IM_KEY_UID_ZSET, ttl, uid)
}

/******************************************************************************
 **函数名称: UsrSvrPingHandler
 **功    能: 客户端PING
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.03 21:40:30 #
 ******************************************************************************/
func UsrSvrPingHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv ping request!")

	/* 1. > 解析PING请求 */
	head := ctx.ping_parse(data)

	/* 2. > PING请求处理 */
	ctx.ping_handler(head)

	return 0
}

/******************************************************************************
 **函数名称: send_kick
 **功    能: 发送踢人操作
 **输入参数:
 **     sid: 会话ID
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
func (ctx *UsrSvrCntx) send_kick(sid uint64, nid uint32, code uint32, errmsg string) int {
	var head comm.MesgHeader

	/* > 设置协议体 */
	rsp := &mesg.MesgKickReq{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(errmsg),
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
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_KICK_REQ
	head.Sid = sid
	head.Nid = nid
	head.Length = uint32(length)

	comm.MesgHeadHton(&head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_KICK_REQ, p.Buff, uint32(len(p.Buff)))

	ctx.log.Debug("Send kick command success! sid:%d nid:%d", sid, nid)
	return 0
}

package ctrl

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/crypt"
	"chat/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: online_req_isvalid
 **功    能: 判断ONLINE是否合法
 **输入参数:
 **     req: ONLINE请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述: 计算TOKEN合法性
 **注意事项:
 **     TOKEN的格式"uid:${uid}:ttl:${ttl}"
 **     uid: 用户ID
 **     ttl: 该token的最大生命时间
 **作    者: # Qifeng.zou # 2016.11.02 10:20:57 #
 ******************************************************************************/
func (ctx *OlsvrCntx) online_req_isvalid(req *mesg.MesgOnlineReq) bool {
	/* > TOKEN解码 */
	cry := crypt.CreateEncodeCtx(ctx.conf.SecretKey)
	token := crypt.Decode(cry, req.GetToken())
	words := strings.Split(token, ":")
	if 4 != len(words) {
		ctx.log.Error("Token format not right! token:%s", token)
		return false
	}

	/* > 验证TOKEN合法性 */
	uid, _ := strconv.ParseInt(words[1], 10, 64)
	ttl, _ := strconv.ParseInt(words[3], 10, 64)
	if ttl < time.Now().Unix() {
		ctx.log.Error("Token is timeout!")
		return false
	} else if uint64(uid) != req.GetUid() {
		ctx.log.Error("Token is invalid! uid:%d/%d", uid, req.GetUid())
		return false
	}

	return true
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
 **实现描述:
 **注意事项: 首先需要调用MesgHeadNtoh()对头部数据进行直接序转换.
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *OlsvrCntx) online_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgOnlineReq) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if comm.CMD_ONLINE_REQ != head.GetCmd() {
		ctx.log.Error("Command type isn't right! cmd:%d", head.GetCmd())
		return nil, nil
	}

	/* > 解析PB协议 */
	req = &mesg.MesgOnlineReq{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal online request failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */
	if !ctx.online_req_isvalid(req) {
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: send_err_online_ack
 **功    能: 发送上线应答
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **     errno: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 ** {
 **     optional uint64 Uid = 1;        // M|用户ID|数字|<br>
 **     optional uint64 Sid = 2;        // M|连接ID|数字|内部使用<br>
 **     optional string App = 3;        // M|APP名|字串|<br>
 **     optional string Version = 4;    // M|APP版本|字串|<br>
 **     optional uint32 Terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
 **     optional uint32 Errnum = 6;     // M|错误码|数字|<br>
 **     optional string Errmsg = 7;     // M|错误描述|字串|<br>
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *OlsvrCntx) send_err_online_ack(
	head *comm.MesgHeader, req *mesg.MesgOnlineReq, errno uint32, errmsg string) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgOnlineAck{
		Uid:      proto.Uint64(req.GetUid()),
		Sid:      proto.Uint64(0),
		App:      proto.String(req.GetApp()),
		Version:  proto.String(req.GetVersion()),
		Terminal: proto.Uint32(req.GetTerminal()),
		ErrNum:   proto.Uint32(errno),
		ErrMsg:   proto.String(errmsg),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+len(body))

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.Send(comm.CMD_ONLINE_ACK, p.Buff, uint32(len(p.Buff)))

	return 0

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
 **     optional uint64 Uid = 1;        // M|用户ID|数字|<br>
 **     optional uint64 Sid = 2;        // M|会话ID|数字|内部使用<br>
 **     optional string App = 3;        // M|APP名|字串|<br>
 **     optional string Version = 4;    // M|APP版本|字串|<br>
 **     optional uint32 Terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
 **     optional uint32 Errnum = 6;     // M|错误码|数字|<br>
 **     optional string Errmsg = 7;     // M|错误描述|字串|<br>
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *OlsvrCntx) send_online_ack(sid uint64, head *comm.MesgHeader, req *mesg.MesgOnlineReq) int {
	/* > 设置协议体 */
	rsp := &mesg.MesgOnlineAck{
		Uid:      proto.Uint64(req.GetUid()),
		Sid:      proto.Uint64(sid),
		App:      proto.String(req.GetApp()),
		Version:  proto.String(req.GetVersion()),
		Terminal: proto.Uint32(req.GetTerminal()),
		ErrNum:   proto.Uint32(0),
		ErrMsg:   proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, binary.Size(comm.MesgHeader{})+len(body))

	comm.MesgHeadHton(head, p)
	copy(p.Buff[binary.Size(comm.MesgHeader{}):], body)

	/* > 发送协议包 */
	ctx.proxy.Send(comm.CMD_ONLINE_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: online_update
 **功    能: 更新数据库
 **输入参数: NONE
 **输出参数: NONE
 **返    回: true:成功 false:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 19:17:01 #
 ******************************************************************************/
func (ctx *OlsvrCntx) online_update(head *comm.MesgHeader, sid uint64) bool {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ttl := time.Now().Unix() + comm.CHAT_SID_TTL
	pl.Send("ZADD", comm.CHAT_KEY_SID_ZSET, ttl, sid)

	return true
}

/******************************************************************************
 **函数名称: online_handler
 **功    能: 上线处理
 **输入参数:
 **     req: 上线请求
 **输出参数: NONE
 **返    回:
 **     sid: 会话ID
 **     err: 异常信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 21:12:36 #
 ******************************************************************************/
func (ctx *OlsvrCntx) online_handler(head *comm.MesgHeader, req *mesg.MesgOnlineReq) (sid uint64, err error) {
	/* > 申请会话SID */
	sid, err = ctx.alloc_sid()
	if nil != err {
		return 0, err
	}

	/* > 更新数据库 */
	if !ctx.online_update(head, sid) {
		return 0, errors.New("Update database failed!")
	}

	return sid, err
}

/******************************************************************************
 **函数名称: OlsvrMesgOnlineReqHandler
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
 **注意事项: 首先需要调用mesg_head_ntoh()对头部数据进行直接序转换.
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func OlsvrMesgOnlineReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	/* 1. > 解析上线请求 */
	head, req := ctx.online_parse(data)
	if nil == head || nil != req {
		ctx.log.Error("Parse online request failed!")
		ctx.send_err_online_ack(head, req, comm.ERR_SVR_PARSE_PARAM, "Parse online request failed!")
		return -1
	}

	/* 2. > 初始化上线环境 */
	sid, err := ctx.online_handler(head, req)
	if nil != err {
		ctx.log.Error("Online handler failed!")
		ctx.send_err_online_ack(head, req, comm.ERR_SYS_SYSTEM, err.Error())
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.send_online_ack(sid, head, req)

	return 0
}

/******************************************************************************
 **函数名称: OlsvrMesgOfflineReqHandler
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
func OlsvrMesgOfflineReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	return 0
}

/******************************************************************************
 **函数名称: OlsvrMesgJoinReqHandler
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
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func OlsvrMesgJoinReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	return 0
}

/******************************************************************************
 **函数名称: OlsvrMesgQuitReqHandler
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
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func OlsvrMesgQuitReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	return 0
}

/******************************************************************************
 **函数名称: OlsvrMesgPingHandler
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
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func OlsvrMesgPingHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	return 0
}

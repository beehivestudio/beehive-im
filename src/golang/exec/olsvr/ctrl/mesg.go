package ctrl

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/mesg/online"
)

/******************************************************************************
 **函数名称: MesgOnlineReqParse
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
func (ctx *OlsvrCntx) MesgOnlineReqParse(data []byte) (head *comm.MesgHeader, req *mesg_online.MesgOnlineReq) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if comm.CMD_ONLINE_REQ != head.GetCmd() {
		ctx.log.Error("Command type isn't right! cmd:%d", head.GetCmd())
		return nil, nil
	}

	/* > 解析PB协议 */
	req = &mesg_online.MesgOnlineReq{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal online request failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */

	return head, req
}

/******************************************************************************
 **函数名称: SendErrOnlineAck
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 ** {
 **     optional uint64 Uid = 1;        // M|用户ID|数字|<br>
 **     optional uint64 Cid = 2;        // M|连接ID|数字|内部使用<br>
 **     optional string App = 3;        // M|APP名|字串|<br>
 **     optional string Version = 4;    // M|APP版本|字串|<br>
 **     optional uint32 Terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
 **     optional uint32 Errnum = 6;     // M|错误码|数字|<br>
 **     optional string Errmsg = 7;     // M|错误描述|字串|<br>
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *OlsvrCntx) SendErrOnlineAck(errno uint32, errmsg string) int {
	rsp := &mesg_online.MesgOnlineAck{}

	rsp.Errnum = &errno
	rsp.Errmsg = &errmsg

	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.proxy.Send(comm.CMD_ONLINE_ACK, body, uint32(len(body)))

	return 0
}

/******************************************************************************
 **函数名称: SendOnlineAck
 **功    能: 发送上线应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 ** {
 **     optional uint64 Uid = 1;        // M|用户ID|数字|<br>
 **     optional uint64 Cid = 2;        // M|连接ID|数字|内部使用<br>
 **     optional string App = 3;        // M|APP名|字串|<br>
 **     optional string Version = 4;    // M|APP版本|字串|<br>
 **     optional uint32 Terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
 **     optional uint32 Errnum = 6;     // M|错误码|数字|<br>
 **     optional string Errmsg = 7;     // M|错误描述|字串|<br>
 ** }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.01 18:37:59 #
 ******************************************************************************/
func (ctx *OlsvrCntx) SendOnlineAck(sid uint64, head *comm.MesgHeader, req *mesg_online.MesgOnlineReq) int {
	var errno uint32 = 0
	errmsg := "Ok"
	cid := head.Sid
	/* > 设置协议头 */
	head.Sid = sid

	/* > 设置协议体 */
	rsp := &mesg_online.MesgOnlineAck{}

	rsp.Uid = req.Uid
	rsp.Cid = &cid
	rsp.App = req.App
	rsp.Version = req.Version
	rsp.Terminal = req.Terminal
	rsp.Errnum = &errno
	rsp.Errmsg = &errmsg

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
 **函数名称: OnlineHandler
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
func (ctx *OlsvrCntx) OnlineHandler(req *mesg_online.MesgOnlineReq) (sid uint64, err error) {
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
	head, req := ctx.MesgOnlineReqParse(data)
	if nil == head || nil != req {
		ctx.log.Error("Parse online request failed!")
		ctx.SendErrOnlineAck(comm.ERR_SVR_PARSE_PARAM, "Parse online request failed!")
		return -1
	}

	/* 2. > 初始化上线环境 */
	sid, err := ctx.OnlineHandler(req)
	if nil != err {
		ctx.log.Error("Online handler failed!")
		ctx.SendErrOnlineAck(comm.ERR_SYS_SYSTEM, err.Error())
		return -1
	}

	/* 3. > 发送上线应答 */
	ctx.SendOnlineAck(sid, head, req)

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

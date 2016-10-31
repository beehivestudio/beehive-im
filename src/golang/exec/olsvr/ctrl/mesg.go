package ctrl

import (
	"github.com/golang/protobuf/proto"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/mesg/online"
)

/******************************************************************************
 **函数名称: OlsvrMesgOnlineReqParse
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

	return head, req
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
 **     1. 解析上线协议合法性
 **     2. 验证数据合法性
 **     3. 申请会话SID
 **注意事项: 首先需要调用mesg_head_ntoh()对头部数据进行直接序转换.
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func OlsvrMesgOnlineReqHandler(cmd uint32, orig uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*OlsvrCntx)
	if false == ok {
		return -1
	}

	head, req := ctx.MesgOnlineReqParse(data)
	if nil == head || nil != req {
		ctx.log.Error("Parse online request failed!")
		return -1
	}

	ctx.log.Debug("Recv online request!")

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

package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndOnlineReqHandler
 **功    能: ONLINE消息的处理
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将ONLINE请求转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 23:10:58 #
 ******************************************************************************/
func LsndOnlineReqHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv online request!")

	/* > 验证当前状态 */
	if CONN_STATUS_READY != conn.status &&
		CONN_STATUS_CHECK != conn.status {
		ctx.log.Error("Drop online request! cid:%d sid:%d status:%d",
			conn.cid, conn.sid, conn.status)
		return 0
	}

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	head.SetSid(conn.cid)
	head.SetNid(ctx.get_nid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), conn.cid)

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	/* > 更新连接状态 */
	conn.status = CONN_STATUS_CHECK

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndPingHandler
 **功    能: PING处理
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 回复PONG, 并转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 23:40:55 #
 ******************************************************************************/
func LsndPingHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv broadcast ack!")

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.get_nid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	/* > 回复PONG */
	head.SetCmd(comm.CMD_PONG)

	ctx.lws.AsyncSend(conn.cid, []byte(head))

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndSubReqHandler
 **功    能: 订阅请求处理
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 直接转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 21:56:56 #
 ******************************************************************************/
func LsndSubReqHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv sub request!")

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.get_nid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndUnsubReqHandler
 **功    能: 取消订阅处理
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 移除订阅消息, 再转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 22:58:12 #
 ******************************************************************************/
func LsndUnsubReqHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv unsub request!")

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.get_nid())

	/* > 移除订阅消息 */
	req := &mesg.MesgUnsubReq{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal unsub request failed! errmsg:%s", err.Error())
		return -1
	}

	ctx.chat.SubDel(head.GetSid(), req.GetSub()) /* 移除订阅消息 */

	/* > 转发给上游模块 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换 */

	ctx.frwder.AsyncSend(cmd, data, length) /* 转发给上游 */

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndMesgForwardHandler
 **功    能: 直接转发给上游模块
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 直接将消息转发给上游模块
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.04 22:49:17 #
 ******************************************************************************/
func LsndMesgForwardHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.get_nid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

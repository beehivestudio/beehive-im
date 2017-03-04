package controllers

import (
	_ "encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	_ "time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
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
 **作    者: # Qifeng.zou # 2017.02.20 23:10:58 #
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
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndBcAckHandler
 **功    能: 广播消息应答处理(待商议)
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
 **作    者: # Qifeng.zou # 2017.03.04 23:40:55 #
 ******************************************************************************/
func LsndBcAckHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
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

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndP2pMsgHandler
 **功    能: 点到点消息的处理
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
 **作    者: # Qifeng.zou # 2016.11.09 21:56:56 #
 ******************************************************************************/
func LsndP2pMsgHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv p2p message!")

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
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndP2pMsgAckHandler
 **功    能: 点到点应答的处理
 **输入参数:
 **     conn: 连接数据
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将离线消息从离线队列中删除.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:58:12 #
 ******************************************************************************/
func LsndP2pMsgAckHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv p2p message ack!")

	/* > 网络->主机字节序 */
	head := comm.MesgHeadNtoh(data)

	head.SetNid(ctx.get_nid())

	/* > 主机->网络字节序 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p)

	/* > 转发给上游模块 */
	ctx.frwder.AsyncSend(cmd, data, length)

	ctx.log.Debug("Header data! cmd:0x%04X sid:%d", head.GetCmd(), head.GetSid())

}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndSyncHandler
 **功    能: 同步请求的处理
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
 **作    者: # Qifeng.zou # 2017.01.14 22:49:17 #
 ******************************************************************************/
func LsndSyncHandler(conn *LsndConnExtra, cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv sync request!")

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
////////////////////////////////////////////////////////////////////////////////

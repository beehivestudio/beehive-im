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

/******************************************************************************
 **函数名称: DownlinkRegister
 **功    能: 下行消息回调注册
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func (ctx *LsndCntx) DownlinkRegister() {
	/* > 未知消息 */
	ctx.frwder.Register(comm.CMD_UNKNOWN, LsndDownlinkCommHandler, ctx)

	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndDownlinkOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_SUB_ACK, LsndDownlinkSubAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_UNSUB_ACK, LsndDownlinkUnsubAckHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CHAT, LsndDownlinkGroupChatHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndDownlinkRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndDownlinkRoomBcHandler, ctx)
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkCommHandler
 **功    能: 通用消息处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:54:26 #
 ******************************************************************************/
func LsndDownlinkCommHandler(cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv command [%d]!", cmd)

	/* > 验证合法性 */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Mesg head is invalid!")
		return -1
	}

	/* > 下发私聊消息 */
	cid := ctx.find_cid_by_sid(head.GetSid())
	if 0 == cid {
		ctx.log.Error("Get cid by sid failed! sid:%d", head.GetSid())
		return -1
	}

	ctx.lws.AsyncSend(cid, data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: LsndDownlinkOnlineAckHandler
 **功    能: ONLINE-ACK消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.07 23:49:03 #
 ******************************************************************************/
func LsndDownlinkOnlineAckHandler(cmd uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv online ack!")

	/* > 字节序转换(网络 -> 主机) */
	head := comm.MesgHeadNtoh(data)
	if !head.IsValid() {
		ctx.log.Error("Header of online-ack is failed!")
		return -1
	}

	cid := head.GetCid()

	/* > 消息ONLINE-ACK的处理 */
	ack = &mesg.MesgOnlineAck{}

	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], ack) /* 解析报体 */
	if nil != err {
		ctx.log.Error("Unmarshal online-ack failed! errmsg:%s", err.Error())
		ctx.lws.Kick(cid)
		return -1
	} else if 0 != ack.GetCode() {
		ctx.log.Error("Logon failed! cid:%d sid:%d code:%d errmsg:%s",
			head.GetCid(), ack.GetSid(), ack.GetCode(), ack.GetErrmsg())
		ctx.lws.Kick(cid)
		return 0
	}

	head.SetSid(ack.GetSid())

	/* > 获取&更新会话状态 */
	param := ctx.chat.SessionGetParam(ack.GetSid())
	if nil == param {
		ctx.log.Error("Didn't find session data! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		ctx.lws.Kick(cid)
		return -1
	}

	session, ok := param.(*LsndSessionExtra)
	if !ok {
		ctx.log.Error("Convert session extra failed! cid:%d sid:%d", head.GetCid(), ack.GetSid())
		ctx.lws.Kick(cid)
		return -1
	}

	session.SetStatus(CONN_STATUS_LOGON) /* 已登录 */

	/* > 下发ONLINE-ACK消息 */
	p := &comm.MesgPacket{Buff: data}

	comm.MesgHeadHton(head, p) /* 字节序转换(主机 - > 网络) */

	ctx.lws.AsyncSend(cid, data)

	return 0
}

////////////////////////////////////////////////////////////////////////////////

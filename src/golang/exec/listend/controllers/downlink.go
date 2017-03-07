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
 **函数名称: LsndP2pMsgHandler
 **功    能: 点到点消息的处理
 **输入参数:
 **     cmd: 消息类型
 **     orig: 帧听层ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断点到点消息的合法性. 如果不合法, 则直接回复错误应答; 如果正常的话, 则
 **        进行进行第2步的处理.
 **     2. 将消息放入离线队列
 **     3. 将消息发送给在线的人员.
 **     4. 回复发送成功应答给发送方.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 21:56:56 #
 ******************************************************************************/
func LsndP2pMsgHandler(cmd uint32, dest uint32,
	data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv group msg ack!")

	return 0
}

////////////////////////////////////////////////////////////////////////////////

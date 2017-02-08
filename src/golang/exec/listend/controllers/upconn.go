package controllers

import (
	"fmt"
)

/******************************************************************************
 **函数名称: frwder_register
 **功    能: 注册处理回调
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 注册回调函数
 **注意事项: 请在调用Launch()前完成此函数调用
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *LsndCntx) frwder_register() {
	/* > 通用消息 */
	ctx.frwder.Register(comm.CMD_ONLINE_ACK, LsndUpconnOnlineAckHandler, ctx)
	ctx.frwder.Register(comm.CMD_OFFLINE_ACK, LsndUpconnP2pMsgHandler, ctx)
	ctx.frwder.Register(comm.CMD_SYNC_ACK, LsndSUpconnyncAckHandler, ctx)

	/* > 私聊消息 */
	ctx.frwder.Register(comm.CMD_CHAT, LsndUpconnChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_CHAT_ACK, LsndUpconnChatAckHandler, ctx)

	/* > 群聊消息 */
	ctx.frwder.Register(comm.CMD_GROUP_CHAT, LsndUpconnGroupChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_GROUP_CHAT_ACK, LsndUpconnGroupChatAckHandler, ctx)

	/* > 聊天室消息 */
	ctx.frwder.Register(comm.CMD_ROOM_CHAT, LsndUpconnRoomChatHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_CHAT_ACK, LsndUpconnRoomChatAckHandler, ctx)

	ctx.frwder.Register(comm.CMD_ROOM_BC, LsndUpconnRoomBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_ROOM_BC_ACK, LsndUpconnRoomBcAckHandler, ctx)

	/* > 推送消息 */
	ctx.frwder.Register(comm.CMD_BC, LsndUpconnBcHandler, ctx)
	ctx.frwder.Register(comm.CMD_BC_ACK, LsndUpconnBcAckHandler, ctx)
}

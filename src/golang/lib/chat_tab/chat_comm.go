package chat_tab

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

/******************************************************************************
 **函数名称: trav_group
 **功    能: 遍历聊天室指定分组
 **输入参数:
 **     room: 聊天室
 **     group: 聊天室群组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.23 20:26:40 #
 ******************************************************************************/
func (ctx *ChatTab) trav_group(room *ChatRoom, group *ChatGroup, proc ChatTravProcCb, param interface{}) int {
	group.RLock()
	for sid, exist := range group.sid_list {
		ctx.session.RLock()
		ssn, ok := ctx.session[sid]
		if !ok {
			ctx.session.RUnlock()
			continue
		}
		proc(ssn, param)
		ctx.session.RUnlock()
	}
	group.RUnlock()
}

/******************************************************************************
 **函数名称: trav_all_group
 **功    能: 遍历聊天室所有分组
 **输入参数:
 **     room: 聊天室
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.23 20:26:40 #
 ******************************************************************************/
func (ctx *ChatTab) trav_all_group(room *ChatRoom, proc ChatTravProcCb, param interface{}) int {
	for gid, group := range room.group {
		ctx.trav_group(room, group, proc, param)
	}
}

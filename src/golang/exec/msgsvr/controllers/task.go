package controllers

import (
	"time"

	"beehive-im/src/golang/lib/chat"
)

/******************************************************************************
 **函数名称: update
 **功    能: 启动update服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 01:00:10 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) update() {
	for {
		ctx.update_rid_to_nid_map()
		ctx.update_gid_to_nid_map()

		time.Sleep(5 * time.Second)
	}
}

/******************************************************************************
 **函数名称: update_rid_to_nid_map
 **功    能: 更新聊天室RID->NID映射表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.05 00:21:54 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) update_rid_to_nid_map() {
	m, err := chat.RoomGetRidToNidMap(ctx.redis)
	if nil != err {
		ctx.log.Error("Get rid to nid map failed! errmsg:%s", err.Error())
		return
	}

	ctx.rid_to_nid_map.Lock()
	ctx.rid_to_nid_map.m = m
	ctx.rid_to_nid_map.Unlock()
}

/******************************************************************************
 **函数名称: update_gid_to_nid_map
 **功    能: 更新群GID->NID映射表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.08 23:15:29 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) update_gid_to_nid_map() {
	m, err := chat.GroupGetGidToNidMap(ctx.redis)
	if nil != err {
		ctx.log.Error("Get gid to nid map failed! errmsg:%s", err.Error())
		return
	}

	ctx.gid_to_nid_map.Lock()
	ctx.gid_to_nid_map.m = m
	ctx.gid_to_nid_map.Unlock()
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task
 **功    能: 启动定时任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 依次启动私聊、群聊、聊天室的存储任务协程
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.27 11:43:03 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) task() {
	go ctx.group_mesg_storage_task()
	go ctx.private_mesg_storage_task()

	go ctx.room_mesg_storage_task()
	go ctx.room_mesg_queue_clean_task()
}

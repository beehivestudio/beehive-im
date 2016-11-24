package controllers

import (
	"time"

	_ "github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/chat"
	_ "chat/src/golang/lib/comm"
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

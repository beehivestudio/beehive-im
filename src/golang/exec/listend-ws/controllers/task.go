package controllers

import (
	"time"
)

/******************************************************************************
 **函数名称: Task
 **功    能: 启动定时任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 启动定时任务
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.09 15:21:36 #
 ******************************************************************************/
func (ctx *LsndCntx) Task() {
	go ctx.task_kick_handler()
}

/******************************************************************************
 **函数名称: task_kick_handler
 **功    能: 指定定时踢除操作
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 获取kick_list中的数据, 并执行kick操作.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.13 00:51:38 #
 ******************************************************************************/
func (ctx *LsndCntx) task_kick_handler() {
	for {
		item, ok := <-ctx.kick_list
		if !ok {
			ctx.log.Error("Kick list was closed!")
			return
		}

		ctm := time.Now().Unix()
		if item.ttl <= ctm {
			ctx.lws.Kick(item.cid)
			ctx.log.Error("Kick connection! cid:%d ctm:%d ttl:%d", item.cid, ctm, item.ttl)
			continue
		}

		diff := time.Duration(item.ttl - ctm)
		time.Sleep(diff * time.Second)

		ctx.lws.Kick(item.cid) /* 执行踢除操作 */

		ctx.log.Error("Kick connection! cid:%d ctm:%d ttl:%d diff:%d", item.cid, ctm, item.ttl, diff)
	}
}

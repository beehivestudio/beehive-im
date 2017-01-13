package controllers

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/chat"
	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: start_task
 **功    能: 开启定时任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 08:20:07 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) start_task() {
	for {
		ctx.update_lsn_list()                      // 更新侦听层列表
		chat.RoomSendUsrNum(ctx.frwder, ctx.redis) // 下发聊天室人数

		time.Sleep(time.Second)
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: update_lsn_list
 **功    能: 更新侦听层列表
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:11:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) update_lsn_list() {
	list := ctx.get_lsn_list()
	if nil == list {
		ctx.log.Error("Get listen list failed!")
		return
	}

	ctx.lsnlist.Lock()
	defer ctx.lsnlist.Unlock()

	ctx.lsnlist.list = list
}

/******************************************************************************
 **函数名称: get_lsn_list
 **功    能: 更新侦听层列表
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: 侦听层IP列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:09:55 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) get_lsn_list() map[string](map[string][]string) {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()
	list := make(map[string](map[string][]string))

	/* > 获取"国家/地区"列表 */
	nations, err := redis.Strings(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_LSN_NATION_ZSET, ctm, "+inf"))
	if nil != err {
		ctx.log.Error("Get nation list failed! errmsg:%s", err.Error())
		return nil
	}

	nation_num := len(nations)
	for m := 0; m < nation_num; m += 1 {
		ctx.log.Debug("Nation:%s", nations[m])
		/* > 获取"国家/地区"对应的"运营商"列表 */
		key := fmt.Sprintf(comm.IM_KEY_LSN_OP_ZSET, nations[m])
		operators, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
		if nil != err {
			ctx.log.Error("Get operator list by nation failed! errmsg:%s", err.Error())
			return nil
		}

		operator_set := make(map[string][]string, 0)

		operator_num := len(operators)
		for n := 0; n < operator_num; n += 1 {
			ctx.log.Debug("    Operator:%s", operators[n])
			/* > 获取"运营商"对应的"IP+PORT"列表 */
			key := fmt.Sprintf(comm.IM_KEY_LSN_IP_ZSET, nations[m], operators[n])
			iplist, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
			if nil != err {
				ctx.log.Error("Get operator list by nation failed! errmsg:%s", err.Error())
				return nil
			}

			iplist_num := len(iplist)
			for k := 0; k < iplist_num; k += 1 {
				ctx.log.Debug("    iplist:%s", iplist[k])
				operator_set[operators[n]] = append(operator_set[operators[n]], iplist[k])
			}
		}

		list[nations[m]] = operator_set
	}

	return list
}

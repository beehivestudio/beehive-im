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
		ctx.listend_update()                       // 更新侦听层列表
		chat.RoomSendUsrNum(ctx.frwder, ctx.redis) // 下发聊天室人数

		time.Sleep(time.Second)
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: listend_update
 **功    能: 更新侦听层列表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: NONE
 **实现描述:
 **     1. 获取所有侦听层网络类型.
 **     2. 根据侦听层网络类型获取侦听层的列表.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:11:08 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) listend_update() {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 获取"网络类型"列表 */
	network, err := redis.Ints(rds.Do("ZRANGEBYSCORE",
		comm.IM_KEY_LSND_NETWORK_ZSET, ctm, "+inf"))
	if nil != err {
		ctx.log.Error("Get network type list failed! errmsg:%s", err.Error())
		return
	}

	ctx.listend.Lock()
	defer ctx.listend.Unlock()

	num := len(network)
	for idx := 0; idx < num; idx += 1 {
		typ := network[idx]

		list := ctx.listend_fetch(typ)
		if nil == list {
			ctx.log.Error("Get listen list failed! network:%d", typ)
			delete(ctx.listend.network, typ)
			continue
		}

		ctx.listend.network[typ] = list
	}
}

/******************************************************************************
 **函数名称: listend_fetch
 **功    能: 获取侦听层列表
 **输入参数:
 **     typ: 网络类型(0:Unkonwn 1:TCP 2:WS)
 **输出参数: NONE
 **返    回: 侦听层IP列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:09:55 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) listend_fetch(typ int) *UsrSvrLsndList {
	rds := ctx.redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	lsnd := &UsrSvrLsndList{
		list: make(map[string](map[string][]string)),
	}

	/* > 获取"国家/地区"列表 */
	key := fmt.Sprintf(comm.IM_KEY_LSND_NATION_ZSET, typ)

	nations, err := redis.Strings(rds.Do("ZRANGEBYSCORE", key, ctm, "+inf"))
	if nil != err {
		ctx.log.Error("Get nation list failed! errmsg:%s", err.Error())
		return nil
	}

	nation_num := len(nations)
	for m := 0; m < nation_num; m += 1 {
		ctx.log.Debug("Nation:%s", nations[m])
		/* > 获取"国家/地区"对应的"运营商"列表 */
		key := fmt.Sprintf(comm.IM_KEY_LSND_OP_ZSET, typ, nations[m])

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
			key := fmt.Sprintf(comm.IM_KEY_LSND_IP_ZSET, typ, nations[m], operators[n])

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

		lsnd.list[nations[m]] = operator_set
	}

	return lsnd
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: get_lsn_id_list
 **功    能: 更新侦听层ID列表
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回: 侦听层IP列表
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 00:09:55 #
 ******************************************************************************/
func get_lsn_id_list() {
}

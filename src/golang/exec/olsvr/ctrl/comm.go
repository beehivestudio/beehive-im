package ctrl

import (
	"chat/src/golang/lib/comm"
	"github.com/garyburd/redigo/redis"
)

/******************************************************************************
 **函数名称: alloc_sid
 **功    能: 申请会话SID
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     sid: 会话SID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 10:48:40 #
 ******************************************************************************/
func (ctx *OlsvrCntx) alloc_sid() (sid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	return redis.Uint64(rds.Do("INCRBY", comm.CHAT_KEY_SID_INCR, 1))
}

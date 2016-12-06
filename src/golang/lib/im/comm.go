package im

import (
	"errors"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: AllocSid
 **功    能: 申请会话SID
 **输入参数:
 **     pool: Redis连接池
 **输出参数: NONE
 **返    回:
 **     sid: 会话SID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 10:48:40 #
 ******************************************************************************/
func AllocSid(pool *redis.Pool) (sid uint64, err error) {
	rds := pool.Get()
	defer rds.Close()

	for {
		sid, err := redis.Uint64(rds.Do("INCRBY", comm.IM_KEY_SID_INCR, 1))
		if nil != err {
			return 0, err
		} else if 0 == sid {
			continue
		}
		return sid, nil
	}

	return 0, errors.New("Alloc sid failed!")
}

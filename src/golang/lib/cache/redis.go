package cache

import (
	"github.com/garyburd/redigo/redis"
)

/******************************************************************************
 **函数名称: CreateRedisPool
 **功    能: 创建连接池
 **输入参数:
 **     addr: IP地址
 **     passwd: 登录密码
 **     max_idle: 最大空闲连接数
 **     max_active: 最大激活连接数
 **输出参数: NONE
 **返    回:
 **     pool: 连接池对象
 **实现描述:
 **注意事项:
 **     1. 如果max配置过小, 可能会出现连接池耗尽的情况.
 **     2. 如果idle配置过小, 可能会出现大量'TIMEWAIT'的TCP状态.
 **作    者: # Qifeng.zou # 2017.03.30 22:18:34 #
 ******************************************************************************/
func CreateRedisPool(addr string, passwd string, max_idle int) *redis.Pool {
	return &redis.Pool{
		MaxIdle: max_idle,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if nil != err {
				panic(err.Error())
				return nil, err
			}

			if 0 != len(passwd) {
				if _, err := c.Do("AUTH", passwd); nil != err {
					c.Close()
					panic(err.Error())
					return nil, err
				}
			}
			return c, err
		},
	}
}

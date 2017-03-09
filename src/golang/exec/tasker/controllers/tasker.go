package controllers

import (
	"errors"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rtmq"

	"beehive-im/src/golang/exec/tasker/controllers/conf"
)

/* Tasker上下文 */
type TaskerCntx struct {
	conf   *conf.TaskerConf    /* 配置信息 */
	log    *logs.BeeLogger     /* 日志对象 */
	frwder *rtmq.RtmqProxyCntx /* 代理对象 */
	redis  *redis.Pool         /* REDIS连接池 */
}

/******************************************************************************
 **函数名称: TaskerInit
 **功    能: 初始化对象
 **输入参数:
 **     conf: 配置信息
 **输出参数: NONE
 **返    回:
 **     ctx: 上下文
 **     err: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func TaskerInit(conf *conf.TaskerConf) (ctx *TaskerCntx, err error) {
	ctx = &TaskerCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "tasker.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.Redis.Addr)
			if nil != err {
				panic(err.Error())
				return nil, err
			}
			if 0 != len(conf.Redis.Passwd) {
				if _, err := c.Do("AUTH", conf.Redis.Passwd); nil != err {
					c.Close()
					panic(err.Error())
					return nil, err
				}
			}
			return c, err
		},
	}
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s passwd:%s",
			conf.Redis.Addr, conf.Redis.Passwd)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	return ctx, nil
}

/******************************************************************************
 **函数名称: Register
 **功    能: 注册处理回调
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 注册回调函数
 **注意事项: 请在调用Launch()前完成此函数调用
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *TaskerCntx) Register() {
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动Tasker服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *TaskerCntx) Launch() {
	go ctx.timer_clean()
	go ctx.timer_update()

	ctx.frwder.Launch()
}

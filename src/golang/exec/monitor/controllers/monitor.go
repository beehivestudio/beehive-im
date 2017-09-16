package controllers

import (
	"errors"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rdb"
	"beehive-im/src/golang/lib/rtmq"

	"beehive-im/src/golang/exec/monitor/controllers/conf"
)

/* MONITOR上下文 */
type MonSvrCntx struct {
	conf   *conf.MonConf   /* 配置信息 */
	log    *logs.BeeLogger /* 日志对象 */
	frwder *rtmq.Proxy     /* 代理对象 */
	redis  *redis.Pool     /* REDIS连接池 */
}

/******************************************************************************
 **函数名称: MonInit
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
func MonInit(conf *conf.MonConf) (ctx *MonSvrCntx, err error) {
	ctx = &MonSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "monitor.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = rdb.CreatePool(conf.Redis.Addr, conf.Redis.Passwd, 512)
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s passwd:%s",
			conf.Redis.Addr, conf.Redis.Passwd)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.Frwder, ctx.log)
	if nil == ctx.frwder {
		ctx.log.Error("Init rtmq proxy failed! addr:%s", conf.Frwder.RemoteAddr)
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
func (ctx *MonSvrCntx) Register() {
	/* > 运维消息 */
	ctx.frwder.Register(comm.CMD_LSND_INFO, MonLsndInfoHandler, ctx)
	ctx.frwder.Register(comm.CMD_FRWD_INFO, MonFrwdInfoHandler, ctx)
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动OLSVR服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *MonSvrCntx) Launch() {
	ctx.frwder.Launch()
}

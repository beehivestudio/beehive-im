package ctrl

import (
	"errors"
	"net/http"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/log"
	"chat/src/golang/lib/rtmq"
)

/* OLS上下文 */
type HttpSvrCntx struct {
	conf   *HttpSvrConf        /* 配置信息 */
	log    *logs.BeeLogger     /* 日志对象 */
	frwder *rtmq.RtmqProxyCntx /* 代理对象 */
	redis  *redis.Pool         /* REDIS连接池 */
}

var httsvr *HttpSvrCntx

func GetHttpCtx() *HttpSvrCntx {
	return httpsvr
}

func SetHttpCtx(ctx *HttpSvrCntx) {
	httpsvr = ctx
}

/******************************************************************************
 **函数名称: HttpSvrInit
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
func HttpSvrInit(conf *HttpSvrConf) (ctx *HttpSvrCntx, err error) {
	ctx = &HttpSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "httpsvr.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.RedisAddr)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s", conf.RedisAddr)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > 初始化RTMQ-PROXY */
	ctx.frwder = rtmq.ProxyInit(&conf.frwder, ctx.log)
	if nil == ctx.frwder {
		return nil, err
	}

	SetHttpCtx(ctx)

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
 **作    者: # Qifeng.zou # 2016.11.20 00:29:41 #
 ******************************************************************************/
func (ctx *HttpSvrCntx) Register() {
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动HTTPSVR服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.20 00:27:03 #
 ******************************************************************************/
func (ctx *HttpSvrCntx) Launch() {
	conf := ctx.conf
	ctx.frwder.Launch()

	ip_port := fmt.Sprintf(":%d", conf.Port)
	http.ListenAndServe(ip_port, nil)
}

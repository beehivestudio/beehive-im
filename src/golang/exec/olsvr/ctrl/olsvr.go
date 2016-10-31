package ctrl

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/logs"

	"chat/src/golang/lib/comm"
	"chat/src/golang/lib/rtmq"
)

/* OLS上下文 */
type OlsvrCntx struct {
	conf  *OlsvrConf          /* 配置信息 */
	log   *logs.BeeLogger     /* 日志对象 */
	proxy *rtmq.RtmqProxyCntx /* 代理对象 */
}

/******************************************************************************
 **函数名称: OlsvrInit
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
func OlsvrInit(conf *OlsvrConf) (ctx *OlsvrCntx, err error) {
	ctx = &OlsvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	if err := ctx.log_init(); nil != err {
		return nil, err
	}

	/* > 初始化RTMQ-PROXY */
	ctx.proxy = rtmq.ProxyInit(&conf.proxy, ctx.log)
	if nil == ctx.proxy {
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
func (ctx *OlsvrCntx) Register() {
	ctx.proxy.Register(comm.CMD_ONLINE_REQ, OlsvrMesgOnlineReqHandler, ctx)
	ctx.proxy.Register(comm.CMD_OFFLINE_REQ, OlsvrMesgOfflineReqHandler, ctx)

	ctx.proxy.Register(comm.CMD_JOIN_REQ, OlsvrMesgJoinReqHandler, ctx)
	ctx.proxy.Register(comm.CMD_QUIT_REQ, OlsvrMesgQuitReqHandler, ctx)

	ctx.proxy.Register(comm.CMD_PING, OlsvrMesgPingHandler, ctx)
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
func (ctx *OlsvrCntx) Launch() {
	ctx.proxy.Launch()
}

/******************************************************************************
 **函数名称: log_init
 **功    能: 初始化日志
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:34:34 #
 ******************************************************************************/
func (ctx *OlsvrCntx) log_init() (err error) {
	conf := ctx.conf

	ctx.log = logs.NewLogger(20000)
	log := ctx.log

	err = os.Mkdir("../log", 0755)
	if nil != err && false == os.IsExist(err) {
		log.Emergency(err.Error())
		return err
	}

	log.SetLogger("file", fmt.Sprintf(`{"filename":"%s/../log/olsvr.log"}`, conf.AppPath))
	log.SetLevel(logs.LevelDebug)
	return nil
}

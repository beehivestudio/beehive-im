package ctrl

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/logs"

	"chat/src/golang/lib/rtmq"
)

/* OLS上下文 */
type OlSvrCntx struct {
	conf       *OlSvrConf          // 配置信息
	log        *logs.BeeLogger     // 日志对象
	rtmq_proxy *rtmq.RtmqProxyCntx // 代理对象
}

/* 初始化对象 */
func OlSvrInit(conf *OlSvrConf) (ctx *OlSvrCntx, err error) {
	ctx = &OlSvrCntx{}

	ctx.conf = conf

	if err := ctx.log_init(); nil != err {
		return nil, err
	}

	ctx.rtmq_proxy = rtmq.RtmqProxyInit(&conf.rtmq_proxy, ctx.log)
	if nil == ctx.rtmq_proxy {
		return nil, err
	}

	return ctx, nil
}

/* 获取配置对象 */
func (ctx *OlSvrCntx) conf_get() (conf *OlSvrConf) {
	return ctx.conf
}

/* 初始化日志 */
func (ctx *OlSvrCntx) log_init() (err error) {
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

/* 启动OLS服务 */
func (ctx *OlSvrCntx) OlSvrLaunch() {
	/* > 启动RTMQ */
}

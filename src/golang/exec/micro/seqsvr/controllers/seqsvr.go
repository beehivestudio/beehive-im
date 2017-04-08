package controllers

import (
	"database/sql"
	"errors"
	_ "fmt"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"

	"beehive-im/src/golang/lib/cache"
	"beehive-im/src/golang/lib/dbase"
	"beehive-im/src/golang/lib/log"

	"beehive-im/src/golang/exec/micro/seqsvr/controllers/conf"
)

/* SeqSvr上下文 */
type SeqSvrCntx struct {
	conf  *conf.SeqSvrConf /* 配置信息 */
	log   *logs.BeeLogger  /* 日志对象 */
	mysql *sql.DB          /* MYSQL数据库 */
	redis *redis.Pool      /* REDIS连接池 */
}

/******************************************************************************
 **函数名称: SeqSvrInit
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
func SeqSvrInit(conf *conf.SeqSvrConf) (ctx *SeqSvrCntx, err error) {
	ctx = &SeqSvrCntx{}

	ctx.conf = conf

	/* > 初始化日志 */
	ctx.log = log.Init(conf.Log.Level, conf.Log.Path, "seqsvr.log")
	if nil == ctx.log {
		return nil, errors.New("Initialize log failed!")
	}

	/* > REDIS连接池 */
	ctx.redis = cache.CreateRedisPool(conf.Redis.Addr, conf.Redis.Passwd, 512)
	if nil == ctx.redis {
		ctx.log.Error("Create redis pool failed! addr:%s passwd:%s",
			conf.Redis.Addr, conf.Redis.Passwd)
		return nil, errors.New("Create redis pool failed!")
	}

	/* > MYSQL连接池 */
	auth := dbase.MySqlAuthStr(conf.Mysql.Usr, conf.Mysql.Passwd, conf.Mysql.Addr, conf.Mysql.Dbname)

	ctx.mysql, err = sql.Open("mysql", auth)
	if nil != err {
		ctx.log.Error("Connect mysql [%s] failed! errmsg:%s", auth, err.Error())
		return nil, err
	}

	err = ctx.mysql.Ping()
	if nil != err {
		ctx.log.Error("Ping [%s] failed! errmsg:%s", auth, err.Error())
		return nil, err
	}

	ctx.mysql.SetMaxIdleConns(1024)
	ctx.mysql.SetMaxOpenConns(1024)

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
func (ctx *SeqSvrCntx) Register() {
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动SeqSvr服务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.30 22:32:23 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) Launch() {
	go ctx.launch_thrift(ctx.conf.Addr)
}

package controllers

import (
	"database/sql"
	"errors"
	_ "fmt"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"

	"beehive-im/src/golang/lib/dbase"
	"beehive-im/src/golang/lib/log"
	"beehive-im/src/golang/lib/rdb"

	"beehive-im/src/golang/exec/micsvr/seqsvr/controllers/conf"
)

const (
	USER_LIST_LEN    = 100
	SECTION_LIST_LEN = 100000
)

/* SeqSvr上下文 */
type SeqSvrCntx struct {
	conf  *conf.SeqSvrConf /* 配置信息 */
	log   *logs.BeeLogger  /* 日志对象 */
	mysql *sql.DB          /* MYSQL数据库 */
	redis *redis.Pool      /* REDIS连接池 */
	ctrl  SectionCtrl      /* SECTION管理表 */
}

/* 段管理 */
type SectionCtrl struct {
	session [USER_LIST_LEN]SessionList    /* 会话列表 */
	section [SECTION_LIST_LEN]SectionList /* SECTION列表 */
}

/* 段管理列表 */
type SectionList struct {
	sync.RWMutex                         /* 读写锁 */
	item         map[uint64]*SectionItem /* 段信息[通过id查找对应段信息] */
}

type SectionItem struct {
	sync.RWMutex        /* 读写锁 */
	min          uint64 /* 最小序列号 */
	max          uint64 /* 最大序列号 */
}

/* 会话列表 */
type SessionList struct {
	sync.RWMutex                         /* 读写锁 */
	item         map[uint64]*SessionItem /* 会话信息 */
}

type SessionItem struct {
	sync.RWMutex        /* 读写锁 */
	sid          uint64 /* 会话SID */
	seq          uint64 /* 当前序列号 */
	max          uint64 /* 最大序列号(注:与对应SECTION中的MAX同步) */
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

	/* > SECTION管理表 */
	for idx := 0; idx < USER_LIST_LEN; idx += 1 {
		ctx.ctrl.session[idx].item = make(map[uint64]*SessionItem)
	}

	for idx := 0; idx < SECTION_LIST_LEN; idx += 1 {
		ctx.ctrl.section[idx].item = make(map[uint64]*SectionItem)
	}

	/* > REDIS连接池 */
	ctx.redis = rdb.CreatePool(conf.Redis.Addr, conf.Redis.Passwd, 512)
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

	/* > 加载各段序列号 */
	ctx.load_seq_from_db()

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

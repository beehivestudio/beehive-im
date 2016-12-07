#include "mesg.h"
#include "redo.h"
#include "listend.h"
#include "cmd_list.h"

#include "mesg_lsn_rpt.pb-c.h"

/******************************************************************************
 **函数名称: lsnd_getopt 
 **功    能: 解析输入参数
 **输入参数: 
 **     argc: 参数个数
 **     argv: 参数列表
 **输出参数:
 **     opt: 参数选项
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **     1. 解析输入参数
 **     2. 验证输入参数
 **注意事项: 
 **     N: 服务名 - 根据服务名可找到配置路径
 **     h: 帮助手册
 **作    者: # Qifeng.zou # 2014.11.15 #
 ******************************************************************************/
int lsnd_getopt(int argc, char **argv, lsnd_opt_t *opt)
{
    int ch;
    const struct option opts[] = {
        {"help",                    no_argument,        NULL, 'h'}
        , {"daemon",                no_argument,        NULL, 'd'}
        , {"log-level",             required_argument,  NULL, 'l'}
        , {"configuration path",    required_argument,  NULL, 'c'}
        , {NULL,                    0,                  NULL, 0}
    };

    memset(opt, 0, sizeof(lsnd_opt_t));

    opt->isdaemon = false;
    opt->log_level = LOG_LEVEL_TRACE;
    opt->conf_path = "../conf/listend.xml";

    /* 1. 解析输入参数 */
    while (-1 != (ch = getopt_long(argc, argv, "l:c:hd", opts, NULL))) {
        switch (ch) {
            case 'c':   /* 配置路径 */
            {
                opt->conf_path = optarg;
                break;
            }
            case 'l':   /* 日志级别 */
            {
                opt->log_level = log_get_level(optarg);
                break;
            }
            case 'd':
            {
                opt->isdaemon = true;
                break;
            }
            case 'h':   /* 显示帮助信息 */
            default:
            {
                return LSND_SHOW_HELP;
            }
        }
    }

    optarg = NULL;
    optind = 1;

    /* 2. 验证输入参数 */
    if (NULL == opt->conf_path) {
        return LSND_SHOW_HELP;
    }

    return 0;
}

/* 显示启动参数帮助信息 */
int lsnd_usage(const char *exec)
{
    printf("\nUsage: %s -l <log level> -L <log key path> -n <node name> [-h] [-d]\n", exec);
    printf("\t-l: Log level\n"
            "\t-n: Node name\n"
            "\t-d: Run as daemon\n"
            "\t-h: Show help\n\n");
    return 0;
}

/* 初始化日志模块 */
log_cycle_t *lsnd_init_log(char *fname)
{
    char path[FILE_NAME_MAX_LEN];

    log_get_path(path, sizeof(path), basename(fname));

    return log_init(LOG_LEVEL_ERROR, path);
}

/******************************************************************************
 **函数名称: lsnd_acc_reg_add
 **功    能: 添加下游数据的处理注册回调
 **输入参数:
 **     ctx: 全局信息
 **     type: 数据类型
 **     proc: 处理回调
 **     args: 附加参数
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.17 22:18:53 #
 ******************************************************************************/
int lsnd_acc_reg_add(lsnd_cntx_t *ctx, int type, lsnd_reg_cb_t proc, void *args)
{
    lsnd_reg_t *reg;

    reg = (lsnd_reg_t *)calloc(1, sizeof(lsnd_reg_t));
    if (NULL == reg) {
        return -1;
    }

    reg->type = type;
    reg->proc = proc;
    reg->args = args;

    if (avl_insert(ctx->reg, reg)) {
        free(reg);
        return -1;
    }

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_kick_trav_cb
 **功    能: 编历被踢连接是否到达执行踢除的时间
 **输入参数:
 **     item: 被踢连接
 **     timeout_list: 超时记录链表
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 16:24:05 #
 ******************************************************************************/
static int lsnd_kick_trav_cb(lsnd_kick_item_t *item, list_t *timeout_list)
{
    time_t ctm = time(NULL);

    if (ctm < item->ttl) {
        return 0; /* 未超时 */
    }

    list_rpush(timeout_list, (void *)item->cid);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_timer_kick_handler
 **功    能: 定时踢除连接
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 16:24:05 #
 ******************************************************************************/
void lsnd_timer_kick_handler(void *_ctx)
{
    void *addr;
    uint64_t cid;
    list_t *timeout_list;
    lsnd_kick_item_t *item, key;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)_ctx;

    timeout_list = list_creat(NULL);
    if (NULL == timeout_list) {
        log_error(ctx->log, "Initialize kick timeout list failed!");
        return;
    }

    /* > 获取已经的被踢连接CID列表 */
    hash_tab_trav(ctx->conn_kick_list, (trav_cb_t)lsnd_kick_trav_cb, timeout_list, RDLOCK);

    /* > 依次执行踢除连接的操作 */
    while (NULL != (addr = list_lpop(timeout_list))) {
        cid = (uint64_t)addr;

        key.cid = cid;
        item = hash_tab_delete(ctx->conn_kick_list, &key, WRLOCK);
        if (NULL == item) {
            continue;
        }

        free(item);

        acc_async_kick(ctx->access, (uint64_t)cid);
    }

    return;
}

/******************************************************************************
 **函数名称: lsnd_kick_insert
 **功    能: 清理踢人列表中的连接
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 18:05:31 #
 ******************************************************************************/
int lsnd_kick_insert(lsnd_cntx_t *ctx, lsnd_conn_extra_t *conn)
{
    conn->loc |= CHAT_EXTRA_LOC_KICK_TAB;
    conn->kick_ttl = time(NULL) + LSND_KICK_TTL;

    hash_tab_insert(ctx->conn_kick_list, conn, WRLOCK);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 比较回调 */
static int lsnd_task_cmp_cb(lsnd_task_item_t *item1, lsnd_task_item_t *item2)
{
    return item1->seq - item2->seq;
}

/******************************************************************************
 **函数名称: lsnd_task_init
 **功    能: 初始化定时任务表
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 初始化读写锁、创建红黑树
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 22:34:25 #
 ******************************************************************************/
int lsnd_task_init(lsnd_cntx_t *ctx)
{
    lsnd_task_t *task = &ctx->task;

    task->seq = 0;
    pthread_rwlock_init(&task->lock, NULL);
    task->list = rbt_creat(NULL, (cmp_cb_t)lsnd_task_cmp_cb);
    if (NULL == task->list) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_task_add
 **功    能: 增加定时任务
 **输入参数:
 **     ctx: 全局信息
 **     proc: 任务回调
 **     start: 第一次执行的间隔
 **     interval: 执行执行的间隔
 **     param: 附件参数
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 21:57:30 #
 ******************************************************************************/
int lsnd_task_add(lsnd_cntx_t *ctx, void (*proc)(void *param), int start, int interval, int times, void *param)
{
    lsnd_task_item_t *item;
    lsnd_task_t *task = &ctx->task;

    item = (lsnd_task_item_t *)calloc(1, sizeof(lsnd_task_item_t));
    if (NULL == item) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    item->interval = interval;
    item->is_limit = (0 == times)? false : true;
    item->times = times;
    item->last = (0 == start)? 0 : time(NULL);

    item->proc = proc;
    item->param = param;

    /* 插入定时任务表 */
    pthread_rwlock_wrlock(&task->lock);
    item->seq = ++task->seq;
    if (rbt_insert(task->list, item)) {
        pthread_rwlock_unlock(&task->lock);
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        free(item);
        return 0;
    }
    pthread_rwlock_unlock(&task->lock);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_task_exec_cb
 **功    能: 执行定时任务
 **输入参数:
 **     ctx: 全局信息
 **     timeout_list: 超时列表
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 21:57:30 #
 ******************************************************************************/
static int lsnd_task_exec_cb(lsnd_task_item_t *item, list_t *timeout_list)
{
    time_t ctm = time(NULL);

    if ((ctm - item->last) >= item->interval) {
        item->proc(item->param); // 执行任务
        item->last = ctm;
        if (item->is_limit) {
            --item->times;
            if (0 == item->times) {
                list_rpush(timeout_list, (void *)item->seq);
            }
        }
    }

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_task_handler
 **功    能: 定时任务处理
 **输入参数:
 **     _ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 21:57:30 #
 ******************************************************************************/
void *lsnd_task_handler(void *_ctx)
{
    void *addr;
    list_t *timeout_list;
    lsnd_task_item_t *item, key;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)_ctx;
    lsnd_task_t *task = &ctx->task;

    timeout_list = list_creat(NULL);
    if (NULL == timeout_list) {
        log_error(ctx->log, "Create timeout list failed!");
        return (void *)-1;
    }

    for (;;) {
        /* > 执行定时任务 */
        pthread_rwlock_rdlock(&task->lock);
        rbt_trav(task->list, (trav_cb_t)lsnd_task_exec_cb, (void *)timeout_list);
        pthread_rwlock_unlock(&task->lock);

        /* > 清理定时任务 */
        pthread_rwlock_wrlock(&task->lock);
        for (;;) {
            addr = (void *)list_lpop(timeout_list);
            if (NULL == addr) {
                break;
            }

            key.seq = (uint64_t)((uint64_t *)addr);

            rbt_delete(task->list, (void *)&key, (void **)&item);
            if (NULL == item) {
                continue;
            }
            free(item);
        }
        pthread_rwlock_unlock(&task->lock);
        Sleep(1);
    }

    return NULL;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_timer_report_handler
 **功    能: 侦听层定时上报
 **输入参数:
 **     _ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 使用PB协议组装接入层上报数据
 **     {
 **         required uint64 nid = 1;    // M|结点ID|数字|<br>
 **         required string nation = 2; // M|所属国家|字串|<br>
 **         required string name = 3;   // M|运营商名称|字串|<br>
 **         required string ipaddr = 4; // M|IP地址|字串|<br>
 **         required uint32 port = 5;   // M|端口号|数字|<br>
 **     }
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 23:23:51 #
 ******************************************************************************/
void lsnd_timer_report_handler(void *_ctx)
{
    void *addr;
    unsigned int len;
    mesg_header_t *head;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)_ctx;
    lsnd_conf_t *conf = &ctx->conf;
    MesgLsnRpt report = MESG_LSN_RPT__INIT;

    /* > 设置上报数据 */
    report.nid = conf->nid;
    report.nation = conf->operator.nation;
    report.name = conf->operator.name;
    report.ipaddr = conf->access.ipaddr;
    report.port = conf->access.port;

    /* > 组装PB协议 */
    len = mesg_lsn_rpt__get_packed_size(&report);

    addr = (void *)calloc(1, sizeof(mesg_header_t) + len);
    if (NULL == addr) {
        log_error(ctx->log, "Alloc memory failed! errmsg:[%d] %s!", errno, strerror(errno));
        return;
    }

    head = (mesg_header_t *)addr;

    head->type = CMD_LSN_RPT;
    head->flag = MSG_FLAG_USR;
    head->length = len;
    head->chksum = MSG_CHKSUM_VAL;
    head->nid = conf->nid;

    MESG_HEAD_HTON(head, head);

    mesg_lsn_rpt__pack(&report, addr + sizeof(mesg_header_t)); /* 组装PB协议 */

    /* > 发送数据 */
    rtmq_proxy_async_send(ctx->frwder, CMD_LSN_RPT, addr, sizeof(mesg_header_t) + len);

    free(addr);
    return;
}

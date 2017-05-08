/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: listend.c
 ** 版本号: 1.0
 ** 描  述: 帧听层服务
 **         负责接收外界请求，并将处理结果返回给外界
 ** 作  者: # Qifeng.zou # 2016.09.20 #
 ******************************************************************************/

#include "sck.h"
#include "lock.h"
#include "mesg.h"
#include "redo.h"
#include "timer.h"
#include "access.h"
#include "listend.h"
#include "mem_ref.h"
#include "cmd_list.h"
#include "hash_alg.h"
#include "lsnd_mesg.h"

#define LSND_MOD_NAME "listend"

static lsnd_cntx_t *lsnd_init(lsnd_conf_t *conf, log_cycle_t *log);
static int lsnd_launch(lsnd_cntx_t *ctx);
static int lsnd_set_reg(lsnd_cntx_t *ctx);

static size_t lsnd_mesg_body_length(mesg_header_t *head)
{
    return ntohl(head->length);
}

/******************************************************************************
 **函数名称: main
 **功    能: 代理服务
 **输入参数:
 **     argc: 参数个数
 **     argv: 参数列表
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 加载配置，再通过配置启动各模块
 **注意事项:
 **作    者: # Qifeng.zou # 2014.11.15 #
 ******************************************************************************/
int main(int argc, char *argv[])
{
    lsnd_opt_t opt;
    lsnd_conf_t conf;
    log_cycle_t *log;
    lsnd_cntx_t *ctx = NULL;
    char path[FILE_PATH_MAX_LEN];

    /* > 解析输入参数 */
    if (lsnd_getopt(argc, argv, &opt)) {
        return lsnd_usage(argv[0]);
    }
    else if (opt.isdaemon) {
        /* int daemon(int nochdir, int noclose);
         *  1． daemon()函数主要用于希望脱离控制台,以守护进程形式在后台运行的程序.
         *  2． 当nochdir为0时,daemon将更改进程的根目录为root(“/”).
         *  3． 当noclose为0是,daemon将进程的STDIN, STDOUT, STDERR都重定向到/dev/null */
        daemon(1, 1);
    }

    umask(0);
    mem_ref_init();

    do {
        /* > 初始化日志 */
        log_get_path(path, sizeof(path), LSND_MOD_NAME);

        log = log_init(opt.log_level, path);
        if (NULL == log) {
            fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
            goto LSND_INIT_ERR;
        }

        /* > 加载配置信息 */
        if (lsnd_load_conf(opt.conf_path, &conf, log)) {
            fprintf(stderr, "Load configuration failed!\n");
            goto LSND_INIT_ERR;
        }

        /* > 初始化侦听 */
        ctx = lsnd_init(&conf, log);
        if (NULL == ctx) {
            fprintf(stderr, "Initialize lsnd failed!\n");
            goto LSND_INIT_ERR;
        }

        /* > 注册回调函数 */
        if (lsnd_set_reg(ctx)) {
            fprintf(stderr, "Set register callback failed!\n");
            goto LSND_INIT_ERR;
        }

        /* > 启动侦听服务 */
        if (lsnd_launch(ctx)) {
            fprintf(stderr, "Startup search-engine failed!\n");
            goto LSND_INIT_ERR;
        }

        while (1) { pause(); }
    } while(0);

LSND_INIT_ERR:
    Sleep(2);
    return -1;
}

/******************************************************************************
 **函数名称: lsnd_proc_lock
 **功    能: 代理服务进程锁(防止同时启动两个服务进程)
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 使用文件锁
 **注意事项:
 **作    者: # Qifeng.zou # 2014.11.15 #
 ******************************************************************************/
static int lsnd_proc_lock(lsnd_conf_t *conf)
{
    int fd;
    char path[FILE_PATH_MAX_LEN];

    /* 1. 获取路径 */
    snprintf(path, sizeof(path), "%s/lsnd.lck", conf->wdir);

    Mkdir2(path, DIR_MODE);

    /* 2. 打开文件 */
    fd = Open(path, OPEN_FLAGS, OPEN_MODE);
    if (fd < 0) {
        return -1;
    }

    /* 3. 尝试加锁 */
    if (proc_try_wrlock(fd) < 0) {
        CLOSE(fd);
        return -1;
    }

    return 0;
}

/* 注册比较回调 */
static int lsnd_acc_reg_cmp_cb(lsnd_reg_t *reg1, lsnd_reg_t *reg2)
{
    return reg1->type - reg2->type;
}

/* CID哈希回调 */
static uint64_t lsnd_conn_cid_hash_cb(lsnd_conn_extra_t *extra)
{
    return extra->cid;
}

/* CID比较回调 */
static int lsnd_conn_cid_cmp_cb(lsnd_conn_extra_t *extra1, lsnd_conn_extra_t *extra2)
{
    return (int)(extra1->cid - extra2->cid);
}

/* SID哈希回调 */
static uint64_t lsnd_conn_sid_hash_cb(lsnd_conn_extra_t *extra)
{
    return extra->sid;
}

/* SID比较回调 */
static int lsnd_conn_sid_cmp_cb(lsnd_conn_extra_t *extra1, lsnd_conn_extra_t *extra2)
{
    int diff;

    diff = (int)(extra1->sid - extra2->sid);
    if (0 == diff) {
        return (int)(extra1->cid - extra2->cid);
    }
    return diff;
}

/* KICK哈希回调 */
static uint64_t lsnd_conn_kick_hash_cb(lsnd_kick_item_t *item)
{
    return (uint64_t)item->cid;
}

/* KICK比较回调 */
static int lsnd_conn_kick_cmp_cb(lsnd_kick_item_t *item1, lsnd_kick_item_t *item2)
{
    return (int)(item1->cid - item2->cid);
}

/******************************************************************************
 **函数名称: lsnd_init
 **功    能: 初始化进程
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.28 23:11:54 #
 ******************************************************************************/
static lsnd_cntx_t *lsnd_init(lsnd_conf_t *conf, log_cycle_t *log)
{
    lsnd_cntx_t *ctx;
    static acc_protocol_t protocol = {
        lsnd_callback,
        sizeof(mesg_header_t),
        (acc_get_packet_body_size_cb_t)lsnd_mesg_body_length,
        sizeof(lsnd_conn_extra_t)
    };

    /* > 加进程锁 */
    if (lsnd_proc_lock(conf)) {
        log_error(log, "errmsg:[%d] %s!", errno, strerror(errno));
        return NULL;
    }

    /* > 创建全局对象 */
    ctx = (lsnd_cntx_t *)calloc(1, sizeof(lsnd_cntx_t));
    if (NULL == ctx) {
        log_error(log, "errmsg:[%d] %s!", errno, strerror(errno));
        return NULL;
    }

    ctx->log = log;
    memcpy(&ctx->conf, conf, sizeof(lsnd_conf_t));  /* 拷贝配置信息 */

    do {
        /* > 初始化回调注册表 */
        ctx->reg = avl_creat(NULL, (cmp_cb_t)lsnd_acc_reg_cmp_cb);
        if (NULL == ctx->reg) {
            log_error(log, "Initialize register table failed!");
            break;
        }

        /* > 初始化连接列表 */
        ctx->conn_list = hash_tab_creat(LSND_CONN_HASH_TAB_LEN,
                (hash_cb_t)lsnd_conn_sid_hash_cb,
                (cmp_cb_t)lsnd_conn_sid_cmp_cb, NULL);
        if (NULL == ctx->conn_list) {
            log_error(log, "Initialize conn sid table failed!");
            break;
        }

        /* > 初始化KICK管理表 */
        ctx->kick_list = hash_tab_creat(LSND_CONN_HASH_TAB_LEN,
                (hash_cb_t)lsnd_conn_kick_hash_cb,
                (cmp_cb_t)lsnd_conn_kick_cmp_cb, NULL);
        if (NULL == ctx->kick_list) {
            log_error(log, "Initialize conn kick table failed!");
            break;
        }

        /* > 初始化定时任务表 */
        ctx->timer = timer_cntx_init();
        if (NULL == ctx->timer) {
            log_error(log, "Initialize timer failed!");
            break;
        }

        /* > 初始化chat表 */
        ctx->chat_tab = chat_tab_init(1024, log);
        if (NULL == ctx->chat_tab) {
            log_error(log, "Initialize chat table failed!");
            break;
        }

        /* > 初始化RTMQ信息 */
        ctx->frwder = rtmq_proxy_init(&conf->frwder, log);
        if (NULL == ctx->frwder) {
            log_error(log, "Initialize real-time-transport-protocol failed!");
            break;
        }

        /* > 定时任务线程池 */
        ctx->timer_task_tp = thread_pool_init(1, NULL, (void *)ctx);
        if (NULL == ctx->timer_task_tp) {
            log_error(log, "Initialize timer task thread pool failed!");
            break;
        }

        /* > 初始化帧听模块 */
        protocol.args = (void *)ctx;
        ctx->access = acc_init(&protocol, &conf->access, log);
        if (NULL == ctx->access) {
            log_error(log, "Initialize access failed!");
            break;
        }

        return ctx;
    } while (0);

    FREE(ctx);
    return NULL;
}

/******************************************************************************
 **函数名称: lsnd_set_timer
 **功    能: 设置定时任务
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 20:56:24 #
 ******************************************************************************/
static void lsnd_set_timer(lsnd_cntx_t *ctx)
{
    timer_task_init(&ctx->timer_kick, lsnd_timer_kick_handler, 5, 5, (void *)ctx);
    timer_task_init(&ctx->timer_info, lsnd_timer_info_handler, 5, 5, (void *)ctx);

    timer_task_add(ctx->timer, &ctx->timer_kick);
    timer_task_add(ctx->timer, &ctx->timer_info);
}

/******************************************************************************
 **函数名称: lsnd_set_reg
 **功    能: 设置注册函数
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.28 23:11:54 #
 ******************************************************************************/
static int lsnd_set_reg(lsnd_cntx_t *ctx)
{
    /* 注册代理数据回调 */
#define LSND_ACC_REG_CB(lsnd, type, proc, args) \
    if (lsnd_acc_reg_add(lsnd, type, (lsnd_reg_cb_t)proc, (void *)args)) { \
        log_error((lsnd)->log, "Register type [0x%0X] failed!", type); \
        return LSND_ERR; \
    }

    LSND_ACC_REG_CB(ctx, CMD_ONLINE, lsnd_mesg_online_handler, ctx);
    LSND_ACC_REG_CB(ctx, CMD_OFFLINE, lsnd_mesg_offline_handler, ctx);
    LSND_ACC_REG_CB(ctx, CMD_PING, lsnd_mesg_ping_handler, ctx);

    LSND_ACC_REG_CB(ctx, CMD_ROOM_JOIN, lsnd_mesg_room_join_handler, ctx);
    LSND_ACC_REG_CB(ctx, CMD_ROOM_QUIT, lsnd_mesg_room_quit_handler, ctx);

    /* 注册队列数据回调 */
#define LSND_RTQ_REG_CB(lsnd, type, proc, args) \
    if (rtmq_proxy_reg_add((lsnd)->frwder, type, (rtmq_reg_cb_t)proc, (void *)args)) { \
        log_error((lsnd)->log, "Register type [0x%0X] failed!", type); \
        return LSND_ERR; \
    }

    LSND_RTQ_REG_CB(ctx, CMD_ONLINE_ACK, lsnd_mesg_online_ack_handler, ctx);
    LSND_RTQ_REG_CB(ctx, CMD_KICK_REQ, lsnd_mesg_kick_handler, ctx);

    LSND_RTQ_REG_CB(ctx, CMD_ROOM_JOIN_ACK, lsnd_mesg_room_join_ack_handler, ctx);
    LSND_ACC_REG_CB(ctx, CMD_ROOM_CHAT, lsnd_mesg_room_chat_handler, ctx);

    /* 注册定时任务回调 */
    lsnd_set_timer(ctx);

    return LSND_OK;
}

/******************************************************************************
 **函数名称: lsnd_launch
 **功    能: 启动侦听服务
 **输入参数:
 **     ctx: 侦听对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.06.20 22:58:16 #
 ******************************************************************************/
static int lsnd_launch(lsnd_cntx_t *ctx)
{
    /* > 启动代理服务 */
    if (acc_launch(ctx->access)) {
        log_error(ctx->log, "Startup agent failed!");
        return LSND_ERR;
    }

    /* > 启动代理服务 */
    if (rtmq_proxy_launch(ctx->frwder)) {
        log_error(ctx->log, "Startup invertd upstream failed!");
        return LSND_ERR;
    }

    /* > 启动定时任务 */
    if (thread_pool_add_worker(ctx->timer_task_tp, timer_task_routine, ctx->timer)) {
        log_error(ctx->log, "Add timeout handler failed!");
        return LSND_ERR;
    }

    return LSND_OK;
}

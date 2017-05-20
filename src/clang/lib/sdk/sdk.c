#include "sdk.h"
#include "comm.h"
#include "lock.h"
#include "redo.h"
#include "syscall.h"
#include "sdk_mesg.h"

/******************************************************************************
 **函数名称: sdk_init
 **功    能: 初始化发送端
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.19 #
 ******************************************************************************/
sdk_cntx_t *sdk_init(const sdk_conf_t *conf)
{
    sdk_cntx_t *ctx;
    log_cycle_t *log;

    log = log_init(conf->log_level, conf->log_path);
    if (NULL == log) {
        return NULL;
    }

    /* > 创建对象 */
    ctx = (sdk_cntx_t *)calloc(1, sizeof(sdk_cntx_t));
    if (NULL == ctx) {
        log_fatal(log, "errmsg:[%d] %s!", errno, strerror(errno));
        return NULL;
    }

    ctx->log = log;

    /* > 加载配置信息 */
    memcpy(&ctx->conf, conf, sizeof(sdk_conf_t));

    do {
        /* > 锁住指定文件 */
        if (sdk_lock_server(conf)) {
            log_fatal(log, "Lock proxy server failed! errmsg:[%d] %s", errno, strerror(errno));
            break;
        }

        /* > 创建处理映射表 */
        ctx->reg = avl_creat(NULL, (cmp_cb_t)sdk_reg_cmp_cb);
        if (NULL == ctx->reg) {
            log_fatal(log, "Create register map failed!");
            break;
        }

        /* > 请求<->应答 映射表 */
        ctx->cmd = avl_creat(NULL, (cmp_cb_t)sdk_ack_cmp_cb);
        if (NULL == ctx->cmd) {
            log_fatal(log, "Create cmd map failed!");
            break;
        }

        /* > 发送单元管理表 */
        if (sdk_send_mgr_init(ctx)) {
            log_fatal(log, "Send mgr table failed!");
            break;
        }

        /* > 创建通信套接字 */
        if (sdk_creat_cmd_usck(ctx)) {
            log_fatal(log, "Create cmd socket failed!");
            break;
        }

        /* > 创建接收队列 */
        if (sdk_queue_init(&ctx->recvq)) {
            log_fatal(log, "Create recv-queue failed!");
            break;
        }

        /* > 创建发送队列 */
        if (sdk_queue_init(&ctx->sendq)) {
            log_fatal(log, "Create send queue failed!");
            break;
        }

        /* > 创建工作线程池 */
        if (sdk_creat_workers(ctx)) {
            log_fatal(ctx->log, "Create work thread pool failed!");
            break;
        }

        /* > 创建发送线程池 */
        if (sdk_creat_sends(ctx)) {
            log_fatal(ctx->log, "Create send thread pool failed!");
            break;
        }

        return ctx;
    } while(0);

    free(ctx);
    return NULL;
}

/******************************************************************************
 **函数名称: sdk_launch
 **功    能: 启动发送端
 **输入参数:
 **     ctx: 全局信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建工作线程池
 **     2. 创建发送线程池
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
int sdk_launch(sdk_cntx_t *ctx)
{
    int idx;
    sdk_conf_t *conf = &ctx->conf;

    /* > 注册Worker线程回调 */
    for (idx=0; idx<conf->work_thd_num; ++idx) {
        thread_pool_add_worker(ctx->worktp, sdk_worker_routine, ctx);
    }

    /* > 注册Send线程回调 */
    thread_pool_add_worker(ctx->sendtp, sdk_ssvr_routine, ctx);

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_cmd_add
 **功    能: 添加命令"应答 -> 请求"的映射
 **输入参数:
 **     ctx: 全局对象
 **     req: 请求命令
 **     ack: 应答命令
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 不允许重复注册
 **作    者: # Qifeng.zou # 2016.11.10 16:50:18 #
 ******************************************************************************/
int sdk_cmd_add(sdk_cntx_t *ctx, uint32_t req, uint32_t ack)
{
    sdk_cmd_ack_t *item;

    item = (sdk_cmd_ack_t *)calloc(1, sizeof(sdk_cmd_ack_t));
    if (NULL == item) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    item->ack = ack;
    item->req = req;

    if (avl_insert(ctx->cmd, item)) {
        log_error(ctx->log, "Register maybe repeat! cmd:%d ack:%d!", req, ack);
        free(item);
        return SDK_ERR_REPEAT_REG;
    }

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_register
 **功    能: 消息处理的注册接口
 **输入参数:
 **     ctx: 全局对象
 **     type: 扩展消息类型 Range:(0 ~ SDK_TYPE_MAX)
 **     proc: 回调函数
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     1. 只能用于注册处理扩展数据类型的处理
 **     2. 不允许重复注册
 **作    者: # Qifeng.zou # 2015.05.19 #
 ******************************************************************************/
int sdk_register(sdk_cntx_t *ctx, uint32_t cmd, sdk_reg_cb_t proc, void *param)
{
    sdk_reg_t *item;

    item = (sdk_reg_t *)calloc(1, sizeof(sdk_reg_t));
    if (NULL == item) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    item->cmd = cmd;
    item->proc = proc;
    item->param = param;

    if (avl_insert(ctx->reg, item)) {
        log_error(ctx->log, "Register maybe repeat! cmd:%d!", cmd);
        free(item);
        return SDK_ERR_REPEAT_REG;
    }

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_async_send
 **功    能: 发送指定数据(对外接口)
 **输入参数:
 **     ctx: 上下文信息
 **     cmd: 数据类型
 **     data: 数据地址
 **     size: 数据长度
 **     timeout: 超时时间
 **     cb: 该包发送结果回调(成功/失败/超时)
 **     param: 回调附加参数
 **输出参数: NONE
 **返    回: 发送包的序列号(-1:失败)
 **实现描述: 将数据按照约定格式放入队列中
 **注意事项:
 **     1. 只能用于发送自定义数据类型, 而不能用于系统数据类型
 **     2. 不用关注变量num在多线程中的值, 因其不影响安全性
 **     3. 只要SSVR未处于未上线成功的状态, 则认为联网失败.
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
uint32_t sdk_async_send(sdk_cntx_t *ctx, uint32_t cmd, 
        const void *data, size_t size, int timeout, sdk_send_cb_t cb, void *param)
{
    int ret;
    void *addr;
    uint64_t seq;
    mesg_header_t *head;
    sdk_send_item_t *item;
    sdk_ssvr_t *ssvr = ctx->ssvr;

    /* > 判断网络是否正常 */
    if (!SDK_SSVR_GET_ONLINE(ssvr)) {
        cb(cmd, data, size, NULL, 0, SDK_STAT_SEND_FAIL, param);
        log_error(ctx->log, "Network is still disconnect!");
        return -1; /* 网络已断开 */
    }

    /* > 申请内存空间 */
    addr = (void *)calloc(1, sizeof(mesg_header_t)+size+Random()%20);
    if (NULL == addr) {
        cb(cmd, data, size, NULL, 0, SDK_STAT_SEND_FAIL, param);
        log_error(ctx->log, "Alloc memory [%d] failed! errmsg:[%d] %s!",
                size+sizeof(mesg_header_t), errno, strerror(errno));
        return -1;
    }

    /* > 设置发送数据 */
    head = (mesg_header_t *)addr;

    head->type = cmd;
    head->length = size;
    head->sid = ctx->sid;
    head->seq = sdk_gen_seq(ctx);

    seq = head->seq;

    memcpy(head+1, data, size);

    log_debug(ctx->log, "Head type:0x%02X sid:%d length:%d seq:%lu!",
            head->type, head->sid, head->length, head->seq);

    /* > 设置发送单元 */
    item = (sdk_send_item_t *)calloc(1, sizeof(sdk_send_item_t));
    if (NULL == item) {
        cb(cmd, data, size, NULL, 0, SDK_STAT_SEND_FAIL, param);
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        FREE(addr);
        return SDK_ERR;
    }

    item->seq = seq;
    item->stat = SDK_STAT_IN_SENDQ;
    item->cmd = cmd;
    item->len = size;
    item->ttl = time(NULL) + timeout;
    item->cb = cb;
    item->data = (void *)addr;
    item->param = param;

    /* > 放入管理表 */
    ret = sdk_send_mgr_insert(ctx, item, WRLOCK);
    if (0 != ret) {
        sdk_send_mgr_unlock(ctx, WRLOCK);
        cb(cmd, data, size, NULL, 0, SDK_STAT_SEND_FAIL, param);
        log_error(ctx->log, "Insert send mgr tab failed! ret:%d", ret);
        FREE(addr);
        FREE(item);
        return -1;
    }

    /* > 放入发送队列 */
    sdk_queue_rpush(&ctx->sendq, (void *)addr);

    sdk_send_mgr_unlock(ctx, WRLOCK);

    /* > 通知发送线程 */
    sdk_cli_cmd_send_req(ctx);

    return seq;
}

/******************************************************************************
 **函数名称: sdk_network_switch
 **功    能: 网络状态切换(对外接口)
 **输入参数:
 **     ctx: 上下文信息
 **     status: 状态(0:关闭 1:开启)
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 发送命令通知接收线程
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.07 09:43:21 #
 ******************************************************************************/
int sdk_network_switch(sdk_cntx_t *ctx, int status)
{
    int ret;
    sdk_cmd_t cmd;
    char path[FILE_NAME_MAX_LEN];
    sdk_conf_t *conf = &ctx->conf;

    memset(&cmd, 0, sizeof(cmd));

    cmd.type = status? SDK_CMD_NETWORK_CONN : SDK_CMD_NETWORK_DISCONN;
    sdk_ssvr_usck_path(conf, path);

    if (spin_trylock(&ctx->cmd_sck_lck)) {
        log_debug(ctx->log, "Try lock failed!");
        return SDK_OK;
    }

    ret = unix_udp_send(ctx->cmd_sck_id, path, &cmd, sizeof(cmd));
    spin_unlock(&ctx->cmd_sck_lck);
    if (ret < 0) {
        if (EAGAIN != errno) {
            log_debug(ctx->log, "errmsg:[%d] %s! path:%s", errno, strerror(errno), path);
        }
        return SDK_ERR;
    }

    return SDK_OK;
}

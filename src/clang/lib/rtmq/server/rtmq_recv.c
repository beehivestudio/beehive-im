/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: rtmq_recv.c
 ** 版本号: 1.0
 ** 描  述: 实时消息队列(Real-Time Message Queue)
 **         1. 主要用于异步系统之间数据消息的传输
 ** 作  者: # Qifeng.zou # 2014.12.29 #
 ******************************************************************************/

#include "log.h"
#include "lock.h"
#include "redo.h"
#include "mem_ref.h"
#include "shm_opt.h"
#include "hash_alg.h"
#include "rtmq_mesg.h"
#include "rtmq_comm.h"
#include "rtmq_recv.h"
#include "thread_pool.h"

static int rtmq_auth_init(rtmq_cntx_t *ctx);

static int rtmq_creat_connq(rtmq_cntx_t *ctx);
static int rtmq_creat_recvq(rtmq_cntx_t *ctx);
static int rtmq_creat_sendq(rtmq_cntx_t *ctx);
static int rtmq_creat_distq(rtmq_cntx_t *ctx);

static int rtmq_creat_recv_cmd_fd(rtmq_cntx_t *ctx);
static int rtmq_creat_work_cmd_fd(rtmq_cntx_t *ctx);
static int rtmq_creat_dist_cmd_fd(rtmq_cntx_t *ctx);

static int rtmq_cmd_send_dist_req(rtmq_cntx_t *ctx);

static int rtmq_creat_recvs(rtmq_cntx_t *ctx);
void rtmq_recvs_destroy(void *_ctx, void *param);

static int rtmq_creat_workers(rtmq_cntx_t *ctx);
void rtmq_workers_destroy(void *_ctx, void *param);

static int rtmq_proc_def_hdl(int type, int orig, char *buff, size_t len, void *param);

static int rtmq_pub_group_trav_cb(void *data, void *args);

/******************************************************************************
 **函数名称: rtmq_init
 **功    能: 初始化SDTP接收端
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **     1. 创建全局对象
 **     2. 备份配置信息
 **     3. 初始化接收端
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
rtmq_cntx_t *rtmq_init(const rtmq_conf_t *cf, log_cycle_t *log)
{
    rtmq_cntx_t *ctx;
    rtmq_conf_t *conf;

    /* > 校验配置合法性 */
    if (!rtmq_conf_isvalid(cf)) {
        log_error(log, "Configuration isn't invalid!");
        return NULL;
    }

    /* > 创建全局对象 */
    ctx = (rtmq_cntx_t *)calloc(1, sizeof(rtmq_cntx_t));
    if (NULL == ctx) {
        log_error(log, "errmsg:[%d] %s!", errno, strerror(errno));
        return NULL;
    }

    ctx->log = log;
    conf = &ctx->conf;
    memcpy(conf, cf, sizeof(rtmq_conf_t));  /* 配置信息 */
    conf->recvq_num = RTMQ_WORKER_HDL_QNUM * cf->work_thd_num;

    do {
        /* > 构建鉴权表 */
        if (rtmq_auth_init(ctx)) {
            log_error(ctx->log, "Initialize auth failed!");
            break;
        }

        /* > 构建NODE->SVR映射表 */
        if (rtmq_node_to_svr_map_init(ctx)) {
            log_error(ctx->log, "Initialize sck-dev map table failed!");
            break;
        }

        /* > 初始化订阅列表 */
        if (rtmq_sub_init(ctx)) {
            log_error(ctx->log, "Initialize sub list failed!");
            break;
        }

        /* > 初始化注册信息 */
        ctx->reg = avl_creat(NULL, (cmp_cb_t)rtmq_reg_cmp_cb);
        if (NULL == ctx->reg) {
            log_error(ctx->log, "Create register map failed!");
            break;
        }

        /* > 创建连接队列 */
        if (rtmq_creat_connq(ctx)) {
            log_error(ctx->log, "Create conn queue failed!");
            break;
        }

        /* > 创建接收队列 */
        if (rtmq_creat_recvq(ctx)) {
            log_error(ctx->log, "Create recv queue failed!");
            break;
        }

        /* > 创建发送队列 */
        if (rtmq_creat_sendq(ctx)) {
            log_error(ctx->log, "Create send queue failed!");
            break;
        }

        /* > 创建分发队列 */
        if (rtmq_creat_distq(ctx)) {
            log_error(ctx->log, "Create distribute queue failed!");
            break;
        }

        /* > 创建接收线程通信FD */
        if (rtmq_creat_recv_cmd_fd(ctx)) {
            log_error(ctx->log, "Create recv cmd fd failed!");
            break;
        }

        /* > 创建工作线程通信FD */
        if (rtmq_creat_work_cmd_fd(ctx)) {
            log_error(ctx->log, "Create work cmd fd failed!");
            break;
        }

        /* > 创建分发线程通信FD */
        if (rtmq_creat_dist_cmd_fd(ctx)) {
            log_error(ctx->log, "Create work cmd fd failed!");
            break;
        }

        /* > 创建接收线程池 */
        if (rtmq_creat_recvs(ctx)) {
            log_error(ctx->log, "Create recv thread pool failed!");
            break;
        }

        /* > 创建工作线程池 */
        if (rtmq_creat_workers(ctx)) {
            log_error(ctx->log, "Create worker thread pool failed!");
            break;
        }

        /* > 初始化侦听服务 */
        if (rtmq_lsn_init(ctx)) {
            log_error(ctx->log, "Create worker thread pool failed!");
            break;
        }

        return ctx;
    } while(0);

    return NULL;
}

/******************************************************************************
 **函数名称: rtmq_launch
 **功    能: 启动SDTP接收端
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
int rtmq_launch(rtmq_cntx_t *ctx)
{
    int idx;
    pthread_t tid;
    thread_pool_t *tp;
    rtmq_listen_t *lsn = &ctx->listen;

    /* > 设置接收线程回调 */
    tp = ctx->recvtp;
    for (idx=0; idx<tp->num; ++idx) {
        thread_pool_add_worker(tp, rtmq_rsvr_routine, ctx);
    }

    /* > 设置工作线程回调 */
    tp = ctx->worktp;
    for (idx=0; idx<tp->num; ++idx) {
        thread_pool_add_worker(tp, rtmq_worker_routine, ctx);
    }

    /* > 创建侦听线程 */
    if (thread_creat(&lsn->tid, rtmq_lsn_routine, ctx)) {
        log_error(ctx->log, "Start listen failed");
        return RTMQ_ERR;
    }

    /* > 创建分发线程 */
    if (thread_creat(&tid, rtmq_dsvr_routine, ctx)) {
        log_error(ctx->log, "Start distribute thread failed");
        return RTMQ_ERR;
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_register
 **功    能: 消息处理的注册接口
 **输入参数:
 **     ctx: 全局对象
 **     type: 扩展消息类型 Range:(0 ~ RTMQ_TYPE_MAX)
 **     proc: 回调函数
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     1. 只能用于注册处理扩展数据类型的处理
 **     2. 不允许重复注册
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
int rtmq_register(rtmq_cntx_t *ctx, int type, rtmq_reg_cb_t proc, void *param)
{
    rtmq_reg_t *item;

    item = (rtmq_reg_t *)calloc(1, sizeof(rtmq_reg_t));
    if (NULL == item) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    item->type = type;
    item->proc = proc;
    item->param = param;

    if (avl_insert(ctx->reg, item)) {
        log_error(ctx->log, "Register maybe repeat! type:%d!", type);
        free(item);
        return RTMQ_ERR_REPEAT_REG;
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_async_send
 **功    能: 接收客户端发送数据
 **输入参数:
 **     ctx: 全局对象
 **     type: 消息类型
 **     dest: 目标结点ID
 **     data: 需要发送的数据
 **     len: 发送数据的长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将数据放入应答队列
 **注意事项: 内存结构: 转发信息(frwd) + 实际数据
 **作    者: # Qifeng.zou # 2015.06.01 #
 ******************************************************************************/
int rtmq_async_send(rtmq_cntx_t *ctx, int type, int dest, void *data, size_t len)
{
    int idx;
    void *addr;
    rtmq_header_t *head;

    idx = rand() % ctx->conf.distq_num;

    /* > 申请队列空间 */
    addr = mem_ref_alloc(sizeof(rtmq_header_t) + len,
            NULL, (mem_alloc_cb_t)mem_alloc, (mem_dealloc_cb_t)mem_dealloc);
    if (NULL == addr) {
        log_error(ctx->log, "Alloc memory failed! errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    head = (rtmq_header_t *)addr;

    head->type = type;
    head->nid = dest;
    head->flag = RTMQ_EXP_MESG;
    head->chksum = RTMQ_CHKSUM_VAL;
    head->length = len;

    memcpy(addr+sizeof(rtmq_header_t), data, len);

    /* > 压入队列空间 */
    if (ring_push(ctx->distq[idx], addr)) {
        mem_ref_decr(addr);
        log_error(ctx->log, "Push into ring failed!");
        return RTMQ_ERR;
    }

    rtmq_cmd_send_dist_req(ctx);

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_publish
 **功    能: 发布消息
 **输入参数:
 **     ctx: 全局对象
 **     type: 消息类型
 **     data: 需要发送的数据
 **     len: 发送数据的长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将数据放入应答队列
 **注意事项: 内存结构: 转发信息(rtmq) + 实际数据
 **作    者: # Qifeng.zou # 2017.06.26 20:50:42 #
 ******************************************************************************/
typedef struct{
    int type;
    void *data;
    size_t len;
    rtmq_cntx_t *ctx;
} rtmq_pub_item_t;

int rtmq_publish(rtmq_cntx_t *ctx, int type, void *data, size_t len)
{
    rtmq_pub_item_t item;
    rtmq_sub_list_t *list, key;

    /* > 查找消息订阅列表 */
    key.type = type;

    list = hash_tab_query(ctx->sub, &key, RDLOCK);
    if (NULL == list) {
        log_error(ctx->log, "No node sub this message! type:0x%04X", type);
        return -1;
    }

    /* > 给各订阅分组发送消息 */
    item.type = type;
    item.data = data;
    item.len = len;
    item.ctx = ctx;

    avl_trav(list->groups, rtmq_pub_group_trav_cb, &item);

    hash_tab_unlock(ctx->sub, &key, RDLOCK);

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_connq
 **功    能: 创建连接队列
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建队列数组
 **     2. 依次创建连接队列
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.21 23:48:12 #
 ******************************************************************************/
static int rtmq_creat_connq(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 创建队列数组 */
    ctx->connq = calloc(conf->recv_thd_num, sizeof(queue_t *));
    if (NULL == ctx->connq) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 依次创建连接队列 */
    for(idx=0; idx<conf->recv_thd_num; ++idx) {
        ctx->connq[idx] = queue_creat(RTMQ_CONNQ_LEN, sizeof(rtmq_conn_item_t));
        if (NULL == ctx->connq[idx]) {
            log_error(ctx->log, "Create conn queue failed!");
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_recvq
 **功    能: 创建接收队列
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建队列数组
 **     2. 依次创建接收队列
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
static int rtmq_creat_recvq(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 创建队列数组 */
    ctx->recvq = calloc(conf->recvq_num, sizeof(queue_t *));
    if (NULL == ctx->recvq) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 依次创建接收队列 */
    for(idx=0; idx<conf->recvq_num; ++idx) {
        ctx->recvq[idx] = queue_creat(conf->recvq.max, sizeof(rtmq_recv_item_t));
        if (NULL == ctx->recvq[idx]) {
            log_error(ctx->log, "Create queue failed! max:%d size:%d",
                    conf->recvq.max, conf->recvq.size);
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_sendq
 **功    能: 创建发送队列
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.22 #
 ******************************************************************************/
static int rtmq_creat_sendq(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 创建队列数组 */
    ctx->sendq = calloc(1, conf->recv_thd_num*sizeof(queue_t *));
    if (NULL == ctx->sendq) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 依次创建发送队列 */
    for(idx=0; idx<conf->recv_thd_num; ++idx) {
        ctx->sendq[idx] = ring_creat(conf->sendq.max);
        if (NULL == ctx->sendq[idx]) {
            log_error(ctx->log, "Create send-queue failed! max:%d size:%d",
                    conf->sendq.max, conf->sendq.size);
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_distq
 **功    能: 创建分发队列
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015-07-06 11:21:28 #
 ******************************************************************************/
static int rtmq_creat_distq(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 申请对象空间 */
    ctx->distq = (ring_t **)calloc(1, conf->distq_num*sizeof(ring_t *));
    if (NULL == ctx->distq) {
        log_error(ctx->log, "Alloc memory failed!");
        return RTMQ_ERR;
    }

    /* > 依次创建队列 */
    for (idx=0; idx<conf->distq_num; ++idx) {
        ctx->distq[idx] = ring_creat(conf->distq.max);
        if (NULL == ctx->distq[idx]) {
            log_error(ctx->log, "Create queue failed!");
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/* 创建分发线程通信管道 */
static int rtmq_creat_dist_cmd_fd(rtmq_cntx_t *ctx)
{
    int idx, total = 1;

    ctx->dist_cmd_fd = (rtmq_pipe_t *)calloc(total, sizeof(rtmq_pipe_t));
    if (NULL == ctx->dist_cmd_fd) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_OK;
    }

    for (idx=0; idx<total; idx+=1) {
        pipe(ctx->dist_cmd_fd[idx].fd);

        fd_set_nonblocking(ctx->dist_cmd_fd[idx].fd[0]);
        fd_set_nonblocking(ctx->dist_cmd_fd[idx].fd[1]);
    }

    return RTMQ_OK;
}

/* 创建接收线程通信管道 */
static int rtmq_creat_recv_cmd_fd(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    ctx->recv_cmd_fd = (rtmq_pipe_t *)calloc(conf->recv_thd_num, sizeof(rtmq_pipe_t));
    if (NULL == ctx->recv_cmd_fd) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_OK;
    }

    for (idx=0; idx<conf->recv_thd_num; idx+=1) {
        pipe(ctx->recv_cmd_fd[idx].fd);

        fd_set_nonblocking(ctx->recv_cmd_fd[idx].fd[0]);
        fd_set_nonblocking(ctx->recv_cmd_fd[idx].fd[1]);
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_recvs
 **功    能: 创建接收线程池
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建线程池
 **     2. 创建接收对象
 **     3. 初始化接收对象
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.01 #
 ******************************************************************************/
static int rtmq_creat_recvs(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_rsvr_t *rsvr;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 创建接收对象 */
    rsvr = (rtmq_rsvr_t *)calloc(conf->recv_thd_num, sizeof(rtmq_rsvr_t));
    if (NULL == rsvr) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 创建线程池 */
    ctx->recvtp = thread_pool_init(conf->recv_thd_num, NULL, (void *)rsvr);
    if (NULL == ctx->recvtp) {
        log_error(ctx->log, "Initialize thread pool failed!");
        free(rsvr);
        return RTMQ_ERR;
    }

    /* > 初始化接收对象 */
    for (idx=0; idx<conf->recv_thd_num; ++idx) {
        if (rtmq_rsvr_init(ctx, rsvr+idx, idx)) {
            log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));

            free(rsvr);
            thread_pool_destroy(ctx->recvtp);
            ctx->recvtp = NULL;

            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_recvs_destroy
 **功    能: 销毁接收线程池
 **输入参数:
 **     ctx: 全局对象
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.01 #
 ******************************************************************************/
void rtmq_recvs_destroy(void *_ctx, void *param)
{
    int idx;
    rtmq_cntx_t *ctx = (rtmq_cntx_t *)_ctx;
    rtmq_rsvr_t *rsvr = (rtmq_rsvr_t *)ctx->recvtp->data;

    for (idx=0; idx<ctx->conf.recv_thd_num; ++idx, ++rsvr) {
        /* > 关闭命令套接字 */
        CLOSE(rsvr->cmd_fd);

        /* > 关闭通信套接字 */
        rtmq_rsvr_del_all_conn_hdl(ctx, rsvr);
    }

    FREE(ctx->recvtp->data);
    thread_pool_destroy(ctx->recvtp);

    return;
}

/* 创建工作线程通信管道 */
static int rtmq_creat_work_cmd_fd(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_conf_t *conf = &ctx->conf;

    ctx->work_cmd_fd = (rtmq_pipe_t *)calloc(conf->work_thd_num, sizeof(rtmq_pipe_t));
    if (NULL == ctx->work_cmd_fd) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_OK;
    }

    for (idx=0; idx<conf->work_thd_num; idx+=1) {
        pipe(ctx->work_cmd_fd[idx].fd);

        fd_set_nonblocking(ctx->work_cmd_fd[idx].fd[0]);
        fd_set_nonblocking(ctx->work_cmd_fd[idx].fd[1]);
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_creat_workers
 **功    能: 创建工作线程池
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建线程池
 **     2. 创建工作对象
 **     3. 初始化工作对象
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.06 #
 ******************************************************************************/
static int rtmq_creat_workers(rtmq_cntx_t *ctx)
{
    int idx;
    rtmq_worker_t *wrk;
    rtmq_conf_t *conf = &ctx->conf;

    /* > 创建工作对象 */
    wrk = (void *)calloc(conf->work_thd_num, sizeof(rtmq_worker_t));
    if (NULL == wrk) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 创建线程池 */
    ctx->worktp = thread_pool_init(conf->work_thd_num, NULL, (void *)wrk);
    if (NULL == ctx->worktp) {
        log_error(ctx->log, "Initialize thread pool failed!");
        free(wrk);
        return RTMQ_ERR;
    }

    /* > 初始化工作对象 */
    for (idx=0; idx<conf->work_thd_num; ++idx) {
        if (rtmq_worker_init(ctx, wrk+idx, idx)) {
            log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
            free(wrk);
            thread_pool_destroy(ctx->recvtp);
            ctx->recvtp = NULL;
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_workers_destroy
 **功    能: 销毁工作线程池
 **输入参数:
 **     ctx: 全局对象
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.06 #
 ******************************************************************************/
void rtmq_workers_destroy(void *_ctx, void *param)
{
    int idx;
    rtmq_cntx_t *ctx = (rtmq_cntx_t *)_ctx;
    rtmq_conf_t *conf = &ctx->conf;
    rtmq_worker_t *wrk = (rtmq_worker_t *)ctx->worktp->data;

    for (idx=0; idx<conf->work_thd_num; ++idx, ++wrk) {
        CLOSE(wrk->cmd_fd);
    }

    FREE(ctx->worktp->data);
    thread_pool_destroy(ctx->recvtp);

    return;
}

/******************************************************************************
 **函数名称: rtmq_proc_def_hdl
 **功    能: 默认消息处理函数
 **输入参数:
 **     type: 消息类型
 **     orig: 源设备ID
 **     buff: 消息内容
 **     len: 内容长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.06 #
 ******************************************************************************/
static int rtmq_proc_def_hdl(int type, int orig, char *buff, size_t len, void *param)
{
    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_cmd_send_dist_req
 **功    能: 通知分发服务
 **输入参数:
 **     cli: 上下文信息
 **     idx: 发送服务ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.03.20 #
 ******************************************************************************/
static int rtmq_cmd_send_dist_req(rtmq_cntx_t *ctx)
{
    rtmq_cmd_t cmd;

    memset(&cmd, 0, sizeof(cmd));

    cmd.type = RTMQ_CMD_DIST_REQ;

    if (write(ctx->dist_cmd_fd[0].fd[1], &cmd, sizeof(cmd)) < 0) {
        log_error(ctx->log, "Send dist command failed! errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    log_trace(ctx->log, "Send dist command success!");

    return RTMQ_OK;
}

/* 遍历鉴权连表 */
static int rtmq_auth_trav_add(rtmq_auth_t *auth, rtmq_cntx_t *ctx)
{
    return rtmq_auth_add(ctx, auth->usr, auth->passwd);
}

/* 鉴权比较 */
static int rtmq_auth_cmp_cb(const rtmq_auth_t *auth1, const rtmq_auth_t *auth2)
{
    return strcmp(auth1->usr, auth2->usr);
}

/******************************************************************************
 **函数名称: rtmq_auth_init
 **功    能: 初始化鉴权表
 **输入参数:
 **     ctx: 全局信息
 **     usr: 用户名
 **     passwd: 密码
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 文件描述符可不关闭
 **作    者: # Qifeng.zou # 2016.07.19 22:03:43 #
 ******************************************************************************/
static int rtmq_auth_init(rtmq_cntx_t *ctx)
{
    list_t *auth = ctx->conf.auth;

    ctx->auth = avl_creat(NULL, (cmp_cb_t)rtmq_auth_cmp_cb);
    if (NULL == ctx->auth) {
        return -1;
    }

    return list_trav(auth, (trav_cb_t)rtmq_auth_trav_add, (void *)ctx);
}

/******************************************************************************
 **函数名称: rtmq_auth_add
 **功    能: 添加鉴权注册
 **输入参数:
 **     ctx: 全局信息
 **     usr: 用户名
 **     passwd: 密码
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 文件描述符可不关闭
 **作    者: # Qifeng.zou # 2016.07.19 22:03:43 #
 ******************************************************************************/
int rtmq_auth_add(rtmq_cntx_t *ctx, char *usr, char *passwd)
{
    rtmq_auth_t *auth, key;

    snprintf(key.usr, sizeof(key.usr), "%s", usr);

    auth = (rtmq_auth_t *)avl_query(ctx->auth, &key);
    if (NULL != auth) {
        return (0 == strcmp(auth->passwd, passwd))? 0 : -1;
    }

    auth = (rtmq_auth_t *)calloc(1, sizeof(rtmq_auth_t));
    if (NULL == auth) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    snprintf(auth->usr, sizeof(auth->usr), "%s", usr);
    snprintf(auth->passwd, sizeof(auth->passwd), "%s", passwd);

    if (avl_insert(ctx->auth, (void *)auth)) {
        free(auth);
        return -1;
    }

    return 0;
}

/******************************************************************************
 **函数名称: rtmq_auth_check
 **功    能: 鉴权检查
 **输入参数:
 **     ctx: 全局信息
 **     usr: 用户名
 **     passwd: 密码
 **输出参数: NONE
 **返    回: true:通过 false:失败
 **实现描述:
 **注意事项: 文件描述符可不关闭
 **作    者: # Qifeng.zou # 2016.07.19 22:03:43 #
 ******************************************************************************/
bool rtmq_auth_check(rtmq_cntx_t *ctx, char *usr, char *passwd)
{
    rtmq_auth_t *auth, key;

    snprintf(key.usr, sizeof(key.usr), "%s", usr);

    auth = avl_query(ctx->auth, &key);
    if (NULL != auth) {
        return (0 == strcmp(auth->passwd, passwd))? true : false;
    }
    return false;
}

/******************************************************************************
 **函数名称: rtmq_pub_group_trav_cb
 **功    能: 发布消息
 **输入参数:
 **     ctx: 全局对象
 **     type: 消息类型
 **     data: 需要发送的数据
 **     len: 发送数据的长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 每组都要发送数据, 但只给每组中的一个结点发送数据.
 **注意事项: 内存结构: 转发信息(rtmq) + 实际数据
 **作    者: # Qifeng.zou # 2017.06.26 20:50:42 #
 ******************************************************************************/
static int rtmq_pub_group_trav_cb(void *data, void *args)
{
    rtmq_sub_node_t *node;
    rtmq_pub_item_t *item = (rtmq_pub_item_t *)args;
    rtmq_cntx_t *ctx = item->ctx;
    rtmq_sub_group_t *group = (rtmq_sub_group_t *)data;

    /* > 随机选择一个结点 */
    node = vector_get(group->nodes, Random()%vector_len(group->nodes));
    if (NULL == node) {
        log_debug(ctx->log, "Get sub item failed! gid:%u type:0x%04X", group->gid, item->type);
        return RTMQ_OK;
    }

    /* > 发送数据 */
    rtmq_async_send(ctx, item->type, node->nid, item->data, item->len);

    log_debug(ctx->log, "Send data! type:0x%04X gid:%u nid:%u sid:%d!",
            item->type, group->gid, node->nid, node->sid);

    return RTMQ_OK;
}

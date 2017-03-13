#include "sck.h"
#include "comm.h"
#include "mesg.h"
#include "redo.h"
#include "search.h"
#include "command.h"
#include "acc_rsvr.h"
#include "acc_lsn.h"

static int acc_lsvr_timeout_hdl(acc_cntx_t *ctx, acc_lsvr_t *lsvr);
static int acc_lsvr_accept(acc_cntx_t *ctx, acc_lsvr_t *lsvr);
static int acc_lsvr_send_add_sck_req(acc_cntx_t *ctx, acc_lsvr_t *lsvr, int idx);

/******************************************************************************
 **函数名称: acc_lsvr_self
 **功    能: 获取代理对象
 **输入参数: 
 **     ctx: 全局信息
 **输出参数: NONE
 **返    回: 侦听对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2015-07-07 21:53:08 #
 ******************************************************************************/
static acc_lsvr_t *acc_lsn_self(acc_cntx_t *ctx)
{
    int id;
    acc_lsvr_t *lsvr;

    id = thread_pool_get_tidx(ctx->lsvr_pool);
    if (id < 0) {
        log_error(ctx->log, "Get self-thread failed!");
        return NULL;
    }

    lsvr = thread_pool_get_args(ctx->lsvr_pool);

    return lsvr + id;
}

/******************************************************************************
 **函数名称: acc_lsvr_routine
 **功    能: 运行侦听线程
 **输入参数:
 **     _ctx: 全局信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 初始化过程在程序启动时已经完成 - 在子线程中初始化，一旦出现异常将不好处理!
 **作    者: # Qifeng.zou # 2014.11.18 #
 ******************************************************************************/
void *acc_lsvr_routine(void *_ctx)
{
    int ret, max;
    fd_set rdset;
    acc_lsvr_t *lsvr;
    struct timeval tv;
    acc_cntx_t *ctx = (acc_cntx_t *)_ctx;

    nice(-20);

    /* > 获取侦听对象 */
    lsvr = acc_lsn_self(ctx);
    if (NULL == lsvr) {
        log_error(ctx->log, "Search listen object failed!");
        return (void *)-1;
    }

    /* > 接收网络连接 */
    while (1) {
        FD_ZERO(&rdset);

        FD_SET(lsvr->cmd_sck_id, &rdset);
        FD_SET(ctx->listen.lsn_sck_id, &rdset);

        max = MAX(ctx->listen.lsn_sck_id, lsvr->cmd_sck_id);

        /* > 等待事件通知 */
        tv.tv_sec = 1;
        tv.tv_usec = 0;
        ret = select(max+1, &rdset, NULL, NULL, &tv);
        if (ret < 0) {
            if (EINTR == errno) { continue; }
            log_error(lsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));
            continue;
        }
        else if (0 == ret) {
            acc_lsvr_timeout_hdl(ctx, lsvr);
            continue;
        }

        /* > 接收网络连接 */
        if (FD_ISSET(ctx->listen.lsn_sck_id, &rdset)) {
            acc_lsvr_accept(ctx, lsvr);
        }
    }
    return (void *)0;
}

/******************************************************************************
 **函数名称: acc_lsvr_init
 **功    能: 初始化侦听线程
 **输入参数:
 **     ctx: 全局信息
 **     lsvr: 侦听对象
 **     idx: 侦听对象索引
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2014.11.19 #
 ******************************************************************************/
int acc_lsvr_init(acc_cntx_t *ctx, acc_lsvr_t *lsvr, int idx)
{
    char path[FILE_NAME_MAX_LEN];
    acc_conf_t *conf = ctx->conf;

    lsvr->id = idx;

    acc_lsvr_cmd_usck_path(conf, idx, path, sizeof(path));

    lsvr->cmd_sck_id = unix_udp_creat(path);
    if (lsvr->cmd_sck_id < 0) {
        log_error(ctx->log, "errmsg:[%d] %s! idx:%d path:%s", errno, strerror(errno), idx, path);
        return ACC_ERR;
    }

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_lsvr_timeout_hdl
 **功    能: 超时处理
 **输入参数:
 **     ctx: 全局信息
 **     lsvr: 侦听对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2015-07-08 07:46:32 #
 ******************************************************************************/
static int acc_lsvr_timeout_hdl(acc_cntx_t *ctx, acc_lsvr_t *lsvr)
{
    int idx;

    for (idx=0; idx<ctx->conf->rsvr_num; ++idx) {
        if (queue_used(ctx->connq[idx])) {
            acc_lsvr_send_add_sck_req(ctx, lsvr, idx);
        }
    }

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_lsvr_accept
 **功    能: 接收连接请求
 **输入参数:
 **     ctx: 全局信息
 **     lsvr: 侦听对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **     1. 接收连接请求
 **     2. 将通信套接字放入队列
 **注意事项: 
 **作    者: # Qifeng.zou # 2014.11.20 #
 ******************************************************************************/
static int acc_lsvr_accept(acc_cntx_t *ctx, acc_lsvr_t *lsvr)
{
    int fd, idx, cid;
    acc_add_sck_t *add;
    struct sockaddr_in cliaddr;

    if (spin_trylock(&ctx->listen.accept_lock)) { /* 加锁 */
        return ACC_ERR;
    }

    /* > 接收连接请求 */
    fd = tcp_accept(ctx->listen.lsn_sck_id, (struct sockaddr *)&cliaddr);
    if (fd < 0) {
        spin_unlock(&ctx->listen.accept_lock);
        log_error(lsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return ACC_OK;
    }

    cid = atomic64_inc(&ctx->listen.cid); /* 计数 */

    spin_unlock(&ctx->listen.accept_lock); /* 解锁 */

    /* > 将通信套接字放入队列 */
    idx = cid % ctx->conf->rsvr_num;

    add = queue_malloc(ctx->connq[idx], sizeof(acc_add_sck_t));
    if (NULL == add) {
        log_error(lsvr->log, "Alloc from queue failed! fd:%d size:%d/%d",
                fd, sizeof(acc_add_sck_t), queue_size(ctx->connq[idx]));
        CLOSE(fd);
        return ACC_ERR;
    }

    add->fd = fd;
    add->cid = cid;
    ftime(&add->crtm);

    log_debug(lsvr->log, "Push data! fd:%d addr:%p cid:%u idx:%d", fd, add, cid, idx);

    queue_push(ctx->connq[idx], add);

    /* > 发送ADD-SCK请求 */
    acc_lsvr_send_add_sck_req(ctx, lsvr, idx);

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_lsvr_send_add_sck_req
 **功    能: 发送ADD-SCK请求
 **输入参数:
 **     ctx: 全局信息
 **     lsvr: 侦听对象
 **     idx: 接收服务的索引号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2015-06-22 21:47:52 #
 ******************************************************************************/
static int acc_lsvr_send_add_sck_req(acc_cntx_t *ctx, acc_lsvr_t *lsvr, int idx)
{
    cmd_data_t cmd;
    char path[FILE_NAME_MAX_LEN];
    acc_conf_t *conf = ctx->conf;

    cmd.type = CMD_ADD_SCK;
    acc_rsvr_cmd_usck_path(conf, idx, path, sizeof(path));

    unix_udp_send(lsvr->cmd_sck_id, path, &cmd, sizeof(cmd));

    return ACC_OK;
}

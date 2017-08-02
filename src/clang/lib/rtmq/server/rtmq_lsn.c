/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: rtmq_rlsn.c
 ** 版本号: 1.0
 ** 描  述: 实时消息队列(Real-Time Message Queue)
 **         1. 主要用于异步系统之间数据消息的传输
 ** 作  者: # Qifeng.zou # 2014.12.29 #
 ******************************************************************************/

#include "redo.h"
#include "rtmq_comm.h"
#include "rtmq_recv.h"
#include "thread_pool.h"

/* 静态函数 */
static int rtmq_lsn_accept(rtmq_cntx_t *ctx, rtmq_listen_t *lsn);

/******************************************************************************
 **函数名称: rtmq_lsn_routine
 **功    能: 启动SDTP侦听线程
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **     1. 初始化侦听
 **     2. 等待请求和命令
 **     3. 接收请求和命令
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
void *rtmq_lsn_routine(void *param)
{
#define RTMQ_LSN_TMOUT_SEC 30
#define RTMQ_LSN_TMOUT_USEC 0
    fd_set rdset;
    int ret, max;
    struct timeval timeout;
    rtmq_cntx_t *ctx = (rtmq_cntx_t *)param;
    rtmq_listen_t *lsn = &ctx->listen;

    for (;;) {
        /* 2. 等待请求和命令 */
        FD_ZERO(&rdset);

        FD_SET(lsn->lsn_sck_id, &rdset);

        max = lsn->lsn_sck_id;

        timeout.tv_sec = RTMQ_LSN_TMOUT_SEC;
        timeout.tv_usec = RTMQ_LSN_TMOUT_USEC;
        ret = select(max+1, &rdset, NULL, NULL, &timeout);
        if (ret < 0) {
            if (EINTR == errno) { continue; }
            log_error(lsn->log, "errmsg:[%d] %s", errno, strerror(errno));
            abort();
            return (void *)-1;
        } else if (0 == ret) {
            continue;
        }

        /* 3. 接收连接请求 */
        if (FD_ISSET(lsn->lsn_sck_id, &rdset)) {
            rtmq_lsn_accept(ctx, lsn);
        }
    }

    pthread_exit(NULL);
    return (void *)-1;
}

/******************************************************************************
 **函数名称: rtmq_lsn_init
 **功    能: 启动SDTP侦听线程
 **输入参数:
 **     conf: 配置信息
 **输出参数: NONE
 **返    回: 侦听对象
 **实现描述:
 **     1. 侦听指定端口
 **     2. 创建CMD套接字
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
int rtmq_lsn_init(rtmq_cntx_t *ctx)
{
    rtmq_listen_t *lsn = &ctx->listen;
    rtmq_conf_t *conf = &ctx->conf;

    lsn->log = ctx->log;

    /* 1. 侦听指定端口 */
    lsn->lsn_sck_id = tcp_listen(conf->port);
    if (lsn->lsn_sck_id < 0) {
        log_error(lsn->log, "Listen special port failed!");
        return RTMQ_ERR;
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_listen_accept
 **功    能: 接收连接请求
 **输入参数:
 **     ctx: 全局对象
 **     lsn: 侦听对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 接收连接请求
 **     2. 发送至接收端
 **注意事项:
 **作    者: # Qifeng.zou # 2014.12.30 #
 ******************************************************************************/
static int rtmq_lsn_accept(rtmq_cntx_t *ctx, rtmq_listen_t *lsn)
{
    int fd, idx;
    socklen_t len;
    rtmq_cmd_t cmd;
    rtmq_conn_item_t *item;
    struct sockaddr_in cliaddr;
    rtmq_conf_t *conf = &ctx->conf;

    /* 1. 接收连接请求 */
    for (;;) {
        memset(&cliaddr, 0, sizeof(cliaddr));

        len = sizeof(struct sockaddr_in);

        fd = accept(lsn->lsn_sck_id, (struct sockaddr *)&cliaddr, &len);
        if (fd >= 0) {
            break;
        } else if (EINTR == errno) {
            continue;
        }

        log_error(lsn->log, "errmsg:[%d] %s", errno, strerror(errno));
        return RTMQ_ERR;
    }

    fd_set_nonblocking(fd);

    /* 2. 将连接放入队列 */
    idx = lsn->sid % conf->recv_thd_num;

    item = queue_malloc(ctx->connq[idx], sizeof(rtmq_conn_item_t));
    if (NULL == item) {
        close(fd);
        log_error(lsn->log, "Alloc from conn queue failed!");
        return RTMQ_ERR;
    }

    item->fd = fd;
    ftime(&item->ctm);
    item->sid = ++lsn->sid;
    snprintf(item->ipaddr, sizeof(item->ipaddr), "%s", inet_ntoa(cliaddr.sin_addr));

    queue_push(ctx->connq[idx], item);

    /* 3. 发送信号给接收服务 */
    memset(&cmd, 0, sizeof(cmd));

    cmd.type = RTMQ_CMD_ADD_SCK;

    write(ctx->recv_cmd_fd[idx].fd[1], &cmd, sizeof(cmd));

    log_trace(lsn->log, "Accept new connection! idx:%d sid:%lu fd:%d ip:%s",
            idx, lsn->sid, fd, item->ipaddr);

    return RTMQ_OK;
}

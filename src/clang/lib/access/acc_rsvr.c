#include "sck.h"
#include "comm.h"
#include "list.h"
#include "mesg.h"
#include "redo.h"
#include "utils.h"
#include "access.h"
#include "mem_ref.h"
#include "command.h"
#include "xml_tree.h"
#include "hash_alg.h"
#include "acc_rsvr.h"
#include "thread_pool.h"

#define AGT_RSVR_DIST_POP_NUM   (1024)

static acc_rsvr_t *acc_rsvr_self(acc_cntx_t *ctx);
static int acc_rsvr_add_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr);
static int acc_rsvr_del_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck);

static int acc_recv_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck);
static int acc_send_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck);

static int acc_rsvr_dist_send_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr);
static socket_t *acc_push_into_send_list(acc_cntx_t *ctx, acc_rsvr_t *rsvr, uint64_t cid, void *addr);

static int acc_rsvr_kick_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr);

static int acc_rsvr_event_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr);
static int acc_rsvr_timeout_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr);

static int acc_rsvr_connection_cmp(const int *cid, const socket_t *sck);

/******************************************************************************
 **函数名称: acc_rsvr_routine
 **功    能: 运行接收线程
 **输入参数:
 **     _ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.11.18 #
 ******************************************************************************/
void *acc_rsvr_routine(void *_ctx)
{
    acc_rsvr_t *rsvr;
    acc_cntx_t *ctx = (acc_cntx_t *)_ctx;

    nice(-20);

    /* > 获取代理对象 */
    rsvr = acc_rsvr_self(ctx);
    if (NULL == rsvr) {
        log_error(rsvr->log, "Get agent failed!");
        pthread_exit((void *)-1);
        return (void *)-1;
    }

    while (1) {
        /* > 等待事件通知 */
        rsvr->fds = epoll_wait(rsvr->epid, rsvr->events,
                ACC_EVENT_MAX_NUM, ACC_TMOUT_MSEC);
        if (rsvr->fds < 0) {
            if (EINTR == errno) {
                continue;
            }

            /* 异常情况 */
            log_error(rsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));
            abort();
            return (void *)-1;
        } else if (0 == rsvr->fds) {
            rsvr->ctm = time(NULL);
            if (rsvr->ctm - rsvr->scan_tm > ACC_TMOUT_SCAN_SEC) {
                rsvr->scan_tm = rsvr->ctm;
                acc_rsvr_timeout_hdl(ctx, rsvr);
            }
            continue;
        }

        /* > 进行事件处理 */
        acc_rsvr_event_hdl(ctx, rsvr);
    }

    return NULL;
}

/* 命令接收处理 */
static int agt_rsvr_recv_cmd_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck)
{
    cmd_data_t cmd;

    while (1) {
        /* > 接收命令 */
        if (read(sck->fd, &cmd, sizeof(cmd)) < 0) {
            return ACC_SCK_AGAIN;
        }

        /* > 处理命令 */
        switch (cmd.type) {
            case CMD_ADD_SCK:
                if (acc_rsvr_add_conn(ctx, rsvr)) {
                    log_error(rsvr->log, "Add connection failed！");
                }
                break;
            case CMD_DIST_DATA:
                if (acc_rsvr_dist_send_data(ctx, rsvr)) {
                    log_error(rsvr->log, "Disturibute data failed！");
                }
                break;
            case CMD_KICK_CONN:
                if (acc_rsvr_kick_conn(ctx, rsvr)) {
                    log_error(rsvr->log, "Kick connection failed！");
                }
                break;
            default:
                log_error(rsvr->log, "Unknown command type [%d]！", cmd.type);
                break;
        }
    }
    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_rsvr_init
 **功    能: 初始化Agent线程
 **输入参数:
 **     ctx: 全局信息
 **     rsvr: 接收对象
 **     idx: 线程索引
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.11.28 #
 ******************************************************************************/
int acc_rsvr_init(acc_cntx_t *ctx, acc_rsvr_t *rsvr, int idx)
{
    struct epoll_event ev;
    acc_socket_extra_t *extra;
    socket_t *cmd_sck = &rsvr->cmd_sck;

    rsvr->id = idx;
    rsvr->log = ctx->log;
    rsvr->recv_seq = 0;

    rsvr->sendq = ctx->sendq[idx];
    rsvr->connq = ctx->connq[idx];
    rsvr->kickq = ctx->kickq[idx];

    do {
        /* > 创建epoll对象 */
        rsvr->epid = epoll_create(ACC_EVENT_MAX_NUM);
        if (rsvr->epid < 0) {
            log_error(rsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));
            break;
        }

        rsvr->events = calloc(1, ACC_EVENT_MAX_NUM * sizeof(struct epoll_event));
        if (NULL == rsvr->events) {
            log_error(rsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));
            break;
        }

        /* > 创建附加信息 */
        extra = calloc(1, sizeof(acc_socket_extra_t));
        if (NULL == extra) {
            log_error(rsvr->log, "Alloc from slab failed!");
            break;
        }

        extra->sck = cmd_sck;
        cmd_sck->extra = extra;
        cmd_sck->fd = ctx->rsvr_cmd_fd[idx].fd[0];

        ftime(&cmd_sck->ctm);
        cmd_sck->wrtm = cmd_sck->rdtm = cmd_sck->ctm.time;
        cmd_sck->recv_cb = (socket_recv_cb_t)agt_rsvr_recv_cmd_hdl;
        cmd_sck->send_cb = NULL;

        /* > 加入事件侦听 */
        memset(&ev, 0, sizeof(ev));

        ev.data.ptr = cmd_sck;
        ev.events = EPOLLIN | EPOLLET; /* 边缘触发 */

        epoll_ctl(rsvr->epid, EPOLL_CTL_ADD, cmd_sck->fd, &ev);

        return ACC_OK;
    } while(0);

    acc_rsvr_destroy(rsvr);
    return ACC_ERR;
}

/******************************************************************************
 **函数名称: acc_rsvr_sck_dealloc
 **功    能: 释放SCK对象的空间
 **输入参数:
 **     pool: 内存池
 **     sck: 需要被释放的套接字对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次释放所有内存空间
 **注意事项: 
 **作    者: # Qifeng.zou # 2015.07.22 21:39:05 #
 ******************************************************************************/
static int acc_rsvr_sck_dealloc(void *pool, socket_t *sck)
{
    acc_socket_extra_t *extra = sck->extra;

    FREE(sck);
    list_destroy(extra->send_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
    FREE(extra);

    return 0;
}

/******************************************************************************
 **函数名称: acc_rsvr_destroy
 **功    能: 销毁Agent线程
 **输入参数:
 **     rsvr: 接收对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次释放所有内存空间
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.11.18 #
 ******************************************************************************/
int acc_rsvr_destroy(acc_rsvr_t *rsvr)
{
    FREE(rsvr->events);
    CLOSE(rsvr->epid);
    CLOSE(rsvr->cmd_sck.fd);
    FREE(rsvr->cmd_sck.extra);
    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_rsvr_self
 **功    能: 获取代理对象
 **输入参数: 
 **     ctx: 全局信息
 **输出参数: NONE
 **返    回: 代理对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.11.26 #
 ******************************************************************************/
static acc_rsvr_t *acc_rsvr_self(acc_cntx_t *ctx)
{
    int id;
    acc_rsvr_t *rsvr;

    id = thread_pool_get_tidx(ctx->rsvr_pool);
    if (id < 0) {
        return NULL;
    }

    rsvr = thread_pool_get_args(ctx->rsvr_pool);

    return rsvr + id;
}

/******************************************************************************
 **函数名称: acc_rsvr_event_hdl
 **功    能: 事件通知处理
 **输入参数: 
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 代理对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.11.28 #
 ******************************************************************************/
static int acc_rsvr_event_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
    int idx, ret;
    socket_t *sck;

    rsvr->ctm = time(NULL);

    /* 1. 依次遍历套接字, 判断是否可读可写 */
    for (idx=0; idx<rsvr->fds; ++idx) {
        sck = (socket_t *)rsvr->events[idx].data.ptr;

        /* 1.1 判断是否可读 */
        if (rsvr->events[idx].events & EPOLLIN) {
            /* 接收网络数据 */
            ret = sck->recv_cb(ctx, rsvr, sck);
            if (ACC_SCK_AGAIN != ret) {
                log_info(rsvr->log, "Delete connection! fd:%d", sck->fd);
                acc_rsvr_del_conn(ctx, rsvr, sck);
                continue; /* 异常-关闭SCK: 不必判断是否可写 */
            }
        }

        /* 1.2 判断是否可写 */
        if (rsvr->events[idx].events & EPOLLOUT) {
            /* 发送网络数据 */
            ret = sck->send_cb(ctx, rsvr, sck);
            if (ACC_ERR == ret) {
                log_info(rsvr->log, "Delete connection! fd:%d", sck->fd);
                acc_rsvr_del_conn(ctx, rsvr, sck);
                continue; /* 异常: 套接字已关闭 */
            }
        }
    }

    /* 2. 超时扫描 */
    if (rsvr->ctm - rsvr->scan_tm > ACC_TMOUT_SCAN_SEC) {
        rsvr->scan_tm = rsvr->ctm;

        acc_rsvr_timeout_hdl(ctx, rsvr);
    }

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_rsvr_get_timeout_conn_list
 **功    能: 将超时连接加入链表
 **输入参数: 
 **     node: 平衡二叉树结点
 **     timeout: 超时链表
 **输出参数: NONE
 **返    回: 代理对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.24 #
 ******************************************************************************/
static int acc_rsvr_get_timeout_conn_list(
        acc_socket_extra_t *extra, acc_conn_timeout_list_t *timeout)
{
#define ACC_SCK_TIMEOUT_SEC (180)
    socket_t *sck = extra->sck;

    /* 判断是否超时，则加入到timeout链表中 */
    if ((timeout->ctm - sck->rdtm <= ACC_SCK_TIMEOUT_SEC)
        || (timeout->ctm - sck->wrtm <= ACC_SCK_TIMEOUT_SEC))
    {
        return ACC_OK; /* 未超时 */
    }

    return list_lpush(timeout->list, sck);
}

/******************************************************************************
 **函数名称: acc_rsvr_conn_timeout
 **功    能: 删除超时连接
 **输入参数: 
 **     ctx: 全局信息
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2015.06.11 15:02:33 #
 ******************************************************************************/
static int acc_rsvr_conn_timeout(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
    socket_t *sck;
    acc_socket_extra_t key;
    acc_conn_timeout_list_t timeout;

    memset(&timeout, 0, sizeof(timeout));

    timeout.ctm = rsvr->ctm;

    do {
        /* > 创建链表 */
        timeout.list = list_creat(NULL);
        if (NULL == timeout.list) {
            log_error(rsvr->log, "Create list failed!");
            break;
        }

        /* > 获取超时连接 */
        key.cid = rsvr->id;

        hash_tab_trav_slot(ctx->conn_cid_tab, &key,
                (trav_cb_t)acc_rsvr_get_timeout_conn_list, &timeout, RDLOCK);

        log_debug(rsvr->log, "Timeout connections: %d!", timeout.list->num);

        /* > 删除超时连接 */
        for (;;) {
            sck = (socket_t *)list_lpop(timeout.list);
            if (NULL == sck) {
                break;
            }
            acc_rsvr_del_conn(ctx, rsvr, sck);
        }
    } while(0);

    /* > 释放内存空间 */
    list_destroy(timeout.list, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_rsvr_timeout_hdl
 **功    能: 事件超时处理
 **输入参数: 
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **     不必依次释放超时链表各结点的空间，只需一次性释放内存池便可释放所有空间.
 **作    者: # Qifeng.zou # 2016.11.28 #
 ******************************************************************************/
static int acc_rsvr_timeout_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
    acc_rsvr_conn_timeout(ctx, rsvr);
    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_rsvr_add_conn
 **功    能: 添加新的连接
 **输入参数: 
 **     ctx: 全局信息
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.17 #
 ******************************************************************************/
static int acc_rsvr_add_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
#define AGT_RSVR_CONN_POP_NUM (1024)
    int num, idx;
    socket_t *sck;
    queue_t *connq;
    struct epoll_event ev;
    time_t ctm = time(NULL);
    acc_socket_extra_t *extra;
    acc_add_sck_t *add[AGT_RSVR_CONN_POP_NUM];

    connq = rsvr->connq;
    while (1) {
        num = MIN(queue_used(connq), AGT_RSVR_CONN_POP_NUM);
        if (0 == num) {
            return ACC_OK;
        }

        /* > 取数据 */
        num = queue_mpop(connq, (void **)add, num);
        if (0 == num) {
            continue;
        }

        for (idx=0; idx<num; ++idx) {
            /* > 申请SCK空间 */
            sck = (socket_t *)calloc(1, sizeof(socket_t));
            if (NULL == sck) {
                log_error(rsvr->log, "Alloc memory failed! cid:%lu", add[idx]->cid);
                CLOSE(add[idx]->fd);
                queue_dealloc(connq, add[idx]);
                continue;
            }

            memset(sck, 0, sizeof(socket_t));

            /* > 创建SCK关联对象 */
            extra = (acc_socket_extra_t *)calloc(1, sizeof(acc_socket_extra_t));
            if (NULL == extra) {
                log_error(rsvr->log, "Alloc memory failed! cid:%lu", add[idx]->cid);
                CLOSE(add[idx]->fd);
                FREE(sck);
                queue_dealloc(connq, add[idx]);
                continue;
            }

            extra->user = (void *)calloc(1, ctx->protocol->per_session_data_size);
            if (NULL == extra->user) {
                log_error(rsvr->log, "Alloc memory failed! cid:%lu", add[idx]->cid);
                CLOSE(add[idx]->fd);
                FREE(sck);
                FREE(extra);
                queue_dealloc(connq, add[idx]);
                continue;
            }

            extra->sck = sck;
            extra->rid = rsvr->id;
            extra->cid = add[idx]->cid;
            extra->send_list = list_creat(NULL);
            if (NULL == extra->send_list) {
                log_error(rsvr->log, "Create send list failed! cid:%lu", add[idx]->cid);
                CLOSE(add[idx]->fd);
                FREE(extra->user);
                FREE(extra);
                FREE(sck);
                queue_dealloc(connq, add[idx]);
                continue;
            }

            sck->extra = extra;

            /* > 设置SCK信息 */
            sck->fd = add[idx]->fd;
            sck->wrtm = sck->rdtm = ctm;/* 记录当前时间 */
            memcpy(&sck->ctm, &add[idx]->ctm, sizeof(add[idx]->ctm)); /* 创建时间 */

            sck->recv.phase = SOCK_PHASE_RECV_INIT;
            sck->recv_cb = (socket_recv_cb_t)acc_recv_data;   /* Recv回调函数 */
            sck->send_cb = (socket_send_cb_t)acc_send_data;   /* Send回调函数*/

            queue_dealloc(connq, add[idx]);      /* 释放连接队列空间 */

            /* > 插入红黑树中(以序列号为主键) */
            if (acc_conn_cid_tab_add(ctx, extra)) {
                log_error(rsvr->log, "Insert into avl failed! fd:%d cid:%lu",
                          sck->fd, extra->cid);
                CLOSE(sck->fd);
                list_destroy(extra->send_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
                FREE(extra->user);
                FREE(sck->extra);
                FREE(sck);
                return ACC_ERR;
            }

            log_debug(rsvr->log, "Insert into avl success! fd:%d cid:%lu",
                      sck->fd, extra->cid);

            ctx->protocol->callback(ctx, sck, ACC_CALLBACK_SCK_CREAT,
                    (void *)extra->user, (void *)NULL, 0, (void *)ctx->protocol->args);

            /* > 加入epoll监听(首先是接收客户端搜索请求, 所以设置EPOLLIN) */
            memset(&ev, 0, sizeof(ev));

            ev.data.ptr = sck;
            ev.events = EPOLLIN | EPOLLET; /* 边缘触发 */

            epoll_ctl(rsvr->epid, EPOLL_CTL_ADD, sck->fd, &ev);
            ++rsvr->conn_total;
        }
    }

    return ACC_ERR;
}

/******************************************************************************
 **函数名称: acc_rsvr_del_conn
 **功    能: 删除指定套接字
 **输入参数:
 **     rsvr: 接收服务
 **     sck: SCK对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次释放套接字对象各成员的空间
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 #
 ******************************************************************************/
static int acc_rsvr_del_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck)
{
    acc_socket_extra_t *extra = sck->extra;

    log_trace(rsvr->log, "fd:%d cid:%ld", sck->fd, extra->cid);

    /* > 剔除CID对象 */
    acc_conn_cid_tab_del(ctx, extra->cid);

    /* > 释放套接字空间 */
    CLOSE(sck->fd);

    ctx->protocol->callback(ctx, sck, ACC_CALLBACK_SCK_CLOSED,
            (void *)extra->user, (void *)NULL, 0, (void *)ctx->protocol->args);
    ctx->protocol->callback(ctx, sck, ACC_CALLBACK_SCK_DESTROY,
            (void *)extra->user, (void *)NULL, 0, (void *)ctx->protocol->args);

    list_destroy(extra->send_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
    if (sck->recv.addr) {
        mem_ref_decr(sck->recv.addr);
    }
    FREE(extra->user);
    FREE(extra);
    FREE(sck);

    --rsvr->conn_total;
    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_recv_head
 **功    能: 接收报头
 **输入参数:
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **     sck: SCK对象
 **     per_packet_head_size: 报头的大小
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.01 #
 ******************************************************************************/
static int acc_recv_head(acc_cntx_t *ctx,
        acc_rsvr_t *rsvr, socket_t *sck, size_t per_packet_head_size)
{
    void *addr;
    int n, left;
    socket_snap_t *recv = &sck->recv;

    addr = recv->addr;

    while (1) {
        /* 1. 计算剩余字节 */
        left = per_packet_head_size - recv->off;

        /* 2. 接收报头数据 */
        n = read(sck->fd, addr + recv->off, left);
        if (n == left) {
            recv->off += n;
            break; /* 接收完毕 */
        } else if (n > 0) {
            recv->off += n;
            continue;
        } else if (0 == n || ECONNRESET == errno) {
            log_info(rsvr->log, "Client disconnected. errmsg:[%d] %s! fd:[%d] n:[%d/%d]",
                    errno, strerror(errno), sck->fd, n, left);
            return ACC_SCK_CLOSE;
        } else if ((n < 0) && (EAGAIN == errno)) {
            return ACC_SCK_AGAIN; /* 等待下次事件通知 */
        } else if (EINTR == errno) {
            continue; 
        }
        log_error(rsvr->log, "errmsg:[%d] %s. fd:[%d]", errno, strerror(errno), sck->fd);
        return ACC_ERR;
    }
    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_recv_body
 **功    能: 接收报体
 **输入参数:
 **     rsvr: 接收服务
 **     sck: SCK对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.02 #
 ******************************************************************************/
static int acc_recv_body(acc_rsvr_t *rsvr, socket_t *sck)
{
    int n, left;
    socket_snap_t *recv = &sck->recv;

    /* 1. 接收报体 */
    while (1) {
        left = recv->total - recv->off;

        n = read(sck->fd, recv->addr + recv->off, left);
        if (n == left) {
            recv->off += n;
            break; /* 接收完毕 */
        } else if (n > 0) {
            recv->off += n;
            continue;
        } else if (0 == n) {
            log_info(rsvr->log, "Client disconnected. errmsg:[%d] %s! fd:[%d] n:[%d/%d]",
                    errno, strerror(errno), sck->fd, n, left);
            return ACC_SCK_CLOSE;
        } else if ((n < 0) && (EAGAIN == errno)) {
            return ACC_SCK_AGAIN;
        }

        if (EINTR == errno) {
            continue;
        }

        return ACC_ERR;
    }

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_recv_post_hdl
 **功    能: 数据接收完毕，进行数据处理
 **输入参数:
 **     ctx: 全局变量
 **     rsvr: 接收服务
 **     sck: SCK对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.17 #
 ******************************************************************************/
static int acc_recv_post_hdl(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck)
{
    acc_protocol_t *protocol = ctx->protocol;
    acc_socket_extra_t *extra = (acc_socket_extra_t *)sck->extra;

    return protocol->callback(ctx, sck, ACC_CALLBACK_RECEIVE,
            (void *)extra->user, (void *)extra->head,
            protocol->get_packet_body_size(extra->head) + protocol->per_packet_head_size,
            (void *)ctx->protocol->args);
}

/******************************************************************************
 **函数名称: acc_recv_data
 **功    能: 接收数据
 **输入参数:
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **     sck: SCK对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: TODO: 此处理流程可进一步进行优化
 **作    者: # Qifeng.zou # 2016.09.17 #
 ******************************************************************************/
static int acc_recv_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck)
{
    int ret;
    socket_snap_t *recv = &sck->recv;
    queue_conf_t *conf = &ctx->conf->recvq;
    acc_protocol_t *protocol = ctx->protocol;
    acc_socket_extra_t *extra = (acc_socket_extra_t *)sck->extra;

    for (;;) {
        switch (recv->phase) {
            case SOCK_PHASE_RECV_INIT: /* 1. 分配空间 */
                recv->addr = mem_ref_alloc(conf->size,
                        NULL, (mem_alloc_cb_t)mem_alloc, (mem_dealloc_cb_t)mem_dealloc);
                if (NULL == recv->addr) {
                    log_error(rsvr->log, "Alloc from queue failed!");
                    return ACC_ERR;
                }

                log_info(rsvr->log, "Alloc memory from queue success!");

                extra->head = (void *)recv->addr;
                extra->body = (void *)(recv->addr + protocol->per_packet_head_size);
                recv->off = 0;
                recv->total = protocol->per_packet_head_size;

                /* 设置下步 */
                recv->phase = SOCK_PHASE_RECV_HEAD;

                goto RECV_HEAD;
            case SOCK_PHASE_RECV_HEAD: /* 2. 接收报头 */
            RECV_HEAD:
                ret = acc_recv_head(ctx, rsvr, sck, protocol->per_packet_head_size);
                switch (ret) {
                    case ACC_OK:
                        if (protocol->get_packet_body_size(recv->addr)) {
                            recv->phase = SOCK_PHASE_READY_BODY; /* 设置下步 */
                        } else {
                            recv->phase = SOCK_PHASE_RECV_POST; /* 设置下步 */
                            goto RECV_POST;
                        }
                        break;      /* 继续后续处理 */
                    case ACC_SCK_AGAIN:
                        return ret; /* 下次继续处理 */
                    default:
                        mem_ref_decr(recv->addr);
                        recv->addr = NULL;
                        return ret; /* 异常情况 */
                }
                goto READY_BODY;
            case SOCK_PHASE_READY_BODY: /* 3. 准备接收报体 */
            READY_BODY:
                recv->total += protocol->get_packet_body_size(recv->addr);

                /* 设置下步 */
                recv->phase = SOCK_PHASE_RECV_BODY;

                goto RECV_BODY;
            case SOCK_PHASE_RECV_BODY: /* 4. 接收报体 */
            RECV_BODY:
                ret = acc_recv_body(rsvr, sck);
                switch (ret) {
                    case ACC_OK:
                        recv->phase = SOCK_PHASE_RECV_POST; /* 设置下步 */
                        break;      /* 继续后续处理 */
                    case ACC_SCK_AGAIN:
                        return ret; /* 下次继续处理 */
                    default:
                        mem_ref_decr(recv->addr);
                        recv->addr = NULL;
                        return ret; /* 异常情况 */
                }
                goto RECV_POST;
            case SOCK_PHASE_RECV_POST: /* 5. 接收完毕: 数据处理 */
            RECV_POST:
                /* 对接收到的数据进行处理 */
                ret = acc_recv_post_hdl(ctx, rsvr, sck);
                mem_ref_decr(recv->addr);
                recv->addr = NULL;
                if (ACC_OK == ret) {
                    recv->phase = SOCK_PHASE_RECV_INIT;
                    recv->addr = NULL;
                    continue; /* 接收下一条数据 */
                }
                return ACC_ERR;
        }
    }

    return ACC_ERR;
}

/******************************************************************************
 **函数名称: acc_send_data
 **功    能: 发送数据
 **输入参数:
 **     rsvr: 接收服务
 **     sck: SCK对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.17 #
 ******************************************************************************/
static int acc_send_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr, socket_t *sck)
{
    int n, left;
    acc_send_item_t *item;
    struct epoll_event ev;
    socket_snap_t *send = &sck->send;
    acc_socket_extra_t *extra = (acc_socket_extra_t *)sck->extra;

    sck->wrtm = time(NULL);

    for (;;) {
        /* 1. 取发送的数据 */
        if (NULL == send->addr) {
            item = list_lpop(extra->send_list);
            if (NULL == item) {
                return ACC_OK; /* 无数据 */
            }

            send->addr = item->data;
            send->off = 0;
            send->total = item->len;

            log_trace(rsvr->log, "cid:%lu!", item->cid);
            free(item);
        }

        /* 2. 发送数据 */
        left = send->total - send->off;

        n = Writen(sck->fd, send->addr+send->off, left);
        if (n != left) {
            if (n > 0) {
                send->off += n;
                return ACC_SCK_AGAIN;
            }

            log_error(rsvr->log, "errmsg:[%d] %s!", errno, strerror(errno));

            /* 释放空间 */
            FREE(send->addr);
            send->addr = NULL;
            return ACC_ERR;
        }

        /* 3. 释放空间 */
        FREE(send->addr);
        send->addr = NULL;

        /* > 设置epoll监听 */
        memset(&ev, 0, sizeof(ev));

        ev.data.ptr = sck;
        ev.events = list_empty(extra->send_list)?
            (EPOLLIN|EPOLLET) : (EPOLLIN|EPOLLOUT|EPOLLET); /* 边缘触发 */

        epoll_ctl(rsvr->epid, EPOLL_CTL_MOD, sck->fd, &ev);
    }

    return ACC_ERR;
}

/******************************************************************************
 **函数名称: acc_rsvr_dist_send_data
 **功    能: 分发发送的数据
 **输入参数:
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 千万勿将共享变量参与MIN()三目运算, 否则可能出现严重错误!!!!且很难找出原因!
 **作    者: # Qifeng.zou # 2015-06-05 17:35:02 #
 ******************************************************************************/
static int acc_rsvr_dist_send_data(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
    int num, idx;
    ring_t *sendq;
    socket_t *sck;
    acc_send_item_t *item;
    struct epoll_event ev;
    void *addr[AGT_RSVR_DIST_POP_NUM];

    sendq = rsvr->sendq;
    while (1) {
        num = MIN(ring_used(sendq), AGT_RSVR_DIST_POP_NUM);
        if (0 == num) {
            break;
        }

        /* > 弹出应答数据 */
        num = ring_mpop(sendq, addr, num);
        if (0 == num) {
            break;
        }

        log_debug(rsvr->log, "Pop data succ! num:%d", num);

        for (idx=0; idx<num; ++idx) {
            item = (acc_send_item_t *)addr[idx];

            /* > 发入发送列表 */
            sck = acc_push_into_send_list(ctx, rsvr, item->cid, addr[idx]); 
            if (NULL == sck) {
                log_error(ctx->log, "Query socket failed! cid:%lu", item->cid);
                FREE(addr[idx]);
                continue;
            }

            /* > 设置epoll监听(添加EPOLLOUT) */
            memset(&ev, 0, sizeof(ev));

            ev.data.ptr = sck;
            ev.events = EPOLLOUT | EPOLLET; /* 边缘触发 */

            epoll_ctl(rsvr->epid, EPOLL_CTL_MOD, sck->fd, &ev);
        }
    }

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_push_into_send_list
 **功    能: 将数据放入发送列表
 **输入参数:
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **     cid: 会话ID
 **     addr: 需要发送的数据
 **输出参数: NONE
 **返    回: 连接对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016-07-24 22:49:02 #
 ******************************************************************************/
static socket_t *acc_push_into_send_list(
        acc_cntx_t *ctx, acc_rsvr_t *rsvr, uint64_t cid, void *addr)
{
    socket_t *sck;
    acc_socket_extra_t *extra, key;

    /* > 查询会话对象 */
    key.cid = cid;

    extra = hash_tab_query(ctx->conn_cid_tab, &key, WRLOCK);
    if (NULL == extra) {
        log_error(ctx->log, "Query connection by cid failed! cid:%lu", cid);
        return NULL;
    }

    sck = extra->sck;

    /* > 放入发送列表 */
    if (list_rpush(extra->send_list, addr)) {
        hash_tab_unlock(ctx->conn_cid_tab, &key, WRLOCK);
        log_error(ctx->log, "Push data into send list failed! cid:%lu", cid);
        return NULL;
    }

    hash_tab_unlock(ctx->conn_cid_tab, &key, WRLOCK);

    return sck;
}

/******************************************************************************
 **函数名称: acc_rsvr_kick_conn
 **功    能: 踢某连接
 **输入参数:
 **     ctx: 全局对象
 **     rsvr: 接收服务
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 千万勿将共享变量参与MIN()三目运算, 否则可能出现严重错误!!!!且很难找出原因!
 **作    者: # Qifeng.zou # 2016-10-01 19:08:57 #
 ******************************************************************************/
static int acc_rsvr_kick_conn(acc_cntx_t *ctx, acc_rsvr_t *rsvr)
{
    int num, idx;
    socket_t *sck;
    queue_t *kickq;
    acc_kick_req_t *kick;
    acc_socket_extra_t *extra, key;
    void *addr[AGT_RSVR_DIST_POP_NUM];

    kickq = rsvr->kickq;
    while (1) {
        num = MIN(queue_used(kickq), AGT_RSVR_DIST_POP_NUM);
        if (0 == num) {
            break;
        }

        /* > 弹出应答数据 */
        num = queue_mpop(kickq, addr, num);
        if (0 == num) {
            break;
        }

        log_debug(rsvr->log, "Pop data succ! num:%d", num);

        for (idx=0; idx<num; ++idx) {
            kick = (acc_kick_req_t *)addr[idx];

            /* > 查询会话对象 */
            key.cid = kick->cid;

            extra = hash_tab_query(ctx->conn_cid_tab, &key, RDLOCK);
            if (NULL == extra) {
                continue;
            }

            sck = extra->sck;

            hash_tab_unlock(ctx->conn_cid_tab, &key, RDLOCK);

            acc_rsvr_del_conn(ctx, rsvr, sck);
        }
    }

    return ACC_OK;
}



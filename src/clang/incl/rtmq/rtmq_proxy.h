#if !defined(__RTMQ_PROXY_H__)
#define __RTMQ_PROXY_H__

#include "rtmq_comm.h"
#include "rtmq_proxy_ssvr.h"

#define RTMQ_IPADD_MAX_NUM  (30)

/* 配置信息 */
typedef struct
{
    int nid;                            /* 结点ID: 唯一值 */
    int gid;                            /* 分组ID */
    char path[FILE_LINE_MAX_LEN];       /* 工作路径 */

    struct {
        char usr[RTMQ_USR_MAX_LEN];     /* 用户名 */
        char passwd[RTMQ_PWD_MAX_LEN];  /* 登录密码 */
    } auth;                             /* 鉴权信息 */

    /* 服务端地址(格式:${IP1}:${PORT1},${IP2}:${PORT2},${IP...x}:${PORT...x}) */
    char ipaddr[RTMQ_IPADD_MAX_NUM*IP_ADDR_MAX_LEN];

    int send_thd_num;                   /* 发送线程数 */
    int work_thd_num;                   /* 工作线程数 */

    size_t recv_buff_size;              /* 接收缓存大小 */

    rtmq_cpu_conf_t cpu;                /* CPU亲和性配置 */

    queue_conf_t sendq;                 /* 发送队列配置 */
    queue_conf_t recvq;                 /* 接收队列配置 */
} rtmq_proxy_conf_t;

/* 全局信息 */
typedef struct
{
    rtmq_proxy_conf_t conf;             /* 配置信息 */
    log_cycle_t *log;                   /* 日志对象 */
    list_t *iplist;                     /* 服务端IP列表(注: 存储iplist_item_t对象) */

    thread_pool_t *sendtp;              /* 发送线程池 */

    thread_pool_t *worktp;              /* 工作线程池 */

    avl_tree_t *reg;                    /* 回调注册对象(注: 存储rtmq_reg_t数据) */

    rtmq_pipe_t *work_cmd_fd;           /* 工作线程通信FD */
    queue_t **recvq;                    /* 接收队列(数组长度与conf->send_thd_num一致) */

    rtmq_pipe_t *send_cmd_fd;           /* 发送线程通信FD */
    queue_t **sendq;                    /* 发送缓存(数组长度与conf->send_thd_num一致) */
} rtmq_proxy_t;

/* 内部接口 */
int rtmq_proxy_ssvr_init(rtmq_proxy_t *pxy,
        rtmq_proxy_ssvr_t *ssvr, int tidx,
        const char *ipaddr, int port, queue_t *sendq, rtmq_pipe_t *pipe);
void *rtmq_proxy_ssvr_routine(void *_ctx);

int rtmq_proxy_worker_init(rtmq_proxy_t *pxy, rtmq_worker_t *worker, int tidx);
void *rtmq_proxy_worker_routine(void *_ctx);

rtmq_worker_t *rtmq_proxy_worker_get_by_idx(rtmq_proxy_t *pxy, int idx);

/* 对外接口 */
rtmq_proxy_t *rtmq_proxy_init(const rtmq_proxy_conf_t *conf, log_cycle_t *log);
int rtmq_proxy_launch(rtmq_proxy_t *pxy);
int rtmq_proxy_reg_add(rtmq_proxy_t *pxy, int type, rtmq_reg_cb_t proc, void *args);
int rtmq_proxy_async_send(rtmq_proxy_t *pxy, int type, const void *data, size_t size);

#endif /*__RTMQ_PROXY_H__*/

#if !defined(__SDK_SSVR_H__)
#define __SDK_SSVR_H__

#include "log.h"
#include "list.h"
#include "avl_tree.h"
#include "sdk_comm.h"
#include "thread_pool.h"

#define SDK_PING_MIN_SEC    (5)
#define SDK_PING_MAX_SEC    (300)
#define SDK_CONN_MAX_SEC    (300)
#define SDK_SLEEP_MAX_SEC   (300)

/* COMM的UNIX-UDP路径 */
#define sdk_comm_usck_path(conf, _path) \
    snprintf(_path, sizeof(_path), "%s/.sdk/%d_comm.usck", (conf)->path, (conf)->nid)
/* SSVR线程的UNIX-UDP路径 */
#define sdk_ssvr_usck_path(conf, _path) \
    snprintf(_path, sizeof(_path), "%s/.sdk/%d_ssvr.usck", (conf)->path, (conf)->nid)
/* WORKER线程的UNIX-UDP路径 */
#define sdk_worker_usck_path(conf, _path, id) \
    snprintf(_path, sizeof(_path), "%s/.sdk/%d_swrk_%d.usck", (conf)->path, (conf)->nid, id+1)
/* 加锁路径 */
#define sdk_lock_path(conf, _path) \
    snprintf(_path, sizeof(_path), "%s/.sdk/%d.lock", (conf)->path, (conf)->nid)

typedef struct
{
    pthread_mutex_t lock;               /* 互斥锁 */
    list_t *list;                       /* 队列 */
} sdk_queue_t;

/* 套接字信息 */
typedef struct
{
    int fd;                             /* 套接字ID */
    time_t wrtm;                        /* 最近写入操作时间 */
    time_t rdtm;                        /* 最近读取操作时间 */

    int kpalive_times;                  /* 保活次数 */
    time_t next_kpalive_tm;             /* 下次发送保活请求的时间 */
    list_t *mesg_list;                  /* 发送链表 */

    sdk_snap_t recv;                   /* 接收快照 */
    wiov_t send;                       /* 发送信息 */
} sdk_sck_t;

/* 连接信息 */
typedef struct
{
    time_t expire;                      /* TOKEN过期时间(更新时间+expire) */
#define SDK_TOKEN_MAX_LEN   (128)
    char token[SDK_TOKEN_MAX_LEN];      /* 鉴权TOKEN */
    list_t *iplist;                     /* IP列表(数据类型ip_port_t) */
    uint64_t sid;                       /* 会话ID */
} sdk_conn_info_t;

/* SND线程上下文 */
typedef struct
{
    void *ctx;                          /* 存储sdk_t对象 */
    log_cycle_t *log;                   /* 日志对象 */
    sdk_queue_t *sendq;                 /* 发送缓存 */

    int cmd_sck_id;                     /* 命令通信套接字ID */
    sdk_sck_t sck;                      /* 发送套接字 */

    int max;                            /* 套接字最大值 */
    fd_set rset;                        /* 读集合 */
    fd_set wset;                        /* 写集合 */

    int try_conn_times;                 /* 尝试连接的次数(连接成功后清零) */
    time_t next_conn_tm;                /* 下一次尝试连接的时间(通过此值控制重连的频率) */

    int try_ping_times;                 /* 尝试PING的次数(收到PONG后清零) */
    time_t next_ping_tm;                /* 下一次发送PING的时间 */

    bool is_online_succ;                /* 上线是否成功 */
    sdk_conn_info_t conn_info;          /* CONN INFO信息 */

    /* 统计信息 */
    uint64_t recv_total;                /* 获取的数据总条数 */
    uint64_t err_total;                 /* 错误的数据条数 */
    uint64_t drop_total;                /* 丢弃的数据条数 */
} sdk_ssvr_t;

#define SDK_SSVR_SET_ONLINE(ssvr, ok) ((ssvr)->is_online_succ = (ok))
#define SDK_SSVR_GET_ONLINE(ssvr) ((ssvr)->is_online_succ)

#endif /*__SDK_SSVR_H__*/

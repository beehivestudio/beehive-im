#if !defined(__ACCESS_H__)
#define __ACCESS_H__

#include "sck.h"
#include "slot.h"
#include "queue.h"
#include "rb_tree.h"
#include "spinlock.h"
#include "avl_tree.h"
#include "shm_queue.h"
#include "acc_comm.h"
#include "acc_lsn.h"
#include "thread_pool.h"

/* 宏定义 */
#define ACC_TMOUT_SCAN_SEC    (15)        /* 超时扫描间隔 */
#define ACC_MSG_TYPE_MAX      (0xFF)      /* 消息最大类型 */

/* 命令路径 */
#define ACC_LSVR_CMD_PATH     "lsvr-cmd-%02d.usck" /* 侦听服务 */
#define ACC_RSVR_CMD_PATH     "rsvr-cmd-%02d.usck" /* 接收服务 */
#define ACC_WSVR_CMD_PATH     "wsvr-cmd-%02d.usck" /* 工作服务 */
#define ACC_CLI_CMD_PATH      "cli_cmd.usck"       /* 客户端 */

#define acc_cli_cmd_usck_path(conf, path, size)       /* 客户端命令路径 */\
    snprintf(path, size, "%s"ACC_CLI_CMD_PATH, (conf)->path)
#define acc_lsvr_cmd_usck_path(conf, idx, path, size)      /* 侦听服务命令路径 */\
    snprintf(path, size, "%s"ACC_LSVR_CMD_PATH, (conf)->path, idx)
#define acc_rsvr_cmd_usck_path(conf, rid, path, size) /* 接收服务命令路径 */\
    snprintf(path, size, "%s"ACC_RSVR_CMD_PATH, (conf)->path, rid)
#define acc_wsvr_cmd_usck_path(conf, wid, path, size) /* 工作服务命令路径 */\
    snprintf(path, size, "%s"ACC_WSVR_CMD_PATH, (conf)->path, wid)

/* 配置信息 */
typedef struct
{
    int nid;                                /* 结点ID */
    char path[FILE_NAME_MAX_LEN];           /* 工作路径 */

    struct {
        int max;                            /* 最大并发数 */
        int timeout;                        /* 连接超时时间 */
        int port;                           /* 侦听端口 */
    } connections;

    int lsn_num;                            /* Listen线程数 */
    int acc_num;                          /* Agent线程数 */
    int worker_num;                         /* Worker线程数 */

    queue_conf_t connq;                     /* 连接队列 */
    queue_conf_t recvq;                     /* 接收队列 */
    queue_conf_t sendq;                     /* 发送队列 */
} acc_conf_t;

/* SID列表 */
typedef struct
{
    spinlock_t lock;                        /* 锁 */
    rbt_tree_t *sids;                       /* SID列表 */
} acc_sid_list_t;

/* 代理对象 */
typedef struct
{
    acc_conf_t *conf;                     /* 配置信息 */
    log_cycle_t *log;                       /* 日志对象 */
    int cmd_sck_id;                         /* 命令套接字 */

    /* 侦听信息 */
    struct {
        int lsn_sck_id;                     /* 侦听套接字 */
        spinlock_t accept_lock;             /* 侦听锁 */
        uint64_t sid;                       /* Session ID */
        acc_lsvr_t *lsvr;                 /* 侦听对象 */
    } listen;

    thread_pool_t *agents;                  /* Agent线程池 */
    thread_pool_t *listens;                 /* Listen线程池 */
    thread_pool_t *workers;                 /* Worker线程池 */
    acc_reg_t reg[ACC_MSG_TYPE_MAX];    /* 消息注册 */

    acc_sid_list_t *connections;          /* SID集合(注:数组长度与Agent相等) */

    queue_t **connq;                        /* 连接队列(注:数组长度与Agent相等) */
    ring_t **recvq;                         /* 接收队列(注:数组长度与Agent相等) */
    ring_t **sendq;                         /* 发送队列(注:数组长度与Agent相等) */
} acc_cntx_t;

#define ACC_GET_NODE_ID(ctx) ((ctx)->conf->nid)

/* 内部接口 */
int acc_listen_init(acc_cntx_t *ctx, acc_lsvr_t *lsn, int idx);

int acc_sid_item_add(acc_cntx_t *ctx, uint64_t sid, socket_t *sck);
socket_t *acc_sid_item_del(acc_cntx_t *ctx, uint64_t sid);
int acc_get_aid_by_sid(acc_cntx_t *ctx, uint64_t sid);

/* 外部接口 */
acc_cntx_t *acc_init(acc_conf_t *conf, log_cycle_t *log);
int acc_launch(acc_cntx_t *ctx);
int acc_reg_add(acc_cntx_t *ctx, unsigned int type, acc_reg_cb_t proc, void *args);
void acc_destroy(acc_cntx_t *ctx);

int acc_async_send(acc_cntx_t *ctx, int type, uint64_t sid, void *data, int len);

#endif /*__ACCESS_H__*/

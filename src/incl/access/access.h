#if !defined(__ACCESS_H__)
#define __ACCESS_H__

#include "sck.h"
#include "slot.h"
#include "queue.h"
#include "rb_tree.h"
#include "spinlock.h"
#include "avl_tree.h"
#include "hash_tab.h"
#include "shm_queue.h"
#include "acc_comm.h"
#include "acc_lsn.h"
#include "thread_pool.h"

/* 宏定义 */
#define ACC_TMOUT_SCAN_SEC    (15)        /* 超时扫描间隔 */

/* 命令路径 */
#define ACC_LSVR_CMD_PATH     "lsvr-cmd-%02d.usck" /* 侦听服务 */
#define ACC_RSVR_CMD_PATH     "rsvr-cmd-%02d.usck" /* 接收服务 */
#define ACC_CLI_CMD_PATH      "cli_cmd.usck"       /* 客户端 */

#define acc_cli_cmd_usck_path(conf, path, size)       /* 客户端命令路径 */\
    snprintf(path, size, "%s"ACC_CLI_CMD_PATH, (conf)->path)
#define acc_lsvr_cmd_usck_path(conf, idx, path, size)      /* 侦听服务命令路径 */\
    snprintf(path, size, "%s"ACC_LSVR_CMD_PATH, (conf)->path, idx)
#define acc_rsvr_cmd_usck_path(conf, rid, path, size) /* 接收服务命令路径 */\
    snprintf(path, size, "%s"ACC_RSVR_CMD_PATH, (conf)->path, rid)

/* 配置信息 */
typedef struct {
    int nid;                        /* 结点ID */
    char path[FILE_NAME_MAX_LEN];   /* 工作路径 */

    struct {
        int max;                    /* 最大并发数 */
        int timeout;                /* 连接超时时间 */
        int port;                   /* 侦听端口 */
    } connections;

    int lsvr_num;                   /* 帧听线程数 */
    int rsvr_num;                   /* 接收线程数 */

    queue_conf_t connq;             /* 连接队列 */
    queue_conf_t recvq;             /* 接收队列 */
    queue_conf_t sendq;             /* 发送队列 */
} acc_conf_t;

typedef struct _acc_cntx_t acc_cntx_t;
typedef int (*acc_callback_t)(acc_cntx_t *ctx, socket_t *asi, int reason, void *user, void *in, int len, void *args);
typedef size_t (*acc_get_packet_body_size_cb_t)(void *head);

/* 帧听协议 */
typedef struct
{
    acc_callback_t callback;        /* 处理回调 */
    size_t per_packet_head_size;    /* 每个包的报头长度 */
    acc_get_packet_body_size_cb_t get_packet_body_size; /* 每个包的报体长度 */
    size_t per_session_data_size;   /* 每个会话的自定义数据大小 */
    void *args;                     /* 附加参数 */
} acc_protocol_t;

/* 代理对象 */
typedef struct _acc_cntx_t {
    acc_conf_t *conf;               /* 配置信息 */
    log_cycle_t *log;               /* 日志对象 */
    int cmd_sck_id;                 /* 命令套接字 */
    avl_tree_t *reg;                /* 函数注册表 */
    acc_protocol_t *protocol;       /* 处理协议 */

    /* 侦听信息 */
    struct {
        int lsn_sck_id;             /* 侦听套接字 */
        spinlock_t accept_lock;     /* 侦听锁 */
        uint64_t cid;               /* Session ID */
        acc_lsvr_t *lsvr;           /* 侦听对象 */
    } listen;

    thread_pool_t *rsvr_pool;       /* 接收线程池 */
    thread_pool_t *lsvr_pool;       /* 帧听线程池 */

    hash_tab_t *conn_cid_tab;       /* CID集合(注:数组长度与Agent相等) */

    queue_t **connq;                /* 连接队列(注:数组长度与Agent相等) */
    ring_t **recvq;                 /* 接收队列(注:数组长度与Agent相等) */
    ring_t **sendq;                 /* 发送队列(注:数组长度与Agent相等) */
} acc_cntx_t;

#define ACC_GET_NODE_ID(ctx) ((ctx)->conf->nid)

/* 内部接口 */
int acc_lsvr_init(acc_cntx_t *ctx, acc_lsvr_t *lsn, int idx);

int acc_conn_cid_tab_add(acc_cntx_t *ctx, socket_t *sck);
socket_t *acc_conn_cid_tab_del(acc_cntx_t *ctx, uint64_t cid);
int acc_get_rid_by_cid(acc_cntx_t *ctx, uint64_t cid);

/* 外部接口 */
acc_cntx_t *acc_init(acc_protocol_t *protocol, acc_conf_t *conf, log_cycle_t *log);
int acc_launch(acc_cntx_t *ctx);
void acc_destroy(acc_cntx_t *ctx);

int acc_async_send(acc_cntx_t *ctx, int type, uint64_t cid, void *data, int len);
uint64_t acc_sck_get_cid(socket_t *sck);

#endif /*__ACCESS_H__*/

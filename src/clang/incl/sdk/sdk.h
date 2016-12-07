#if !defined(__SDK_H__)
#define __SDK_H__

#include "log.h"
#include "lock.h"
#include "rb_tree.h"
#include "spinlock.h"
#include "sdk_ssvr.h"

#define SDK_DEV_ID_LEN      (64)        /* 设备ID长度 */
#define SDK_CLIENT_ID_LEN   (64)        /* 客户端ID长度 */
#define SDK_APP_KEY_LEN     (128)       /* 应用KEY长度 */

/* 发送状态 */
typedef enum
{
    SDK_STAT_UNKNOWN       = 0          /* 未知状态 */
    , SDK_STAT_IN_SENDQ     = 1         /* 发送队列中... */
    , SDK_STAT_SENDING      = 2         /* 正在发送... */
    , SDK_STAT_SEND_SUCC    = 3         /* 发送成功 */
    , SDK_STAT_SEND_FAIL    = 4         /* 发送失败 */
    , SDK_STAT_SEND_TIMEOUT = 5         /* 发送超时 */
    , SDK_STAT_ACK_SUCC     = 6         /* 应答成功 */
    , SDK_STAT_ACK_TIMEOUT  = 7         /* 应答超时 */
} sdk_send_stat_e;

/* 发送结果回调
 *  data: 被发送的数据
 *  len: 数据长度
 *  reason: 回调原因(0:发送成功 -1:发送失败 -2:超时未发送 -3:发送后超时未应答)
 * 作用: 发送成功还是失败都会调用此回调 */
typedef int (*sdk_send_cb_t)(uint32_t cmd, const void *orig, size_t size,
        char *ack, size_t ack_len, sdk_send_stat_e stat, void *param);

/* 发送单元 */
typedef struct
{
    uint64_t serial;                    /* 序列号 */
    sdk_send_stat_e stat;               /* 处理状态 */

    uint32_t cmd;                       /* 命令类型 */
    int len;                            /* 报体长度 */
    time_t ttl;                         /* 超时时间 */
    void *data;                         /* 发送数据 */
    sdk_send_cb_t cb;                   /* 发送回调 */
    void *param;                        /* 回调参数 */
} sdk_send_item_t;

/* 发送管理表 */
typedef struct
{
    time_t next_trav_tm;                /* 下一次遍历时间 */
    pthread_rwlock_t lock;              /* 读写锁 */
    uint64_t serial;                    /* 序列号 */
    rbt_tree_t *tab;                    /* 管理表 */
} sdk_send_mgr_t;

/* 配置信息 */
typedef struct
{
    int nid;                            /* 设备ID: 唯一值 */
    char path[FILE_LINE_MAX_LEN];       /* 工作路径 */

    struct {
        uint64_t uid;                   /* 用户ID */
        uint64_t sid;             /* 会话ID(备选) */
        int terminal;                   /* 终端类型 */
        char app[SDK_APP_KEY_LEN];      /* 应用名称 */
        char version[SDK_CLIENT_ID_LEN];    /* 客户端自身版本号(留做统计用) */
    };

    int log_level;                      /* 日志级别 */
    char log_path[FILE_LINE_MAX_LEN];   /* 日志路径(路径+文件名) */

    char httpsvr[IP_ADDR_MAX_LEN];      /* HTTP端(IP+端口/域名) */

    int work_thd_num;                   /* 工作线程数 */

    size_t recv_buff_size;              /* 接收缓存大小 */

    int sendq_len;                      /* 发送队列配置 */
    int recvq_len;                      /* 接收队列配置 */
} sdk_conf_t;

/* 全局信息 */
typedef struct
{
    uint64_t sid;                       /* 会话SID */
    sdk_conf_t conf;                    /* 配置信息 */
    log_cycle_t *log;                   /* 日志对象 */

    int cmd_sck_id;                     /* 命令套接字 */
    spinlock_t cmd_sck_lck;             /* 命令套接字锁 */

    sdk_ssvr_t *ssvr;                   /* 发送服务 */
    thread_pool_t *sendtp;              /* 发送线程池 */
    thread_pool_t *worktp;              /* 工作线程池 */

    avl_tree_t *cmd;                    /* "请求<->应答"对应表 */
    avl_tree_t *reg;                    /* 回调注册对象(注: 存储sdk_reg_t数据) */
    sdk_send_mgr_t mgr;                 /* 发送管理表 */

    sdk_queue_t recvq;                  /* 接收队列 */
    sdk_queue_t sendq;                  /* 发送队列 */
} sdk_cntx_t;

/* 内部接口 */
int sdk_creat_workers(sdk_cntx_t *ctx);
int sdk_creat_sends(sdk_cntx_t *ctx);
int sdk_creat_cmd_usck(sdk_cntx_t *ctx);
int sdk_cli_cmd_send_req(sdk_cntx_t *ctx);
int sdk_lock_server(const sdk_conf_t *conf);

int sdk_ssvr_init(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr);
void *sdk_ssvr_routine(void *_ctx);

int sdk_worker_init(sdk_cntx_t *ctx, sdk_worker_t *worker, int tidx);
void *sdk_worker_routine(void *_ctx);

sdk_worker_t *sdk_worker_get_by_idx(sdk_cntx_t *ctx, int idx);

int sdk_mesg_send_ping_req(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr);
int sdk_mesg_send_online_req(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr);
int sdk_mesg_send_sync_req(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr, sdk_sck_t *sck);

int sdk_mesg_pong_handler(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr, sdk_sck_t *sck);
int sdk_mesg_ping_handler(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr, sdk_sck_t *sck);
int sdk_mesg_online_ack_handler(sdk_cntx_t *ctx, sdk_ssvr_t *ssvr, sdk_sck_t *sck, void *addr);

uint64_t sdk_gen_serial(sdk_cntx_t *ctx);
int sdk_send_mgr_init(sdk_cntx_t *ctx);
int sdk_send_mgr_insert(sdk_cntx_t *ctx, sdk_send_item_t *item, lock_e lock);
int sdk_send_mgr_delete(sdk_cntx_t *ctx, uint64_t serial);
bool sdk_send_mgr_empty(sdk_cntx_t *ctx);
sdk_send_item_t *sdk_send_mgr_query(sdk_cntx_t *ctx, uint64_t serial, lock_e lock);
int sdk_send_mgr_unlock(sdk_cntx_t *ctx, lock_e lock);
int sdk_send_mgr_trav(sdk_cntx_t *ctx);

int sdk_send_succ_hdl(sdk_cntx_t *ctx, void *addr, size_t len);
int sdk_send_fail_hdl(sdk_cntx_t *ctx, void *addr, size_t len);
bool sdk_send_data_is_timeout_and_hdl(sdk_cntx_t *ctx, void *addr);
bool sdk_ack_succ_hdl(sdk_cntx_t *ctx, uint64_t serial, void *ack);

int sdk_queue_init(sdk_queue_t *q);
int sdk_queue_length(sdk_queue_t *q);
int sdk_queue_rpush(sdk_queue_t *q, void *addr);
void *sdk_queue_lpop(sdk_queue_t *q);
bool sdk_queue_empty(sdk_queue_t *q);

/* 对外接口 */
sdk_cntx_t *sdk_init(const sdk_conf_t *conf);
int sdk_cmd_add(sdk_cntx_t *ctx, uint32_t cmd, uint32_t ack);
int sdk_register(sdk_cntx_t *ctx, uint32_t cmd, sdk_reg_cb_t proc, void *args);
int sdk_launch(sdk_cntx_t *ctx);
uint32_t sdk_async_send(sdk_cntx_t *ctx, uint32_t cmd, const void *data, size_t size, int timeout, sdk_send_cb_t cb, void *param);
int sdk_network_switch(sdk_cntx_t *ctx, int status);

#endif /*__SDK_H__*/

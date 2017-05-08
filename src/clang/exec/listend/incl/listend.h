#if !defined(__LISTEND_H__)
#define __LISTEND_H__

#include "log.h"
#include "comm.h"
#include "timer.h"
#include "chat.h"
#include "access.h"
#include "listend.h"
#include "rb_tree.h"
#include "avl_tree.h"
#include "lsnd_conf.h"
#include "thread_pool.h"

#define LSND_DEF_CONF_PATH      "../conf/listend.xml"     /* 默认配置路径 */
#define LSND_CONN_HASH_TAB_LEN  (999)       /* 哈希表长度 */
#define LSND_KICK_TTL           (5)

#define CHAT_APP_NAME_LEN       (64)        /* APP名长度 */
#define CHAT_APP_VERS_LEN       (32)        /* APP版本长度 */
#define CHAT_JSON_STR_LEN       (1024)      /* JSON字符长度 */

/* 错误码 */
typedef enum
{
    LSND_OK = 0                             /* 正常 */
    , LSND_SHOW_HELP                        /* 显示帮助信息 */

    , LSND_ERR = ~0x7FFFFFFF                /* 失败、错误 */
} lsnd_err_code_e;

/* 网络类型 */
typedef enum
{
    LSND_TYPE_UNKNOWN                        /* 未知网络 */
    , LSND_TYPE_TCP                          /* 网络类型: TCP */
    , LSND_TYPE_WS                           /* 网络类型: WS */
} lsnd_type_e;

/* 输入参数 */
typedef struct
{
    int log_level;                          /* 日志级别 */
    bool isdaemon;                          /* 是否后台运行 */
    char *conf_path;                        /* 配置路径 */
} lsnd_opt_t;

typedef struct _lsnd_cntx_t lsnd_cntx_t;

typedef enum
{
    CHAT_TERM_TYPE_UNKNOWN          /* 未知类型 */
    , CHAT_TERM_TYPE_PC             /* PC版 */
    , CHAT_TERM_TYPE_WEB            /* 网页版 */
    , CHAT_TERM_TYPE_IPHONE         /* IPHONE版 */
    , CHAT_TERM_TYPE_IPAD           /* IPAD版 */
    , CHAT_TERM_TYPE_ANDROID        /* ANDROID版 */
    , CHAT_TERM_TYPE_TOTAL
} lsnd_terminal_type_e;

/* 连接状态 */
typedef enum
{
    CHAT_CONN_STAT_UNKNOWN          /* 未知 */
    , CHAT_CONN_STAT_ESTABLISH      /* 创建 */
    , CHAT_CONN_STAT_ONLINE         /* 上线 */
    , CHAT_CONN_STAT_KICK           /* 被踢 */
    , CHAT_CONN_STAT_OFFLINE        /* 下线 */
    , CHAT_CONN_STAT_CLOSEING       /* 正在关闭... */
    , CHAT_CONN_STAT_CLOSED         /* 已关闭 */
} lsnd_conn_stat_e;

/* 会话扩展数据 */
typedef struct
{
    socket_t *sck;                  /* 所属TCP连接 */
    lsnd_cntx_t *ctx;               /* 全局上下文 */
    pthread_rwlock_t lock;          /* 读写锁 */

    uint64_t sid;                   /* 会话ID */
    uint64_t seq;                   /* 消息序列号 */
    uint64_t cid;                   /* 连接ID */
    uint64_t uid;                   /* 用户ID */
    lsnd_conn_stat_e stat;          /* 连接状态 */

    time_t create_time;             /* 创建时间 */
    time_t recv_time;               /* 最近接收数据时间 */
    time_t send_time;               /* 最近发送数据时间 */
    time_t keepalive_time;          /* 保活时间 */
    time_t kick_ttl;                /* 踢除TTL */

    char app_name[CHAT_APP_NAME_LEN]; /* 应用名 */
    char app_vers[CHAT_APP_VERS_LEN]; /* 应用版本 */
    lsnd_terminal_type_e terminal;  /* 终端类型 */
} lsnd_conn_extra_t;

/* 注册回调 */
typedef int (*lsnd_reg_cb_t)(lsnd_conn_extra_t *conn, unsigned int type, void *data, size_t len, void *args);

/* 注册项 */
typedef struct
{
    int type;
    lsnd_reg_cb_t proc;
    void *args;
} lsnd_reg_t;

/* 用户ID信息 */
typedef struct
{
    uint64_t uid;                   /* 用户ID */
    rbt_tree_t *sid_list;           /* 该用户相关的SID列表(以SID为主键) */
} lsnd_uid_item_t;

/* 被踢连接 */
typedef struct
{
    uint64_t cid;                   /* 连接ID */
    time_t ttl;                     /* 被踢时间 */
} lsnd_kick_item_t;

/* 全局对象 */
typedef struct _lsnd_cntx_t
{
    lsnd_conf_t conf;               /* 配置信息 */
    log_cycle_t *log;               /* 日志对象 */
    avl_tree_t *reg;                /* 注册回调 */

    acc_cntx_t *access;             /* 帧听层模块 */
    rtmq_proxy_t *frwder;           /* FRWDER服务 */

    timer_cntx_t *timer;            /* 定时任务表 */
    chat_tab_t *chat_tab;           /* 聊天室组织表 */
    hash_tab_t *uid_sid_tab;        /* 用户ID管理表(以UID为主键, 数据:lsnd_uid_item_t) */

    /* 定时任务 */
    timer_task_t timer_kick;        /* 定时清理被踢列表 */
    timer_task_t timer_info;        /* 定时上报帧听层信息 */
    /* 注意: 以下三个表互斥, 共同个管理类为lsnd_conn_extra_t的数据  */
    hash_tab_t *conn_list;          /* 连接管理表(以(SID+CID)为主键, 数据:lsnd_conn_extra_t) */
    hash_tab_t *kick_list;          /* 被踢管理表(以SCK为主键, 数据:lsnd_kick_item_t) */

    thread_pool_t *timer_task_tp;   /* 定时任务线程池 */
} lsnd_cntx_t;

int lsnd_getopt(int argc, char **argv, lsnd_opt_t *opt);
int lsnd_usage(const char *exec);
int lsnd_acc_reg_add(lsnd_cntx_t *ctx, int type, lsnd_reg_cb_t proc, void *args);
uint64_t lsnd_gen_cid(lsnd_cntx_t *ctx);

int lsnd_kick_insert(lsnd_cntx_t *ctx, lsnd_conn_extra_t *conn);

void lsnd_timer_kick_handler(void *_ctx);
void lsnd_timer_info_handler(void *_ctx);

void lsnd_conn_update_seq(lsnd_conn_extra_t *extra, uint64_t seq);

#endif /*__LISTEND_H__*/

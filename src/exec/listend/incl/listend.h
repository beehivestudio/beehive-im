#if !defined(__LISTEND_H__)
#define __LISTEND_H__

#include "log.h"
#include "comm.h"
#include "chat.h"
#include "access.h"
#include "listend.h"
#include "rb_tree.h"
#include "avl_tree.h"
#include "lsnd_conf.h"

#define LSND_DEF_CONF_PATH      "../conf/listend.xml"     /* 默认配置路径 */

/* 错误码 */
typedef enum
{
    LSND_OK = 0                             /* 正常 */
    , LSND_SHOW_HELP                        /* 显示帮助信息 */

    , LSND_ERR = ~0x7FFFFFFF                /* 失败、错误 */
} lsnd_err_code_e;

/* 输入参数 */
typedef struct
{
    int log_level;                          /* 日志级别 */
    bool isdaemon;                          /* 是否后台运行 */
    char *conf_path;                        /* 配置路径 */
} lsnd_opt_t;

/* 注册回调 */
typedef int (*lsnd_reg_cb_t)(unsigned int type, void *data, size_t len, void *args);

/* 注册项 */
typedef struct
{
    int type;
    lsnd_reg_cb_t proc;
    void *args;
} lsnd_reg_t;

/* 全局对象 */
typedef struct
{
    uint64_t cid;                           /* CID序列(递增) */
    lsnd_conf_t conf;                       /* 配置信息 */
    log_cycle_t *log;                       /* 日志对象 */
    avl_tree_t *reg;                        /* 注册回调 */

    acc_cntx_t *access;                     /* 帧听层模块 */
    rtmq_proxy_t *frwder;                   /* FRWDER服务 */

    chat_tab_t *chat_tab;                   /* 聊天组织表 */
    hash_tab_t *conn_sid_tab;               /* 连接管理表(以SID为主键) */
    hash_tab_t *conn_cid_tab;               /* 连接管理表(以CID为主键) */
} lsnd_cntx_t;

int lsnd_getopt(int argc, char **argv, lsnd_opt_t *opt);
int lsnd_usage(const char *exec);
int lsnd_acc_reg_add(lsnd_cntx_t *ctx, int type, lsnd_reg_cb_t proc, void *args);
uint64_t lsnd_gen_cid(lsnd_cntx_t *ctx);

#endif /*__LISTEND_H__*/

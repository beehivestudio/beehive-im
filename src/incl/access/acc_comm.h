#if !defined(__ACC_COMM_H__)
#define __ACC_COMM_H__

#include "comm.h"
#include "list.h"
#include "mem_pool.h"

/* 错误码 */
typedef enum
{
    ACC_OK = 0                             /* 正常 */
    , ACC_SHOW_HELP                        /* 显示帮助信息 */
    , ACC_DONE                             /* 完成 */
    , ACC_SCK_AGAIN                        /* 出现EAGAIN提示 */
    , ACC_SCK_CLOSE                        /* 套接字关闭 */

    , ACC_ERR = ~0x7FFFFFFF                /* 失败、错误 */
} acc_errno_e;

/* 注册回调类型 */
typedef int (*acc_reg_cb_t)(unsigned int type, char *buff, size_t len, void *args);

/* 注册对象 */
typedef struct
{
    unsigned int type:8;                    /* 数据类型 范围:(0 ~ ACC_MSG_TYPE_MAX) */
#define ACC_REG_FLAG_UNREG     (0)          /* 0: 未注册 */
#define ACC_REG_FLAG_REGED     (1)          /* 1: 已注册 */
    unsigned int flag:8;                    /* 注册标志 范围:(0: 未注册 1: 已注册) */
    acc_reg_cb_t proc;                      /* 对应数据类型的处理函数 */
    void *args;                             /* 附加参数 */
} acc_reg_t;

/* 新增套接字对象 */
typedef struct
{
    int fd;                                 /* 套接字 */
    struct timeb crtm;                      /* 创建时间 */
    uint64_t sid;                           /* Session ID */
} acc_add_sck_t;

/* 超时连接链表 */
typedef struct
{
    time_t ctm;                             /* 当前时间 */
    list_t *list;                           /* 超时链表 */
} acc_conn_timeout_list_t;

#endif /*__ACC_COMM_H__*/

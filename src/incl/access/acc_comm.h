#if !defined(__ACC_COMM_H__)
#define __ACC_COMM_H__

#include "comm.h"
#include "list.h"
#include "mem_pool.h"

/* 错误码 */
typedef enum
{
    ACC_OK = 0                  /* 正常 */
    , ACC_SHOW_HELP             /* 显示帮助信息 */
    , ACC_DONE                  /* 完成 */
    , ACC_SCK_AGAIN             /* 出现EAGAIN提示 */
    , ACC_SCK_CLOSE             /* 套接字关闭 */

    , ACC_ERR = ~0x7FFFFFFF     /* 失败、错误 */
} acc_errno_e;

/* 处罚回调的原因 */
typedef enum
{
    ACC_CALLBACK_SCK_CREAT      /* 创建SCK对象 */
    , ACC_CALLBACK_SCK_CLOSED   /* 连接关闭 */
    , ACC_CALLBACK_SCK_DESTROY  /* 销毁SCK对象 */

    , ACC_CALLBACK_RECEIVE      /* 收到数据 */
    , ACC_CALLBACK_WRITEABLE    /* 可写事件 */

    , ACC_CALLBACK_ADD_POLL_FD  /* 将套接字加入事件监听 */
    , ACC_CALLBACK_DEL_POLL_FD  /* 将套接字移除事件监听 */
    , ACC_CALLBACK_CHANGE_MODE_POLL_FD  /* 改变监听的事件 */
    , ACC_CALLBACK_LOCK_POLL    /* 对POLL加锁 */
    , ACC_CALLBACK_UNLOCK_POLL  /* 对POLL解锁 */
} acc_callback_reason_e;

/* 新增套接字对象 */
typedef struct
{
    int fd;                 /* 套接字 */
    struct timeb crtm;      /* 创建时间 */
    uint64_t cid;           /* Connection ID */
} acc_add_sck_t;

/* 超时连接链表 */
typedef struct
{
    time_t ctm;             /* 当前时间 */
    list_t *list;           /* 超时链表 */
} acc_conn_timeout_list_t;

#endif /*__ACC_COMM_H__*/

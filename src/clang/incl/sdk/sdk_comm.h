#if !defined(__SDK_COMM_H__)
#define __SDK_COMM_H__

#include "sck.h"
#include "comm.h"
#include "iovec.h"
#include "sdk_mesg.h"

#define SDK_RECONN_INTV        (2)     /* 连接重连间隔 */
#define SDK_KPALIVE_INTV       (250)   /* 保活间隔 */

#define SDK_WORKER_HDL_QNUM    (2)     /* 各Worker线程负责的队列数 */

/* 返回码 */
typedef enum
{
    SDK_OK = 0                         /* 成功 */
    , SDK_DONE                         /* 处理完成 */
    , SDK_RECONN                       /* 重连处理 */
    , SDK_AGAIN                        /* 未完成 */
    , SDK_SCK_DISCONN                  /* 连接断开 */

    , SDK_ERR = ~0x7FFFFFFF            /* 失败 */
    , SDK_ERR_NETWORK_DISCONN          /* 网络断开连接 */
    , SDK_ERR_CALLOC                   /* Calloc错误 */
    , SDK_ERR_QALLOC                   /* 申请Queue空间失败 */
    , SDK_ERR_QSIZE                    /* Queue的单元空间不足 */
    , SDK_ERR_QUEUE_NOT_ENOUGH         /* 队列空间不足 */
    , SDK_ERR_DATA_TYPE                /* 错误的数据类型 */
    , SDK_ERR_RECV_CMD                 /* 命令接收失败 */
    , SDK_ERR_REPEAT_REG               /* 重复注册 */
    , SDK_ERR_TOO_LONG                 /* 数据太长 */
    , SDK_ERR_UNKNOWN_CMD              /* 未知命令类型 */
} sdk_err_e;

/* 鉴权配置 */
typedef struct
{
    int nid;                            /* 结点ID */
    char usr[SDK_USR_MAX_LEN];         /* 用户名 */
    char passwd[SDK_PWD_MAX_LEN];      /* 登录密码 */
} sdk_auth_conf_t;

/* 接收/发送快照 */
typedef struct
{
    /*  |<------------       size       --------------->|
     *  | 已发送 |          未发送           | 空闲空间 |
     *  | 已处理 |          已接收           | 空闲空间 |
     *   -----------------------------------------------
     *  |XXXXXXXX|///////////////////////////|          |
     *  |XXXXXXXX|///////////////////////////|<--left-->|
     *  |XXXXXXXX|///////////////////////////|          |
     *   -----------------------------------------------
     *  ^        ^                           ^          ^
     *  |        |                           |          |
     * base     optr                        iptr       end
     */
    char *base;                         /* 起始地址 */
    char *end;                          /* 结束地址 */

    size_t size;                        /* 缓存大小 */

    char *optr;                         /* 发送偏移 */
    char *iptr;                         /* 输入偏移 */
} sdk_snap_t;

#define sdk_snap_setup(snap, _addr, _size) \
   (snap)->base = (_addr);  \
   (snap)->end = (_addr) + (_size); \
   (snap)->size = (_size); \
   (snap)->optr = (_addr);  \
   (snap)->iptr = (_addr); 

#define sdk_snap_reset(snap)           /* 重置标志 */\
   (snap)->optr = (snap)->base;  \
   (snap)->iptr = (snap)->base; 

/* 工作对象 */
typedef struct
{
    int id;                             /* 对象索引 */
    log_cycle_t *log;                   /* 日志对象 */

    int cmd_sck_id;                     /* 命令套接字 */

    int max;                            /* 套接字最大值 */
    fd_set rdset;                       /* 可读套接字集合 */

    uint64_t proc_total;                /* 已处理条数 */
    uint64_t drop_total;                /* 丢弃条数 */
    uint64_t err_total;                 /* 错误条数 */
} sdk_worker_t;

/******************************************************************************
 **函数名称: sdk_reg_cb_t
 **功    能: 回调注册类型
 **输入参数:
 **     cmd: 消息类型
 **     from: 源ID
 **     data: 数据
 **     len: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 ******************************************************************************/
typedef int (*sdk_reg_cb_t)(int cmd, uint64_t from, char *data, size_t len, void *param);
typedef struct
{
    int cmd;                            /* 消息类型 */
    sdk_reg_cb_t proc;                  /* 回调函数指针 */
    void *param;                        /* 附加参数 */
} sdk_reg_t;

/******************************************************************************
 **函数名称: sdk_reg_cmp_cb
 **功    能: 注册比较函数回调
 **输入参数:
 **     reg1: 注册项1
 **     reg2: 注册项2
 **输出参数:
 **返    回: 0:相等 <0:小于 >0:大于
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.08.09 17:19:24 #
 ******************************************************************************/
static inline int sdk_reg_cmp_cb(const sdk_reg_t *reg1, const sdk_reg_t *reg2)
{
    return (reg1->cmd - reg2->cmd);
}

/* 命令项 */
typedef struct
{
    uint32_t ack;                           /* "应答"命令 */
    uint32_t req;                           /* 对应的"请求"命令 */
} sdk_cmd_ack_t;

/******************************************************************************
 **函数名称: sdk_ack_cmp_cb
 **功    能: 命令比较函数回调
 **输入参数:
 **     cmd1: 命令项1
 **     cmd2: 命令项2
 **输出参数:
 **返    回: 0:相等 <0:小于 >0:大于
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 16:37:13 #
 ******************************************************************************/
static inline int sdk_ack_cmp_cb(const sdk_cmd_ack_t *cmd1, const sdk_cmd_ack_t *cmd2)
{
    return (cmd1->ack - cmd2->ack);
}

#endif /*__SDK_COMM_H__*/

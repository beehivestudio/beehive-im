#if !defined(__SDK_MESG_H__)
#define __SDK_MESG_H__

#include "comm.h"
#include "mesg.h"

/* 宏定义 */
#define SDK_USR_MAX_LEN    (32)        /* 用户名 */
#define SDK_PWD_MAX_LEN    (16)        /* 登录密码 */

/* 系统数据类型 */
typedef enum
{
    SDK_CMD_UNKNOWN                    = 0x0000  /* 未知命令 */
    , SDK_CMD_PROC_REQ                 = 0x0001  /* 处理客户端数据-请求 */
    , SDK_CMD_SEND                     = 0x0002  /* 发送数据-请求 */
    , SDK_CMD_SEND_ALL                 = 0x0003  /* 发送所有数据-请求 */
    , SDK_CMD_NETWORK_CONN             = 0x0004  /* 网络已开启-请求 */
    , SDK_CMD_NETWORK_DISCONN          = 0x0005  /* 网络已断开-请求 */
} sdk_mesg_e;

/* 链路鉴权请求 */
typedef struct
{
    char usr[SDK_USR_MAX_LEN];         /* 用户名 */
    char passwd[SDK_PWD_MAX_LEN];      /* 登录密码 */
} sdk_link_auth_req_t;

/* 链路鉴权应答 */
typedef struct
{
#define SDK_LINK_AUTH_FAIL     (0)
#define SDK_LINK_AUTH_SUCC     (1)
    uint32_t is_succ;                        /* 应答码(0:失败 1:成功) */
} sdk_link_auth_rsp_t;

/* 订阅请求 */
typedef struct
{
    uint32_t type;                      /* 订阅内容: 订阅的消息类型 */
} sdk_sub_req_t;

#define SDK_SUB_REQ_HTON(n, h) do {    /* 主机 -> 网络 */\
    (n)->type = htonl((h)->type); \
} while(0)

#define SDK_SUB_REQ_NTOH(h, n) do {    /* 网络 -> 主机 */\
    (h)->type = ntohl((n)->type); \
} while(0)

/* 添加套接字请求的相关参数 */
typedef struct
{
    int sckid;                          /* 套接字 */
    uint64_t sid;                       /* Session ID */
    char ipaddr[IP_ADDR_MAX_LEN];       /* IP地址 */
} sdk_cmd_add_sck_t;

/* 处理数据请求的相关参数 */
typedef struct
{
    uint32_t num;                       /* 需要处理的数据条数 */
} sdk_cmd_proc_req_t;

/* 发送数据请求的相关参数 */
typedef struct
{
    /* No member */
} sdk_cmd_send_req_t;

/* 配置信息 */
typedef struct
{
    int nid;                            /* 结点ID: 不允许重复 */
    char path[FILE_NAME_MAX_LEN];       /* 工作路径 */
    int port;                           /* 侦听端口 */

    int work_thd_num;                   /* 工作线程数 */

    int qmax;                           /* 队列长度 */
    int qsize;                          /* 队列大小 */
} sdk_cmd_conf_t;

/* Recv状态信息 */
typedef struct
{
    uint32_t connections;               /* 总连接数 */
    uint64_t recv_total;                /* 接收数据总数 */
    uint64_t drop_total;                /* 丢弃数据总数 */
    uint64_t err_total;                 /* 异常数据总数 */
} sdk_cmd_recv_stat_t;

/* Work状态信息 */
typedef struct
{
    uint64_t proc_total;                /* 已处理数据总数 */
    uint64_t drop_total;                /* 放弃处理数据总数 */
    uint64_t err_total;                 /* 处理数据异常总数 */
} sdk_cmd_proc_stat_t;

/* 各命令所附带的数据 */
typedef union
{
    sdk_cmd_add_sck_t add_sck_req;
    sdk_cmd_proc_req_t proc_req;
    sdk_cmd_send_req_t send_req;
    sdk_cmd_proc_stat_t proc_stat;
    sdk_cmd_recv_stat_t recv_stat;
    sdk_cmd_conf_t conf;
} sdk_cmd_param_t;

/* 命令数据信息 */
typedef struct
{
    uint32_t type;                      /* 命令类型 Range: sdk_cmd_e */
    char src_path[FILE_NAME_MAX_LEN];   /* 命令源路径 */
    sdk_cmd_param_t param;              /* 其他数据信息 */
} sdk_cmd_t;

#endif /*__SDK_MESG_H__*/

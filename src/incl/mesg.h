/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: mesg.h
 ** 版本号: 1.0
 ** 描  述: 消息类型的定义
 ** 作  者: # Qifeng.zou # Fri 08 May 2015 10:43:30 PM CST #
 ******************************************************************************/
#if !defined(__MESG_H__)
#define __MESG_H__

#include "uri.h"

/* 消息类型 */
typedef enum
{
    CMD_UNKNOWN                             /* 未知消息 */

    /* 请求/应答类消息 */
    , CMD_ONLINE_REQ            = 0x0101    /* 上线请求 */
    , CMD_ONLINE_ACK            = 0x0102    /* 上线请求应答 */

    , CMD_OFFLINE_REQ           = 0x0103    /* 下线请求 */
    , CMD_OFFLINE_ACK           = 0x0104    /* 下线请求应答 */

    , CMD_JOIN_REQ              = 0x0105    /* 加入聊天室 */
    , CMD_JOIN_ACK              = 0x0106    /* 加入聊天室应答 */

    , CMD_QUIT_REQ              = 0x0107    /* 退出聊天室 */
    , CMD_QUIT_ACK              = 0x0108    /* 退出聊天室应答 */

    , CMD_CLIENT_PING           = 0x0109    /* 客户端心跳 */
    , CMD_CLIENT_PONG           = 0x010A    /* 客户端心跳应答 */

    , CMD_SUB_REQ               = 0x010B    /* 订阅请求 */
    , CMD_SUB_ACK               = 0x010C    /* 订阅应答 */

    , CMD_UNSUB_REQ             = 0x010D    /* 取消订阅 */
    , CMD_UNSUB_ACK             = 0x010E    /* 取消订阅应答 */

    , CMD_GROUP_MSG             = 0x0110    /* 群聊消息 */
    , CMD_GROUP_MSG_ACK         = 0x0111    /* 群聊消息应答 */

    , CMD_PRVT_MSG              = 0x0112    /* 私聊消息 */
    , CMD_PRVT_MSG_ACK          = 0x0113    /* 私聊消息应答 */

    , CMD_BC_MSG                = 0x0114    /* 广播消息 */
    , CMD_BC_MSG_ACK            = 0x0115    /* 广播消息应答 */

    , CMD_P2P_MSG               = 0x0116    /* 点到点消息 */
    , CMD_P2P_MSG_ACK           = 0x0117    /* 点到点消息应答 */

    , CMD_ROOM_MSG              = 0x0118    /* 聊天室消息 */
    , CMD_ROOM_MSG_ACK          = 0x0119    /* 聊天室消息应答 */

    , CMD_ROOM_BC_MSG           = 0x011A    /* 聊天室广播消息 */
    , CMD_ROOM_BC_MSG_ACK       = 0x011B    /* 聊天室广播消息应答 */

    , CMD_EXCEPT_MSG            = 0x011C    /* 通用异常消息 */
    , CMD_EXCEPT_MSG_ACK        = 0x011D    /* 通用异常消息应答 */

    , CMD_ROOM_USR_NUM          = 0x011E    /* 聊天室人数 */
    , CMD_ROOM_USR_NUM_ACK      = 0x0120    /* 聊天室人数应答 */

    , CMD_SYNC_MSG              = 0x0121    /* 同步消息 */
    , CMD_SYNC_MSG_ACK          = 0x0122    /* 同步消息应答 */

    /* 通知类消息 */
    , CMD_ONLINE_NTC            = 0x0301    /* 上线通知 */
    , CMD_OFFLINE_NTC           = 0x0302    /* 下线通知 */
    , CMD_JOIN_NTC              = 0x0303    /* 加入聊天室通知 */
    , CMD_QUIT_NTC              = 0x0304    /* 退出聊天室通知 */
    , CMD_BAN_NTC               = 0x0305    /* 禁言通知 */
    , CMD_KICK_NTC              = 0x0306    /* 踢人通知 */

    /* 系统内部消息 */
    , CMD_HB                    = 0x0401    /* 内部心跳 */
    , CMD_HB_ACK                = 0x0402    /* 内部心跳应答 */
    , CMD_LSN_RPT               = 0x0403    /* 帧听层上报 */
    , CMD_LSN_RPT_ACK           = 0x0404    /* 帧听层上报应答 */
    , CMD_FRWD_LIST             = 0x0405    /* 转发层列表 */
    , CMD_FRWD_LIST_ACK         = 0x0406    /* 转发层列表应答 */

    ////////////////////////////////////////////
    , MSG_PING                          /* PING-请求 */
    , MSG_PONG                          /* PONG-应答 */

    , MSG_SEARCH_REQ                    /* 搜索关键字-请求 */
    , MSG_SEARCH_RSP                    /* 搜索关键字-应答 */

    , MSG_INSERT_WORD_REQ               /* 插入关键字-请求 */
    , MSG_INSERT_WORD_RSP               /* 插入关键字-应答 */

    , MSG_PRINT_INVT_TAB_REQ            /* 打印倒排表-请求 */
    , MSG_PRINT_INVT_TAB_RSP            /* 打印倒排表-应答 */

    , MSG_QUERY_CONF_REQ                /* 查询配置信息-请求 */
    , MSG_QUERY_CONF_RSP                /* 反馈配置信息-应答 */

    , MSG_QUERY_WORKER_STAT_REQ         /* 查询工作信息-请求 */
    , MSG_QUERY_WORKER_STAT_RSP         /* 反馈工作信息-应答 */

    , MSG_QUERY_WORKQ_STAT_REQ          /* 查询工作队列信息-请求 */
    , MSG_QUERY_WORKQ_STAT_RSP          /* 反馈工作队列信息-应答 */

    , MSG_SWITCH_SCHED_REQ              /* 切换调度-请求 */
    , MSG_SWITCH_SCHED_RSP              /* 反馈切换调度信息-应答 */

    , MSG_SUB_REQ                       /* 订阅-请求 */
    , MSG_SUB_RSP                       /* 订阅-应答 */

    , MSG_TYPE_TOTAL                    /* 消息类型总数 */
} mesg_type_e;

/* 通用协议头 */
typedef struct
{
    uint32_t type;                      /* 消息类型 Range: mesg_type_e */
#define MSG_FLAG_SYS   (-1)             /* -1: 系统数据类型 */
#define MSG_FLAG_USR   (1)              /* 1: 自定义数据类型 */
    uint32_t flag;                      /* 标识量(0:系统数据类型 1:自定义数据类型) */
    uint32_t length;                    /* 报体长度 */
#define MSG_CHKSUM_VAL   (0x1ED23CB4)
    uint32_t chksum;                    /* 校验值 */

    uint64_t sid;                       /* 会话ID */
    uint32_t nid;                       /* 结点ID */

    uint64_t serial;                    /* 流水号(注: 全局唯一流水号) */
    char body[0];                       /* 消息体 */
} mesg_header_t;

/* 字节序转换 */
#define MESG_HEAD_HTON(h, n) do { /* 主机->网络 */\
    (n)->type = htonl((h)->type); \
    (n)->flag = htonl((h)->flag); \
    (n)->length = htonl((h)->length); \
    (n)->chksum = htonl((h)->chksum); \
    (n)->sid = hton64((h)->sid); \
    (n)->nid = htonl((h)->nid); \
    (n)->serial = hton64((h)->serial); \
} while(0)

#define MESG_HEAD_NTOH(n, h) do { /* 网络->主机*/\
    (h)->type = ntohl((n)->type); \
    (h)->flag = ntohl((n)->flag); \
    (h)->length = ntohl((n)->length); \
    (h)->chksum = ntohl((n)->chksum); \
    (h)->sid = ntoh64((n)->sid); \
    (h)->nid = ntohl((n)->nid); \
    (h)->serial = ntoh64((n)->serial); \
} while(0)

#define MESG_HEAD_SET(head, _type, _sid, _nid, _serial, _len) do { /* 设置协议头 */\
    (head)->type = (_type); \
    (head)->flag = MSG_FLAG_USR; \
    (head)->length = (_len); \
    (head)->chksum = MSG_CHKSUM_VAL; \
    (head)->sid = (_sid); \
    (head)->nid = (_nid); \
    (head)->serial = (_serial); \
} while(0)

#define MESG_TOTAL_LEN(body_len) (sizeof(mesg_header_t) + body_len)
#define MESG_CHKSUM_ISVALID(head) (MSG_CHKSUM_VAL == (head)->chksum)

#define MESG_HEAD_PRINT(log, head) \
    log_debug((log), "Call %s()! type:%d len:%d chksum:0x%X/0x%X sid:%lu nid:%u serial:%lu", \
            __func__, (head)->type, (head)->length, (head)->chksum, MSG_CHKSUM_VAL, \
            (head)->sid, (head)->nid, (head)->serial);

////////////////////////////////////////////////////////////////////////////////
/* 搜索消息结构 */
#define SRCH_WORD_LEN       (128)
typedef struct
{
    char words[SRCH_WORD_LEN];          /* 搜索关键字 */
} mesg_search_req_t;

////////////////////////////////////////////////////////////////////////////////
/* 插入关键字-请求 */
typedef struct
{
    char word[SRCH_WORD_LEN];           /* 关键字 */
    char url[URL_MAX_LEN];              /* 关键字对应的URL */
    int freq;                           /* 频率 */
} mesg_insert_word_req_t;

/* 插入关键字-应答 */
typedef struct
{
#define MESG_INSERT_WORD_FAIL   (0)
#define MESG_INSERT_WORD_SUCC   (1)
    int code;                           /* 应答码 */
    char word[SRCH_WORD_LEN];           /* 关键字 */
} mesg_insert_word_rsp_t;

#define mesg_insert_word_resp_hton(rsp) do { \
    (rsp)->code = htonl((rsp)->code); \
} while(0)

#define mesg_insert_word_resp_ntoh(rsp) do { \
    (rsp)->code = ntohl((rsp)->code); \
} while(0)

////////////////////////////////////////////////////////////////////////////////
/* 订阅-请求 */
typedef struct
{
    mesg_type_e type;                   /* 订阅类型(消息类型) */
    int nid;                            /* 结点ID */
} mesg_sub_req_t;

#define mesg_sub_req_hton(req) do { /* 主机 -> 网络 */\
    (req)->type = htonl((req)->type); \
    (req)->nid = htonl((req)->nid); \
} while(0)

#define mesg_sub_req_ntoh(req) do { /* 网络 -> 主机 */\
    (req)->type = ntohl((req)->type); \
    (req)->nid = ntohl((req)->nid); \
} while(0)

/* 订阅-应答 */
typedef struct
{
    int code;                           /* 应答码 */
    mesg_type_e type;                   /* 订阅类型(消息类型) */
} mesg_sub_rsp_t;

#define mesg_sub_rsp_hton(rsp) do { /* 主机 -> 网络 */\
    (rsp)->code = htonl((rsp)->code); \
    (rsp)->type = htonl((rsp)->type); \
} while(0)

#define mesg_sub_rsp_ntoh(rsp) do { /* 网络 -> 主机 */\
    (rsp)->code = ntohl((rsp)->code); \
    (rsp)->type = ntohl((rsp)->type); \
} while(0)

#endif /*__MESG_H__*/

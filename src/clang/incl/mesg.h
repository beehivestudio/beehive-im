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

#include "comm.h"

/* 通用协议头 */
typedef struct
{
    uint32_t type;                      /* 消息类型 */
#define MSG_FLAG_SYS   (0)              /* 0: 系统数据类型 */
#define MSG_FLAG_USR   (1)              /* 1: 自定义数据类型 */
    uint32_t flag;                      /* 标识量(0:系统数据类型 1:自定义数据类型) */
    uint32_t length;                    /* 报体长度 */
#define MSG_CHKSUM_VAL   (0x1ED23CB4)
    uint32_t chksum;                    /* 校验值 */

    uint64_t sid;                       /* 会话ID */
    uint32_t nid;                       /* 结点ID */

    uint64_t serial;                    /* 流水号(注: 全局唯一流水号) */
    char body[0];                       /* 消息体 */
}__attribute__ ((__packed__))mesg_header_t;

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
    log_debug((log), "type:0x%04X len:%d chksum:0x%X/0x%X sid:%lu nid:%u serial:%lu", \
            (head)->type, (head)->length, (head)->chksum, MSG_CHKSUM_VAL, \
            (head)->sid, (head)->nid, (head)->serial);

#endif /*__MESG_H__*/

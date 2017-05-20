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
    uint32_t length;                    /* 报体长度 */

    uint64_t sid;                       /* 会话ID */
    uint64_t cid;                       /* 连接ID */
    uint32_t nid;                       /* 结点ID */

    uint64_t seq;                       /* 消息序列号(注: 全局唯一流水号) */

    uint64_t dsid;                      /* 目的: 会话SID */
    uint64_t dseq;                      /* 目的: 消息序列号 */

    char body[0];                       /* 消息体 */
}__attribute__ ((__packed__))mesg_header_t;

/* 字节序转换 */
#define MESG_HEAD_HTON(h, n) do { /* 主机->网络 */\
    (n)->type = htonl((h)->type); \
    (n)->length = htonl((h)->length); \
    (n)->sid = hton64((h)->sid); \
    (n)->cid = hton64((h)->cid); \
    (n)->nid = htonl((h)->nid); \
    (n)->seq = hton64((h)->seq); \
    (n)->dsid = hton64((h)->dsid); \
    (n)->dseq = hton64((h)->dseq); \
} while(0)

#define MESG_HEAD_NTOH(n, h) do { /* 网络->主机*/\
    (h)->type = ntohl((n)->type); \
    (h)->length = ntohl((n)->length); \
    (h)->sid = ntoh64((n)->sid); \
    (h)->cid = ntoh64((n)->cid); \
    (h)->nid = ntohl((n)->nid); \
    (h)->seq = ntoh64((n)->seq); \
    (h)->dsid = ntoh64((n)->dsid); \
    (h)->dseq = ntoh64((n)->dseq); \
} while(0)

 /* 设置协议头 */
#define MESG_HEAD_SET(head, _type, _sid, _cid, _nid, _seq, _len, _dsid, _dseq) do { \
    (head)->type = (_type); \
    (head)->length = (_len); \
    (head)->sid = (_sid); \
    (head)->cid = (_cid); \
    (head)->nid = (_nid); \
    (head)->seq = (_seq); \
    (head)->dsid = (_dsid); \
    (head)->dseq = (_dseq); \
} while(0)

#define MESG_TOTAL_LEN(body_len) (sizeof(mesg_header_t) + body_len)

#define MESG_HEAD_PRINT(log, head) \
    log_debug((log), "type:0x%04X len:%d sid:%lu cid:%lu nid:%u seq:%lu", \
            (head)->type, (head)->length, \
            (head)->sid, (head)->cid, (head)->nid, (head)->seq);

#endif /*__MESG_H__*/

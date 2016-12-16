/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: cmd_list.h
 ** 版本号: 1.0
 ** 描  述: 消息类型的定义
 ** 作  者: # Qifeng.zou # Fri 08 May 2015 10:43:30 PM CST #
 ******************************************************************************/
#if !defined(__CMD_LIST_H__)
#define __CMD_LIST_H__

/* 消息类型 */
typedef enum
{
    CMD_UNKNOWN                             /* 未知消息 */

    /* 请求/应答类消息 */
    , CMD_ONLINE_REQ            = 0x0101    /* 上线请求(服务端) */
    , CMD_ONLINE_ACK            = 0x0102    /* 上线请求应答(客户端) */

    , CMD_OFFLINE_REQ           = 0x0103    /* 下线请求(服务端) */
    , CMD_OFFLINE_ACK           = 0x0104    /* 下线请求应答(客户端) */

    , CMD_JOIN_REQ              = 0x0105    /* 加入聊天室(服务端) */
    , CMD_JOIN_ACK              = 0x0106    /* 加入聊天室应答(客户端) */

    , CMD_UNJOIN_REQ            = 0x0107    /* 退出聊天室(服务端) */
    , CMD_UNJOIN_ACK            = 0x0108    /* 退出聊天室应答(客户端) */

    , CMD_PING                  = 0x0109    /* 客户端心跳(服务端) */
    , CMD_PONG                  = 0x010A    /* 客户端心跳应答(客户端) */

    , CMD_SUB_REQ               = 0x010B    /* 订阅请求(服务端) */
    , CMD_SUB_ACK               = 0x010C    /* 订阅应答(客户端) */

    , CMD_UNSUB_REQ             = 0x010D    /* 取消订阅(服务端) */
    , CMD_UNSUB_ACK             = 0x010E    /* 取消订阅应答(客户端) */

    , CMD_GROUP_MSG             = 0x0110    /* 群聊消息(服务端&客户端) */
    , CMD_GROUP_MSG_ACK         = 0x0111    /* 群聊消息应答(服务端&客户端) */

    , CMD_PRVT_MSG              = 0x0112    /* 私聊消息(服务端&客户端) */
    , CMD_PRVT_MSG_ACK          = 0x0113    /* 私聊消息应答(客户端&服务端) */

    , CMD_BC_MSG                = 0x0114    /* 广播消息(服务端&客户端) */
    , CMD_BC_MSG_ACK            = 0x0115    /* 广播消息应答(服务端) */

    , CMD_P2P_MSG               = 0x0116    /* 点到点消息(暂时不需要) */
    , CMD_P2P_MSG_ACK           = 0x0117    /* 点到点消息应答(客户端&服务端) */

    , CMD_ROOM_MSG              = 0x0118    /* 聊天室消息(服务端&客户端) */
    , CMD_ROOM_MSG_ACK          = 0x0119    /* 聊天室消息应答(客户端&服务端) */

    , CMD_ROOM_BC_MSG           = 0x011A    /* 聊天室广播消息(客户端) */
    , CMD_ROOM_BC_MSG_ACK       = 0x011B    /* 聊天室广播消息应答(服务端) */

    , CMD_EXCEPT_MSG            = 0x011C    /* 通用异常消息 */
    , CMD_EXCEPT_MSG_ACK        = 0x011D    /* 通用异常消息应答 */

    , CMD_ROOM_USR_NUM          = 0x011E    /* 聊天室人数(客户端) */
    , CMD_ROOM_USR_NUM_ACK      = 0x0120    /* 聊天室人数应答(服务端) */

    , CMD_SYNC_MSG              = 0x0121    /* 同步消息 */
    , CMD_SYNC_MSG_ACK          = 0x0122    /* 同步消息应答(客户端) */

    /* 通知类消息 */
    , CMD_ONLINE_NTC            = 0x0301    /* 上线通知 */
    , CMD_OFFLINE_NTC           = 0x0302    /* 下线通知 */
    , CMD_JOIN_NTC              = 0x0303    /* 加入聊天室通知 */
    , CMD_QUIT_NTC              = 0x0304    /* 退出聊天室通知 */
    , CMD_BAN_ADD_NTC           = 0x0305    /* 禁言通知 */
    , CMD_BAN_DEL_NTC           = 0x0306    /* 移除禁言通知 */
    , CMD_BLACKLIST_ADD_NTC     = 0x0307    /* 加入黑名单通知 */
    , CMD_BLACKLIST_DEL_NTC     = 0x0308    /* 移除黑名单通知 */

    /* 系统内部消息 */
    , CMD_LSN_RPT               = 0x0401    /* 帧听层上报 */
    , CMD_LSN_RPT_ACK           = 0x0402    /* 帧听层上报应答 */
    , CMD_FRWD_RPT              = 0x0403    /* 转发层上报 */
    , CMD_FRWD_RPT_ACK          = 0x0404    /* 转发层上报应答 */
    , CMD_KICK_REQ              = 0x0405    /* 踢人请求 */
    , CMD_KICK_ACK              = 0x0406    /* 踢人应答 */
} mesg_type_e;

#endif /*__CMD_LIST_H__*/

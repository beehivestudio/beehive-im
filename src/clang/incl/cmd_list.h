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

    /* 通用消息 */
    , CMD_ONLINE                = 0x0101    /* 上线请求(服务端) */
    , CMD_ONLINE_ACK            = 0x0102    /* 上线请求应答(客户端) */

    , CMD_OFFLINE               = 0x0103    /* 下线请求(服务端) */
    , CMD_OFFLINE_ACK           = 0x0104    /* 下线请求应答(客户端) */

    , CMD_PING                  = 0x0105    /* 客户端心跳(服务端) */
    , CMD_PONG                  = 0x0106    /* 客户端心跳应答(客户端) */

    , CMD_SUB                   = 0x0107    /* 订阅请求(服务端) */
    , CMD_SUB_ACK               = 0x0108    /* 订阅应答(客户端) */

    , CMD_UNSUB                 = 0x0109    /* 取消订阅(服务端) */
    , CMD_UNSUB_ACK             = 0x010A    /* 取消订阅应答(客户端) */

    , CMD_EXCEPT_MSG            = 0x010B    /* 通用异常消息 */
    , CMD_EXCEPT_MSG_ACK        = 0x010C    /* 通用异常消息应答 */

    , CMD_SYNC                  = 0x010D    /* 同步消息 */
    , CMD_SYNC_ACK              = 0x010E    /* 同步消息应答(客户端) */

    //, CMD_ALLOC_SEQ             = 0x0110    /* 申请序列号 */
    //, CMD_ALLOC_SEQ_ACK         = 0x0111    /* 申请序列号应答 */

    , CMD_KICK_REQ              = 0x0112    /* 踢人请求 */
    , CMD_KICK_ACK              = 0x0113    /* 踢人应答 */

    , CMD_ONLINE_NTC            = 0x0151    /* 上线通知 */
    , CMD_OFFLINE_NTC           = 0x0152    /* 下线通知 */

    /* 私聊消息 */
    , CMD_CHAT                  = 0x0201    /* 私聊消息 */
    , CMD_CHAT_ACK              = 0x0202    /* 私聊消息应答 */
    , CMD_FRIEND_ADD            = 0x0203    /* 添加好友 */
    , CMD_FRIEND_ADD_ACK        = 0x0204    /* 添加好友应答 */
    , CMD_FRIEND_DEL            = 0x0205    /* 移除好友 */
    , CMD_FRIEND_DEL_ACK        = 0x0206    /* 移除好友应答 */
    , CMD_BLACKLIST_ADD         = 0x0207    /* 加入黑名单 */
    , CMD_BLACKLIST_ADD_ACK     = 0x0208    /* 加入黑名单应答 */
    , CMD_BLACKLIST_DEL         = 0x0209    /* 移除黑名单 */
    , CMD_BLACKLIST_DEL_ACK     = 0x020A    /* 移除黑名单应答 */
    , CMD_GAG_ADD               = 0x020B    /* 设置禁言 */
    , CMD_GAG_ADD_ACK           = 0x020C    /* 设置禁言应答 */
    , CMD_GAG_DEL               = 0x020D    /* 移除禁言 */
    , CMD_GAG_DEL_ACK           = 0x020E    /* 移除禁言应答 */
    , CMD_MARK_ADD              = 0x0210    /* 设置备注 */
    , CMD_MARK_ADD_ACK          = 0x0211    /* 设置备注应答 */
    , CMD_MARK_DEL              = 0x0212    /* 移除备注 */
    , CMD_MARK_DEL_ACK          = 0x0213    /* 移除备注应答 */

    /* 群聊消息 */
    , CMD_GROUP_CREAT           = 0x0301    /* 创建群组 */
    , CMD_GROUP_CREAT_ACK       = 0x0302    /* 创建群组应答 */

    , CMD_GROUP_DISMISS         = 0x0303    /* 解散群组 */
    , CMD_GROUP_DISMISS_ACK     = 0x0304    /* 解散群组应答 */

    , CMD_GROUP_JOIN            = 0x0305    /* 加入群组 */
    , CMD_GROUP_JOIN_ACK        = 0x0306    /* 加入群组应答 */

    , CMD_GROUP_QUIT            = 0x0307    /* 退群 */
    , CMD_GROUP_QUIT_ACK        = 0x0308    /* 退群应答 */

    , CMD_GROUP_INVITE          = 0x0309    /* 邀请入群 */
    , CMD_GROUP_INVITE_ACK      = 0x030A    /* 邀请入群应答 */

    , CMD_GROUP_CHAT            = 0x030B    /* 群聊消息 */
    , CMD_GROUP_CHAT_ACK        = 0x030C    /* 群聊消息应答 */

    , CMD_GROUP_KICK            = 0x030D    /* 群组踢人 */
    , CMD_GROUP_KICK_ACK        = 0x030E    /* 群组踢人应答 */

    , CMD_GROUP_GAG_ADD         = 0x0310    /* 群组禁言 */
    , CMD_GROUP_GAG_ADD_ACK     = 0x0311    /* 群组禁言应答 */

    , CMD_GROUP_GAG_DEL         = 0x0312    /* 解除群组禁言 */
    , CMD_GROUP_GAG_DEL_ACK     = 0x0313    /* 解除群组禁言应答 */

    , CMD_GROUP_BL_ADD          = 0x0314    /* 加入群组黑名单 */
    , CMD_GROUP_BL_ADD_ACK      = 0x0315    /* 加入群组黑名单应答 */

    , CMD_GROUP_BL_DEL          = 0x0316    /* 解除群组黑名单 */
    , CMD_GROUP_BL_DEL_ACK      = 0x0317    /* 解除群组黑名单应答 */

    , CMD_GROUP_MGR_ADD         = 0x0318    /* 添加群组管理员 */
    , CMD_GROUP_MGR_ADD_ACK     = 0x0319    /* 添加群组管理员应答 */

    , CMD_GROUP_MGR_DEL         = 0x031A    /* 解除群组管理员 */
    , CMD_GROUP_MGR_DEL_ACK     = 0x031B    /* 解除群组管理员应答 */

    , CMD_GROUP_USR_LIST         = 0x031C    /* 群组成员列表 */
    , CMD_GROUP_USR_LIST_ACK     = 0x031D    /* 群组成员列表应答 */

    , CMD_GROUP_JOIN_NTC        = 0x0350    /* 入群通知 */
    , CMD_GROUP_QUIT_NTC        = 0x0351    /* 退群通知 */
    , CMD_GROUP_KICK_NTC        = 0x0352    /* 踢人通知 */
    , CMD_GROUP_GAG_ADD_NTC     = 0x0353    /* 禁言通知 */
    , CMD_GROUP_GAG_DEL_NTC     = 0x0354    /* 解除禁言通知 */
    , CMD_GROUP_BL_ADD_NTC      = 0x0355    /* 添加群组黑名单通知 */
    , CMD_GROUP_BL_DEL_NTC      = 0x0356    /* 解除群组黑名单通知 */
    , CMD_GROUP_MGR_ADD_NTC     = 0x0357    /* 添加群组管理员通知 */
    , CMD_GROUP_MGR_DEL_NTC     = 0x0358    /* 解除群组管理员通知 */

    /* 聊天室消息 */
    , CMD_ROOM_CREAT            = 0x0401    /* 创建聊天室 */
    , CMD_ROOM_CREAT_ACK        = 0x0402    /* 创建聊天室应答 */

    , CMD_ROOM_DISMISS          = 0x0403    /* 解散聊天室 */
    , CMD_ROOM_DISMISS_ACK      = 0x0404    /* 解散聊天室应答 */

    , CMD_ROOM_JOIN             = 0x0405    /* 加入聊天室 */
    , CMD_ROOM_JOIN_ACK         = 0x0406    /* 加入聊天室应答 */

    , CMD_ROOM_QUIT             = 0x0407    /* 退出聊天室 */
    , CMD_ROOM_QUIT_ACK         = 0x0408    /* 退出聊天室应答 */

    , CMD_ROOM_KICK             = 0x0409    /* 踢出聊天室 */
    , CMD_ROOM_KICK_ACK         = 0x040A    /* 踢出聊天室应答 */

    , CMD_ROOM_CHAT             = 0x040B    /* 聊天室消息 */
    , CMD_ROOM_CHAT_ACK         = 0x040C    /* 聊天室消息应答 */

    , CMD_ROOM_BC               = 0x040D    /* 聊天室广播消息 */
    , CMD_ROOM_BC_ACK           = 0x040E    /* 聊天室广播消息应答 */

    , CMD_ROOM_USR_NUM          = 0x0410    /* 聊天室人数 */
    , CMD_ROOM_USR_NUM_ACK      = 0x0411    /* 聊天室人数应答 */

    , CMD_ROOM_JOIN_NTC         = 0x0450    /* 加入聊天室通知 */
    , CMD_ROOM_QUIT_NTC         = 0x0451    /* 退出聊天室通知 */
    , CMD_ROOM_KICK_NTC         = 0x0452    /* 踢出聊天室通知 */

    /* 推送消息 */
    , CMD_BC                    = 0x0501    /* 广播消息 */
    , CMD_BC_ACK                = 0x0502    /* 广播消息应答 */

    , CMD_P2P                   = 0x0503    /* 点到点消息(暂时不需要) */
    , CMD_P2P_ACK               = 0x0504    /* 点到点消息应答(客户端&服务端) */

    /* 系统内部消息 */
    , CMD_LSND_INFO             = 0x0601    /* 帧听层信息上报 */
    , CMD_LSND_INFO_ACK         = 0x0602    /* 帧听层信息上报应答 */

    , CMD_FRWD_INFO             = 0x0603    /* 转发层信息上报 */
    , CMD_FRWD_INFO_ACK         = 0x0604    /* 转发层信息上报应答 */
} mesg_type_e;

#endif /*__CMD_LIST_H__*/

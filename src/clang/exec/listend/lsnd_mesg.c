/******************************************************************************
 ** Coypright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: lsnd_mesg.c
 ** 版本号: 1.0
 ** 描  述: 侦听相关的消息处理函数的定义
 ** 作  者: # Qifeng.zou # Thu 16 Jul 2015 01:08:20 AM CST #
 ******************************************************************************/

#include "chat.h"
#include "mesg.h"
#include "access.h"
#include "listend.h"
#include "cmd_list.h"
#include "lsnd_mesg.h"

#include "mesg.pb-c.h"

static int lsnd_callback_creat_handler(lsnd_cntx_t *lsnd, socket_t *sck, lsnd_conn_extra_t *extra);
static int lsnd_callback_destroy_handler(lsnd_cntx_t *lsnd, socket_t *sck, lsnd_conn_extra_t *extra);
static int lsnd_callback_recv_handler(lsnd_cntx_t *lsnd, socket_t *sck, lsnd_conn_extra_t *extra, void *in, int len);

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//
/******************************************************************************
 **函数名称: lsnd_mesg_def_handler
 **功    能: 消息默认处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 直接将消息转发给上游服务.
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int lsnd_mesg_def_handler(lsnd_conn_extra_t *conn, unsigned int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    head->sid = conn->sid;
    head->cid = conn->cid;
    head->nid = conn->nid;


    log_debug(lsnd->log, "Call default handler! type:0x%04X sid:%lu cid:%d nid:%d seq:%lu len:%d!",
            head->type, head->sid, head->cid, head->nid, head->seq, len);

    MESG_HEAD_NTOH(head, head);

    /* > 转发数据 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_online_handler
 **功    能: ONLINE请求处理
 **输入参数:
 **     conn: 连接信息
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 无需提取任何信息, 直接转发给上游服务.
 **协议格式:
 **     {
 **        optional uint64 uid = 1;         // M|用户ID|数字|
 **        optional string token = 2;       // M|鉴权TOKEN|字串|
 **        optional string app = 3;         // M|APP名|字串|
 **        optional string version = 4;     // M|APP版本|字串|
 **        optional uint32 terminal = 5;    // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     }
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int lsnd_mesg_online_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);

    conn->sid = head->sid;
    head->cid = conn->cid;
    head->nid = conf->nid;

    /* > 插入SID表(SID+CID为主键) */
    if (hash_tab_insert(lsnd->conn_list, conn, WRLOCK)) {
        log_error(lsnd->log, "Insert into sid table failed! sid:%lu cid:%lu seq:%lu len:%d!",
                conn->sid, conn->cid, head->seq, len);
        return -1;
    }

    log_debug(lsnd->log, "Head is valid! sid:%lu cid:%lu seq:%lu len:%d!",
            head->sid, head->cid, head->seq, len);

    MESG_HEAD_HTON(head, head);

    /* > 转发ONLINE请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_online_ack_logic
 **功    能: ONLINE应答逻辑处理
 **输入参数:
 **     lsnd: 全局对象
 **     ack: 上线应答
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答信息中提取UID, SID等信息, 并构建索引关系.
 **注意事项: 更新变量extra的相关数据时, 无需加锁.
 **作    者: # Qifeng.zou # 2016.10.01 21:06:07 #
 ******************************************************************************/
static int lsnd_mesg_online_ack_logic(lsnd_cntx_t *lsnd, MesgOnlineAck *ack, uint64_t cid)
{
    uint64_t old_cid;
    lsnd_conn_extra_t *extra, key, *old_extra;

    /* > 查找扩展数据 */
    key.sid = ack->sid;
    key.cid = cid;

    extra = hash_tab_query(lsnd->conn_list, &key, WRLOCK);
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find connection! sid:%lu cid:%lu", ack->sid, cid);
        return -1;
    } else if (CHAT_CONN_STAT_ESTABLISH != extra->stat) {
        log_error(lsnd->log, "Connection status isn't establish! sid:%lu cid:%lu", ack->sid, cid);
        lsnd_kick_add(lsnd, extra);
        hash_tab_unlock(lsnd->conn_list, &key, WRLOCK);
        return -1;
    } else if (0 == ack->sid) { /* SID分配失败 */
        lsnd_kick_add(lsnd, extra);
        hash_tab_unlock(lsnd->conn_list, &key, WRLOCK);
        log_error(lsnd->log, "Alloc sid failed! kick this connection! sid:%lu cid:%lu errmsg:%s",
                ack->sid, cid, ack->errmsg);
        return -1;
    }

    /* > 更新扩展数据 */
    extra->sid = ack->sid;
    extra->seq = ack->seq;
    extra->stat = CHAT_CONN_STAT_ONLINE;

    snprintf(extra->app_name, sizeof(extra->app_name), "%s", ack->app);
    snprintf(extra->app_vers, sizeof(extra->app_vers), "%s", ack->version);
    extra->terminal = ack->terminal;

    hash_tab_unlock(lsnd->conn_list, &key, WRLOCK);

    /* > 踢除老连接 */
    old_cid = chat_get_cid_by_sid(lsnd->chat_tab, key.sid);
    if ((0 != old_cid) && (old_cid != cid)) {
        key.sid = key.sid;
        key.cid = old_cid;
        old_extra = hash_tab_query(lsnd->conn_list, &key, WRLOCK);
        if (NULL != old_extra) {
            lsnd_kick_add(lsnd, old_extra);
            hash_tab_unlock(lsnd->conn_list, &key, WRLOCK);
        }
    }

    chat_set_sid_to_cid(lsnd->chat_tab, key.sid, cid);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_upmesg_online_ack_handler
 **功    能: ONLINE应答处理
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答信息中提取UID, SID等信息, 并构建索引关系.
 ** {
 **     required uint64 uid = 1;        // M|用户ID|数字|
 **     required uint64 sid = 2;        // M|会话ID|数字|
 **     required uint64 seq = 3;        // M|消息序列号|数字|
 **     required string app = 4;        // M|APP名|字串|
 **     required string version = 5;    // M|APP版本|字串|
 **     optional uint32 terminal = 6;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     required uint32 code = 7;     // M|错误码|数字|
 **     required string errmsg = 8;     // M|错误描述|字串|
 ** }
 **注意事项: 此时head.sid为cid.
 **作    者: # Qifeng.zou # 2016.09.20 23:38:38 #
 ******************************************************************************/
int lsnd_upmesg_online_ack_handler(int type, int orig, char *data, size_t len, void *args)
{
    MesgOnlineAck *ack;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv online ack!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 提取有效信息 */
    ack = mesg_online_ack__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == ack) {
        log_error(lsnd->log, "Unpack online ack failed! body:%s", head->body);
        return -1;
    }

    if (lsnd_mesg_online_ack_logic(lsnd, ack, hhead.cid)) {
        mesg_online_ack__free_unpacked(ack, NULL);
        return -1;
    }

    mesg_online_ack__free_unpacked(ack, NULL);

    /* 下发应答请求 */
    return acc_async_send(lsnd->access, type, hhead.cid, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_offline_handler
 **功    能: 下线请求处理
 **输入参数:
 **     conn: 连接信息
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败(注: 该函数始终返回-1)
 **实现描述: 修改连接状态 + 并释放相关资源.
 **注意事项:
 **     1.需要将协议头转换为"本机"字节序
 **     2.无需直接转发下线请求给上游模块, 待执行CLOSE操作时在上报上游模块.
 **作    者: # Qifeng.zou # 2016.10.01 09:15:01 #
 ******************************************************************************/
int lsnd_mesg_offline_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_conn_extra_t *extra, key;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data, hhead; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    head->cid = hton64(conn->cid);
    head->nid = htonl(conf->nid);

    log_debug(lsnd->log, "sid:%lu seq:%lu len:%d body:%s!",
            hhead.sid, hhead.seq, len, hhead.body);

    /* > 查找扩展数据 */
    key.sid = hhead.sid;
    key.cid = conn->cid;

    extra = hash_tab_query(lsnd->conn_list, &key, WRLOCK); // 加写锁
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find connection! sid:%lu cid:%lu", hhead.sid, conn->cid);
        return -1;
    }

    extra->stat = CHAT_CONN_STAT_OFFLINE;

    hash_tab_unlock(lsnd->conn_list, &key, WRLOCK); // 解锁

    /* > 发送数据 */
    rtmq_proxy_async_send(lsnd->frwder, CMD_OFFLINE, data, len);

    return -1; /* 强制下线 */
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_room_join_handler
 **功    能: JOIN请求处理
 **输入参数:
 **     conn: 连接信息
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 请求数据的内存结构: 流水信息 + 消息头 + 消息体
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int lsnd_mesg_room_join_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);

    head->nid = conf->nid;

    log_debug(lsnd->log, "Head is valid! sid:%lu seq:%lu len:%d.",
            head->sid, head->seq, len);

    MESG_HEAD_HTON(head, head);

    /* > 转发JOIN请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_upmesg_room_join_ack_handler
 **功    能: JOIN应答处理
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答中获取加入聊天室是否成功. 如果成功则构建索引.
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2016.09.20 23:40:12 #
 ******************************************************************************/
int lsnd_upmesg_room_join_ack_handler(int type, int orig, char *data, size_t len, void *args)
{
    uint32_t gid;
    uint64_t cid;
    MesgRoomJoinAck *ack;
    lsnd_conn_extra_t *extra, key;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)
    log_debug(lsnd->log, "Recv room join ack! sid:%lu seq:%lu", hhead.sid, hhead.seq);

    /* > 提取应答信息 */
    ack = mesg_room_join_ack__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == ack) {
        log_error(lsnd->log, "Unpack join ack body failed!");
        return 0;
    }

    /* > 查找扩展数据 */
    key.sid = hhead.sid;
    key.cid = chat_get_cid_by_sid(lsnd->chat_tab, hhead.sid);

    extra = hash_tab_query(lsnd->conn_list, &key, WRLOCK); // 加写锁
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find socket from sid table! sid:%lu", hhead.sid);
        mesg_room_join_ack__free_unpacked(ack, NULL);
        return 0;
    } else if (CHAT_CONN_STAT_ONLINE != extra->stat) {
        hash_tab_unlock(lsnd->conn_list, &key, WRLOCK); // 解锁
        mesg_room_join_ack__free_unpacked(ack, NULL);
        log_error(lsnd->log, "Connection status isn't online! sid:%lu", hhead.sid);
        return 0;
    }

    cid = extra->cid;

    /* > 更新扩展数据 */
    extra->stat = CHAT_CONN_STAT_ONLINE;

    /* 将SID加入聊天室 */
    gid = chat_room_add_session(lsnd->chat_tab, ack->rid, ack->gid, extra->sid, cid);
    if ((uint32_t)-1 == gid) {
        log_error(lsnd->log, "Add into chat room failed! sid:%lu rid:%lu gid:%u",
                hhead.sid, ack->rid, ack->gid);
        hash_tab_unlock(lsnd->conn_list, &key, WRLOCK); // 解锁
        mesg_room_join_ack__free_unpacked(ack, NULL);
        return -1;
    }

    hash_tab_unlock(lsnd->conn_list, &key, WRLOCK); // 解锁
    mesg_room_join_ack__free_unpacked(ack, NULL);

    /* 下发应答请求 */
    return acc_async_send(lsnd->access, type, cid, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_room_quit_handler
 **功    能: ROOM-QUIT请求处理(退出聊天室)
 **输入参数:
 **     conn: 连接信息
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 请求数据的内存结构: 流水信息 + 消息头 + 消息体
 **  {
 **     optional uint64 uid = 1;    // M|用户ID|数字|
 **     optional uint64 rid = 2;    // M|聊天室ID|数字|
 **  }
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int lsnd_mesg_room_quit_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    MesgRoomQuit *quit;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);

    head->cid = conn->cid;
    head->nid = conf->nid;

    log_debug(lsnd->log, "Head is valid! sid:%lu cid:%lu seq:%lu len:%d.",
            head->sid, head->cid, head->seq, len);

    /* > 解析退出请求 */
    quit = mesg_room_quit__unpack(NULL, head->length, (void *)(head + 1));
    if (NULL == quit) {
        log_error(lsnd->log, "Quit room is invalid! sid:%lu seq:%lu len:%d.",
                head->sid, head->seq, len);
        return -1;
    }

    log_debug(lsnd->log, "Quit room! uid:%lu rid:%lu sid:%lu cid:%lu!",
            quit->uid, quit->rid, head->sid, head->cid);

    /* > 从聊天室删除此会话 */
    chat_room_del_session(lsnd->chat_tab, quit->rid, conn->sid, conn->cid);

    mesg_room_quit__free_unpacked(quit, NULL);

    /* > 转发UNJOIN请求 */
    MESG_HEAD_HTON(head, head);

    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_mesg_ping_handler
 **功    能: PING请求处理(心跳)
 **输入参数:
 **     conn: 连接信息
 **     type: 全局对象
 **     data: 数据内容
 **     len: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 请求数据的内存结构: 流水信息 + 消息头 + 消息体
 **  {
 **     optional uint64 uid = 1;    // M|用户ID|数字|
 **     optional uint64 rid = 2;    // M|聊天室ID|数字|
 **  }
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int lsnd_mesg_ping_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);

    log_debug(lsnd->log, "cid:%lu sid:%lu seq:%lu len:%d.",
            conn->cid, head->sid, head->seq, len);

    head->nid = conf->nid;
    head->type = CMD_PONG;

    /* > 发送PONG应答 */
    MESG_HEAD_HTON(head, head);

    acc_async_send(lsnd->access, CMD_PONG, conn->cid, head, sizeof(mesg_header_t));

    return 0; 
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_room_mesg_trav_send_handler
 **功    能: 依次针对各SESSION下发聊天室消息
 **输入参数:
 **     sid: 会话ID
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 待完善
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.10.01 20:25:42 #
 ******************************************************************************/
typedef struct
{
    void *data;                 // 被发送数据
    size_t length;              // 被发数据长度
    lsnd_cntx_t *lsnd;          // 帧听层对象
    mesg_header_t *hhead;       // 主机套接字
} lsnd_room_mesg_param_t;

static int lsnd_room_mesg_trav_send_handler(chat_sid2cid_item_t *item, lsnd_room_mesg_param_t *param)
{
    lsnd_cntx_t *lsnd = param->lsnd;
    mesg_header_t *head = param->hhead;

    /* > 下发数据给指定连接 */
    acc_async_send(lsnd->access, head->type, item->cid, param->data, param->length);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_chat_handler
 **功    能: 下发聊天室消息(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2016.09.25 01:24:45 #
 ******************************************************************************/
int lsnd_upmesg_room_chat_handler(int type, int orig, void *data, size_t len, void *args)
{
    uint32_t gid;
    MesgRoomChat *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_chat__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room message failed!");
        return -1;
    }

    gid = mesg->gid;

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, gid,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_chat__free_unpacked(mesg, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_bc_handler
 **功    能: 下发聊天室广播消息(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2017.05.08 22:29:58 #
 ******************************************************************************/
int lsnd_upmesg_room_bc_handler(int type, int orig, void *data, size_t len, void *args)
{
    MesgRoomBc *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv chat room broadcast!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_bc__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room broadcast failed!");
        return -1;
    }

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, 0,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_bc__free_unpacked(mesg, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_usr_num_handler
 **功    能: 下发聊天室用户人数消息(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.28 18:27:51 #
 ******************************************************************************/
int lsnd_upmesg_room_usr_num_handler(int type, int orig, void *data, size_t len, void *args)
{
    MesgRoomUsrNum *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv chat room usr num!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_usr_num__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room usr num failed!");
        return -1;
    }

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, 0,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_usr_num__free_unpacked(mesg, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_quit_ntf_handler
 **功    能: 下发聊天室下线通知(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 11:20:57 #
 ******************************************************************************/
int lsnd_upmesg_room_quit_ntf_handler(int type, int orig, void *data, size_t len, void *args)
{
    MesgRoomQuitNtf *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv chat room quit ntf!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_quit_ntf__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room quit ntf failed!");
        return -1;
    }

    log_debug(lsnd->log, "Room quit ntf! rid:%lu uid:%lu", mesg->rid, mesg->uid);

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, 0,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_quit_ntf__free_unpacked(mesg, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_kick_ntf_handler
 **功    能: 下发聊天室踢人通知(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.04 11:20:57 #
 ******************************************************************************/
int lsnd_upmesg_room_kick_ntf_handler(int type, int orig, void *data, size_t len, void *args)
{
    MesgRoomKickNtf *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv chat room kick ntf!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_kick_ntf__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room kick ntf failed!");
        return -1;
    }

    log_debug(lsnd->log, "Room kick ntf! rid:%lu uid:%lu", mesg->rid, mesg->uid);

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, 0,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_kick_ntf__free_unpacked(mesg, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_upmesg_room_join_ntf_handler
 **功    能: 下发聊天室上线通知(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2017.07.04 11:20:57 #
 ******************************************************************************/
int lsnd_upmesg_room_join_ntf_handler(int type, int orig, void *data, size_t len, void *args)
{
    MesgRoomJoinNtf *mesg;
    lsnd_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    log_debug(lsnd->log, "Recv chat room join ntf!");

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 解压PROTO-BUF */
    mesg = mesg_room_join_ntf__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == mesg) {
        log_error(lsnd->log, "Unpack chat room join ntf failed!");
        return -1;
    }

    log_debug(lsnd->log, "Room join ntf! rid:%lu uid:%lu", mesg->rid, mesg->uid);

    /* > 给制定聊天室和分组发送消息 */
    param.lsnd = lsnd;
    param.data = data;
    param.length = len;
    param.hhead = &hhead;

    chat_room_trav_session(lsnd->chat_tab, mesg->rid, 0,
            (trav_cb_t)lsnd_room_mesg_trav_send_handler, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room_join_ntf__free_unpacked(mesg, NULL);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_upmesg_kick_handler
 **功    能: 将某连接KICK下线(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2016.12.17 06:27:21 #
 ******************************************************************************/
int lsnd_upmesg_kick_handler(int type, int orig, void *data, size_t len, void *args)
{
    uint64_t cid;
    MesgKick *kick;
    lsnd_conn_extra_t *conn, key;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 打印被踢原因 */
    kick = mesg_kick__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == kick) {
        log_error(lsnd->log, "Unpack kick command failed!");
        return -1;
    }

    log_debug(lsnd->log, "Kick session [%d]! code:%d errmsg:%s", hhead.sid, kick->code, kick->errmsg);

    mesg_kick__free_unpacked(kick, NULL);

    /* > 查找对应的连接 */
    key.sid = hhead.sid;
    key.cid = (0 == hhead.cid)? chat_get_cid_by_sid(lsnd->chat_tab, hhead.sid) : hhead.cid;

    conn = hash_tab_delete(lsnd->conn_list, &key, WRLOCK);
    if (NULL == conn) {
        log_error(lsnd->log, "Didn't find socket from sid table! sid:%lu", hhead.sid);
        return -1;
    }

    cid = conn->cid;
    lsnd_kick_add(lsnd, conn); /* 放入被踢列表 */

    /* > 转发被踢原因 */
    acc_async_send(lsnd->access, type, cid, data, len);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_upmesg_def_handler
 **功    能: 默认消息处理(下行)
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 注意hash tab加锁时, 不要造成死锁的情况.
 **作    者: # Qifeng.zou # 2017.05.28 14:12:43 #
 ******************************************************************************/
int lsnd_upmesg_def_handler(int type, int orig, void *data, size_t len, void *args)
{
    uint64_t cid;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 查找对应的连接 */
    cid = (0 == hhead.cid)? chat_get_cid_by_sid(lsnd->chat_tab, hhead.sid) : hhead.cid;

    /* > 转发被踢原因 */
    acc_async_send(lsnd->access, type, cid, data, len);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_callback
 **功    能: CHAT处理回调
 **输入参数:
 **     acc: Access
 **     sck: 套接字
 **     reason: 回调的原因
 **     user: 扩展数据
 **     in: 接收数据
 **     len: 接收数据长度
 **     args: 附加数据. 当前为lsnd_cntx_t对象.
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 各连接上下文由固定线程维护. 因此, 该连接对象数据的删除只能由固定线程
 **          操作, 故无需加锁.
 **作    者: # Qifeng.zou # 2016.09.20 22:03:02 #
 ******************************************************************************/
int lsnd_callback(acc_cntx_t *acc,
        socket_t *sck, int reason, void *user, void *in, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conn_extra_t *extra = (lsnd_conn_extra_t *)user;

    switch (reason) {
        case ACC_CALLBACK_SCK_CREAT:
            return lsnd_callback_creat_handler(lsnd, sck, extra);
        case ACC_CALLBACK_SCK_CLOSED:
            return 0;
        case ACC_CALLBACK_SCK_DESTROY:
            return lsnd_callback_destroy_handler(lsnd, sck, extra);
        case ACC_CALLBACK_RECEIVE:
            return lsnd_callback_recv_handler(lsnd, sck, extra, in, len);
        case ACC_CALLBACK_WRITEABLE:
        default:
            break;
    }
    return 0;
}

/* 聊天室ID哈希回调 */
static uint64_t lsnd_rid_list_hash_cb(uint64_t *rid)
{
    return (uint64_t)rid;
}

/* 聊天室ID比较回调 */
static int lsnd_rid_list_cmp_cb(uint64_t *rid1, uint64_t *rid2)
{
    return (int)((uint64_t)rid1 - (uint64_t)rid2);
}

/******************************************************************************
 **函数名称: lsnd_callback_creat_handler
 **功    能: 连接创建的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     extra: 扩展数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 将新建连接放入CONN_CID_TAB维护起来, 待分配了SID后再转移到CONN_SID_TAB中.
 **作    者: # Qifeng.zou # 2016.09.20 21:30:53 #
 ******************************************************************************/
static int lsnd_callback_creat_handler(lsnd_cntx_t *lsnd, socket_t *sck, lsnd_conn_extra_t *extra)
{
    time_t ctm = time(NULL);
    lsnd_conf_t *conf = &lsnd->conf;

    /* 初始化设置 */
    extra->ctx = lsnd;
    extra->sck = sck;
    pthread_rwlock_init(&extra->lock, NULL);

    extra->sid = 0;
    extra->cid = acc_sck_get_cid(sck);
    extra->nid = conf->nid;
    extra->create_time = ctm;
    extra->recv_time = ctm;
    extra->send_time = ctm;
    extra->keepalive_time = ctm;
    extra->stat = CHAT_CONN_STAT_ESTABLISH;

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_callback_destroy_handler
 **功    能: 连接销毁的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     user: 扩展数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **     1. 释放extra对象内存的所有空间, 但是请勿释放extra对象本身.
 **     2. 释放extra前, 必须将该对象从其他各表中删除, 否则存在多线程同时操作一块的风险.
 **     3. 对象extra的内存空间由access模块框架释放
 **作    者: # Qifeng.zou # 2016.09.20 21:43:13 #
 ******************************************************************************/
static int lsnd_callback_destroy_handler(lsnd_cntx_t *lsnd, socket_t *sck, lsnd_conn_extra_t *extra)
{
    lsnd_conn_extra_t key;

    /* > 从CONN列表中删除 */
    key.sid = extra->sid;
    key.cid = extra->cid;
    hash_tab_delete(lsnd->conn_list, &key, WRLOCK);

    /* > 清理相关数据 */
    pthread_rwlock_destroy(&extra->lock);

    extra->stat = CHAT_CONN_STAT_CLOSED;
    chat_del_session(lsnd->chat_tab, extra->sid, extra->cid);

    /* > 发送下线指令 */
    lsnd_offline_notify(lsnd, extra->sid, extra->cid, extra->nid);

    log_debug(lsnd->log, "Connection was closed! sid:%lu cid:%lu", extra->sid, extra->cid);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_callback_recv_handler
 **功    能: 接收数据的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     conn: 连接数据
 **     in: 收到的数据
 **     len: 收到数据的长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **     1. 暂无需加锁. 原因: 注册表在程序启动时, 就已固定不变.
 **     2. 本函数收到的数据是一条完整的数据, 且其内容网络字节序.
 **     3. 消息序列号必须依次递增
 **作    者: # Qifeng.zou # 2016.09.20 21:44:40 #
 ******************************************************************************/
static int lsnd_callback_recv_handler(lsnd_cntx_t *lsnd,
    socket_t *sck, lsnd_conn_extra_t *conn, void *in, int len)
{
    lsnd_reg_t *reg, key;
    mesg_header_t hhead, *head = (mesg_header_t *)in;

    /* > 字节序转换 */
    MESG_HEAD_NTOH(head, &hhead);

    /* > 更新序列号 */
    pthread_rwlock_wrlock(&conn->lock);
    if (CHAT_CONN_STAT_KICK == conn->stat) {
        pthread_rwlock_unlock(&conn->lock);
        log_debug(lsnd->log, "This connection was kicked! sid:%lu cid:%lu",
                conn->sid, conn->cid);
        return 0;
    } else if (conn->seq >= hhead.seq) {
        pthread_rwlock_unlock(&conn->lock);
        log_debug(lsnd->log, "Message seq is invalid! cmd:0x%04X sid:%lu seq:%lu/%lu len:%d",
                hhead.type, hhead.sid, hhead.seq, conn->seq, len);
        return -1;
    }
    conn->seq = hhead.seq;
    pthread_rwlock_unlock(&conn->lock);

    log_debug(lsnd->log, "Recv data! cid:%lu", conn->cid);

    /* > 查找回调函数 */
    key.type = hhead.type;

    reg = avl_query(lsnd->reg, &key);
    if (NULL == reg) {
        if (CHAT_CONN_STAT_ONLINE != conn->stat) {
            log_warn(lsnd->log, "Drop unknown data! type:0x%04X", key.type);
            return 0;
        }

        /* > 查找默认处理回调 */
        key.type = CMD_UNKNOWN;

        reg = avl_query(lsnd->reg, &key);
        if (NULL == reg) {
            log_warn(lsnd->log, "Drop unknown data! type:0x%04X", key.type);
            return 0;
        }
    }

    return reg->proc(conn, hhead.type, in, len, reg->args);
}

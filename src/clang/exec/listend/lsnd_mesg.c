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

#include "mesg_room.pb-c.h"
#include "mesg_online.pb-c.h"
#include "mesg_online_ack.pb-c.h"
#include "mesg_join_ack.pb-c.h"

static int chat_callback_creat_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra);
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra);
static int chat_callback_recv_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra, void *in, int len);

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//
/******************************************************************************
 **函数名称: chat_mesg_def_hdl
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
int chat_mesg_def_hdl(chat_conn_extra_t *conn, unsigned int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t hhead, *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, &hhead);

    log_debug(lsnd->log, "sid:%lu serial:%lu len:%d body:%s!",
            hhead.sid, hhead.serial, len, hhead.body);

    /* > 转发数据 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_online_req_hdl
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
int chat_mesg_online_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);
    if (!MESG_CHKSUM_ISVALID(head)) {
        log_error(lsnd->log, "Head is invalid! sid:%lu serial:%lu len:%d chksum:0x%08X!",
                head->sid, head->serial, len, head->chksum);
        return -1;
    }

    head->sid = conn->cid;
    head->nid = conf->nid;

    log_debug(lsnd->log, "Head is invalid! sid:%lu serial:%lu len:%d chksum:0x%08X!",
            head->sid, head->serial, len, head->chksum);

    MESG_HEAD_HTON(head, head);

    /* > 转发ONLINE请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

/******************************************************************************
 **函数名称: chat_mesg_online_ack_logic_hdl
 **功    能: ONLINE应答逻辑处理
 **输入参数:
 **     lsnd: 全局对象
 **     ack: 上线应答
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答信息中提取UID, SID等信息, 并构建索引关系.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.01 21:06:07 #
 ******************************************************************************/
static int chat_mesg_online_ack_logic_hdl(lsnd_cntx_t *lsnd, MesgOnlineAck *ack, uint64_t cid)
{
    chat_conn_extra_t *extra, key;

    /* > 查找扩展数据 */
    key.cid = cid;

    extra = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find socket from cid table! cid:%lu", cid);
        return -1;
    }

    extra->loc &= ~CHAT_EXTRA_LOC_CID_TAB;

    if (CHAT_CONN_STAT_ESTABLISH != extra->stat) {
        log_error(lsnd->log, "Connection status isn't establish! cid:%lu", cid);
        return -1;
    }
    else if (0 == ack->sid) { /* SID分配失败 */
        lsnd_kick_insert(lsnd, extra);
        log_error(lsnd->log, "Alloc sid failed! kick this connection! cid:%lu errmsg:%s", cid, ack->errmsg);
        return 0;
    }

    extra->sid = ack->sid;
    extra->loc |= CHAT_EXTRA_LOC_SID_TAB;
    extra->stat = CHAT_CONN_STAT_ONLINE;

    snprintf(extra->app_name, sizeof(extra->app_name), "%s", ack->app);
    snprintf(extra->app_vers, sizeof(extra->app_vers), "%s", ack->version);
    extra->terminal = ack->terminal;

    /* 插入SID管理表 */
    if (hash_tab_insert(lsnd->conn_sid_tab, extra, WRLOCK)) {
        log_error(lsnd->log, "Connection is in sid table!");
        assert(0);
        return 0;
    }

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_online_ack_hdl
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
 **     required string app = 3;        // M|APP名|字串|
 **     required string version = 4;    // M|APP版本|字串|
 **     optional uint32 terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     required uint32 code = 6;     // M|错误码|数字|
 **     required string errmsg = 7;     // M|错误描述|字串|
 ** }
 **注意事项: 此时head.sid为cid.
 **作    者: # Qifeng.zou # 2016.09.20 23:38:38 #
 ******************************************************************************/
int chat_mesg_online_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    uint64_t cid;
    MesgOnlineAck *ack;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    cid = hhead.sid;

    MESG_HEAD_PRINT(lsnd->log, &hhead)

    /* > 提取有效信息 */
    ack = mesg_online_ack__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == ack) {
        log_error(lsnd->log, "Unpack online ack failed! body:%s", head->body);
        return -1;
    }

    if (chat_mesg_online_ack_logic_hdl(lsnd, ack, cid)) {
        mesg_online_ack__free_unpacked(ack, NULL);
        log_error(lsnd->log, "Miss required field!");
        return -1;
    }

    /* 下发应答请求 */
    acc_async_send(lsnd->access, type, cid, data, len);

    mesg_online_ack__free_unpacked(ack, NULL);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_offline_req_hdl
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
 **注意事项: 需要将协议头转换为"本机"字节序
 **作    者: # Qifeng.zou # 2016.10.01 09:15:01 #
 ******************************************************************************/
int chat_mesg_offline_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    chat_conn_extra_t *extra, key;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data, hhead; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    head->nid = ntohl(conf->nid);

    log_debug(lsnd->log, "sid:%lu serial:%lu len:%d body:%s!",
            hhead.sid, hhead.serial, len, hhead.body);

    /* > 查找扩展数据 */
    key.sid = head->sid;

    extra = hash_tab_query(lsnd->conn_sid_tab, &key, WRLOCK); // 加写锁
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find socket from sid table! sid:%lu", head->sid);
        return -1;
    }

    extra->stat = CHAT_CONN_STAT_OFFLINE;

    hash_tab_unlock(lsnd->conn_sid_tab, &key, WRLOCK); // 解锁

    /* > 转发下线请求 */
    rtmq_proxy_async_send(lsnd->frwder, type, data, len);

    return -1; /* 强制下线 */
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_join_req_hdl
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
int chat_mesg_join_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data, hhead; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    head->nid = ntohl(conf->nid);

    log_debug(lsnd->log, "sid:%lu serial:%lu len:%d body:%s!",
            hhead.sid, hhead.serial, len, hhead.body);

    /* > 转发JOIN请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_join_ack_hdl
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
int chat_mesg_join_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    uint32_t gid;
    uint64_t cid;
    MesgJoinAck *ack;
    chat_conn_extra_t *extra, key;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)
    log_debug(lsnd->log, "body:%s", head->body);

    /* > 提取应答信息 */
    ack = mesg_join_ack__unpack(NULL, hhead.length, (void *)(head + 1));
    if (NULL == ack) {
        log_error(lsnd->log, "Unpack join ack body failed!");
        return 0;
    }

    /* > 查找扩展数据 */
    key.sid = hhead.sid;

    extra = hash_tab_query(lsnd->conn_sid_tab, &key, WRLOCK); // 加写锁
    if (NULL == extra) {
        log_error(lsnd->log, "Didn't find socket from sid table! sid:%lu", hhead.sid);
        mesg_join_ack__free_unpacked(ack, NULL);
        return 0;
    }
    else if (CHAT_CONN_STAT_ONLINE != extra->stat) {
        hash_tab_unlock(lsnd->conn_sid_tab, &key, WRLOCK); // 解锁
        mesg_join_ack__free_unpacked(ack, NULL);
        log_error(lsnd->log, "Connection status isn't online! sid:%lu", hhead.sid);
        return 0;
    }

    cid = extra->cid;

    /* > 更新扩展数据 */
    extra->loc = CHAT_EXTRA_LOC_SID_TAB;
    extra->stat = CHAT_CONN_STAT_ONLINE;

    /* 将SID加入聊天室 */
    gid = chat_room_add_session(lsnd->chat_tab, ack->rid, ack->gid, extra->sid);
    if ((uint32_t)-1 == gid) {
        log_error(lsnd->log, "Add into chat room failed! sid:%lu rid:%lu gid:%u",
                hhead.sid, ack->rid, ack->gid);
        hash_tab_unlock(lsnd->conn_sid_tab, &key, WRLOCK); // 解锁
        mesg_join_ack__free_unpacked(ack, NULL);
        return -1;
    }

    hash_tab_unlock(lsnd->conn_sid_tab, &key, WRLOCK); // 解锁
    mesg_join_ack__free_unpacked(ack, NULL);

    /* 下发应答请求 */
    return acc_async_send(lsnd->access, type, cid, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_unjoin_req_hdl
 **功    能: UNJOIN请求处理(退出聊天室)
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
int chat_mesg_unjoin_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t hhead, *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    head->nid = ntohl(conf->nid);

    log_debug(lsnd->log, "sid:%lu serial:%lu len:%d body:%s!",
            hhead.sid, hhead.serial, len, hhead.body);

    /* > 从聊天室中删除此会话 */
    chat_del_session(lsnd->chat_tab, hhead.sid);

    /* > 转发UNJOIN请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, len);
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_mesg_ping_req_hdl
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
int chat_mesg_ping_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    lsnd_conf_t *conf = &lsnd->conf;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    /* > 转换字节序 */
    MESG_HEAD_NTOH(head, head);

    log_debug(lsnd->log, "cid:%lu sid:%lu serial:%lu len:%d chksum:0x%08X!",
            conn->cid, head->sid, head->serial, len, head->chksum);

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
 **函数名称: chat_room_mesg_trav_send_hdl
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
} chat_room_mesg_param_t;

static int chat_room_mesg_trav_send_hdl(uint64_t *sid, chat_room_mesg_param_t *param)
{
    uint64_t cid;
    chat_conn_extra_t *extra, key;
    lsnd_cntx_t *lsnd = param->lsnd;
    mesg_header_t *head = param->hhead;

    /* > 查找扩展数据 */
    key.sid = (uint64_t)sid;

    extra = hash_tab_query(lsnd->conn_sid_tab, &key, RDLOCK);
    if (NULL == extra) {
        return 0;
    }

    cid = extra->cid;

    hash_tab_unlock(lsnd->conn_sid_tab, &key, RDLOCK);

    /* > 下发数据给指定连接 */
    acc_async_send(lsnd->access, head->type, cid, param->data, param->length);

    return 0;
}

/******************************************************************************
 **函数名称: chat_mesg_room_mesg_hdl
 **功    能: 下发聊天室消息
 **输入参数:
 **     conn: 连接信息
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
int chat_mesg_room_mesg_hdl(chat_conn_extra_t *conn, int type, int orig, void *data, size_t len, void *args)
{
    uint32_t gid;
    MesgRoom *mesg;
    chat_room_mesg_param_t param;
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(lsnd->log, &hhead)
    log_debug(lsnd->log, "body:%s", head->body);

    /* > 解压PROTO-BUF */
    mesg = mesg_room__unpack(NULL, hhead.length, (void *)(head + 1));
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

    chat_room_trav(lsnd->chat_tab, mesg->rid, gid,
            (trav_cb_t)chat_room_mesg_trav_send_hdl, (void *)&param);

    /* > 释放PROTO-BUF空间 */
    mesg_room__free_unpacked(mesg, NULL);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: chat_callback
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
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 22:03:02 #
 ******************************************************************************/
int chat_callback(acc_cntx_t *acc,
        socket_t *sck, int reason, void *user, void *in, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    chat_conn_extra_t *extra = (chat_conn_extra_t *)user;

    switch (reason) {
        case ACC_CALLBACK_SCK_CREAT:
            return chat_callback_creat_hdl(lsnd, sck, extra);
        case ACC_CALLBACK_SCK_CLOSED:
            return 0;
        case ACC_CALLBACK_SCK_DESTROY:
            return chat_callback_destroy_hdl(lsnd, sck, extra);
        case ACC_CALLBACK_RECEIVE:
            return chat_callback_recv_hdl(lsnd, sck, extra, in, len);
        case ACC_CALLBACK_WRITEABLE:
        default:
            break;
    }
    return 0;
}

/* 聊天室ID哈希回调 */
static uint64_t chat_rid_list_hash_cb(uint64_t *rid)
{
    return (uint64_t)rid;
}

/* 聊天室ID比较回调 */
static int chat_rid_list_cmp_cb(uint64_t *rid1, uint64_t *rid2)
{
    return (int)((uint64_t)rid1 - (uint64_t)rid2);
}

/******************************************************************************
 **函数名称: chat_callback_creat_hdl
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
static int chat_callback_creat_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra)
{
    time_t ctm = time(NULL);

    /* 初始化设置 */
    extra->ctx = lsnd;
    extra->sck = sck;

    extra->sid = 0;
    extra->cid = acc_sck_get_cid(sck);
    extra->create_time = ctm;
    extra->recv_time = ctm;
    extra->send_time = ctm;
    extra->keepalive_time = ctm;
    extra->loc = CHAT_EXTRA_LOC_UNKNOWN;
    extra->stat = CHAT_CONN_STAT_ESTABLISH;

    /* 加入CID管理表 */
    if (hash_tab_insert(lsnd->conn_cid_tab, (void *)extra, WRLOCK)) {
        log_error(lsnd->log, "Insert cid table failed! cid:%lu", extra->cid);
        return -1;
    }

    extra->loc |= CHAT_EXTRA_LOC_CID_TAB;

    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_destroy_hdl
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
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra)
{
    chat_conn_extra_t key, *item;

    extra->stat = CHAT_CONN_STAT_CLOSED;
    chat_del_session(lsnd->chat_tab, extra->sid);

    if (extra->loc & CHAT_EXTRA_LOC_CID_TAB) {
        key.cid = extra->cid;
        item = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
        if (item != extra) {
            assert(0);
        }
        extra->loc &= ~CHAT_EXTRA_LOC_CID_TAB;
    }

    if (extra->loc & CHAT_EXTRA_LOC_SID_TAB) {
        key.sid = extra->sid;
        item = hash_tab_delete(lsnd->conn_sid_tab, &key, WRLOCK);
        if (item != extra) {
            assert(0);
        }
        extra->loc &= ~CHAT_EXTRA_LOC_SID_TAB;
    }

    if (extra->loc & CHAT_EXTRA_LOC_KICK_TAB) {
        key.sck = sck;
        item = hash_tab_delete(lsnd->conn_kick_list, &key, WRLOCK);
        if (item != extra) {
            assert(0);
        }

        extra->loc &= ~CHAT_EXTRA_LOC_KICK_TAB;
    }

    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_recv_hdl
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
 **作    者: # Qifeng.zou # 2016.09.20 21:44:40 #
 ******************************************************************************/
static int chat_callback_recv_hdl(lsnd_cntx_t *lsnd,
    socket_t *sck, chat_conn_extra_t *conn, void *in, int len)
{
    lsnd_reg_t *reg, key;
    mesg_header_t *head = (mesg_header_t *)in;

    log_debug(lsnd->log, "Recv data! cid:%lu", conn->cid);

    key.type = ntohl(head->type);

    reg = avl_query(lsnd->reg, &key);
    if (NULL == reg) {
        if (CHAT_CONN_STAT_ONLINE != conn->stat) {
            log_warn(lsnd->log, "Drop unknown data! type:0x%X", key.type);
            return 0;
        }
        log_warn(lsnd->log, "Forward unknown data! type:0x%X", key.type);
        return chat_mesg_def_hdl(conn, key.type, in, len, (void *)lsnd);
    }

    return reg->proc(conn, reg->type, in, len, reg->args);
}

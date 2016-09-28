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
#include "lsnd_mesg.h"
#include "cjson/cJSON.h"

static int chat_callback_creat_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra);
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra);
static int chat_callback_recv_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra, void *in, int len);

/******************************************************************************
 **函数名称: chat_mesg_def_hdl
 **功    能: 消息默认处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 直接将消息转发给上游服务.
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_mesg_def_hdl(unsigned int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发数据 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_online_parse_hdl
 **功    能: ONLINE应答处理
 **输入参数:
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **协议格式:
 **     {
 **        "uid":${uid},               // M|用户ID|数字|
 **        "roomid":${roomid},         // M|聊天室ID|数字|
 **        "app":"${app}",             // M|APP名|字串|
 **        "version":"${version}",     // M|APP版本|字串|
 **        "terminal":${terminal}      // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **        "errno":${errno},           // M|错误码|数字|
 **        "errmsg":"${errmsg}"        // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.24 13:01:31 #
 ******************************************************************************/
typedef struct
{
    uint64_t uid;                       // 用户ID
    char app[CHAT_APP_NAME_LEN];        // 应用名
    char version[CHAT_APP_VERS_LEN];    // 版本号
    chat_terminal_type_e terminal;      // 终端类型
} chat_online_req_t;

static int chat_online_parse_hdl(lsnd_cntx_t *ctx,
        const char *body, uint64_t len, chat_online_req_t *req)
{
    char json_str[CHAT_JSON_STR_LEN];
    cJSON *json, *uid, *app, *version, *terminal;

    if (len >= sizeof(json_str)) {
        log_error(ctx->log, "Body is too long! len:%d", len);
        return -1;
    }

    memcpy(json_str, (const void *)body, len);
    json_str[len] = '\0';

    /* 解析JSON */
    json = cJSON_Parse(json_str);
    if (NULL == json) {
        log_error(ctx->log, "Parse join ack failed!");
        return -1;
    }

    do {
        /* 定位各结点 */
        uid = cJSON_GetObjectItem(json, "uid");
        if (NULL == uid || 0 == uid->valueint) {
            break;
        }

        app = cJSON_GetObjectItem(json, "app");
        if (NULL == app) {
            break;
        }

        version = cJSON_GetObjectItem(json, "version");
        if (NULL == version) {
            break;
        }

        terminal = cJSON_GetObjectItem(json, "terminal");
        if (NULL == terminal) {
            break;
        }

        /* 提取有效信息 */
        req->uid = uid->valueint;
        snprintf(req->app, sizeof(req->app), "%s", app->valuestring);
        snprintf(req->version, sizeof(req->version), "%s", version->valuestring);
        req->terminal = terminal->valueint;

        /* 释放内存空间 */
        cJSON_Delete(json);

        return 0;
    } while(0);

    cJSON_Delete(json);
    return -1;
}

/******************************************************************************
 **函数名称: chat_online_req_hdl
 **功    能: ONLINE请求处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 无需提取任何信息, 直接转发给上游服务.
 **协议格式:
 **     {
 **        "uid":${uid},               // M|用户ID|数字|
 **        "roomid":${roomid},         // M|聊天室ID|数字|
 **        "app":"${app}",             // M|APP名|字串|
 **        "version":"${version}",     // M|APP版本|字串|
 **        "terminal":${terminal}      // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **     }
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_online_req_hdl(int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发搜索请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_online_ack_parse_hdl
 **功    能: ONLINE应答处理
 **输入参数:
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **协议格式:
 **     {
 **        "uid":${uid},               // M|用户ID|数字|
 **        "app":"${app}",             // M|APP名|字串|
 **        "version":"${version}",     // M|APP版本|字串|
 **        "terminal":${terminal}      // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|
 **        "errno":${errno},           // M|错误码|数字|
 **        "errmsg":"${errmsg}"        // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.24 13:01:31 #
 ******************************************************************************/
typedef struct
{
    uint64_t uid;                       // 用户ID
    char app[CHAT_APP_NAME_LEN];        // 应用名
    char version[CHAT_APP_VERS_LEN];    // 版本号
    chat_terminal_type_e terminal;      // 终端类型
    int errcode;                        // 错误码
    char errmsg[ERR_MSG_MAX_LEN];       // 错误描述
} chat_online_ack_t;

static int chat_online_ack_parse_hdl(lsnd_cntx_t *ctx,
        const char *body, uint64_t len, chat_online_ack_t *ack)
{
    char ack_str[CHAT_JSON_STR_LEN];
    cJSON *json, *uid, *app, *version, *terminal, *errcode, *errmsg;

    if (len >= sizeof(ack_str)) {
        log_error(ctx->log, "Body is too long! len:%d", len);
        return -1;
    }

    memcpy(ack_str, (const void *)body, len);
    ack_str[len] = '\0';

    /* 解析JSON */
    json = cJSON_Parse(ack_str);
    if (NULL == json) {
        log_error(ctx->log, "Parse join ack failed!");
        return -1;
    }

    do {
        /* 定位各结点 */
        uid = cJSON_GetObjectItem(json, "uid");
        if (NULL == uid || 0 == uid->valueint) {
            break;
        }

        app = cJSON_GetObjectItem(json, "app");
        if (NULL == app) {
            break;
        }

        version = cJSON_GetObjectItem(json, "version");
        if (NULL == version) {
            break;
        }

        terminal = cJSON_GetObjectItem(json, "terminal");
        if (NULL == terminal) {
            break;
        }

        errcode = cJSON_GetObjectItem(json, "errno");
        if (NULL == errcode || 0 != errcode->valueint) {
            break;
        }

        errmsg = cJSON_GetObjectItem(json, "errmsg");
        if (NULL == errmsg) {
            break;
        }

        /* 提取有效信息 */
        ack->uid = uid->valueint;
        snprintf(ack->app, sizeof(ack->app), "%s", app->valuestring);
        snprintf(ack->version, sizeof(ack->version), "%s", version->valuestring);
        ack->terminal = terminal->valueint;
        ack->errcode = errcode->valueint;
        snprintf(ack->errmsg, sizeof(ack->errmsg), "%s", errmsg->valuestring);

        /* 释放内存空间 */
        cJSON_Delete(json);

        return 0;
    } while(0);

    cJSON_Delete(json);
    return -1;
}

/******************************************************************************
 **函数名称: chat_online_ack_hdl
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
 **注意事项:
 **     1. 当为ONLINE-ACK时, 序列号表示的就是CID.
 **作    者: # Qifeng.zou # 2016.09.20 23:38:38 #
 ******************************************************************************/
int chat_online_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    int ret;
    uint64_t cid;
    chat_online_ack_t ack;
    chat_conn_extra_t *extra, key;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(ctx->log, &hhead)
    log_debug(ctx->log, "body:%s", head->body);

    /* > 提取有效信息 */
    if (chat_online_ack_parse_hdl(ctx, head->body, hhead.length, &ack)) {
        log_error(ctx->log, "Parse online ack failed! body:%s", head->body);
        return -1;
    }

    cid = hhead.serial;

    /* > 查找扩展数据 */
    key.cid = cid;

    extra = hash_tab_delete(ctx->conn_cid_tab, &key, WRLOCK);
    if (NULL == extra) {
        log_error(ctx->log, "Didn't find socket from cid table! cid:%lu", cid);
        return 0;
    }
    else if (CHAT_CONN_STAT_ESTABLISH != extra->stat) {
        log_error(ctx->log, "Connection status isn't establish! cid:%lu", cid);
        return 0;
    }
    else if (0 == hhead.sid) { /* SID分配失败 */
        extra->loc = CHAT_EXTRA_LOC_KICK_TAB;
        hash_tab_insert(ctx->conn_kick_tab, extra, WRLOCK);
        log_error(ctx->log, "Alloc sid failed! kick this connection! cid:%lu", cid);
        return 0;
    }

    extra->sid = hhead.sid;
    extra->loc = CHAT_EXTRA_LOC_SID_TAB;
    extra->stat = CHAT_CONN_STAT_ONLINE;

    snprintf(extra->app_name, sizeof(extra->app_name), "%s", ack.app);
    snprintf(extra->app_vers, sizeof(extra->app_vers), "%s", ack.version);
    extra->terminal = ack.terminal;

    /* 插入SID管理表 */
    ret = hash_tab_insert(ctx->conn_sid_tab, extra, WRLOCK);
    if (0 != ret) {
        if (RBT_NODE_EXIST != ret) {
            log_error(ctx->log, "Insert into kick table! cid:%lu sid:%lu", cid, hhead.sid);
            extra->loc = CHAT_EXTRA_LOC_KICK_TAB;
            hash_tab_insert(ctx->conn_kick_tab, extra, WRLOCK);
            return 0;
        }
        assert(0);
        return 0;
    }

    /* 下发应答请求 */
    return acc_async_send(ctx->access, type, cid, data, len);
}

/******************************************************************************
 **函数名称: chat_join_req_hdl
 **功    能: JOIN请求处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 请求数据的内存结构: 流水信息 + 消息头 + 消息体
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_join_req_hdl(int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发搜索请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_join_ack_parse_hdl
 **功    能: JOIN应答处理
 **输入参数:
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **协议格式:
 **     {
 **        "uid":${uid},               // M|用户ID|数字|
 **        "roomid":${roomid},         // M|聊天室ID|数字|
 **        "groupid":${groupid},       // M|分组ID|数字|
 **        "errno":${errno},           // M|错误码|数字|
 **        "errmsg":"${errmsg}"        // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.24 13:01:31 #
 ******************************************************************************/
typedef struct
{
    uint64_t uid;                 // 用户ID
    uint64_t rid;                 // 聊天室ID
    uint64_t gid;                 // 分组ID
    int errcode;                   // 错误码
    char errmsg[ERR_MSG_MAX_LEN]; // 错误描述
} chat_join_ack_t;

static int chat_join_ack_parse_hdl(lsnd_cntx_t *ctx,
        const char *body, uint64_t len, chat_join_ack_t *ack)
{
    char json_str[CHAT_JSON_STR_LEN];
    cJSON *json, *uid, *rid, *gid, *errcode, *errmsg;

    if (len >= sizeof(json_str)) {
        log_error(ctx->log, "Body is too long! len:%d", len);
        return -1;
    }

    memcpy(json_str, (const void *)body, len);
    json_str[len] = '\0';

    /* 解析JSON */
    json = cJSON_Parse(json_str);
    if (NULL == json) {
        log_error(ctx->log, "Parse join ack failed!");
        return -1;
    }

    do {
        /* 定位各结点 */
        uid = cJSON_GetObjectItem(json, "uid");
        if (NULL == uid || 0 == uid->valueint) {
            break;
        }

        rid = cJSON_GetObjectItem(json, "rid");
        if (NULL == rid || 0 == rid->valueint) {
            break;
        }

        gid = cJSON_GetObjectItem(json, "groupid");
        if (NULL == gid) {
            break;
        }

        errcode = cJSON_GetObjectItem(json, "errno");
        if (NULL == errcode || 0 != errcode->valueint) {
            break;
        }

        errmsg = cJSON_GetObjectItem(json, "errmsg");
        if (NULL == errmsg) {
            break;
        }

        /* 提取有效信息 */
        ack->uid = uid->valueint;
        ack->rid = rid->valueint;
        ack->gid = rid->valueint;
        ack->errcode = errcode->valueint;
        snprintf(ack->errmsg, sizeof(ack->errmsg), "%s", errmsg->valuestring);

        /* 释放内存空间 */
        cJSON_Delete(json);

        return 0;
    } while(0);

    cJSON_Delete(json);
    return -1;
}

/******************************************************************************
 **函数名称: chat_join_ack_hdl
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
int chat_join_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    uint64_t cid;
    chat_join_ack_t ack;
    chat_conn_extra_t *extra, key;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(ctx->log, &hhead)
    log_debug(ctx->log, "body:%s", head->body);

    /* > 提取应答信息 */
    if (chat_join_ack_parse_hdl(ctx, head->body, head->length, &ack)) {
        log_error(ctx->log, "Json join ack body failed! body:%s", (char *)head->body);
        return 0;
    }

    /* > 查找扩展数据 */
    key.sid = hhead.sid;

    extra = hash_tab_query(ctx->conn_sid_tab, &key, WRLOCK); // 加写锁
    if (NULL == extra) {
        log_error(ctx->log, "Didn't find socket from sid table! sid:%lu", hhead.sid);
        return 0;
    }
    else if (CHAT_CONN_STAT_ONLINE != extra->stat) {
        hash_tab_unlock(ctx->conn_sid_tab, &key, WRLOCK); // 解锁
        log_error(ctx->log, "Connection status isn't online! sid:%lu", hhead.sid);
        return 0;
    }

    cid = extra->cid;

    /* > 更新扩展数据 */
    extra->loc = CHAT_EXTRA_LOC_SID_TAB;
    extra->stat = CHAT_CONN_STAT_ONLINE;

    hash_tab_insert(extra->rid_list, (void *)ack.rid, WRLOCK);
    chat_add_session(ctx->chat_tab, ack.rid, ack.gid, extra->sid);

    hash_tab_unlock(ctx->conn_sid_tab, &key, WRLOCK); // 解锁

    /* 下发应答请求 */
    return acc_async_send(ctx->access, type, cid, data, len);
}

/******************************************************************************
 **函数名称: chat_room_mesg_trav_send_hdl
 **功    能: 依次针对各SESSION下发聊天室消息
 **输入参数:
 **     ssn: 会话对象 
 **     ctx: 上下文对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 待完善
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.28 14:44:05 #
 ******************************************************************************/
static int chat_room_mesg_trav_send_hdl(chat_session_t *ssn, lsnd_cntx_t *ctx)
{
    return 0;
}

/******************************************************************************
 **函数名称: chat_room_mesg_hdl
 **功    能: 下发聊天室消息
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
int chat_room_mesg_hdl(int type, int orig, char *data, size_t len, void *args)
{
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(ctx->log, &hhead)
    log_debug(ctx->log, "body:%s", head->body);

    return chat_room_trav(ctx->chat_tab,
            hhead.sid, 0, (trav_cb_t)chat_room_mesg_trav_send_hdl, (void *)ctx);
}

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
    extra->sid = 0;
    extra->cid = acc_sck_get_cid(sck);
    extra->sck = sck;
    extra->create_time = ctm;
    extra->recv_time = ctm;
    extra->send_time = ctm;
    extra->keepalive_time = ctm;
    extra->loc = CHAT_EXTRA_LOC_UNKNOWN;
    extra->stat = CHAT_CONN_STAT_ESTABLISH;
    extra->rid_list = hash_tab_creat(5,
            (hash_cb_t)chat_rid_list_hash_cb,
            (cmp_cb_t)chat_rid_list_cmp_cb, NULL);
    if (NULL == extra->rid_list) {
        log_error(lsnd->log, "Create rid list failed!");
        return -1;
    }

    /* 加入CID管理表 */
    if (hash_tab_insert(lsnd->conn_cid_tab, (void *)extra, WRLOCK)) {
        log_error(lsnd->log, "Insert cid table failed!");
        hash_tab_destroy(extra->rid_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
        extra->rid_list = NULL;
        return -1;
    }

    extra->loc = CHAT_EXTRA_LOC_CID_TAB;

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
 **作    者: # Qifeng.zou # 2016.09.20 21:43:13 #
 ******************************************************************************/
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_extra_t *extra)
{
    chat_conn_extra_t key, *item;

    extra->stat = CHAT_CONN_STAT_CLOSED;
    if (NULL != extra->rid_list) {
        hash_tab_destroy(extra->rid_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
    }

    switch (extra->loc) {
        case CHAT_EXTRA_LOC_CID_TAB:
            key.cid = extra->cid;
            item = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
            if (item != extra) {
                assert(0);
            }
        case CHAT_EXTRA_LOC_SID_TAB:
            key.sid = extra->sid;
            item = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
            if (item != extra) {
                assert(0);
            }
        case CHAT_EXTRA_LOC_KICK_TAB:
            key.sck = sck;
            item = hash_tab_delete(lsnd->conn_kick_tab, &key, WRLOCK);
            if (item != extra) {
                assert(0);
            }
        default:
            assert(0);
    }

    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_recv_hdl
 **功    能: 接收数据的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     extra: 扩展数据
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
    socket_t *sck, chat_conn_extra_t *extra, void *in, int len)
{
    lsnd_reg_t *reg, key;
    mesg_header_t *head = (mesg_header_t *)in;

    key.type = ntohl(head->type);

    reg = avl_query(lsnd->reg, &key);
    if (NULL == reg) {
        if (CHAT_CONN_STAT_ONLINE != extra->stat) {
            log_warn(lsnd->log, "Drop unknown data! type:0x%X", key.type);
            return 0;
        }
        log_warn(lsnd->log, "Forward unknown data! type:0x%X", key.type);
        return chat_mesg_def_hdl(key.type, in, len, (void *)lsnd);
    }

    return reg->proc(reg->type, in, len, reg->args);
}

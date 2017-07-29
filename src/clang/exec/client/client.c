#include "sdk.h"
#include "redo.h"
#include "cmd_list.h"
#include "mesg.pb-c.h"

int room_creat(sdk_cntx_t *ctx);

/* > 设置配置信息 */
void client_set_conf(sdk_conf_t *conf)
{
    memset(conf, 0, sizeof(sdk_conf_t));

    conf->nid = 0; /* 设备ID: 唯一值 */
    snprintf(conf->path, sizeof(conf->path), "."); /* 工作路径 */
    conf->uid = 18600522324;                 /* 用户ID */
    conf->sid = 1234;                 /* 会话ID(备选) */
    conf->terminal = 1;                /* 终端类型 */
    snprintf(conf->app, sizeof(conf->app), "beehive chat");
    snprintf(conf->version, sizeof(conf->version), "1.0.0");    /* 客户端自身版本号(留做统计用) */

    conf->log_level = LOG_LEVEL_TRACE;                      /* 日志级别 */
    snprintf(conf->log_path, sizeof(conf->log_path), "client.log");   /* 日志路径(路径+文件名) */
    snprintf(conf->httpsvr, sizeof(conf->httpsvr), "127.0.0.1:8000");      /* HTTP端(IP+端口) */

    conf->work_thd_num = 1;  /* 工作线程数 */
    conf->recv_buff_size = 1024 * 1024;  /* 接收缓存大小 */
    conf->sendq_len = 1024;  /* 发送队列配置 */
    conf->recvq_len = 1024;  /* 接收队列配置 */

    return;
}

/* > 应答消息PONG的处理 */
int sdk_cmd_pong_handler(int cmd, uint64_t from, char *data, size_t len, void *param)
{
    fprintf(stderr, "Call %s() cmd:0x%04X\n", __func__, cmd);
    return 0;
}

/* > 发送结果回调 */
int sdk_send_cb(uint16_t cmd, const void *orig, size_t size,
        char *ack, size_t ack_len, sdk_send_stat_e stat, void *param)
{
    switch (stat) {
        case SDK_STAT_IN_SENDQ: /* 发送队列中... */
            fprintf(stderr, "Call %s() cmd:0x%04X is in sendq.\n", __func__, cmd);
            break;
        case SDK_STAT_SENDING:  /* 正在发送... */
            fprintf(stderr, "Call %s() cmd:0x%04X is sending.\n", __func__, cmd);
            break;
        case SDK_STAT_SEND_SUCC:   /* 发送成功 */
            fprintf(stderr, "Call %s() cmd:0x%04X is send success.\n", __func__, cmd);
            break;
        case SDK_STAT_SEND_FAIL:  /* 发送失败 */
            fprintf(stderr, "Call %s() cmd:0x%04X is send fail.\n", __func__, cmd);
            break;
        case SDK_STAT_SEND_TIMEOUT:  /* 发送超时 */
            fprintf(stderr, "Call %s() cmd:0x%04X is send timeout.\n", __func__, cmd);
            break;
        case SDK_STAT_ACK_SUCC:     /* 应答成功 */
            fprintf(stderr, "Call %s() cmd:0x%04X is ack success.\n", __func__, cmd);
            break;
        case SDK_STAT_ACK_TIMEOUT:   /* 应答超时 */
            fprintf(stderr, "Call %s() cmd:0x%04X is ack timeout.\n", __func__, cmd);
            break;
        case SDK_STAT_UNKNOWN:      /* 未知状态 */
        default:
            fprintf(stderr, "Call %s() cmd:0x%04X is unknown.\n", __func__, cmd);
            break;
    }
    return 0;
}

/* 私聊消息 */
int chat_send(sdk_cntx_t *ctx)
{
    void *addr;
    size_t size;
    sdk_conf_t *conf = &ctx->conf;
    MesgChat chat = MESG_CHAT__INIT;

    /* > 设置ONLINE字段 */
    chat.suid = conf->uid;
    chat.duid = 18610485118;
    chat.level = 0;
    chat.time = (uint64_t)time(NULL);
    chat.text = "This is just a test!";

    /* > 申请内存空间 */
    size = mesg_chat__get_packed_size(&chat);

    addr = (void *)calloc(1, size);
    if (NULL == addr) {
        return SDK_ERR;
    }

    mesg_chat__pack(&chat, addr);

    /* > 发起CHAT请求 */
    sdk_async_send(ctx, CMD_CHAT, addr, size, 3, (sdk_send_cb_t)sdk_send_cb, NULL);

    free(addr);

    return 0;
}

/* 创建聊天室 */
int room_creat(sdk_cntx_t *ctx)
{
    void *addr;
    size_t size;
    sdk_conf_t *conf = &ctx->conf;
    MesgRoomCreat creat = MESG_ROOM_CREAT__INIT;

    /* > 设置ONLINE字段 */
    creat.uid = conf->uid;
    creat.name = "Qifeng";
    creat.desc = "Friend";

    /* > 申请内存空间 */
    size = mesg_room_creat__get_packed_size(&creat);

    addr = (void *)calloc(1, size);
    if (NULL == addr) {
        return SDK_ERR;
    }

    mesg_room_creat__pack(&creat, addr);

    /* > 发起CREAT请求 */
    sdk_async_send(ctx, CMD_ROOM_CREAT, addr, size, 3, (sdk_send_cb_t)sdk_send_cb, NULL);

    free(addr);

    return 0;
}

/* 加入指定聊天室 */
int room_join(sdk_cntx_t *ctx, uint64_t rid)
{
    void *addr;
    size_t size;
    sdk_conf_t *conf = &ctx->conf;
    MesgRoomJoin join = MESG_ROOM_JOIN__INIT;

    /* > 设置ONLINE字段 */
    join.uid = conf->uid;
    join.rid = rid;

    /* > 申请内存空间 */
    size = mesg_room_join__get_packed_size(&join);

    addr = (void *)calloc(1, size);
    if (NULL == addr) {
        return SDK_ERR;
    }

    mesg_room_join__pack(&join, addr);

    /* > 发起JOIN请求 */
    sdk_async_send(ctx, CMD_ROOM_JOIN, addr, size, 3, (sdk_send_cb_t)sdk_send_cb, NULL);

    free(addr);

    return 0;
}

/* 发送聊天室消息 */
int room_chat(sdk_cntx_t *ctx, uint64_t rid)
{
    void *addr;
    size_t size;
    sdk_conf_t *conf = &ctx->conf;
    MesgRoomChat chat = MESG_ROOM_CHAT__INIT;

    /* > 设置ONLINE字段 */
    chat.uid = conf->uid;
    chat.rid = rid;
    chat.gid = 0;
    chat.level = 0;
    chat.time = time(NULL);
    chat.text = "This is room chat";

    /* > 申请内存空间 */
    size = mesg_room_chat__get_packed_size(&chat);

    addr = (void *)calloc(1, size);
    if (NULL == addr) {
        return SDK_ERR;
    }

    mesg_room_chat__pack(&chat, addr);

    /* > 发起ROOM-CHAT请求 */
    sdk_async_send(ctx, CMD_ROOM_CHAT, addr, size, 3, (sdk_send_cb_t)sdk_send_cb, NULL);

    free(addr);

    return 0;
}

/* 聊天室KICK */
int room_kick(sdk_cntx_t *ctx, uint64_t rid)
{
    void *addr;
    size_t size;
    MesgRoomKick chat = MESG_ROOM_KICK__INIT;

    /* > 设置ONLINE字段 */
    chat.rid = rid;
    chat.uid = 18610485118;

    /* > 申请内存空间 */
    size = mesg_room_kick__get_packed_size(&chat);

    addr = (void *)calloc(1, size);
    if (NULL == addr) {
        return SDK_ERR;
    }

    mesg_room_kick__pack(&chat, addr);

    /* > 发起ROOM-KICK请求 */
    sdk_async_send(ctx, CMD_ROOM_KICK, addr, size, 3, (sdk_send_cb_t)sdk_send_cb, NULL);

    free(addr);

    return 0;
}

int main(int argc, char *argv[])
{
    sdk_conf_t conf;
    sdk_cntx_t *ctx;

    client_set_conf(&conf);

    ctx = sdk_init(&conf);
    if (NULL == ctx) {
        fprintf(stderr, "Initialize sdk failed!\n");
        return -1;
    }

    /* > 设置应答命令 */
    sdk_cmd_add(ctx, CMD_PING, CMD_PONG);
    sdk_cmd_add(ctx, CMD_CHAT, CMD_CHAT_ACK);
    sdk_cmd_add(ctx, CMD_ROOM_CREAT, CMD_ROOM_CREAT_ACK);
    sdk_cmd_add(ctx, CMD_ROOM_JOIN, CMD_ROOM_JOIN_ACK);
    sdk_cmd_add(ctx, CMD_ROOM_CHAT, CMD_ROOM_CHAT_ACK);
    sdk_cmd_add(ctx, CMD_ROOM_KICK, CMD_ROOM_KICK_ACK);
    sdk_cmd_add(ctx, CMD_ROOM_QUIT, CMD_ROOM_QUIT_ACK);

    sdk_register(ctx, CMD_PONG, (sdk_reg_cb_t)sdk_cmd_pong_handler, NULL);

    sdk_launch(ctx);

    Sleep(1);

    while (1) {
        //room_creat(ctx);
        Sleep(1);
        room_join(ctx, 1000000015);
        Sleep(1);
        chat_send(ctx);
        Sleep(1);
        room_chat(ctx, 1000000015);
        Sleep(1);
        room_kick(ctx, 1000000015);
        Sleep(1);
        sdk_async_send(ctx, CMD_PING, NULL, 0, 3, (sdk_send_cb_t)sdk_send_cb, NULL);
        Sleep(1);
    }

    return 0;
}

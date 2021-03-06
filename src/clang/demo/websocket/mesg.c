#include <time.h>
#include <stdio.h>
#include <stdlib.h>
#include <getopt.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>
#include <libwebsockets.h>

#include "sck.h"
#include "list.h"
#include "mesg.h"
#include "atomic.h"
#include "client.h"
#include "cmd_list.h"

#include "mesg.pb-c.h"

/* 发送ONLINE请求 */
int lws_mesg_online_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    list_opt_t lo;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgOnline online = MESG_ONLINE__INIT;

    /* 创建发送队列 */
    memset(&lo, 0, sizeof(lo));

    lo.pool = NULL;
    lo.alloc = mem_alloc;
    lo.dealloc = mem_dealloc;

    session->send_list = list_creat(&lo);
    if (NULL == session->send_list) {
        fprintf(stderr, "Create list failed!\n");
        return -1;
    }

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置ONLINE消息 */
    online.uid = AWS_UID;
    online.sid = AWS_SID;
    online.token = "ASUDbFI1CjsGYgEyAjUFZA40BGZbOAM3V2ABYAdpUDhUcQx+VjlZaFczCjQHNFYwCTMPYVEzVGxWZlIyAGxWdwE5A2FSawo0BmkBbwJtBTE=";
    online.app = "beehive-im";
    online.version = "v.0.1";
    online.has_terminal = true;
    online.terminal = 1;

    len = mesg_online__get_packed_size(&online);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_online__pack(&online, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_ONLINE);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    atomic64_xinc(&ctx->seq);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 处理ONLINE-ACK消息 */
int lws_mesg_online_ack_handler(lws_cntx_t *ctx, mesg_header_t *head, void *body)
{
    MesgOnlineAck *ack;

    /* > 提取有效信息 */
    ack = mesg_online_ack__unpack(NULL, head->length, (void *)(head + 1));
    if (NULL == ack) {
        fprintf(stderr, "Unpack online ack failed!\n");
        return -1;
    }

    fprintf(stderr, "Unpack online ack success!\n");
    fprintf(stderr, "uid:%lu sid:%lu seq:%lu app:%s version:%s code:%d errmsg:%s\n",
            ack->uid, ack->sid, ack->seq, ack->app, ack->version, ack->code, ack->errmsg);
    ctx->seq = ack->seq;
    atomic64_xinc(&ctx->seq);

    mesg_online_ack__free_unpacked(ack, NULL);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////

/* 发送PING请求 */
int lws_mesg_ping_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    mesg_header_t *head;
    lws_send_item_t *item;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    item->len = sizeof(mesg_header_t);

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);

    /* 设置通用头部 */
    head->type = htonl(CMD_PING);
    head->length = htonl(0);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 处理PONG消息 */
int lws_mesg_pong_handler(mesg_header_t *head, void *body)
{
    fprintf(stdout, "Call %s()!\n", __func__);
    return 0;
}

////////////////////////////////////////////////////////////////////////////////
/* 发送订阅请求 */
int lws_mesg_sub(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session, uint32_t cmd)
{
    size_t len;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgSub sub = MESG_SUB__INIT;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置SUB消息 */
    sub.cmd = cmd;

    len = mesg_sub__get_packed_size(&sub);

    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_sub__pack(&sub, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_SUB);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////

/* 发送ROOM-JOIN请求 */
int lws_mesg_room_join_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgRoomJoin join = MESG_ROOM_JOIN__INIT;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置ONLINE消息 */
    join.uid = AWS_UID;
    join.rid = 1000000015;

    len = mesg_room_join__get_packed_size(&join);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_room_join__pack(&join, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_ROOM_JOIN);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 处理ROOM-JOIN-ACK */
int lws_mesg_room_join_ack_handler(mesg_header_t *head, void *body)
{
    MesgRoomJoinAck *ack;

    /* > 提取有效信息 */
    ack = mesg_room_join_ack__unpack(NULL, head->length, (void *)(head + 1));
    if (NULL == ack) {
        fprintf(stderr, "Unpack room join ack failed!\n");
        return -1;
    }

    fprintf(stderr, "Unpack room join ack success!\n");
    fprintf(stderr, "uid:%lu rid:%lu gid:%d code:%d errmsg:%s\n",
            ack->uid, ack->rid, ack->gid, ack->code, ack->errmsg);

    mesg_room_join_ack__free_unpacked(ack, NULL);

    return 0;
}

/* 发送ROOM-QUIT请求 */
int lws_mesg_room_quit_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgRoomQuit quit = MESG_ROOM_QUIT__INIT;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置ONLINE消息 */
    quit.uid = AWS_UID;
    quit.rid = 1000000015;

    len = mesg_room_quit__get_packed_size(&quit);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_room_quit__pack(&quit, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_ROOM_QUIT);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 发送ROOM消息 */
int lws_mesg_room_chat_send_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgRoomChat chat = MESG_ROOM_CHAT__INIT;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置ROOM-CHAT消息 */
    chat.uid = AWS_UID;
    chat.rid = 1000000015;
    chat.gid = 0;
    chat.level = 0;
    chat.time = time(NULL);
    chat.text = "Hello world";

    len = mesg_room_chat__get_packed_size(&chat);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_room_chat__pack(&chat, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_ROOM_CHAT);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 发送CHAT消息 */
int lws_mesg_chat_send_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    uint8_t *body;
    mesg_header_t *head;
    lws_send_item_t *item;
    MesgChat chat = MESG_CHAT__INIT;

    fprintf(stderr, "Call %s()\n", __func__);

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置CHAT消息 */
    chat.suid = AWS_UID;
    chat.duid = 18600522324;
    chat.level = 0;
    chat.time = (uint64_t)time(NULL);
    chat.text = "This is just a test!";

    len = mesg_chat__get_packed_size(&chat);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_chat__pack(&chat, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_CHAT);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->seq = hton64(atomic64_xinc(&ctx->seq));

    if (list_rpush(session->send_list, item)) {
        fprintf(stderr, "Send chat failed!\n");
        free(item);
        return -1;
    }

    lws_callback_on_writable(lws, wsi);

    return 0;
}

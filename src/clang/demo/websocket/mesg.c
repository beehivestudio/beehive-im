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

    fprintf(stderr, "callback_im: LWS_CALLBACK_CLIENT_ESTABLISHED\n");

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    /* 设置ONLINE消息 */
    online.uid = 18600522324;
    online.sid = AWS_SID;
    online.token = "UnYGaVUyDTxdOVJhUWZXNwA6AmEBaAw/BjMCYFEzUzsBJAZ0AG9dbFA0CzRQaVE4XGNWOVI5BTVVYwZgVTlUdVJqBmRVbA0zXTJSPFE+V2M=";
    online.app = "beehive-im";
    online.version = "v.0.0.0.1";
    online.has_terminal = true;
    online.terminal = 1;

    len = mesg_online__get_packed_size(&online);
    item->len = sizeof(mesg_header_t) + (uint32_t)len;

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);
    body = (uint8_t *)(void *)(head + 1);

    mesg_online__pack(&online, body);

    /* 设置通用头部 */
    head->type = htonl(CMD_ONLINE);
    head->flag = htonl(1);
    head->length = htonl(len);
    head->sid = hton64(AWS_SID);
    head->chksum = htonl(MSG_CHKSUM_VAL);

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

/* 处理ONLINE-ACK消息 */
int lws_mesg_online_ack_handler(mesg_header_t *head, void *body)
{
    MesgOnlineAck *ack;

    /* > 提取有效信息 */
    ack = mesg_online_ack__unpack(NULL, head->length, (void *)(head + 1));
    if (NULL == ack) {
        fprintf(stderr, "Unpack online ack failed!\n");
        return -1;
    }

    fprintf(stderr, "Unpack online ack success!\n");
    fprintf(stderr, "uid:%lu sid:%lu app:%s version:%s code:%d errmsg:%s\n",
            ack->uid, ack->sid, ack->app, ack->version, ack->code, ack->errmsg);

    mesg_online_ack__free_unpacked(ack, NULL);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////

/* 发送PING请求 */
int lws_mesg_ping_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    list_opt_t lo;
    mesg_header_t *head;
    lws_send_item_t *item;

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

    fprintf(stderr, "callback_im: LWS_CALLBACK_CLIENT_ESTABLISHED\n");

    /* 创建发送单元 */
    item = (lws_send_item_t *)calloc(1, sizeof(lws_send_item_t));
    if (NULL == item) {
        fprintf(stderr, "errmsg:[%d] %s!\n", errno, strerror(errno));
        return -1;
    }

    head = (mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING);

    /* 设置通用头部 */
    head->type = htonl(CMD_PING);
    head->flag = htonl(1);
    head->length = htonl(0);
    head->sid = hton64(AWS_SID);
    head->chksum = htonl(MSG_CHKSUM_VAL);

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

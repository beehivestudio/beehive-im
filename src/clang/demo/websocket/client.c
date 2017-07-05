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

static lws_cntx_t g_wsc_cntx; /* 全局对象 */

/* 接收数据的处理 */
int lws_recv_handler(struct lws_context *lws,
        struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session, void *in)
{
    char *body;
    time_t ctm;
    struct tm lctm;
    mesg_header_t *head;

    head = (mesg_header_t *)in;
    body = (char *)(head + 1);

    MESG_HEAD_NTOH(head, head);

    fprintf(stderr, "Recv data. cmd:0x%04X sid:%lu cid:%lu nid:%d len:%d\n",
            head->type, head->sid, head->cid, head->nid, head->length);

    ctm = time(NULL);
    localtime_r(&ctm, &lctm);
    fprintf(stderr, "year:%u mon:%u day:%u hour:%u min:%u sec:%u\n",
            lctm.tm_year+1900, lctm.tm_mon+1, lctm.tm_mday,
            lctm.tm_hour, lctm.tm_min, lctm.tm_sec);

    switch (head->type) {
        case CMD_ONLINE_ACK:
            if (lws_mesg_online_ack_handler(ctx, head, body)) {
                return -1;
            }
            return lws_mesg_room_join_handler(lws, wsi, ctx, session);
        case CMD_PONG:
            return lws_mesg_pong_handler(head, body);
        case CMD_ROOM_JOIN_ACK:
            if (!lws_mesg_room_join_ack_handler(head, body)) {
                lws_mesg_sub(lws, wsi, ctx, session, CMD_ROOM_USR_NUM);
                return lws_mesg_room_chat_send_handler(lws, wsi, ctx, session);
            }
            return -1;
        case CMD_CHAT:
            return lws_mesg_chat_send_handler(lws, wsi, ctx, session);
        default:
            return 0;
    }

    return 0;
}

/* 发送数据的处理 */
int lws_send_handler(struct lws_context *lws,
        struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session, void *in)
{
    int n;
    mesg_header_t head;
    lws_send_item_t *item;
    static time_t ping_tm;
    time_t ctm = time(NULL);
    static int ping_times = 0;

    if (ctm - ping_tm > 2) {
        ping_tm = ctm;
        if (0 != ping_times) {
            lws_mesg_ping_handler(lws, wsi, ctx, session);
            lws_mesg_sub(lws, wsi, ctx, session, CMD_ROOM_USR_NUM);
            lws_mesg_chat_send_handler(lws, wsi, ctx, session);
            lws_mesg_room_chat_send_handler(lws, wsi, ctx, session);
            //lws_mesg_room_quit_handler(lws, wsi, ctx, session);
        }
        ++ping_times;
    }

    fprintf(stderr, "Call %s()\n", __func__);

    while (1) {
        item = (lws_send_item_t *)list_lpop(session->send_list);
        if (NULL == item) {
            break;
        }

        MESG_HEAD_NTOH((mesg_header_t *)(item->addr + LWS_SEND_BUFFER_PRE_PADDING), &head);
        fprintf(stderr, "Send data! cmd:0x%04X\n", head.type);


        n = lws_write(wsi, (unsigned char *)item->addr+LWS_SEND_BUFFER_PRE_PADDING,
                item->len, LWS_WRITE_BINARY);
        if (n < 0) {
            fprintf(stderr, "Send data failed!\n");
            return -1;
        } else if (n < (int)item->len) {
            lwsl_err("Partial write LWS_CALLBACK_CLIENT_WRITEABLE\n");
            return -1;
        }
        free(item);
    }
    return 0;
}

/* dumb_im protocol */
static int callback_im(
        struct lws_context *lws,
        struct lws *wsi,
        enum lws_callback_reasons reason,
        void *user, void *in, size_t len)
{
    lws_cntx_t *ctx = &g_wsc_cntx;
    lws_session_data_t *session = (lws_session_data_t *)user;

    switch (reason) {
        case LWS_CALLBACK_CLIENT_ESTABLISHED:
            return lws_mesg_online_handler(lws, wsi, ctx, session);
        case LWS_CALLBACK_CLIENT_CONFIRM_EXTENSION_SUPPORTED:
            break;
        case LWS_CALLBACK_CLIENT_RECEIVE:
            return lws_recv_handler(lws, wsi, ctx, session, in);
        case LWS_CALLBACK_CLIENT_WRITEABLE:
            return lws_send_handler(lws, wsi, ctx, session, in);
        case LWS_CALLBACK_WSI_DESTROY:
            fprintf(stderr, "callback_im: LWS_CALLBACK_WSI_DESTROY\n");
            ctx->is_closed = true;
        case LWS_CALLBACK_CLIENT_CONNECTION_ERROR:
            fprintf(stderr, "errmsg:%s\n", (char *)in);
            return -1;
        default:
            break;
    }

    return 0;
}

/* list of supported g_protocols and callbacks */
static struct lws_protocols g_protocols[] = {
    {
        "im",
        callback_im,
        sizeof(lws_session_data_t),
        0,
    },
    { NULL, NULL, 0, 0 } /* end */
};

void sighandler(int sig)
{
    lws_cntx_t *ctx = &g_wsc_cntx;

    ctx->is_force_exit = true;
}

/* Get input options for websocket client */
static int wsc_getopt(int argc, char *argv[], wsc_opt_t *opt)
{
    int n = 0;
    const struct option wsc_options[] = {
        {"host",        required_argument, NULL,   'h'},
        {"port",        required_argument,  NULL,   'p'},
        {"debug",       required_argument,  NULL,   'd'},
        {"ssl",         no_argument,        NULL,   's'},
        {"version",     required_argument,  NULL,   'v'},
        {"undeflated",  no_argument,        NULL,   'u'},
        {"num",         no_argument,        NULL,   'n'},
        {"longlived",   no_argument,        NULL,   'l'},
        {"rid",         required_argument,  NULL,   'r'},
        {NULL,          0,                  0,      0}
    };

    memset(opt, 0, sizeof(wsc_opt_t));

    opt->num = 1;
    opt->ietf_version = 13;
    lws_set_log_level(0xFFFFFFFF, NULL);

    while (n >= 0) {
        n = getopt_long(argc, argv, "h:n:uv:i:r:sp:d:l", wsc_options, NULL);
        if (n < 0) {
            continue;
        }
        switch (n) {
            case 'd':
                lws_set_log_level(atoi(optarg), NULL);
                break;
            case 's':
                opt->use_ssl = 2; /* 2 = allow selfsigned */
                break;
            case 'p':
                opt->port = atoi(optarg);
                break;
            case 'r':
                opt->rid = atoi(optarg);
                break;
            case 'l':
                opt->longlived = 1;
                break;
            case 'v':
                opt->ietf_version = atoi(optarg);
                break;
            case 'u':
                opt->deny_deflate = 1;
                break;
            case 'n':
                opt->num = atoi(optarg);
                break;
            case 'h':
                opt->ipaddr = optarg;
                break;
            default:
                return -1;
        }
    }

    if (0 == opt->port
        || NULL == opt->ipaddr) {
        return -1;
    }

    return 0;
}

int main(int argc, char **argv)
{
    struct lws *wsi;
    int n = 0, ret = 0;
    struct lws_context *lws;
    lws_cntx_t *ctx = &g_wsc_cntx;
    wsc_opt_t *opt = &ctx->opt;
    struct lws_context_creation_info info;

    fprintf(stderr, "libwebsockets test client\n"
            "(C) Copyright 2010-2015 Qifeng.zou <Qifeng.zou.job@hotmail.com> "
            "licensed under LGPL2.1\n");

    if (wsc_getopt(argc, argv, opt)) {
        goto usage;
    }

    //signal(SIGINT, sighandler);

    /*
     * create the websockets lws.  This tracks open connections and
     * knows how to route any traffic and which protocol version to use,
     * and if each connection is client or server side.
     *
     * For lws client-only demo, we tell it to not listen on any port.
     */
    memset(&info, 0, sizeof info);

    info.port = CONTEXT_PORT_NO_LISTEN;
    info.protocols = g_protocols;
#ifndef LWS_NO_EXTENSIONS
    info.extensions = lws_get_internal_extensions();
#endif
    info.gid = -1;
    info.uid = -1;

    lws = lws_create_context(&info);
    if (NULL == lws) {
        fprintf(stderr, "Creating libwebsocket lws failed\n");
        return -1;
    }

    /* create a client websocket using im protocol */
    for (n=0; n<opt->num; n++) {
        wsi = lws_client_connect(lws, opt->ipaddr, opt->port, 0,
                "/im", opt->ipaddr, opt->ipaddr,
                g_protocols[PROTOCOL_IM].name, opt->ietf_version);
        if (NULL == wsi) {
            fprintf(stderr, "libwebsocket connect failed\n");
            ret = 1;
            goto bail;
        }
    }

    fprintf(stderr, "Waiting for connect...\n");

    /*
     * sit there servicing the websocket lws to handle incoming
     * packets, and drawing random circles on the mirror protocol websocket
     * nothing happens until the client websocket connection is
     * asynchronously established
     */

    n = 0;
    while (n >= 0 && !ctx->is_force_exit && !ctx->is_closed) {
        n = lws_service(lws, 10);
    }

bail:
    fprintf(stderr, "Exiting\n");

    lws_context_destroy(lws);

    return ret;

usage:
    fprintf(stderr, "Usage: libwebsockets-test-client "
            "<server ip> [--port=<p>] "
            "[--ssl] [-k] [-v <ver>] "
            "[-d <log bitfield>] [-l]\n");
    return 1;
}

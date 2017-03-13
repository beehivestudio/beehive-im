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
#include "cmd_list.h"

#include "mesg.pb-c.h"

#define AWS_SEND_BUF_LEN   (2048)  // 发送缓存的长度

/* 输入选项 */
typedef struct
{
    int num;                        /* 最大并发数 */
    uint32_t rid;                   /* 聊天室ID */
    int use_ssl;
    int port;
    int longlived;
    int ietf_version;
    int deny_deflate;
    char *ipaddr;
} wsc_opt_t;

/* 发送项 */
typedef struct
{
    size_t len;                     // 长度
    char addr[AWS_SEND_BUF_LEN];    // 发送内容
} lws_send_item_t;

typedef struct
{
    list_t *send_list;
} lws_session_data_t;

/* 全局对象 */
typedef struct
{
    wsc_opt_t opt;

    struct lws_context *lws;

    // 其他标志
    bool is_closed;
    bool is_force_exit;
} lws_cntx_t;

static lws_cntx_t g_wsc_cntx; /* 全局对象 */

/* 协议类型 */
typedef enum
{
    PROTOCOL_IM,

    /* always last */
    PROTOCOL_DEMO_COUNT
} demo_protocols;

/* 发送ONLINE请求 */
int lws_send_online_handler(struct lws_context *lws,
        struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session)
{
    size_t len;
    uint64_t sid;
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
    sid = 123456;
    online.uid = 18600522324;
    online.sid = sid;
    online.token = "This is a token!";
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
    head->sid = hton64(sid);
    head->chksum = htonl(MSG_CHKSUM_VAL);

    list_rpush(session->send_list, item);

    lws_callback_on_writable(lws, wsi);

    return 0;
}

int lws_online_ack_handler(mesg_header_t *head, void *body)
{
    fprintf(stderr, "Call %s()\n", __func__);
    return 0;
}

/* 接收数据的处理 */
int lws_recv_handler(char *data)
{
    char *body;
    time_t ctm;
    struct tm lctm;
    mesg_header_t *head;

    head = (mesg_header_t *)data;
    body = (char *)(head + 1);

    MESG_HEAD_NTOH(head, head);


    fprintf(stderr, "Recv data. cmd:0x%04X len:%d flag:%d chksum:0x%08X body:%p\n",
            head->type, head->length, head->flag, head->chksum, body);

    ctm = time(NULL);
    localtime_r(&ctm, &lctm);
    fprintf(stderr, "year:%u mon:%u day:%u hour:%u min:%u sec:%u\n",
            lctm.tm_year+1900, lctm.tm_mon+1, lctm.tm_mday,
            lctm.tm_hour, lctm.tm_min, lctm.tm_sec);

    switch (head->type) {
        case CMD_ONLINE_ACK:
            return lws_online_ack_handler(head, body);
        default:
            return 0;
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
    int n;
    lws_send_item_t *item;
    lws_cntx_t *ctx = &g_wsc_cntx;
    lws_session_data_t *session = (lws_session_data_t *)user;

    switch (reason) {
        case LWS_CALLBACK_CLIENT_ESTABLISHED:
            return lws_send_online_handler(lws, wsi, ctx, session);
        case LWS_CALLBACK_CLIENT_CONFIRM_EXTENSION_SUPPORTED:
            break;
        case LWS_CALLBACK_CLIENT_RECEIVE:
            return lws_recv_handler(in);
        case LWS_CALLBACK_CLIENT_WRITEABLE:
            fprintf(stderr, "callback_im: LWS_CALLBACK_CLIENT_WRITEABLE\n");
            while (1) {
                item = (lws_send_item_t *)list_lpop(session->send_list);
                if (NULL == item) {
                    break;
                }

                n = lws_write(wsi, (unsigned char *)item->addr+LWS_SEND_BUFFER_PRE_PADDING,
                        item->len, LWS_WRITE_BINARY);
                if (n < 0) {
                    return -1;
                }
                else if (n < (int)item->len) {
                    lwsl_err("Partial write LWS_CALLBACK_CLIENT_WRITEABLE\n");
                    return -1;
                }
                free(item);
            }
            break;
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
        || NULL == opt->ipaddr)
    {
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

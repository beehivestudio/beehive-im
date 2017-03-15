#if !defined(__CLIENT_H__)
#define __CLIENT_H__

#define AWS_SID             (5) // 会话ID
#define AWS_SEND_BUF_LEN    (2048)  // 发送缓存的长度


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

/* 协议类型 */
typedef enum
{
    PROTOCOL_IM,

    /* always last */
    PROTOCOL_DEMO_COUNT
} demo_protocols;



int lws_mesg_online_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session);
int lws_mesg_online_ack_handler(mesg_header_t *head, void *body);

int lws_mesg_ping_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session);
int lws_mesg_pong_handler(mesg_header_t *head, void *body);

int lws_mesg_room_join_handler(struct lws_context *lws, struct lws *wsi, lws_cntx_t *ctx, lws_session_data_t *session);
int lws_mesg_room_join_ack_handler(mesg_header_t *head, void *body);

#endif /*__CLIENT_H__*/

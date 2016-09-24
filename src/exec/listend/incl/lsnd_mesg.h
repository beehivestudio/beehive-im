#if !defined(__LSND_MESG_H__)
#define __LSND_MESG_H__

/* 会话数据由哪个表维护 */
typedef enum
{
    CHAT_EXTRA_LOC_UNKNOWN          /* 未知 */
    , CHAT_EXTRA_LOC_CID_TAB        /* CID表 */
    , CHAT_EXTRA_LOC_SID_TAB        /* SID表 */
    , CHAT_EXTRA_LOC_KICK_TAB       /* KICK表 */
} chat_extra_loc_tab_e;

/* 连接状态 */
typedef enum
{
    CHAT_CONN_STAT_UNKNOWN          /* 未知 */
    , CHAT_CONN_STAT_ESTABLISH      /* 创建 */
    , CHAT_CONN_STAT_ONLINE         /* 上线 */
    , CHAT_CONN_STAT_KICK           /* 被踢 */
    , CHAT_CONN_STAT_OFFLINE        /* 下线 */
    , CHAT_CONN_STAT_CLOSEING       /* 正在关闭... */
    , CHAT_CONN_STAT_CLOSED         /* 已关闭 */
} chat_conn_stat_e;

/* 会话扩展数据 */
typedef struct
{
    uint64_t sid;                   /* 会话ID */
    uint64_t cid;                   /* 连接ID */
    chat_conn_stat_e stat;          /* 连接状态 */
    chat_extra_loc_tab_e  loc;      /* 用户数据由哪个表维护 */

    time_t create_time;             /* 创建时间 */
    time_t recv_time;               /* 最近接收数据时间 */
    time_t send_time;               /* 最近发送数据时间 */
    time_t keepalive_time;          /* 保活时间 */

    socket_t *sck;                  /* TCP连接实例 */
} chat_conn_extra_t;

int chat_callback(acc_cntx_t *ctx, socket_t *sck, int reason, void *user, void *in, int len, void *args);

int lsnd_mesg_def_hdl(int type, void *data, int length, void *args);

int chat_online_req_hdl(int type, void *data, int length, void *args);
int chat_online_ack_hdl(int type, int orig, char *data, size_t len, void *args);

int chat_join_req_hdl(int type, void *data, int length, void *args);
int chat_join_ack_hdl(int type, int orig, char *data, size_t len, void *args);

#endif /*__LSND_MESG_H__*/

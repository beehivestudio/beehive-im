#if !defined(__LSND_MESG_H__)
#define __LSND_MESG_H__

/* 会话数据由哪个表维护 */
typedef enum
{
    LSND_DATA_LOC_UNKNOWN           /* 未知 */
    , LSND_DATA_LOC_CID_TAB         /* CID表 */
    , LSND_DATA_LOC_SID_TAB         /* SID表 */
    , LSND_DATA_LOC_KICK_TAB        /* KICK表 */
} lsnd_user_loc_tab_e;

/* 连接状态 */
typedef enum
{
    LSND_CONN_STAT_UNKNOWN          /* 未知 */
    , LSND_CONN_STAT_ESTABLIST      /* 创建 */
    , LSND_CONN_STAT_LOGIN          /* 登录 */
    , LSND_CONN_STAT_KICK           /* 被踢 */
    , LSND_CONN_STAT_READY_CLOSE    /* 待关闭 */
    , LSND_CONN_STAT_CLOSED         /* 已关闭 */
} lsnd_conn_stat_e;

/* 会话数据 */
typedef struct
{
    uint64_t sid;                   /* 会话ID */
    uint64_t cid;                   /* 连接ID */
    lsnd_conn_stat_e stat;          /* 连接状态 */
    lsnd_user_loc_tab_e  loc;       /* 用户数据由哪个表维护 */

    time_t create_time;             /* 创建时间 */

    socket_t *tsi;                  /* TCP连接实例 */
} lsnd_conn_user_data_t;

int lsnd_acc_callback(acc_cntx_t *ctx, socket_t *sck, int reason, void *user, void *in, int len, void *args);

int lsnd_mesg_def_hdl(unsigned int type, void *data, int length, void *args);

int lsnd_join_req_hdl(unsigned int type, void *data, int length, void *args);
int lsnd_search_rsp_hdl(int type, int orig, char *data, size_t len, void *args);

int lsnd_insert_word_req_hdl(unsigned int type, void *data, int length, void *args);
int lsnd_insert_word_rsp_hdl(int type, int orig, char *data, size_t len, void *args);

#endif /*__LSND_MESG_H__*/

#if !defined(__LSND_MESG_H__)
#define __LSND_MESG_H__

/* 会话数据 */
typedef struct
{
    uint64_t sid;
} lsnd_session_data_t;

int lsnd_acc_callback(acc_cntx_t *ctx, socket_t *sck, int reason, void *user, void *in, int len);

int lsnd_search_req_hdl(unsigned int type, void *data, int length, void *args);
int lsnd_search_rsp_hdl(int type, int orig, char *data, size_t len, void *args);

int lsnd_insert_word_req_hdl(unsigned int type, void *data, int length, void *args);
int lsnd_insert_word_rsp_hdl(int type, int orig, char *data, size_t len, void *args);

#endif /*__LSND_MESG_H__*/

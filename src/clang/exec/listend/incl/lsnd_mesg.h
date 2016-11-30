#if !defined(__LSND_MESG_H__)
#define __LSND_MESG_H__

#include "access.h"
#include "hash_tab.h"

int chat_callback(acc_cntx_t *ctx, socket_t *sck, int reason, void *user, void *in, int len, void *args);

int lsnd_mesg_def_hdl(int type, void *data, int len, void *args);

int chat_mesg_online_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args);
int chat_mesg_online_ack_hdl(int type, int orig, char *data, size_t len, void *args);

int chat_mesg_offline_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args);
int chat_mesg_offline_ack_hdl(int type, int orig, char *data, size_t len, void *args);

int chat_mesg_join_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args);
int chat_mesg_join_ack_hdl(int type, int orig, char *data, size_t len, void *args);

int chat_mesg_unjoin_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args);
int chat_mesg_ping_req_hdl(chat_conn_extra_t *conn, int type, void *data, int len, void *args);

int chat_mesg_room_mesg_hdl(chat_conn_extra_t *conn, int type, int orig, void *data, size_t len, void *args);

#endif /*__LSND_MESG_H__*/

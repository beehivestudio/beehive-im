#if !defined(__LSND_MESG_H__)
#define __LSND_MESG_H__

#include "access.h"
#include "hash_tab.h"

int lsnd_callback(acc_cntx_t *ctx, socket_t *sck, int reason, void *user, void *in, int len, void *args);

int lsnd_mesg_online_req_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args);
int lsnd_mesg_online_ack_handler(int type, int orig, char *data, size_t len, void *args);

int lsnd_mesg_offline_req_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args);
int lsnd_mesg_offline_ack_handler(int type, int orig, char *data, size_t len, void *args);

int lsnd_mesg_room_join_req_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args);
int lsnd_mesg_room_join_ack_handler(int type, int orig, char *data, size_t len, void *args);

int lsnd_mesg_unjoin_req_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args);
int lsnd_mesg_ping_req_handler(lsnd_conn_extra_t *conn, int type, void *data, int len, void *args);

int lsnd_mesg_room_mesg_handler(int type, int orig, void *data, size_t len, void *args);
int lsnd_mesg_kick_handler(int type, int orig, void *data, size_t len, void *args);

#endif /*__LSND_MESG_H__*/

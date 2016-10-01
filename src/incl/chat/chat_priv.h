#if !defined(__CHAT_PRIV_H__)
#define __CHAT_PRIV_H__

#include "lock.h"
#include "chat.h"

int chat_room_add_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid);
int chat_room_del_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid);

int chat_group_trav(chat_tab_t *chat, chat_room_t *room, uint16_t gid, trav_cb_t proc, void *args);

int chat_del_room(chat_tab_t *chat, uint64_t rid);
int chat_del_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp);
int chat_session_tab_add(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid);

#endif /*__CHAT_PRIV_H__*/

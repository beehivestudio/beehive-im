#if !defined(__CHAT_PRIV_H__)
#define __CHAT_PRIV_H__

#include "lock.h"
#include "chat.h"

int _chat_room_add_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid);
int _chat_room_del_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid);
int _chat_room_trav_del_session(chat_room_item_t *room, chat_session_trav_room_t *param);

int chat_group_trav(chat_tab_t *chat, chat_room_t *room, uint16_t gid, trav_cb_t proc, void *args);

int chat_del_room(chat_tab_t *chat, uint64_t rid);
int chat_del_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp);
int chat_session_tab_add(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid);

#endif /*__CHAT_PRIV_H__*/

#if !defined(__CHAT_PRIV_H__)
#define __CHAT_PRIV_H__

#include "chat.h"

chat_room_t *chat_add_room(chat_tab_t *chat, uint64_t rid);
int chat_del_room(chat_tab_t *chat, chat_room_t *room);
chat_group_t *chat_group_alloc(chat_tab_t *chat, chat_room_t *room, uint16_t gid);
int chat_del_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp);
int chat_session_ref(chat_tab_t *chat, chat_session_t *ssn);
chat_group_t *chat_group_find_by_gid(chat_tab_t *chat, chat_room_t *room, uint16_t gid);
chat_session_t *chat_group_add_session(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp, uint64_t sid);
int chat_group_del_session(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp, uint64_t sid);
int chat_room_trav_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp, trav_cb_t proc, void *args);
int chat_room_trav_all_group(chat_tab_t *chat, chat_room_t *room, trav_cb_t proc, void *args);

#endif /*__CHAT_PRIV_H__*/

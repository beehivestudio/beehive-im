#if !defined(__CHAT_H__)
#define __CHAT_H__

#include "log.h"
#include "comm.h"
#include "list.h"
#include "rb_tree.h"
#include "avl_tree.h"
#include "hash_tab.h"

/* 订阅项 */
typedef struct
{
    uint16_t cmd;               /* 命令类型 */
} chat_sub_item_t;

typedef struct
{
    uint64_t rid;               /* 聊天室ID */
    uint32_t gid;               /* 分组ID */
} chat_room_item_t;

/* 聊天室会话信息 */
typedef struct
{
    uint64_t sid;               /* 会话SID(SID+CID主键) */
    uint64_t cid;               /* 连接CID(SID+CID主键) */

    hash_tab_t *room;           /* 聊天室列表 */
    hash_tab_t *sub;            /* 订阅消息列表 */
} chat_session_t;

/* 聊天室分组信息 */
typedef struct
{
    uint16_t gid;               /* 聊天室分组ID */

    uint64_t sid_num;           /* SID总数 */
    hash_tab_t *sid_set;        /* 聊天室分组中SID列表
                                   (以SID+CID为主键, 存储的是chat_sid2cid_item_t值) */
} chat_group_t;

/* 聊天室信息 */
typedef struct
{
    uint64_t rid;               /* ROOMID(主键) */

    uint64_t sid_num;           /* SID总数 */
    uint32_t grp_num;           /* 分组总数 */

    time_t create_tm;           /* 创建时间 */

    hash_tab_t *groups;         /* 聊天室分组管理表(以gid为组建 存储chat_group_t数据) */
} chat_room_t;

/* SID->CID映射 */
typedef struct {
    uint64_t sid;               /* 会话SID */
    uint64_t cid;               /* 连接CID */
} chat_sid2cid_item_t;

/* 全局信息 */
typedef struct
{
    log_cycle_t *log;           /* 日志对象 */

    hash_tab_t *rooms;          /* 聊天室列表(注: ROOMID为主键 存储数据chat_room_t) */
    hash_tab_t *sessions;       /* SESSION信息(注: (SID+CID)为主键存储数据chat_session_t)
                                   注意: 此处的存储数据对象在group->sid_list被引用,
                                   释放时千万不能释放多次 */
    hash_tab_t *sid2cids;       /* SID->CID信息 */
} chat_tab_t;

/* 遍历会话中聊天室列表的参数 */
typedef struct
{
    chat_tab_t *chat;
    chat_session_t *session;
} chat_session_trav_room_t;

chat_tab_t *chat_tab_init(int len, log_cycle_t *log); // OK

uint32_t chat_room_add_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid); // OK
uint32_t chat_room_del_session(chat_tab_t *chat, uint64_t rid, uint64_t sid, uint64_t cid);
int chat_del_session(chat_tab_t *chat, uint64_t sid, uint64_t cid); // OK

uint64_t chat_get_cid_by_sid(chat_tab_t *chat, uint64_t sid); // OK
int chat_set_sid_to_cid(chat_tab_t *chat, uint64_t sid, uint64_t cid); // OK
int chat_del_sid_to_cid(chat_tab_t *chat, uint64_t sid, uint64_t cid); // OK

int chat_add_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd); // OK
int chat_del_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd); // OK
bool chat_has_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd); // OK

int chat_room_trav_session(chat_tab_t *chat, uint64_t rid, uint16_t gid, trav_cb_t proc, void *args); // OK
int chat_room_trav(chat_tab_t *chat, trav_cb_t proc, void *args);
int chat_clean_hdl(chat_tab_t *chat);

#endif /*__CHAT_H__*/

#if !defined(__CHAT_H__)
#define __CHAT_H__

#include "log.h"
#include "comm.h"
#include "list.h"
#include "rb_tree.h"
#include "avl_tree.h"

#define CHAT_USR_DEF_NUM        (1250)              /* 聊天室用户默认数 */
#define CHAT_TMOUT_MAX_SEC      (7200)              /* 聊天室最大超时时间 */
#define CHAT_TMOUT_DELAY_SEC    (300)               /* 当超过最大超时时间后,
                                                       这个时间内仍然有消息, 则以此时间递延 */
/* 订阅项 */
typedef struct
{
    uint16_t cmd;               /* 命令类型 */
} chat_sub_item_t;

/* 聊天室会话信息 */
typedef struct
{
    uint64_t sid;               /* 会话ID(主键) */

    uint64_t rid;               /* 会话所属ROOM ID */
    uint16_t gid;               /* 会话所属聊天室分组ID */

    uint32_t flags;             /* 属性标签(暂时未使用) */
    avl_tree_t *sub;            /* 订阅消息列表 */
} chat_session_t;

/* 聊天室分组信息 */
typedef struct
{
    uint16_t gid;               /* 聊天室分组ID */

    uint64_t sid_num;           /* SID总数 */
    rbt_tree_t *sid_list;       /* 聊天室分组中SID列表
                                   (以SID为主键, 存储的也是SID值) */
} chat_group_t;

/* 聊天室信息 */
typedef struct
{
    uint64_t rid;               /* ROOMID(主键) */

    uint64_t sid_num;           /* SID总数 */
    uint32_t grp_num;           /* 分组总数 */
    uint16_t grp_idx;           /* 下一个可分配的组ID(只增不减) */

    time_t create_tm;           /* 创建时间 */
    time_t update_tm;           /* 更新时间 */
    time_t timeout_tm;          /* 过期时间 */
    time_t delay_tm;            /* 递延时间 */

    avl_tree_t *group_tab;      /* 聊天室分组管理表(以gid为组建 存储chat_group_t数据) */
} chat_room_t;

/* 全局信息 */
typedef struct
{
    log_cycle_t *log;           /* 日志对象 */
    uint64_t grp_usr_max_num;   /* 分组中用户最大数 */

    avl_tree_t *room_tab;       /* 聊天室列表(注: ROOMID为主键 存储数据chat_room_t) */
    rbt_tree_t *session_tab;    /* SESSION信息(注: SID为主键存储数据chat_session_t)
                                   另外: 此处的存储数据对象在group->sid_list被引用,
                                   释放时千万不能释放多次 */
} chat_tab_t;

chat_tab_t *chat_tab_init(int grp_usr_num, log_cycle_t *log);
int chat_tab_set_grp_usr_num(chat_tab_t *chat, uint64_t max);

int chat_add_session(chat_tab_t *chat, uint64_t rid, uint64_t sid, uint32_t gid);
int chat_del_session(chat_tab_t *chat, uint64_t sid);

int chat_add_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd);
static bool _chat_has_sub(chat_session_t *ssn, uint16_t cmd);
int chat_del_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd);
bool chat_has_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd);

int chat_room_trav(chat_tab_t *chat, uint64_t rid, uint16_t gid, trav_cb_t proc, void *args);
static inline void chat_set_usr_max_num(chat_tab_t *chat, uint32_t max)
{
    chat->grp_usr_max_num = (0 == max)? CHAT_USR_DEF_NUM : max;
}
#endif /*__CHAT_H__*/

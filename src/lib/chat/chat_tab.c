#include "log.h"
#include "redo.h"
#include "chat.h"
#include "atomic.h"
#include "chat_priv.h"

/* 静态函数 */
static void chat_group_del_item(chat_group_t *grp);
static void chat_group_loop_del_item(void *pool, chat_group_t *grp);
static int chat_room_del_all_group(chat_tab_t *chat, chat_room_t *room);

/* 聊天室ID哈希回调 */
static uint64_t chat_room_hash_cb(chat_room_t *r)
{
    return r->rid;
}

/* 聊天室ID比较回调 */
static int chat_room_cmp_cb(chat_room_t *r1, chat_room_t *r2)
{
    return (int)(r1->rid - r2->rid);
}

/* 会话ID哈希回调 */
static uint64_t chat_session_hash_cb(chat_session_t *s)
{
    return s->sid;
}

/* 会话ID比较回调 */
static int chat_session_cmp_cb(chat_session_t *s1, chat_session_t *s2)
{
    return (int)(s1->sid - s2->sid);
}

/* 分组ID哈希回调 */
static uint64_t chat_group_hash_cb(chat_group_t *g)
{
    return g->gid;
}

/* 分组ID比较回调 */
static int chat_group_cmp_cb(chat_group_t *g1, chat_group_t *g2)
{
    return (g1->gid - g2->gid);
}

/******************************************************************************
 **函数名称: chat_tab_init
 **功    能: 初始化上下文
 **输入参数:
 **     len: 槽的长度
 **     log: 日志对象
 **输出参数: NONE
 **返    回: CHAT对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 10:38:44 #
 ******************************************************************************/
chat_tab_t *chat_tab_init(int len, log_cycle_t *log)
{
    chat_tab_t *chat;

    /* > 创建全局对象 */
    chat = (chat_tab_t *)calloc(1, sizeof(chat_tab_t));
    if (NULL == chat) {
        return NULL;
    }

    chat->log = log;

    do {
        /* > 初始化聊天室表 */
        chat->room_tab = hash_tab_creat(len,
                (hash_cb_t)chat_room_hash_cb,
                (cmp_cb_t)chat_room_cmp_cb, NULL);
        if (NULL == chat->room_tab) {
            break;
        }

        /* > 初始化SESSION表 */
        chat->session_tab = hash_tab_creat(len,
                (hash_cb_t)chat_session_hash_cb,
                (cmp_cb_t)chat_session_cmp_cb, NULL);
        if (NULL == chat->session_tab) {
            break;
        }
        return chat;
    } while(0);

    /* > 释放内存 */
    hash_tab_destroy(chat->room_tab, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
    hash_tab_destroy(chat->session_tab, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
    FREE(chat);

    return NULL;
}

/******************************************************************************
 **函数名称: chat_add_room
 **功    能: 添加一个聊天室
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 21:27:00 #
 ******************************************************************************/
int chat_add_room(chat_tab_t *chat, uint64_t rid)
{
    chat_room_t *room;

    /* > 创建聊天室对象 */
    room = (chat_room_t *)calloc(1, sizeof(chat_room_t));
    if (NULL == room) {
        return -1;
    }

    room->rid = rid;
    room->create_tm = time(NULL);

    /* > 创建分组列表 */
    room->group_tab = hash_tab_creat(99,
            (hash_cb_t)chat_group_hash_cb, (cmp_cb_t)chat_group_cmp_cb, NULL);
    if (NULL == room->group_tab) {
        free(room);
        return -1;
    }

    /* > 将ROOM放入管理表 */
    if (hash_tab_insert(chat->room_tab, (void *)room, WRLOCK)) {
        hash_tab_destroy(room->group_tab, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
        free(room);
        return -1;
    }

    return 0;
}

/******************************************************************************
 **函数名称: chat_del_room_by_rid
 **功    能: 删除指定聊天室
 **输入参数: 
 **     chat: 全局对象
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static int chat_del_room_by_rid(chat_tab_t *chat, uint64_t rid)
{
    chat_room_t *room, key;

    /* > 查找聊天室对象 */
    key.rid = rid;

    room = hash_tab_query(chat->room_tab, (void *)&key, RDLOCK);
    if (NULL == room) {
        return 0; /* 无此聊天室 */
    }
    else if (0 != room->sid_num) {
        hash_tab_unlock(chat->room_tab, (void *)&key, RDLOCK);
        return 0; /* 聊天室还有人, 不能删除 */
    }

    room = hash_tab_delete(chat->room_tab, (void *)&key, WRLOCK);

    /* > 删除分组信息 */
    chat_room_del_all_group(chat, room);
    FREE(room);
    return 0;
}

/******************************************************************************
 **函数名称: chat_del_room
 **功    能: 删除指定聊天室
 **输入参数: 
 **     chat: 全局对象
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.08 14:26:46 #
 ******************************************************************************/
int chat_del_room(chat_tab_t *chat, uint64_t rid)
{
    chat_room_t key, *room;

    key.rid = rid;

    room = hash_tab_query(chat->room_tab, (void *)&key, WRLOCK);
    if (NULL == room) {
        return 0;
    }
    else if ((0 != room->sid_num) || (0 != room->grp_num)) {
        hash_tab_unlock(chat->room_tab, (void *)&key, WRLOCK);
        log_error(chat->log, "Delete room failed! rid:%lu sid num:%d gid num:",
                room->rid, room->sid_num, room->grp_num);
        return 0; /* 聊天室还有人, 不能删除 */
    }

    log_info(chat->log, "Delete room [%u]", room->rid);

    hash_tab_delete(chat->room_tab, (void *)&key, NONLOCK);

    hash_tab_unlock(chat->room_tab, (void *)&key, WRLOCK);

    /* > 清理分组信息 */
    chat_room_del_all_group(chat, room);

    FREE(room);

    return 0;
}

/******************************************************************************
 **函数名称: chat_del_group
 **功    能: 删除指定分组
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     grp: 聊天分组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.21 20:28:08 #
 ******************************************************************************/
int chat_del_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp)
{
    chat_group_t *item, key;

    if (0 != grp->sid_num) {
        return 0; // 还存在用户, 无法删除
    }
    
    log_info(chat->log, "Delete gid:%u", grp->gid);

    key.gid = grp->gid;

    item = hash_tab_delete(room->group_tab, &key, WRLOCK);
    assert(grp == item);
    chat_group_del_item(grp);

    --room->grp_num; /* 分组数减1 */

    return 0;
}

/******************************************************************************
 **函数名称: chat_group_del_sid_item
 **功    能: 删除分组中的SESSION列表项
 **输入参数: 
 **     pool: 内存池
 **     sid: 会话ID
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 释放各SESSION内存空间
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.21 12:53:28 #
 ******************************************************************************/
static void chat_group_del_sid_item(void *pool, chat_session_t *ssn)
{
    fprintf(stderr, "sid:%lu gid:%u rid:%lu\n", ssn->sid, ssn->gid, ssn->rid);
    hash_tab_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
    FREE(ssn);
}

/******************************************************************************
 **函数名称: chat_group_del_item
 **功    能: 删除聊天室中的分组
 **输入参数: 
 **     grp: 聊天室分组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次删除SESSION对应的连接和内存
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static void chat_group_del_item(chat_group_t *grp)
{
    /* > 销毁SID列表 */
    hash_tab_destroy(grp->sid_set, (mem_dealloc_cb_t)chat_group_del_sid_item, NULL);
    FREE(grp);
}

/******************************************************************************
 **函数名称: chat_group_loop_del_item
 **功    能: 删除聊天室中的分组
 **输入参数: 
 **     grp: 聊天室分组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 复用函数chat_group_del_item()
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static void chat_group_loop_del_item(void *pool, chat_group_t *grp)
{
    chat_group_del_item(grp);
}

/******************************************************************************
 **函数名称: chat_room_del_all_group
 **功    能: 删除聊天室中的分组
 **输入参数: 
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次删除SESSION对应的连接和内存
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static int chat_room_del_all_group(chat_tab_t *chat, chat_room_t *room)
{
    room->grp_num = 0;
    room->sid_num = 0;
    hash_tab_destroy(room->group_tab, (mem_dealloc_cb_t)chat_group_loop_del_item, NULL);
    return 0;
}

/******************************************************************************
 **函数名称: chat_group_add_by_gid
 **功    能: 通过GID创建分组
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 11:08:27 #
 ******************************************************************************/
int chat_group_add_by_gid(chat_tab_t *chat, chat_room_t *room, uint32_t gid)
{
    chat_group_t *grp;

    /* > 创建gid分组 */
    grp = (chat_group_t *)calloc(1, sizeof(chat_group_t));
    if (NULL == grp) {
        return -1;
    }

    grp->sid_set = hash_tab_creat(999,
            (hash_cb_t)chat_session_hash_cb,
            (cmp_cb_t)chat_session_cmp_cb, NULL);
    if (NULL == grp->sid_set) {
        FREE(grp);
        return -1;
    }

    grp->gid = gid;

    /* > 将新建分组加入组表 */
    if (hash_tab_insert(room->group_tab, (void *)grp, WRLOCK)) {
        hash_tab_destroy(grp->sid_set, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
        free(grp);
        return -1;
    }

    atomic32_inc(&room->grp_num); /* 分组ID总数 */

    return 0;
}

/* 订阅ID哈希回调 */
static uint64_t chat_sub_hash_cb(chat_sub_item_t *item)
{
    return item->cmd;
}

/* 订阅比较函数回调 */
static int chat_sub_cmp_cb(chat_sub_item_t *item1, chat_sub_item_t *item2)
{
    return item1->cmd - item2->cmd;
}

/******************************************************************************
 **函数名称: chat_group_add_session
 **功    能: 在分组中添加一个SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     grp: 分组对象
 **     sid: 需要添加的SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 17:13:19 #
 ******************************************************************************/
chat_session_t *chat_group_add_session(chat_tab_t *chat,
        chat_room_t *room, chat_group_t *grp, uint64_t sid)
{
    chat_session_t *ssn;

    ssn = (chat_session_t *)calloc(1, sizeof(chat_session_t));
    if (NULL == ssn) {
        return NULL;
    }

    ssn->sid = sid;
    ssn->gid = grp->gid;
    ssn->rid = room->rid;

    ssn->sub = hash_tab_creat(1,
            (hash_cb_t)chat_sub_hash_cb,
            (cmp_cb_t)chat_sub_cmp_cb, NULL);
    if (NULL == ssn->sub) {
        FREE(ssn);
        return NULL;
    }

    if (hash_tab_insert(grp->sid_set, (void *)ssn, WRLOCK)) {
        hash_tab_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
        free(ssn);
        return NULL;
    }

    atomic64_inc(&grp->sid_num);
    atomic64_inc(&room->sid_num);

    return ssn;
}

/******************************************************************************
 **函数名称: chat_group_del_session
 **功    能: 删除分组中指定SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     grp: 分组对象
 **     sid: 需要删除的SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 17:28:42 #
 ******************************************************************************/
int chat_group_del_session(chat_tab_t *chat,
        chat_room_t *room, chat_group_t *grp, uint64_t sid)
{
    chat_session_t *ssn, key;

    /* > 删除会话信息 */
    key.sid = sid;

    ssn = hash_tab_delete(grp->sid_set, (void *)&key, WRLOCK);
    if (NULL == ssn) {
        return 0;
    }

    if (NULL != ssn->sub) {
        hash_tab_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
    }
    FREE(ssn);

    atomic64_dec(&grp->sid_num);
    atomic64_dec(&room->sid_num);

    return 0;
}

/******************************************************************************
 **函数名称: chat_session_ref
 **功    能: 给聊天室添加一个用户
 **输入参数: 
 **     chat: CHAT对象
 **     ssn: 会话对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: ssn此时已在grp表中被引用, 切记!
 **作    者: # Qifeng.zou # 2016.09.21 19:59:38 #
 ******************************************************************************/
int chat_session_ref(chat_tab_t *chat, chat_session_t *ssn)
{
    return hash_tab_insert(chat->session_tab, (void *)ssn, WRLOCK);
}

/******************************************************************************
 **函数名称: _chat_room_trav_all_group
 **功    能: 遍历处理聊天室所有分组成员
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     proc: 遍历处理回调
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     > 当gid为0时, 表示遍历所有聊天室成员
 **     > 当gid不为0时, 表示遍历聊天室GID分组所有成员
 **作    者: # Qifeng.zou # 2016.09.21 23:19:08 #
 ******************************************************************************/
typedef struct
{
    void *args;
    trav_cb_t proc;
} chat_trav_group_args_t;

static int _chat_room_trav_all_group(chat_group_t *grp, chat_trav_group_args_t *trav)
{
    hash_tab_trav(grp->sid_set, (trav_cb_t)trav->proc, trav->args, RDLOCK);
    return 0;
}

/******************************************************************************
 **函数名称: chat_room_trav_all_group
 **功    能: 遍历处理聊天室所有分组成员
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     proc: 遍历处理回调
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     > 当gid为0时, 表示遍历所有聊天室成员
 **     > 当gid不为0时, 表示遍历聊天室GID分组所有成员
 **作    者: # Qifeng.zou # 2016.09.21 23:19:08 #
 ******************************************************************************/
int chat_room_trav_all_group(chat_tab_t *chat,
        chat_room_t *room, trav_cb_t proc, void *args, lock_e lock)
{
    chat_trav_group_args_t trav;

    memset(&trav, 0, sizeof(trav));

    trav.args = (void *)args;
    trav.proc = (trav_cb_t)proc;

    hash_tab_trav(room->group_tab, (trav_cb_t)_chat_room_trav_all_group, (void *)&trav, lock);

    return 0;
}

/******************************************************************************
 **函数名称: chat_del_session_from_group
 **功    能: 从组中删除会话
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     proc: 遍历处理回调
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     > 当gid为0时, 表示遍历所有聊天室成员
 **     > 当gid不为0时, 表示遍历聊天室GID分组所有成员
 **作    者: # Qifeng.zou # 2016.09.21 23:19:08 #
 ******************************************************************************/
int chat_del_session_from_group(chat_tab_t *chat, chat_room_t *room, int gid, uint64_t sid)
{
    chat_group_t *grp, grp_key;

    /* > 查收聊天分组 */
    grp_key.gid = gid;

    grp = hash_tab_query(room->group_tab, &grp_key, RDLOCK);
    if (NULL == grp) {
        log_error(chat->log, "Didn't find group! sid:%lu rid:%u gid:%u.",
                sid, room->rid, gid);
        return -1; /* 未找到分组 */
    }

    /* > 从分组中删除SID */
    chat_group_del_session(chat, room, grp, sid);

    hash_tab_unlock(room->group_tab, &grp_key, RDLOCK);

    return 0;
}

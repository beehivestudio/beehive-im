#include "log.h"
#include "redo.h"
#include "chat.h"
#include "atomic.h"
#include "chat_priv.h"

/* 静态函数 */
static void chat_group_destroy(chat_group_t *grp);
static void _chat_group_destroy(void *pool, chat_group_t *grp);
static int chat_room_destroy(chat_tab_t *chat, chat_room_t *room);

static int chat_group_add_session(chat_tab_t *chat, chat_room_t *room, uint32_t gid, uint64_t sid, uint64_t cid);
static int chat_group_del_session(chat_tab_t *chat, chat_room_t *room, uint32_t gid, uint64_t sid, uint64_t cid);

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
static int chat_add_room(chat_tab_t *chat, uint64_t rid)
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
    room->groups = hash_tab_creat(99,
            (hash_cb_t)chat_group_hash_cb, (cmp_cb_t)chat_group_cmp_cb, NULL);
    if (NULL == room->groups) {
        free(room);
        return -1;
    }

    /* > 将ROOM放入管理表 */
    if (hash_tab_insert(chat->rooms, (void *)room, WRLOCK)) {
        hash_tab_destroy(room->groups, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
        free(room);
        return -1;
    }

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

    room = hash_tab_query(chat->rooms, (void *)&key, WRLOCK);
    if (NULL == room) {
        return 0;
    }
    else if ((0 != room->sid_num) || (0 != room->grp_num)) {
        hash_tab_unlock(chat->rooms, (void *)&key, WRLOCK);
        log_error(chat->log, "Delete room failed! rid:%lu sid num:%d gid num:",
                room->rid, room->sid_num, room->grp_num);
        return 0; /* 聊天室还有人, 不能删除 */
    }

    log_info(chat->log, "Delete room [%u]", room->rid);

    hash_tab_delete(chat->rooms, (void *)&key, NONLOCK);

    hash_tab_unlock(chat->rooms, (void *)&key, WRLOCK);

    /* > 销毁聊天室 */
    return chat_room_destroy(chat, room);
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

    item = hash_tab_delete(room->groups, &key, WRLOCK);
    assert(grp == item);
    chat_group_destroy(grp);

    --room->grp_num; /* 分组数减1 */

    return 0;
}

/******************************************************************************
 **函数名称: chat_group_del_sid
 **功    能: 删除分组中的SID列表项
 **输入参数: 
 **     pool: 内存池
 **     item: SID->CID映射
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项: 
 **     由于分组中的sid list挂的是只是SID, 因此无空间需要释放.
 **作    者: # Qifeng.zou # 2016.09.21 12:53:28 #
 ******************************************************************************/
static void chat_group_del_sid(void *pool, chat_sid2cid_item_t *item)
{
    free(item);
    return;
}

/******************************************************************************
 **函数名称: chat_group_destroy
 **功    能: 销毁聊天室分组
 **输入参数: 
 **     grp: 聊天室分组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次删除SESSION对应的连接和内存
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static void chat_group_destroy(chat_group_t *grp)
{
    /* > 销毁SID列表 */
    hash_tab_destroy(grp->sid_set, (mem_dealloc_cb_t)chat_group_del_sid, NULL);
    FREE(grp);
}

/******************************************************************************
 **函数名称: _chat_group_destroy
 **功    能: 销毁聊天室中的分组
 **输入参数: 
 **     grp: 聊天室分组
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 复用函数chat_group_destroy()
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static void _chat_group_destroy(void *pool, chat_group_t *grp)
{
    chat_group_destroy(grp);
}

/******************************************************************************
 **函数名称: chat_room_destroy
 **功    能: 删除聊天室
 **输入参数: 
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 依次删除各组和各SESSION内存
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 22:47:11 #
 ******************************************************************************/
static int chat_room_destroy(chat_tab_t *chat, chat_room_t *room)
{
    room->grp_num = 0;
    room->sid_num = 0;

    hash_tab_destroy(room->groups, (mem_dealloc_cb_t)_chat_group_destroy, NULL);

    FREE(room);
    return 0;
}

/* 会话ID哈希回调 */
static uint64_t chat_sid_hash_cb(uint64_t *sid)
{
    return (uint64_t)sid;
}

/* 会话ID比较回调 */
static int chat_sid_cmp_cb(uint64_t *s1, uint64_t *s2)
{
    return (int)((uint64_t)s1 - (uint64_t)s2);
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
static int chat_group_add_by_gid(chat_tab_t *chat, chat_room_t *room, uint32_t gid)
{
    chat_group_t *grp;

    /* > 创建gid分组 */
    grp = (chat_group_t *)calloc(1, sizeof(chat_group_t));
    if (NULL == grp) {
        return -1;
    }

    grp->sid_set = hash_tab_creat(999,
            (hash_cb_t)chat_sid_hash_cb,
            (cmp_cb_t)chat_sid_cmp_cb, NULL);
    if (NULL == grp->sid_set) {
        FREE(grp);
        return -1;
    }

    grp->gid = gid;

    /* > 将新建分组加入组表 */
    if (hash_tab_insert(room->groups, (void *)grp, WRLOCK)) {
        hash_tab_destroy(grp->sid_set, (mem_dealloc_cb_t)mem_dummy_dealloc, NULL);
        free(grp);
        return -1;
    }

    atomic32_inc(&room->grp_num); /* 分组ID总数 */

    return 0;
}

/* 聊天室ID哈希回调 */
static uint64_t chat_room_hash_cb(chat_sub_item_t *item)
{
    return item->cmd;
}

/* 聊天室比较函数回调 */
static int chat_room_cmp_cb(chat_room_item_t *item1, chat_room_item_t *item2)
{
    return item1->rid - item2->rid;
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
 **函数名称: _chat_group_add_session
 **功    能: 在分组中添加一个SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     grp: 分组对象
 **     sid: 需要添加的SID
 **     cid: 需要添加的CID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 17:13:19 #
 ******************************************************************************/
static int _chat_group_add_session(chat_tab_t *chat,
        chat_room_t *room, chat_group_t *grp, uint64_t sid, uint64_t cid)
{
    int ret;
    chat_sid2cid_item_t *item;

    item = calloc(1, sizeof(chat_sid2cid_item_t));
    if (NULL == item) {
        return -1;
    }

    item->sid = sid;
    item->cid = cid;

    ret = hash_tab_insert(grp->sid_set, (void *)item, WRLOCK); 
    if (RBT_OK != ret) {
        free(item);
        return (RBT_NODE_EXIST == ret)? 0 : -1;
    }

    atomic64_inc(&grp->sid_num);
    atomic64_inc(&room->sid_num);

    return 0;
}

/******************************************************************************
 **函数名称: _chat_group_del_session
 **功    能: 删除分组中指定SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     grp: 分组对象
 **     sid: 需要删除的SID
 **     cid: 需要删除的CID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.21 17:28:42 #
 ******************************************************************************/
static int _chat_group_del_session(chat_tab_t *chat,
        chat_room_t *room, chat_group_t *grp, uint64_t sid, uint64_t cid)
{
    chat_sid2cid_item_t *item, key;

    /* > 删除会话信息 */
    key.sid = sid;
    key.cid = cid;

    item = hash_tab_delete(grp->sid_set, (void *)&key, WRLOCK);
    if (NULL == item) {
        return 0; /* Didn't find */
    }

    FREE(item);

    atomic64_dec(&grp->sid_num);
    atomic64_dec(&room->sid_num);

    return 0;
}

/******************************************************************************
 **函数名称: chat_session_tab_add
 **功    能: 给聊天室添加一个用户
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.21 19:59:38 #
 ******************************************************************************/
int chat_session_tab_add(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid)
{
    chat_session_t *session;
    chat_room_item_t *room;

    /* > 新建会话对象 */
    session = (chat_session_t *)calloc(1, sizeof(chat_session_t));
    if (NULL == session) {
        return -1;
    }

    session->sid = sid;
    session->cid = cid;

    /* > 新建ROOM对象 */
    room = (chat_room_item_t *)calloc(1, sizeof(chat_room_item_t));
    if (NULL == session) {
        FREE(session);
        return -1;
    }

    room->rid = rid;
    room->gid = gid;

    /* > 新建ROOM对象 */
    session->room = hash_tab_creat(5, (hash_cb_t)chat_room_hash_cb, (cmp_cb_t)chat_room_cmp_cb, NULL);
    if (NULL == session->room) {
        FREE(session);
        return -1;
    }

    if (hash_tab_insert(session->room, room, WRLOCK)) {
        FREE(room);
    }

    /* > 新建SUB对象 */
    session->sub = hash_tab_creat(5, (hash_cb_t)chat_sub_hash_cb, (cmp_cb_t)chat_sub_cmp_cb, NULL);
    if (NULL == session->sub) {
        hash_tab_destroy(session->room, (mem_dealloc_cb_t)mem_dealloc, NULL);
        FREE(session);
        return -1;
    }

    /* > 插入会话表 */
    if (hash_tab_insert(chat->sessions, (void *)session, WRLOCK)) {
        hash_tab_destroy(session->room, (mem_dealloc_cb_t)mem_dealloc, NULL);
        hash_tab_destroy(session->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
        FREE(session);
        return -1;
    }

    return 0;
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
static int chat_room_trav_all_group(chat_tab_t *chat,
        chat_room_t *room, trav_cb_t proc, void *args, lock_e lock)
{
    chat_trav_group_args_t trav;

    memset(&trav, 0, sizeof(trav));

    trav.args = (void *)args;
    trav.proc = (trav_cb_t)proc;

    hash_tab_trav(room->groups, (trav_cb_t)_chat_room_trav_all_group, (void *)&trav, lock);

    return 0;
}

/******************************************************************************
 **函数名称: _chat_room_add_session
 **功    能: 聊天室添加会话
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.01 16:19:49 #
 ******************************************************************************/
int _chat_room_add_session(chat_tab_t *chat,
        uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid)
{
    int ret;
    chat_room_t *room, key;

    /* > 判断聊天室是否存在
     * 注: 不存在就创建, 知道创建成功. */
QUERY_ROOM:
    key.rid = rid;

    room = (chat_room_t *)hash_tab_query(chat->rooms, (void *)&key, RDLOCK);
    if (NULL == room) {
        chat_add_room(chat, rid);
        goto QUERY_ROOM;
    }

    ret = chat_group_add_session(chat, room, gid, sid, cid);

    hash_tab_unlock(chat->rooms, (void *)&key, RDLOCK);

    return ret;
}

/******************************************************************************
 **函数名称: chat_group_add_session
 **功    能: 在分组中添加一个SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
 **     sid: 会话SID
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.01 16:25:20 #
 ******************************************************************************/
static int chat_group_add_session(chat_tab_t *chat,
        chat_room_t *room, uint32_t gid, uint64_t sid, uint64_t cid)
{
    int ret;
    chat_group_t *grp, key;

    /* > 查找指定分组 */
QUERY_GROUP:
    key.gid = gid;

    grp = (chat_group_t *)hash_tab_query(room->groups, &key, RDLOCK);
    if (NULL == grp) {
        chat_group_add_by_gid(chat, room, gid);
        goto QUERY_GROUP;
    }

    /* > 将此SID加入分组列表 */
    ret = _chat_group_add_session(chat, room, grp, sid, cid);

    hash_tab_unlock(room->groups, (void *)&key, RDLOCK);
    return ret;
}

/******************************************************************************
 **函数名称: _chat_room_del_session
 **功    能: 从聊天室中删除某会话
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.10.01 16:10:10 #
 ******************************************************************************/
int _chat_room_del_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid, uint64_t cid)
{
    chat_room_t *room, key;

    /* > 查找对应的聊天室 */
    key.rid = rid;

    room = hash_tab_query(chat->rooms, (void *)&key, RDLOCK);
    if (NULL == room) {
        log_error(chat->log, "Didn't find room! sid:%u rid:%lu", sid, rid);
        return 0; /* Didn't find */
    }

    /* > 从对应的分组中删除会话 */
    chat_group_del_session(chat, room, gid, sid, cid);

    hash_tab_unlock(chat->rooms, (void *)&key, RDLOCK);

    return 0;
}

/******************************************************************************
 **函数名称: _chat_room_trav_del_session
 **功    能: 从聊天室中删除某会话
 **输入参数: 
 **     room: 聊天室信息
 **     param: 遍历参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.15 22:23:40 #
 ******************************************************************************/
int _chat_room_trav_del_session(chat_room_item_t *room, chat_session_trav_room_t *param)
{
    chat_tab_t *chat;
    chat_session_t *session;

    chat = param->chat;
    session = param->session;

    _chat_room_del_session(chat, room->rid, room->gid, session->sid, session->cid);

    return 0;
}

/******************************************************************************
 **函数名称: chat_group_del_session
 **功    能: 从组中删除会话
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     > 当gid为0时, 表示遍历所有聊天室成员
 **     > 当gid不为0时, 表示遍历聊天室GID分组所有成员
 **作    者: # Qifeng.zou # 2016.09.21 23:19:08 #
 ******************************************************************************/
static int chat_group_del_session(chat_tab_t *chat,
        chat_room_t *room, uint32_t gid, uint64_t sid, uint64_t cid)
{
    chat_group_t *grp, key;

    /* > 查收聊天分组 */
    key.gid = gid;

    grp = hash_tab_query(room->groups, &key, RDLOCK);
    if (NULL == grp) {
        log_error(chat->log, "Didn't find group! sid:%lu rid:%u gid:%u.",
                sid, room->rid, gid);
        return -1; /* 未找到分组 */
    }

    /* > 从分组中删除SID */
    _chat_group_del_session(chat, room, grp, sid, cid);

    hash_tab_unlock(room->groups, &key, RDLOCK);

    return 0;
}

/******************************************************************************
 **函数名称: chat_group_trav
 **功    能: 遍历某个分组所有SID
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
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
int chat_group_trav(chat_tab_t *chat,
        chat_room_t *room, uint16_t gid, trav_cb_t proc, void *args)
{
    chat_group_t *grp, key;

    if (0 == gid) { /* 遍历聊天室所有成员 */
        return chat_room_trav_all_group(chat, room, proc, args, RDLOCK);
    }

    /* 只发给同一聊天室分组的成员 */
    key.gid = gid;

    grp = (chat_group_t *)hash_tab_query(room->groups, &key, RDLOCK);
    if (NULL == grp) {
        return 0; /* Didn't find */
    }

    hash_tab_trav(grp->sid_set, (trav_cb_t)proc, args, RDLOCK);

    hash_tab_unlock(room->groups, (void *)&key, RDLOCK);

    return 0;
}

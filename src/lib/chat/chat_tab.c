#include "log.h"
#include "redo.h"
#include "chat.h"
#include "chat_priv.h"

/* 静态函数 */
static void chat_group_del_item(chat_group_t *grp);
static void chat_group_loop_del_item(void *pool, chat_group_t *grp);
static int chat_room_del_all_group(chat_tab_t *chat, chat_room_t *room);
static chat_group_t *chat_group_add_by_gid(chat_tab_t *chat, chat_room_t *room, uint16_t gid);

/******************************************************************************
 **函数名称: chat_room_cmp_cb
 **功    能: 聊天室比较回调
 **输入参数: 
 **     room1: 聊天室1
 **     room2: 聊天室2
 **输出参数: NONE
 **返    回: 0:相等 <0:小于 >0:大于
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.08.09 18:02:19 #
 ******************************************************************************/
static int chat_room_cmp_cb(chat_room_t *r1, chat_room_t *r2)
{
    return (r1->rid - r2->rid);
}

/******************************************************************************
 **函数名称: chat_group_cmp_cb
 **功    能: 聊天室分组比较回调
 **输入参数: 
 **     group1: 分组1
 **     group2: 分组2
 **输出参数: NONE
 **返    回: 0:相等 <0:小于 >0:大于
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.08.09 18:02:19 #
 ******************************************************************************/
static int chat_group_cmp_cb(chat_group_t *g1, chat_group_t *g2)
{
    return (g1->gid - g2->gid);
}

/******************************************************************************
 **函数名称: chat_session_cmp_cb
 **功    能: 会话分组比较回调
 **输入参数: 
 **     group1:1
 **     group2: 分组2
 **输出参数: NONE
 **返    回: 0:相等 <0:小于 >0:大于
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.08.09 18:02:19 #
 ******************************************************************************/
static int chat_session_cmp_cb(chat_session_t *s1, chat_session_t *s2)
{
    return (s1->sid - s2->sid);
}

/******************************************************************************
 **函数名称: chat_tab_init
 **功    能: 初始化上下文
 **输入参数:
 **     grp_usr_num: 各分组中的人数
 **     log: 日志对象
 **输出参数: NONE
 **返    回: CHAT对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.01.05 19:59:38 #
 ******************************************************************************/
chat_tab_t *chat_tab_init(int grp_usr_num, log_cycle_t *log)
{
    chat_tab_t *chat;

    /* > 创建全局对象 */
    chat = (chat_tab_t *)calloc(1, sizeof(chat_tab_t));
    if (NULL == chat) {
        return NULL;
    }

    chat->log = log;
    chat->grp_usr_max_num = (0 == grp_usr_num)? CHAT_USR_DEF_NUM : grp_usr_num;

    do {
        /* > 初始化聊天室表 */
        chat->room_tab = avl_creat(NULL, (cmp_cb_t)chat_room_cmp_cb);
        if (NULL == chat->room_tab) {
            break;
        }

        /* > 初始化SESSION表 */
        chat->session_tab = rbt_creat(NULL, (cmp_cb_t)chat_session_cmp_cb);
        if (NULL == chat->session_tab) {
            break;
        }
        return chat;
    } while(0);

    /* > 释放内存 */
    FREE(chat->room_tab);
    FREE(chat->session_tab);
    FREE(chat);

    return NULL;
}

/******************************************************************************
 **函数名称: chat_tab_set_grp_usr_num
 **功    能: 设置聊天室各组最大人数
 **输入参数:
 **     num: 各分组中的人数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.07.13 11:38:14 #
 ******************************************************************************/
int chat_tab_set_grp_usr_num(chat_tab_t *chat, uint64_t max)
{
    if (0 == max) {
        return -1;
    }

    chat->grp_usr_max_num = max;

    return 0;
}

/******************************************************************************
 **函数名称: chat_add_room
 **功    能: 添加一个聊天室
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: ROOM对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.01.05 21:27:00 #
 ******************************************************************************/
chat_room_t *chat_add_room(chat_tab_t *chat, uint64_t rid)
{
    time_t tm;
    chat_room_t *room = NULL, key;

    /* > 判断聊天室是否已存在 */
    key.rid = rid;

    room = avl_query(chat->room_tab, (void *)&key);
    if (NULL != room) {
        return room;
    }

    tm = time(NULL);

    do {
        /* > 创建聊天室对象 */
        room = (chat_room_t *)calloc(1, sizeof(chat_room_t));
        if (NULL == room) {
            break;
        }

        room->rid = rid;
        room->create_tm = tm;

        /* > 创建分组列表 */
        room->group_tab = avl_creat(NULL, (cmp_cb_t)chat_group_cmp_cb);
        if (NULL == room->group_tab) {
            break;
        }

        /* > 将ROOM放入管理表 */
        if (avl_insert(chat->room_tab, (void *)room)) {
            break;
        }
        return room;
    } while(0);

    /* > 释放内存空间 */
    if (NULL != room) {
        if (NULL != room->group_tab) {
            avl_destroy(room->group_tab, (mem_dealloc_cb_t)chat_group_loop_del_item, NULL);
        }
        FREE(room);
    }
    return NULL;
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
 **作    者: # Qifeng.zou # 2016.01.05 22:47:11 #
 ******************************************************************************/
static int chat_del_room_by_rid(chat_tab_t *chat, uint64_t rid)
{
    chat_room_t *room, key;

    do {
        /* > 查找聊天室对象 */
        key.rid = rid;

        room = avl_query(chat->room_tab, (void *)&key);
        if (NULL == room) {
            break; /* 无此聊天室 */
        }
        else if (0 != room->sid_num) {
            break; /* 聊天室还有人, 不能删除 */
        }

        avl_delete(chat->room_tab, (void *)&key, (void *)&room);

        /* > 删除分组信息 */
        chat_room_del_all_group(chat, room);
    } while(0);

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
 **作    者: # Qifeng.zou # 2016.01.08 14:26:46 #
 ******************************************************************************/
int chat_del_room(chat_tab_t *chat, chat_room_t *room)
{
    chat_room_t key, *addr;

    if (0 != room->sid_num || 0 != room->grp_num) {
        return 0; /* 聊天室还有人, 不能删除 */
    }

    log_info(chat->log, "Delete room [%u]", room->rid);

    do {
        key.rid = room->rid;

        avl_delete(chat->room_tab, (void *)&key, (void *)&addr);

        /* > 删除分组信息 */
        chat_room_del_all_group(chat, room);
    } while(0);

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
 **作    者: # Qifeng.zou # 2016.01.06 20:28:08 #
 ******************************************************************************/
int chat_del_group(chat_tab_t *chat, chat_room_t *room, chat_group_t *grp)
{
    chat_group_t *item, key;

    if (0 != grp->sid_num) {
        return 0; // 还存在用户, 无法删除
    }
    
    log_info(chat->log, "Delete gid:%u", grp->gid);

    key.gid = grp->gid;

    avl_delete(room->group_tab, &key, (void *)&item);
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
 **作    者: # Qifeng.zou # 2016.01.06 12:53:28 #
 ******************************************************************************/
static void chat_group_del_sid_item(void *pool, chat_session_t *ssn)
{
    fprintf(stderr, "sid:%lu gid:%u rid:%lu\n", ssn->sid, ssn->gid, ssn->rid);
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
 **作    者: # Qifeng.zou # 2016.01.05 22:47:11 #
 ******************************************************************************/
static void chat_group_del_item(chat_group_t *grp)
{
    /* > 销毁SID列表 */
    rbt_destroy(grp->sid_list, (mem_dealloc_cb_t)chat_group_del_sid_item, NULL);
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
 **作    者: # Qifeng.zou # 2016.01.05 22:47:11 #
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
 **作    者: # Qifeng.zou # 2016.01.05 22:47:11 #
 ******************************************************************************/
static int chat_room_del_all_group(chat_tab_t *chat, chat_room_t *room)
{
    avl_destroy(room->group_tab, (mem_dealloc_cb_t)chat_group_loop_del_item, NULL);
    room->grp_num = 0;
    room->sid_num = 0;

    return 0;
}

/******************************************************************************
 **函数名称: chat_group_find_not_full
 **功    能: 查找人数未满的分组
 **输入参数: 
 **     grp: GROUP对象
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: true:已找到 false:未找到, 请继续查找...
 **实现描述:
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.02.14 11:16:50 #
 ******************************************************************************/
bool chat_group_find_not_full(chat_group_t *grp, chat_tab_t *chat)
{
    return (grp->sid_num >= chat->grp_usr_max_num)? false : true;
}

/******************************************************************************
 **函数名称: chat_group_alloc
 **功    能: 筛选某人数未满的分组
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID(gid为-1时, 表示随机分配组ID; 为其他值时查找或创建gid分组)
 **输出参数: NONE
 **返    回: ROOM分组
 **实现描述:
 **注意事项:
 **     > 为防止组ID溢出, 需要实现组ID的复用
 **     > 虚拟分组ID从0开始算起, 可有效防止有些聊天室过于冷清
 **作    者: # Qifeng.zou # 2016.01.06 15:52:10 #
 ******************************************************************************/
chat_group_t *chat_group_alloc(chat_tab_t *chat, chat_room_t *room, uint16_t gid)
{
    chat_group_t *grp;

    /* > 查找/创建指定gid分组 */
    if ((uint16_t)-1 != gid) {
        grp = chat_group_find_by_gid(chat, room, gid);
        if (NULL != grp) {
            return grp;
        }

        return chat_group_add_by_gid(chat, room, gid);
    }

    /* > 查找是否存在未满的分组 */
    grp = avl_find(room->group_tab, (find_cb_t)chat_group_find_not_full, (void *)chat);
    if (NULL != grp) {
        return grp;
    }

    /* > 所有分组人数均满(注: 需新建新的分组) */
    return chat_group_add_by_gid(chat, room, room->grp_idx);
}

/******************************************************************************
 **函数名称: chat_group_find_by_gid
 **功    能: 通过GID查找分组
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
 **输出参数: NONE
 **返    回: ROOM分组
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.01.06 15:52:10 #
 ******************************************************************************/
chat_group_t *chat_group_find_by_gid(chat_tab_t *chat, chat_room_t *room, uint16_t gid)
{
    chat_group_t key;

    key.gid = gid;

    return (chat_group_t *)avl_query(room->group_tab, &key);
}

/******************************************************************************
 **函数名称: chat_group_add_by_gid
 **功    能: 通过GID创建分组
 **输入参数: 
 **     chat: CHAT对象
 **     room: ROOM对象
 **     gid: 分组ID
 **输出参数: NONE
 **返    回: ROOM分组
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.05.19 13:09:42 #
 ******************************************************************************/
static chat_group_t *chat_group_add_by_gid(chat_tab_t *chat, chat_room_t *room, uint16_t gid)
{
    chat_group_t *grp;

    /* > 创建gid分组 */
    grp = (chat_group_t *)calloc(1, sizeof(chat_group_t));
    if (NULL == grp) {
        return NULL;
    }

    grp->sid_list = rbt_creat(NULL, (cmp_cb_t)chat_session_cmp_cb);
    if (NULL == grp->sid_list) {
        FREE(grp);
        return NULL;
    }

    ++room->grp_num; /* 分组ID总数 */
    ++room->grp_idx; /* 分组序列号(此时gid为未被使用的组ID) */
    grp->gid = gid;

    /* > 将新建分组加入到AVL表 */
    if (avl_insert(room->group_tab, (void *)grp)) {
        rbt_destroy(grp->sid_list, (mem_dealloc_cb_t)mem_dealloc, NULL);
        --room->grp_num;
        --room->grp_idx;
        free(grp);
        return NULL;
    }

    return grp;
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
 **作    者: # Qifeng.zou # 2016.01.06 17:13:19 #
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

    ssn->sub = avl_creat(NULL, (cmp_cb_t)chat_sub_cmp_cb);
    if (NULL == ssn->sub) {
        FREE(ssn);
        return NULL;
    }

    if (rbt_insert(grp->sid_list, (void *)ssn)) {
        avl_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
        FREE(ssn);
        return NULL;
    }
    ++grp->sid_num;
    ++room->sid_num;
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
 **作    者: # Qifeng.zou # 2016.01.06 17:28:42 #
 ******************************************************************************/
int chat_group_del_session(chat_tab_t *chat,
        chat_room_t *room, chat_group_t *grp, uint64_t sid)
{
    chat_session_t *ssn, key;

    key.sid = sid;

    rbt_delete(grp->sid_list, (void *)&key, (void **)&ssn);
    if (NULL == ssn) {
        return 0;
    }

    if (NULL == ssn->sub) {
        avl_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
    }
    FREE(ssn);

    --grp->sid_num;
    --room->sid_num;

    if (0 == grp->sid_num) { /* 释放分组空间 */
        chat_del_group(chat, room, grp);
    }

    if (0 == room->sid_num || 0 == room->grp_num) { /* 释放聊天室空间 */
        chat_del_room(chat, room);
    }

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
 **作    者: # Qifeng.zou # 2016.01.05 19:59:38 #
 ******************************************************************************/
int chat_session_ref(chat_tab_t *chat, chat_session_t *ssn)
{
    return rbt_insert(chat->session_tab, (void *)ssn);
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
 **作    者: # Qifeng.zou # 2016.01.06 23:19:08 #
 ******************************************************************************/
typedef struct
{
    void *args;
    trav_cb_t proc;
} chat_trav_group_args_t;

static int _chat_room_trav_all_group(chat_group_t *grp, chat_trav_group_args_t *trav)
{
    rbt_trav(grp->sid_list, (trav_cb_t)trav->proc, trav->args);
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
 **作    者: # Qifeng.zou # 2016.01.06 23:19:08 #
 ******************************************************************************/
int chat_room_trav_all_group(chat_tab_t *chat, chat_room_t *room, trav_cb_t proc, void *args)
{
    chat_trav_group_args_t trav;

    memset(&trav, 0, sizeof(trav));

    trav.args = (void *)args;
    trav.proc = (trav_cb_t)proc;

    avl_trav(room->group_tab, (trav_cb_t)_chat_room_trav_all_group, (void *)&trav);

    return 0;
}

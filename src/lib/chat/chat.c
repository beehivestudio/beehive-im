#include "log.h"
#include "redo.h"
#include "chat.h"
#include "atomic.h"
#include "chat_priv.h"

static bool _chat_has_sub(chat_session_t *ssn, uint16_t cmd);

/******************************************************************************
 **函数名称: chat_add_session
 **功    能: 给聊天室添加一个用户
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **     sid: 会话ID
 **     gid: 分组ID
 **输出参数: NONE
 **返    回: 分组ID
 **实现描述:
 **     1. 如果此SID存在, 则验证数据合法性
 **     2. 如果此SID不存在
 **         > 判断是否添加聊天室
 **         > 判断是否添加聊天分组
 **注意事项: TODO: 单个SID可以加入多个聊天室.
 **     1. 在此处不删除人数为0的聊天室和其下的分组, 删除操作由定时任务统一处理.
 **        理由: 降低程序的复杂度.
 **     2. 尽量进少使用写锁的次数, 尽量降低写锁的粒度.
 **作    者: # Qifeng.zou # 2016.09.20 10:53:18 #
 ******************************************************************************/
int chat_add_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid)
{
    chat_room_t *room, room_key;
    chat_group_t *grp = NULL, grp_key;
    chat_session_t *ssn, ssn_key;

    /* > 判断该SID是否已经存在 */
    ssn_key.sid = sid;

    ssn = (chat_session_t *)hash_tab_query(chat->session_tab, (void *)&ssn_key, RDLOCK);
    if (NULL != ssn) {
        if (ssn->rid == rid) {
            gid = ssn->gid;
            hash_tab_unlock(chat->session_tab, (void *)&ssn_key, RDLOCK);
            return gid; /* 已存在 */
        }
        hash_tab_unlock(chat->session_tab, (void *)&ssn_key, RDLOCK);
        return -1; /* 失败: SID所在聊天室与申请的聊天室冲突 */
    }

    /* > 判断聊天室是否存在
     * 注: 不存在就创建, 知道创建成功. */
QUERY_ROOM:
    room_key.rid = rid;

    room = (chat_room_t *)hash_tab_query(chat->room_tab, (void *)&room_key, RDLOCK);
    if (NULL == room) {
        chat_add_room(chat, rid);
        goto QUERY_ROOM;
    }

    do {
        /* > 查找指定分组 */
    QUERY_ROOM_GROUP:
        grp_key.gid = gid;

        grp = (chat_group_t *)hash_tab_query(room->group_tab, &grp_key, RDLOCK);
        if (NULL == grp) {
            chat_group_add_by_gid(chat, room, gid);
            goto QUERY_ROOM_GROUP;
        }

        /* > 将此SID加入分组列表 */
    ADD_SESSION:
        chat_group_add_session(chat, room, grp, sid); 

        /* > 判断该SID是否已经存在 */
        ssn_key.sid = sid;

        ssn = (chat_session_t *)hash_tab_query(grp->sid_set, (void *)&ssn_key, RDLOCK);
        if (NULL == ssn) {
            goto ADD_SESSION;
        }
        else if (ssn->rid != rid) {
            break; /* 失败: SID所在聊天室与申请的聊天室冲突 */
        }

        /* > 构建SID索引表 */
        if (chat_session_ref(chat, ssn)) {
            hash_tab_unlock(grp->sid_set, (void *)&ssn_key, RDLOCK);
            chat_group_del_session(chat, room, grp, sid);
            hash_tab_unlock(room->group_tab, (void *)&grp_key, RDLOCK);
            hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
            return -1;
        }
        gid = ssn->gid;

        hash_tab_unlock(grp->sid_set, (void *)&ssn_key, RDLOCK);
        hash_tab_unlock(room->group_tab, (void *)&grp_key, RDLOCK);
        hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
        return gid;
    } while(0);

    hash_tab_unlock(grp->sid_set, (void *)&ssn_key, RDLOCK);
    hash_tab_unlock(room->group_tab, (void *)&grp_key, RDLOCK);
    hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
    return -1;
}

/******************************************************************************
 **函数名称: chat_del_session
 **功    能: 给聊天室添加一个用户
 **输入参数: 
 **     chat: CHAT对象
 **     sid: 会话ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     1. session_tab中的chat_session_t对象的内存在group->sid_list中被引用.
 **作    者: # Qifeng.zou # 2016.09.20 19:59:38 #
 ******************************************************************************/
int chat_del_session(chat_tab_t *chat, uint64_t sid)
{
    chat_room_t *room, room_key;
    chat_session_t *ssn, ssn_key;

    /* > 删除SID索引 */
    ssn_key.sid = sid;

    ssn = hash_tab_delete(chat->session_tab, (void *)&ssn_key, WRLOCK);
    if (NULL == ssn) {
        log_error(chat->log, "Didn't find sid[%u]. ptr:%p", sid, ssn);
        return 0; /* 此SID已删除 */
    }

    /* > 查找对应的聊天室 */
    room_key.rid = ssn->rid;

    room = hash_tab_query(chat->room_tab, (void *)&room_key, RDLOCK);
    if (NULL == room) {
        log_error(chat->log, "Didn't find room! sid:%u rid:%u.\n", sid, ssn->rid);
        return 0;
    }

    /* > 从对应的分组中删除会话 */
    chat_del_session_from_group(chat, room, ssn->gid, ssn->sid);

    hash_tab_unlock(chat->room_tab, (void *)&room_key, WRLOCK);

    // FREE(ssn); /* 内存已经被chat_group_del_session释放 */
    return 0;
}

/******************************************************************************
 **函数名称: chat_room_trav
 **功    能: 遍历处理聊天室
 **输入参数: 
 **     chat: CHAT对象
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     proc: 遍历处理回调
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     > 当gid为0时, 表示遍历所有聊天室成员
 **     > 当gid不为0时, 表示遍历聊天室GID分组所有成员
 **作    者: # Qifeng.zou # 2016.09.20 23:19:08 #
 ******************************************************************************/
int chat_room_trav(chat_tab_t *chat, uint64_t rid, uint16_t gid, trav_cb_t proc, void *args)
{
    chat_group_t *grp, grp_key;
    chat_room_t *room, room_key;

    room_key.rid = rid;

    room = (chat_room_t *)hash_tab_query(chat->room_tab, (void *)&room_key, RDLOCK);
    if (NULL == room) {
        return 0; // 聊天室不存在
    }

    if (0 == gid) { // 遍历聊天室所有成员
        chat_room_trav_all_group(chat, room, proc, args, RDLOCK);
        hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
        return 0;
    }

    // 只发给同一聊天室分组的成员
    grp_key.gid = gid;

    grp = (chat_group_t *)hash_tab_query(room->group_tab, &grp_key, RDLOCK);
    if (NULL == grp) {
        hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
        return 0;
    }

    hash_tab_trav(grp->sid_set, (trav_cb_t)proc, args, RDLOCK);

    hash_tab_unlock(room->group_tab, (void *)&grp_key, RDLOCK);
    hash_tab_unlock(chat->room_tab, (void *)&room_key, RDLOCK);
    return 0;
}

/******************************************************************************
 **函数名称: chat_add_sub
 **功    能: 添加订阅消息
 **输入参数: 
 **     chat: TAB对象
 **     sid: 会话ID
 **     cmd: 命令类型
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
int chat_add_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd)
{
    int ret;
    chat_sub_item_t *item;
    chat_session_t *ssn, key;

    /* 准备数据 */
    item = (chat_sub_item_t *)calloc(1, sizeof(chat_sub_item_t));
    if (NULL == item) {
        return -1;
    }

    item->cmd = cmd;

    /* 查找对象 */
    key.sid = sid;

    ssn = (chat_session_t *)hash_tab_query(chat->session_tab, &key, RDLOCK);
    if (NULL == ssn) {
        free(item);
        return -1;
    }

    ret = hash_tab_insert(ssn->sub, (void *)item, WRLOCK);
    if (AVL_OK != ret) {
        hash_tab_unlock(chat->session_tab, ssn, RDLOCK);
        free(item);
        return (AVL_NODE_EXIST == ret)? 0 : -1;
    }

    hash_tab_unlock(chat->session_tab, (void *)&key, RDLOCK);

    return 0;
}

/******************************************************************************
 ***函数名称: chat_del_sub
 **功    能: 取消订阅消息
 **输入参数:
 **     chat: TAB对象
 **     sid: 会话ID
 **     cmd: 命令类型
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
int chat_del_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd)
{
    chat_session_t *ssn, key;
    chat_sub_item_t sub_key, *item;

    /* > 查找会话对象 */
    key.sid = sid;

    ssn = (chat_session_t *)hash_tab_query(chat->session_tab, &key, RDLOCK);
    if (NULL == ssn) {
        return -1;
    }

    /* > 删除订阅信息 */
    sub_key.cmd = cmd;

    item = hash_tab_delete(ssn->sub, &sub_key, WRLOCK);
    if (NULL == item) {
        hash_tab_unlock(chat->session_tab, &key, RDLOCK);
        return  0;
    }

    hash_tab_unlock(chat->session_tab, &key, RDLOCK);

    free(item);

    return 0;
}

/******************************************************************************
 **函数名称: chat_has_sub
 **功    能: check the specified command has sub or not
 **输入参数:
 **     chat: TAB对象
 **     sid: 会话ID
 **     cmd: 命令类型
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 17:44:28 #
 ******************************************************************************/
bool chat_has_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd)
{
    bool has_sub = false;
    chat_session_t *ssn, key;

    key.sid = sid;

    ssn = (chat_session_t *)hash_tab_query(chat->session_tab, &key, RDLOCK);
    if (NULL == ssn) {
        return false;
    }

    has_sub = _chat_has_sub(ssn, cmd);

    hash_tab_unlock(chat->session_tab, &key, RDLOCK);

    return has_sub;
}

/******************************************************************************
 **函数名称: _chat_has_sub
 **功    能: 是否订阅消息
 **输入参数:
 **     ssn: 会话对象
 **     cmd: 命令类型
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
static bool _chat_has_sub(chat_session_t *ssn, uint16_t cmd)
{
    chat_sub_item_t *item, key;

    key.cmd = cmd;

    item = hash_tab_query(ssn->sub, (void *)&key, RDLOCK);
    if (NULL == item) {
        hash_tab_unlock(ssn->sub, (void *)&key, RDLOCK);
        return false;
    }
    hash_tab_unlock(ssn->sub, (void *)&key, RDLOCK);

    return true;
}

/******************************************************************************
 **函数名称: chat_room_get_clean_list
 **功    能: 获取超时的聊天室ID
 **输入参数:
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: 
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
static int chat_room_get_clean_list(chat_room_t *room, void *args)
{
    list_t *clean_list = (list_t *)args;

    if (0 == room->sid_num) {
        list_rpush(clean_list, (void *)room->rid);
    }

    return 0;
}

/******************************************************************************
 **函数名称: chat_timeout_hdl
 **功    能: 清理超时的数据
 **输入参数:
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: 
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
int chat_timeout_hdl(chat_tab_t *chat)
{
    uint64_t *rid;
    list_t *clean_list;

    clean_list = list_creat(NULL);
    if (NULL == clean_list) {
        return 0;
    }

    hash_tab_trav(chat->room_tab, (trav_cb_t)chat_room_get_clean_list, clean_list, RDLOCK);

    while (1) {
        rid = list_lpop(clean_list);
        if (NULL == rid) {
            break;
        }
        chat_del_room(chat, (uint64_t)rid);
    }

    return 0;
}

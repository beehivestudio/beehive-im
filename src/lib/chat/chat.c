#include "log.h"
#include "redo.h"
#include "chat.h"
#include "chat_priv.h"

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
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 19:59:38 #
 ******************************************************************************/
int chat_add_session(chat_tab_t *chat, uint64_t rid, uint64_t sid, uint32_t gid)
{
    chat_room_t *room, room_key;
    chat_group_t *grp = NULL;
    chat_session_t *ssn, ssn_key;

    /* > 判断该SID是否已经存在 */
    ssn_key.sid = sid;

    ssn = (chat_session_t *)rbt_query(chat->session_tab, (void *)&ssn_key);
    if (NULL != ssn) {
        if (ssn->rid == rid) {
            return ssn->gid; /* 已存在 */
        }
        return -1; /* 失败: SID所在聊天室与申请的聊天室冲突 */
    }

    /* > 判断聊天室是否存在 */
    room_key.rid = rid;

    room = (chat_room_t *)avl_query(chat->room_tab, (void *)&room_key);
    if (NULL == room) {
        room = chat_add_room(chat, rid);
        if (NULL == room) {
            return -1; /* 失败 */
        }
    }

    do {
        /* > 申请空闲的组 */
        grp = chat_group_alloc(chat, room, gid);
        if (NULL == grp) {
            break; /* 失败 */
        }

        /* > 将此SID加入分组列表 */
        ssn = chat_group_add_session(chat, room, grp, sid); 
        if (NULL == ssn) {
            break; /* 添加失败 */
        }

        /* > 构建SID索引表 */
        if (chat_session_ref(chat, ssn)) {
            chat_group_del_session(chat, room, grp, sid);
            break; /* 添加失败 */
        }
        return ssn->gid; /* 成功 */
    } while(0);

    /* > 释放内存空间 */
    if ((NULL != grp)
        && (0 == grp->sid_num))
    {
        chat_del_group(chat, room, grp);
    }

    if (0 == room->grp_num) {
        chat_del_room(chat, room);
    }

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
    chat_group_t *grp;
    chat_session_t *ssn, ssn_key;

    /* > 删除SID索引 */
    ssn_key.sid = sid;

    rbt_delete(chat->session_tab, (void *)&ssn_key, (void **)&ssn);
    if (NULL == ssn) {
        log_error(chat->log, "Didn't find sid[%u]. ptr:%p\n", sid, ssn);
        return 0; /* 此SID已删除 */
    }

    do {
        /* > 删除聊天室中的SID信息 */
        room_key.rid = ssn->rid;

        room = avl_query(chat->room_tab, (void *)&room_key);
        if (NULL == room) {
            log_error(chat->log, "Didn't find room! sid:%u rid:%u.\n", sid, ssn->rid);
            break;  /* 未找到 */
        }

        grp = chat_group_find_by_gid(chat, room, ssn->gid);
        if (NULL == grp) {
            log_error(chat->log, "Didn't find group! sid:%u rid:%u gid:%u.\n", sid, ssn->rid, ssn->gid);
            break; /* 未找到分组 */
        }

        /* > 从分组中删除SID */
        chat_group_del_session(chat, room, grp, sid);
    } while(0);

    // FREE(ssn); /* 内存已经被chat_group_del_session释放 */

    return 0;
}

/******************************************************************************
 **函数名称: chat_query_session
 **功    能: 查询SESSION对象
 **输入参数: NONE
 **输出参数: NONE
 **返    回: CHAT对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 19:59:38 #
 ******************************************************************************/
chat_session_t *chat_query_session(chat_tab_t *chat, uint64_t sid)
{
    chat_session_t key;

    key.sid = sid;

    return (chat_session_t *)rbt_query(chat->session_tab, (void *)&key);
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
    chat_group_t *grp;
    chat_room_t *room, key;

    key.rid = rid;

    room = (chat_room_t *)avl_query(chat->room_tab, (void *)&key);
    if (NULL == room
        || 0 == room->grp_num)
    {
        return 0; // 聊天室不存在
    }

    if (0 == gid) { // 遍历聊天室所有成员
        return chat_room_trav_all_group(chat, room, proc, args);
    }

    // 只发给同一聊天室分组的成员
    grp = chat_group_find_by_gid(chat, room, gid);
    if (NULL == grp) {
        return 0;
    }

    return rbt_trav(grp->sid_list, (trav_cb_t)proc, args);
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

    key.sid = sid;

    ssn = (chat_session_t *)rbt_query(chat->session_tab, &key);
    if (NULL == ssn) {
        return -1;
    }

    item = (chat_sub_item_t *)calloc(1, sizeof(chat_sub_item_t));
    if (NULL == item) {
        return -1;
    }

    item->cmd = cmd;

    ret = avl_insert(ssn->sub, (void *)item);
    if (AVL_OK != ret) {
        FREE(item);
        return (AVL_NODE_EXIST == ret)? 0 : -1;
    }

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
 **作    者: # yunchuan.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
int chat_del_sub(chat_tab_t *chat, uint64_t sid, uint16_t cmd)
{
    int ret;
    chat_sub_item_t *item;
    chat_session_t *ssn, key;

    key.sid = sid;

    ssn = (chat_session_t *)rbt_query(chat->session_tab, &key);
    if (NULL == ssn) {
        return -1;
    }

    item = (chat_sub_item_t *)calloc(1, sizeof(chat_sub_item_t));
    if (NULL == item) {
        return -1;
    }

    item->cmd = cmd;

    ret = avl_delete(ssn->sub, &key, (void*)&item);
    if (AVL_OK != ret) {
        FREE(item);
        return  -1;
    }

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
    chat_session_t *ssn, key;

    key.sid = sid;

    ssn = (chat_session_t *)rbt_query(chat->session_tab, &key);
    if (NULL == ssn) {
        return false;
    }

    return  _chat_has_sub(ssn, cmd);
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
bool _chat_has_sub(chat_session_t *ssn, uint16_t cmd)
{
    chat_sub_item_t *item, key;

    key.cmd = cmd;

    item = avl_query(ssn->sub, (void *)&key);

    return (NULL == item)? false : true;
}

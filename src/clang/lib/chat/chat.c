#include "log.h"
#include "redo.h"
#include "chat.h"
#include "atomic.h"
#include "chat_priv.h"

static bool _chat_has_sub(chat_session_t *ssn, uint16_t cmd);

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
 **函数名称: chat_room_add_session
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
 **     2. 如果此SID不存在, 加入聊天室
 **注意事项: 
 **     1. 在此处不删除人数为0的聊天室和其下的分组, 删除操作由定时任务统一处理.
 **        理由: 降低程序的复杂度.
 **     2. 尽量进少使用写锁的次数, 尽量降低写锁的粒度.
 **     3. 防止锁出现交错的情况, 从而造成死锁的情况.
 **作    者: # Qifeng.zou # 2016.09.20 10:53:18 #
 ******************************************************************************/
uint32_t chat_room_add_session(chat_tab_t *chat, uint64_t rid, uint32_t gid, uint64_t sid)
{
    chat_session_t *ssn, key;

    /* > 判断该SID是否已经存在 */
    key.sid = sid;

    ssn = (chat_session_t *)hash_tab_query(chat->session_tab, (void *)&key, RDLOCK);
    if (NULL != ssn) {
        if (ssn->rid == rid) {
            gid = ssn->gid;
            hash_tab_unlock(chat->session_tab, (void *)&key, RDLOCK);
            return gid; /* 已存在 */
        }
        hash_tab_unlock(chat->session_tab, (void *)&key, RDLOCK);
        return -1; /* 失败: SID所在聊天室与申请的聊天室冲突 */
    }

    /* > 将会话加入聊天室 */
    if (_chat_room_add_session(chat, rid, gid, sid)) {
        log_error(chat->log, "Chat room add sid failed. rid:%lu gid:%u sid:%lu",
                rid, gid, sid);
        return -1;
    }

    /* > 构建SID索引表 */
    if (chat_session_tab_add(chat, rid, gid, sid)) {
        _chat_room_del_session(chat, rid, gid, sid);
        log_error(chat->log, "Add sid into session table failed. rid:%lu gid:%u sid:%lu",
                rid, gid, sid);
        return -1;
    }

    return 0;
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
    chat_session_t *ssn, key;

    /* > 删除SID索引 */
    key.sid = sid;

    ssn = hash_tab_delete(chat->session_tab, (void *)&key, WRLOCK);
    if (NULL == ssn) {
        log_error(chat->log, "Didn't find sid[%u]. ptr:%p", sid, ssn);
        return 0; /* Didn't find */
    }

    /* > 从聊天室剔除 */
    _chat_room_del_session(chat, ssn->rid, ssn->gid, sid);

    /* > 释放会话对象 */
    hash_tab_destroy(ssn->sub, (mem_dealloc_cb_t)mem_dealloc, NULL);
    FREE(ssn);
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
    chat_room_t *room, key;

    key.rid = rid;

    room = (chat_room_t *)hash_tab_query(chat->room_tab, (void *)&key, RDLOCK);
    if (NULL == room) {
        return 0; /* Didn't find */
    }

    chat_group_trav(chat, room, gid, proc, args);

    hash_tab_unlock(chat->room_tab, (void *)&key, RDLOCK);

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
 **     加入链表的是RID, 而不是room对象. 否则可能出现程序CRASH的现象.
 **     原因: 加入链表后, 其他线程可能释放room空间.
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
 **函数名称: chat_clean_hdl
 **功    能: 清理数据
 **输入参数:
 **     chat: CHAT对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 15:48:28 #
 ******************************************************************************/
int chat_clean_hdl(chat_tab_t *chat)
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

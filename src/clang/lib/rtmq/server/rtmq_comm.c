#include "redo.h"
#include "rtmq_comm.h"
#include "rtmq_recv.h"

/******************************************************************************
 **函数名称: rtmq_conf_isvalid
 **功    能: 校验配置合法性
 **输入参数:
 **     conf: 配置数据
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.28 16:20:40 #
 ******************************************************************************/
bool rtmq_conf_isvalid(const rtmq_conf_t *conf)
{
    if ((0 == conf->nid)
        || (0 == strlen(conf->path))
        || (NULL == conf->auth)
        || ((0 == conf->port) || (conf->port > 65535))
        || (0 == conf->recv_thd_num)
        || (0 == conf->work_thd_num)
        || (0 == conf->recvq_num)
        || (0 == conf->distq_num)
        || ((0 == conf->recvq.max) || (0 == conf->recvq.size))
        || ((0 == conf->sendq.max) || (0 == conf->sendq.size))
        || ((0 == conf->distq.max) || (0 == conf->distq.size))) {
        return false;
    }
    return true;
}

/******************************************************************************
 **函数名称: rtmq_cmd_to_rsvr
 **功    能: 发送命令到指定的接收线程
 **输入参数:
 **     ctx: 全局对象
 **     cmd_sck_id: 命令套接字
 **     cmd: 处理命令
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 随机选择接收线程
 **     2. 发送命令至接收线程
 **注意事项: 如果发送失败，最多重复3次发送!
 **作    者: # Qifeng.zou # 2015.01.09 #
 ******************************************************************************/
int rtmq_cmd_to_rsvr(rtmq_cntx_t *ctx, int cmd_sck_id, const rtmq_cmd_t *cmd, int idx)
{
    char path[FILE_PATH_MAX_LEN];

    rtmq_rsvr_usck_path(&ctx->conf, path, idx);

    /* 发送命令至接收线程 */
    if (unix_udp_send(cmd_sck_id, path, cmd, sizeof(rtmq_cmd_t)) < 0) {
        if (EAGAIN != errno) {
            log_error(ctx->log, "errmsg:[%d] %s! path:%s type:%d",
                      errno, strerror(errno), path, cmd->type);
        }
        return RTMQ_ERR;
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_link_auth_check
 **功    能: 链路鉴权检测
 **输入参数:
 **     ctx: 全局对象
 **     auth: 鉴权请求
 **输出参数: NONE
 **返    回: succ:成功 fail:失败
 **实现描述: 检测用户名和密码是否正确
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.22 #
 ******************************************************************************/
int rtmq_link_auth_check(rtmq_cntx_t *ctx, rtmq_link_auth_req_t *auth)
{
    return rtmq_auth_check(ctx, auth->usr, auth->passwd)?
        RTMQ_LINK_AUTH_SUCC : RTMQ_LINK_AUTH_FAIL;
}

/******************************************************************************
 **函数名称: rtmq_node_to_svr_map_init
 **功    能: 创建NODE与SVR的映射表
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 构建平衡二叉树
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.30 20:29:26 #
 ******************************************************************************/
static int rtmq_node_to_svr_map_cmp_cb(
        const rtmq_node_to_svr_map_t *map1, const rtmq_node_to_svr_map_t *map2)
{
    return (map1->nid - map2->nid);
}

int rtmq_node_to_svr_map_init(rtmq_cntx_t *ctx)
{
    /* > 创建映射表 */
    ctx->node_to_svr_map = avl_creat(NULL, (cmp_cb_t)rtmq_node_to_svr_map_cmp_cb);
    if (NULL == ctx->node_to_svr_map) {
        log_error(ctx->log, "Initialize dev->svr map failed!");
        return RTMQ_ERR;
    }

    /* > 初始化读写锁 */
    pthread_rwlock_init(&ctx->node_to_svr_map_lock, NULL);

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_node_to_svr_map_add
 **功    能: 添加NODE->SVR映射
 **输入参数:
 **     ctx: 全局对象
 **     nid: 结点ID(主键)
 **     rsvr_id: 接收服务索引
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 注册NODEID与RSVR的映射关系, 为自定义数据的应答做铺垫!
 **作    者: # Qifeng.zou # 2015.05.30 #
 ******************************************************************************/
int rtmq_node_to_svr_map_add(rtmq_cntx_t *ctx, int nid, int rsvr_id)
{
    rtmq_node_to_svr_map_t *map, key;

    key.nid = nid;

    pthread_rwlock_wrlock(&ctx->node_to_svr_map_lock); /* 加锁 */

    /* > 查找是否已经存在 */
    map = avl_query(ctx->node_to_svr_map, &key);
    if (NULL == map) {
        map = (rtmq_node_to_svr_map_t *)calloc(1, sizeof(rtmq_node_to_svr_map_t));
        if (NULL == map) {
            pthread_rwlock_unlock(&ctx->node_to_svr_map_lock); /* 解锁 */
            log_error(ctx->log, "Alloc memory failed!");
            return RTMQ_ERR;
        }

        map->num = 0;
        map->nid = nid;

        if (avl_insert(ctx->node_to_svr_map, (void *)map)) {
            pthread_rwlock_unlock(&ctx->node_to_svr_map_lock); /* 解锁 */
            FREE(map);
            log_error(ctx->log, "Insert into dev2sck table failed! nid:%d rsvr_id:%d",
                      nid, rsvr_id);
            return RTMQ_ERR;
        }
    }

    /* > 插入NODE -> SVR列表 */
    if (map->num >= RTRD_NODE_TO_SVR_MAX_LEN) {
        pthread_rwlock_unlock(&ctx->node_to_svr_map_lock); /* 解锁 */
        log_error(ctx->log, "Node to svr map is full! nid:%d", nid);
        return RTMQ_ERR;
    }

    map->rsvr_id[map->num++] = rsvr_id; /* 插入 */

    pthread_rwlock_unlock(&ctx->node_to_svr_map_lock); /* 解锁 */

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_node_to_svr_map_del
 **功    能: 删除NODE -> SVR映射
 **输入参数:
 **     ctx: 全局对象
 **     nid: 结点ID
 **     rsvr_id: 接收服务索引
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 从链表中找出sck_serial结点, 并删除!
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.30 22:25:20 #
 ******************************************************************************/
int rtmq_node_to_svr_map_del(rtmq_cntx_t *ctx, int nid, int rsvr_id)
{
    int idx;
    rtmq_node_to_svr_map_t *map, key;

    key.nid = nid;

    pthread_rwlock_wrlock(&ctx->node_to_svr_map_lock);

    /* > 查找映射表 */
    map = avl_query(ctx->node_to_svr_map, &key);
    if (NULL == map) {
        pthread_rwlock_unlock(&ctx->node_to_svr_map_lock);
        log_error(ctx->log, "Query nid [%d] failed!", nid);
        return RTMQ_ERR;
    }

    /* > 删除处理 */
    for (idx=0; idx<map->num; ++idx) {
        if (map->rsvr_id[idx] == rsvr_id) {
            map->rsvr_id[idx] = map->rsvr_id[--map->num]; /* 删除:使用最后一个值替代当前值 */
            if (0 == map->num) {
                avl_delete(ctx->node_to_svr_map, &key, (void *)&map);
                FREE(map);
            }
            break;
        }
    }

    pthread_rwlock_unlock(&ctx->node_to_svr_map_lock);
    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_node_to_svr_map_rand
 **功    能: 随机选择NODE -> SVR映射
 **输入参数:
 **     ctx: 全局对象
 **     nid: 结点ID
 **输出参数: NONE
 **返    回: 接收线程索引
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.30 22:25:20 #
 ******************************************************************************/
int rtmq_node_to_svr_map_rand(rtmq_cntx_t *ctx, int nid)
{
    int rsvr_id;
    rtmq_node_to_svr_map_t *map, key;

    key.nid = nid;

    pthread_rwlock_rdlock(&ctx->node_to_svr_map_lock);

    /* > 获取映射表 */
    map = avl_query(ctx->node_to_svr_map, &key);
    if (NULL == map) {
        pthread_rwlock_unlock(&ctx->node_to_svr_map_lock);
        log_error(ctx->log, "Query nid [%d] failed!", nid);
        return -1;
    }

    /* > 选择服务ID */
    rsvr_id = map->rsvr_id[rand() % map->num]; /* 随机选择 */

    pthread_rwlock_unlock(&ctx->node_to_svr_map_lock);

    return rsvr_id;
}

////////////////////////////////////////////////////////////////////////////////
// 订阅表的操作

/* 订阅哈希回调 */
static uint64_t rtmq_sub_tab_hash_cb(const rtmq_sub_list_t *list)
{
    return list->type;
}

/* 订阅比较回调 */
static int rtmq_sub_tab_cmp_cb(const rtmq_sub_list_t *list1, const rtmq_sub_list_t *list2)
{
    return (list1->type - list2->type);
}

/******************************************************************************
 **函数名称: rtmq_sub_init
 **功    能: 初始化订阅表
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 使用hash表管理订阅列表
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.28 22:11:53 #
 ******************************************************************************/
int rtmq_sub_init(rtmq_cntx_t *ctx)
{
    /* > 创建订阅表 */
    ctx->sub = hash_tab_creat(100,
            (hash_cb_t)rtmq_sub_tab_hash_cb,
            (cmp_cb_t)rtmq_sub_tab_cmp_cb, NULL);
    if (NULL == ctx->sub) {
        return RTMQ_ERR;
    }

    return RTMQ_OK;
}

/* 订阅分组比较 */
static int rtmq_sub_group_cmp_cb(rtmq_sub_group_t *g1, rtmq_sub_group_t *g2)
{
    return (int)(g1->gid - g2->gid);
}

/******************************************************************************
 **函数名称: rtmq_sub_list_alloc
 **功    能: 申请订阅列表
 **输入参数:
 **     type: 消息类型
 **输出参数: NONE
 **返    回: 订阅列表
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.04.09 07:07:26 #
 ******************************************************************************/
static rtmq_sub_list_t *rtmq_sub_list_alloc(uint32_t type)
{
    rtmq_sub_list_t *list;

    list = (rtmq_sub_list_t *)calloc(1, sizeof(rtmq_sub_list_t));
    if (NULL == list) {
        return NULL;
    }

    list->groups = avl_creat(NULL, (cmp_cb_t)rtmq_sub_group_cmp_cb);
    if (NULL == list->groups) {
        free(list);
        return NULL;
    }

    list->type = type;

    return list;
}

/* 释放订阅列表空间 */
static int rtmq_sub_group_trav_dealloc_cb(void *data, void *args)
{
    rtmq_sub_group_t *group = (rtmq_sub_group_t *)data;

    vector_destroy(group->nodes, mem_dealloc, NULL);

    return 0;
}

/******************************************************************************
 **函数名称: rtmq_sub_list_dealloc
 **功    能: 释放订阅列表
 **输入参数:
 **     list: 订阅列表
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 回收订阅列表的所有空间
 **注意事项:
 **作    者: # Qifeng.zou # 2016.04.09 07:07:26 #
 ******************************************************************************/
void rtmq_sub_list_dealloc(rtmq_sub_list_t *list)
{
    avl_trav(list->groups, rtmq_sub_group_trav_dealloc_cb, NULL);
    avl_destroy(list->groups, mem_dealloc, NULL);
    free(list);
}

/******************************************************************************
 **函数名称: rtmq_sub_group_alloc
 **功    能: 申请订阅组
 **输入参数:
 **     type: 消息类型
 **输出参数: NONE
 **返    回: 订阅列表
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.01 00:20:41 #
 ******************************************************************************/
static rtmq_sub_group_t *rtmq_sub_group_alloc(uint32_t gid)
{
    rtmq_sub_group_t *group;

    /* 1. 创建group对象 */
    group = (rtmq_sub_group_t *)calloc(1, sizeof(rtmq_sub_group_t));
    if (NULL == group) {
        return NULL;
    }

    group->gid = gid;

    /* 2. 创建结点列表 */
    group->nodes = vector_creat(512, 128);
    if (NULL == group->nodes) {
        free(group);
        return NULL;
    }

    return group;
}

/******************************************************************************
 **函数名称: rtmq_sub_group_dealloc
 **功    能: 释放订阅组
 **输入参数:
 **     group: 订阅分组
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 回收订阅分组对象的空间
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.01 00:20:41 #
 ******************************************************************************/
void rtmq_sub_group_dealloc(rtmq_sub_group_t *group)
{
    vector_destroy(group->nodes, mem_dealloc, NULL);
    free(group);
}

/******************************************************************************
 **函数名称: rtmq_sub_node_alloc
 **功    能: 申请订阅结点
 **输入参数:
 **     type: 消息类型
 **输出参数: NONE
 **返    回: 订阅结点
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.01 00:40:15 #
 ******************************************************************************/
static rtmq_sub_node_t *rtmq_sub_node_alloc(uint32_t nid, uint64_t sid)
{
    rtmq_sub_node_t *node;

    node = (rtmq_sub_node_t *)calloc(1, sizeof(rtmq_sub_node_t));
    if (NULL == node) {
        return NULL;
    }

    node->sid = sid;
    node->nid = nid;

    return node;
}

/******************************************************************************
 **函数名称: rtmq_sub_add
 **功    能: 添加订阅列表
 **输入参数:
 **     ctx: 全局对象
 **     sck: 连接对象
 **     type: 订阅消息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.04.09 07:07:26 #
 ******************************************************************************/
int rtmq_sub_add(rtmq_cntx_t *ctx, rtmq_sck_t *sck, int type)
{
    rtmq_sub_node_t *node;
    rtmq_sub_list_t *list, key;
    rtmq_sub_group_t *group, gkey;

    /* 1. 根据type查找/添加订阅列表 */
QUERY_SUB_TAB:
    key.type = type;

    list = (rtmq_sub_list_t *)hash_tab_query(ctx->sub, (void *)&key, WRLOCK);
    if (NULL == list) {
        list = (rtmq_sub_list_t *)rtmq_sub_list_alloc(type);
        if (NULL == list) {
            log_error(ctx->log, "Alloc sub list failed! type:0x%04X", type);
            return RTMQ_ERR;
        }

        if (hash_tab_insert(ctx->sub, (void *)list, WRLOCK)) {
            rtmq_sub_list_dealloc(list);
            log_error(ctx->log, "Insert sub table failed! type:0x%04X", type);
            return RTMQ_ERR;
        }
        goto QUERY_SUB_TAB;
    }

    /* 2. 根据gid查找/添加订阅列表中的分组 */
QUERY_SUB_GROUP:
    gkey.gid = sck->gid;

    group = avl_query(list->groups, &gkey);
    if (NULL == group) {
        group = rtmq_sub_group_alloc(sck->gid);
        if (NULL == group) {
            hash_tab_unlock(ctx->sub, &key, WRLOCK);
            log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
            return RTMQ_ERR;
        }

        if (avl_insert(list->groups, (void *)group)) {
            rtmq_sub_group_dealloc(group);
            goto QUERY_SUB_GROUP;
        }
        goto QUERY_SUB_GROUP;
    }

    /* 3. 根据sid确定是否连接已经订阅此消息 */
    node = vector_find(group->nodes,
            (find_cb_t)rtmq_sub_group_find_sid_cb, (void *)&sck->sid);
    if (NULL != node) {
        hash_tab_unlock(ctx->sub, &key, WRLOCK);
        return RTMQ_OK; /* 已订阅 */
    }

    /* 4. 将连接加入订阅列表的分组中... */
    node = rtmq_sub_node_alloc(sck->nid, sck->sid);
    if (NULL == node) {
        hash_tab_unlock(ctx->sub, &key, WRLOCK);
        log_error(ctx->log, "Alloc sub node failed! nid:%d sid:%d", sck->nid, sck->sid);
        return RTMQ_ERR;
    }

    if (vector_append(group->nodes, (void *)node)) {
        hash_tab_unlock(ctx->sub, &key, WRLOCK);
        rtmq_sub_node_dealloc(node);
        log_error(ctx->log, "Add sub node failed! nid:%d sid:%d", sck->nid, sck->sid);
        return RTMQ_ERR;
    }

    hash_tab_unlock(ctx->sub, &key, WRLOCK);

    log_debug(ctx->log, "Add sub success! type:0x%04X gid:%u nid:%u",
            type, sck->gid, sck->nid);
    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_sub_del
 **功    能: 删除订阅数据
 **输入参数:
 **     ctx: 全局对象
 **     sck: 连接对象
 **     type: 订阅消息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.04.09 07:07:26 #
 ******************************************************************************/
int rtmq_sub_del(rtmq_cntx_t *ctx, rtmq_sck_t *sck, int type)
{
    rtmq_sub_node_t *node;
    rtmq_sub_list_t *list, key;
    rtmq_sub_group_t *group, gkey;

    /* 1. 查询订阅列表 */
    key.type = type;
    list = (rtmq_sub_list_t *)hash_tab_query(ctx->sub, &key, WRLOCK);
    if (NULL == list) {
        return 0; /* 无数据 */
    }

    /* 2. 查询订阅列表group分组 */
    gkey.gid = sck->gid;
    group = avl_query(list->groups, &gkey); 
    if (NULL == group) {
        return 0; /* 无数据 */
    }

    /* 3. 从订阅列表group分组中删除指定连接 */
    node = vector_find_and_del(group->nodes,
            (find_cb_t)rtmq_sub_group_find_sid_cb, (void *)&sck->sid);
    if (NULL == node) {
        hash_tab_unlock(ctx->sub, &key, WRLOCK);
        return 0; /* 未订阅 */
    }

    /* 4. 回收内存空间 */
    rtmq_sub_node_dealloc(node);

    if (0 == vector_len(group->nodes)) {
        avl_delete(list->groups, &gkey, (void **)&group);
        rtmq_sub_group_dealloc(group);
        if (0 == avl_num(list->groups)) {
            hash_tab_delete(ctx->sub, &key, NONLOCK);
            rtmq_sub_list_dealloc(list);
        }
    }
    hash_tab_unlock(ctx->sub, &key, WRLOCK);

    return 0;
}

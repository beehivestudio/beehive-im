#include "sdk.h"
#include "mesg.h"
#include "redo.h"
#include "rb_tree.h"

/******************************************************************************
 **函数名称: sdk_creat_workers
 **功    能: 创建工作线程线程池
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.08.19 #
 ******************************************************************************/
int sdk_creat_workers(sdk_cntx_t *ctx)
{
    int idx;
    sdk_worker_t *worker;
    sdk_conf_t *conf = &ctx->conf;

    /* > 创建对象 */
    worker = (sdk_worker_t *)calloc(conf->work_thd_num, sizeof(sdk_worker_t));
    if (NULL == worker) {
        return SDK_ERR;
    }

    /* > 创建线程池 */
    ctx->worktp = thread_pool_init(conf->work_thd_num, NULL, (void *)worker);
    if (NULL == ctx->worktp) {
        free(worker);
        return SDK_ERR;
    }

    /* > 初始化线程 */
    for (idx=0; idx<conf->work_thd_num; ++idx) {
        if (sdk_worker_init(ctx, worker+idx, idx)) {
            free(worker);
            thread_pool_destroy(ctx->worktp);
            return SDK_ERR;
        }
    }

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_creat_sends
 **功    能: 创建发送线程线程池
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.08.19 #
 ******************************************************************************/
int sdk_creat_sends(sdk_cntx_t *ctx)
{
    /* > 创建对象 */
    ctx->ssvr = (sdk_ssvr_t *)calloc(1, sizeof(sdk_ssvr_t));
    if (NULL == ctx->ssvr) {
        log_error(ctx->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return SDK_ERR;
    }

    /* > 创建线程池 */
    ctx->sendtp = thread_pool_init(1, NULL, (void *)ctx->ssvr);
    if (NULL == ctx->sendtp) {
        log_error(ctx->log, "Initialize thread pool failed!");
        FREE(ctx->ssvr);
        return SDK_ERR;
    }

    /* > 初始化线程 */
    if (sdk_ssvr_init(ctx, ctx->ssvr)) {
        log_fatal(ctx->log, "Initialize send thread failed!");
        FREE(ctx->ssvr);
        thread_pool_destroy(ctx->sendtp);
        return SDK_ERR;
    }

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_lock_server
 **功    能: 锁住指定路径(注: 防止路径和结点ID相同的配置)
 **输入参数:
 **     conf: 配置信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 文件描述符可不用关闭
 **作    者: # Qifeng.zou # 2016.05.02 21:14:39 #
 ******************************************************************************/
int sdk_lock_server(const sdk_conf_t *conf)
{
    int fd;
    char path[FILE_NAME_MAX_LEN];

    sdk_lock_path(conf, path);

    Mkdir2(path, DIR_MODE);

    fd = Open(path, O_CREAT|O_RDWR, OPEN_MODE);
    if (fd < 0) {
        return -1;
    }

    if (proc_try_wrlock(fd)) {
        close(fd);
        return -1;
    }
    return 0;
}

/******************************************************************************
 **函数名称: sdk_cli_cmd_send_req
 **功    能: 通知Send服务线程
 **输入参数:
 **     ctx: 上下文信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
int sdk_cli_cmd_send_req(sdk_cntx_t *ctx)
{
    sdk_cmd_t cmd;
    char path[FILE_NAME_MAX_LEN];
    sdk_conf_t *conf = &ctx->conf;

    memset(&cmd, 0, sizeof(cmd));

    cmd.type = SDK_CMD_SEND_ALL;
    sdk_ssvr_usck_path(conf, path);

    if (spin_trylock(&ctx->cmd_sck_lck)) {
        log_debug(ctx->log, "Try lock failed!");
        return SDK_OK;
    }

    if (unix_udp_send(ctx->cmd_sck_id, path, &cmd, sizeof(cmd)) < 0) {
        spin_unlock(&ctx->cmd_sck_lck);
        if (EAGAIN != errno) {
            log_debug(ctx->log, "errmsg:[%d] %s! path:%s", errno, strerror(errno), path);
        }
        return SDK_ERR;
    }

    spin_unlock(&ctx->cmd_sck_lck);

    return SDK_OK;
}

/******************************************************************************
 **函数名称: sdk_creat_cmd_usck
 **功    能: 创建命令套接字
 **输入参数:
 **     ctx: 上下文信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
int sdk_creat_cmd_usck(sdk_cntx_t *ctx)
{
    char path[FILE_NAME_MAX_LEN];

    sdk_comm_usck_path(&ctx->conf, path);

    spin_lock_init(&ctx->cmd_sck_lck);
    ctx->cmd_sck_id = unix_udp_creat(path);
    if (ctx->cmd_sck_id < 0) {
        log_error(ctx->log, "errmsg:[%d] %s! path:%s", errno, strerror(errno), path);
        return SDK_ERR;
    }

    return SDK_OK;
}



////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: sdk_queue_init
 **功    能: 队列初始化(内部接口)
 **输入参数:
 **     q: 队列
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:47:09 #
 ******************************************************************************/
int sdk_queue_init(sdk_queue_t *q)
{
    pthread_mutex_init(&q->lock, NULL);
    q->list = list_creat(NULL);
    if (NULL == q->list) {
        pthread_mutex_destroy(&q->lock);
        return -1;
    }

    return 0;
}

/******************************************************************************
 **函数名称: sdk_queue_length
 **功    能: 获取队列长度
 **输入参数:
 **     q: 队列
 **输出参数: NONE
 **返    回: 队列长度
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:47:16 #
 ******************************************************************************/
int sdk_queue_length(sdk_queue_t *q)
{
    int len;

    pthread_mutex_lock(&q->lock);
    len = list_length(q->list);
    pthread_mutex_unlock(&q->lock);

    return len;
}

/******************************************************************************
 **函数名称: sdk_queue_rpush
 **功    能: 插入队尾
 **输入参数:
 **     q: 队列
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:47:23 #
 ******************************************************************************/
int sdk_queue_rpush(sdk_queue_t *q, void *data)
{
    int ret;

    pthread_mutex_lock(&q->lock);
    ret = list_lpush(q->list, data);
    pthread_mutex_unlock(&q->lock);

    return ret;
}

/******************************************************************************
 **函数名称: sdk_queue_lpop
 **功    能: 弹队头
 **输入参数:
 **     q: 队列
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:47:23 #
 ******************************************************************************/
void *sdk_queue_lpop(sdk_queue_t *q)
{
    void *data;

    pthread_mutex_lock(&q->lock);
    data = list_lpop(q->list);
    pthread_mutex_unlock(&q->lock);

    return data;
}

/******************************************************************************
 **函数名称: sdk_queue_remove
 **功    能: 移除某数据
 **输入参数:
 **     q: 队列
 **     data: 数据指针
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 14:34:11 #
 ******************************************************************************/
int sdk_queue_remove(sdk_queue_t *q, void *data)
{
    int ret;

    pthread_mutex_lock(&q->lock);
    ret = list_remove(q->list, data);
    pthread_mutex_unlock(&q->lock);

    return ret;
}

/******************************************************************************
 **函数名称: sdk_queue_empty
 **功    能: 队列是否为空
 **输入参数:
 **     q: 队列
 **输出参数: NONE
 **返    回: true:空 false:非空
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:52:58 #
 ******************************************************************************/
bool sdk_queue_empty(sdk_queue_t *q)
{
    bool ret;

    pthread_mutex_lock(&q->lock);
    ret = list_empty(q->list);
    pthread_mutex_unlock(&q->lock);

    return ret;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 比较回调 */
static int sdk_send_mgr_cmp_cb(sdk_send_item_t *item1, sdk_send_item_t *item2)
{
    return item1->seq - item2->seq;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_init
 **功    能: 初始化发送管理
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:06:53 #
 ******************************************************************************/
int sdk_send_mgr_init(sdk_cntx_t *ctx)
{
    sdk_send_mgr_t *mgr = &ctx->mgr;

    mgr->seq = 0;
    pthread_rwlock_init(&mgr->lock, NULL);
    mgr->tab = rbt_creat(NULL, (cmp_cb_t)sdk_send_mgr_cmp_cb);
    if (NULL == mgr->tab) {
        return -1;
    }

    return 0;
}

/******************************************************************************
 **函数名称: sdk_gen_seq
 **功    能: 生成序列号
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 序列号
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.09 17:52:58 #
 ******************************************************************************/
uint64_t sdk_gen_seq(sdk_cntx_t *ctx)
{
    uint64_t seq;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    seq = atomic64_inc(&mgr->seq);

    return seq;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_insert
 **功    能: 新增发送项
 **输入参数:
 **     ctx: 全局对象
 **     item: 发送项
 **     lock: 加锁类型
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 外部需要主动调用sdk_send_mgr_unlock()进行解锁操作
 **作    者: # Qifeng.zou # 2016.11.10 10:09:45 #
 ******************************************************************************/
int sdk_send_mgr_insert(sdk_cntx_t *ctx, sdk_send_item_t *item, lock_e lock)
{
    sdk_send_mgr_t *mgr = &ctx->mgr;

    pthread_rwlock_wrlock(&mgr->lock);

    return rbt_insert(mgr->tab, item);
}

/******************************************************************************
 **函数名称: sdk_send_mgr_delete
 **功    能: 删除发送项
 **输入参数:
 **     ctx: 全局对象
 **     seq: 序列号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:09:45 #
 ******************************************************************************/
int sdk_send_mgr_delete(sdk_cntx_t *ctx, uint64_t seq)
{
    sdk_send_item_t key, *item;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    memset(&key, 0, sizeof(key));

    key.seq = seq;

    pthread_rwlock_wrlock(&mgr->lock);
    rbt_delete(mgr->tab, (void *)&key, (void **)&item);
    pthread_rwlock_unlock(&mgr->lock);
    if (NULL == item) {
        return -1;
    }

    free(item->data);
    free(item);

    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_item_clean_timeout_hdl
 **功    能: 超时发送项的处理
 **输入参数:
 **     ctx: 全局对象
 **     item: 发送单元
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 15:04:50 #
 ******************************************************************************/
static int sdk_send_item_clean_timeout_hdl(sdk_cntx_t *ctx, sdk_send_item_t *item)
{
    void *data;
    sdk_send_item_t key, *temp;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    memset(&key, 0, sizeof(key));

    data = (void *)(item->data + sizeof(mesg_header_t));

    switch (item->stat) {
        case SDK_STAT_IN_SENDQ:     /* 依然在发送队列 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            sdk_queue_remove(&ctx->sendq, item->data);
            item->stat = SDK_STAT_SEND_TIMEOUT; /* 发送超时 */
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
        case SDK_STAT_SENDING:      /* 正在发送中...(无法撤回) */
            /* 注意: 由于数据在wiov中, 因此此处绝对禁止对数据进行释放操作 */
            return SDK_OK;
        case SDK_STAT_SEND_SUCC:    /* 已发送成功 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            item->stat = SDK_STAT_ACK_TIMEOUT;
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
        case SDK_STAT_SEND_FAIL:    /* 发送失败 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
        case SDK_STAT_SEND_TIMEOUT: /* 发送超时 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
        case SDK_STAT_ACK_SUCC:     /* 已收到ACK应答 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
        case SDK_STAT_ACK_TIMEOUT:  /* ACK超时 */
        case SDK_STAT_UNKNOWN:      /* 未知状态 */
            key.seq = item->seq;
            rbt_delete(mgr->tab, (void *)&key, (void **)&temp);
            if (item->cb) {
                item->cb(item->cmd, data, item->len, NULL, 0, item->stat, item->param);
            }
            FREE(item->data);
            FREE(item);
            return SDK_OK;
    }

    return SDK_OK;
}

/* 获取超时发送项 */
static int sdk_send_mgr_trav_timeout_cb(sdk_send_item_t *item, list_t *list)
{
    time_t tm = time(NULL);

    if (tm < item->ttl) {
        return 0; /* 未超时 */
    }
    else if (SDK_STAT_SENDING == item->stat) {
        return 0;
    }

    list_rpush(list, item);

    return 0;
}

/* 计算下一次遍历发送管理表的时间 */
static int sdk_send_mgr_update_next_trav_tm_cb(sdk_send_item_t *item, time_t *next_trav_tm)
{
    *next_trav_tm = (*next_trav_tm < item->ttl)? *next_trav_tm : item->ttl;
    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_trav
 **功    能: 遍历发送项
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:23:34 #
 ******************************************************************************/
int sdk_send_mgr_trav(sdk_cntx_t *ctx)
{
    list_t *list;
    sdk_send_item_t *item;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    log_debug(ctx->log, "Trav send item table!");

    mgr->next_trav_tm = time(NULL) + SDK_SLEEP_MAX_SEC;

    list = list_creat(NULL);
    if (NULL == list) {
        log_error(ctx->log, "Create timeout list failed!");
        return -1;
    }

    pthread_rwlock_wrlock(&mgr->lock);
    /* > 清除已超时的数据列表 */
    rbt_trav(mgr->tab, (trav_cb_t)sdk_send_mgr_trav_timeout_cb, (void *)list);
    while (1) {
        item = list_lpop(list);
        if (NULL == item) {
            break;
        }

        log_debug(ctx->log, "Clean timeout item! cmd:%d seq:%d", item->cmd, item->seq);

        sdk_send_item_clean_timeout_hdl(ctx, item);
    }
    list_destroy(list, NULL, mem_dummy_dealloc);

    /* > 查找下一个超时时间 */
    rbt_trav(mgr->tab, (trav_cb_t)sdk_send_mgr_update_next_trav_tm_cb, (void *)&mgr->next_trav_tm);

    pthread_rwlock_unlock(&mgr->lock);

    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_query
 **功    能: 更新发送项
 **输入参数:
 **     ctx: 全局对象
 **     seq: 序列号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:09:45 #
 ******************************************************************************/
sdk_send_item_t *sdk_send_mgr_query(sdk_cntx_t *ctx, uint64_t seq, lock_e lock)
{
    sdk_send_item_t key, *item;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    memset(&key, 0, sizeof(key));

    key.seq = seq;

    if (RDLOCK == lock) {
        pthread_rwlock_rdlock(&mgr->lock);
    }
    else if (WRLOCK == lock) {
        pthread_rwlock_wrlock(&mgr->lock);
    }

    item = rbt_query(mgr->tab, (void *)&key);
    if (NULL == item) {
        pthread_rwlock_unlock(&mgr->lock);
        return NULL;
    }

    return item;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_unlock
 **功    能: 解锁发送项
 **输入参数:
 **     ctx: 全局对象
 **     set: 序列号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:23:34 #
 ******************************************************************************/
int sdk_send_mgr_unlock(sdk_cntx_t *ctx, lock_e lock)
{
    sdk_send_mgr_t *mgr = &ctx->mgr;

    pthread_rwlock_unlock(&mgr->lock);

    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_mgr_empty
 **功    能: 发送管理表是否为空
 **输入参数:
 **     ctx: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:23:34 #
 ******************************************************************************/
bool sdk_send_mgr_empty(sdk_cntx_t *ctx)
{
    int num;
    sdk_send_mgr_t *mgr = &ctx->mgr;

    pthread_rwlock_rdlock(&mgr->lock);
    num = rbt_num(mgr->tab);
    pthread_rwlock_unlock(&mgr->lock);

    return (bool)num;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: sdk_send_succ_hdl
 **功    能: 发送成功后的处理
 **输入参数:
 **     ctx: 全局对象
 **     set: 序列号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:23:34 #
 ******************************************************************************/
int sdk_send_succ_hdl(sdk_cntx_t *ctx, void *addr, size_t len)
{
    void *data;
    sdk_send_item_t *item;
    mesg_header_t *head = (mesg_header_t *)addr, hhead;


    MESG_HEAD_NTOH(head, &hhead);

    log_debug(ctx->log, "Send success! type:%d seq:%d", hhead.type, hhead.seq);

    item = sdk_send_mgr_query(ctx, hhead.seq, WRLOCK);
    if (NULL == item) {
        log_error(ctx->log, "Not found! type:%d seq:%d", hhead.type, hhead.seq);
        assert(0);
        return 0;
    }

    item->stat = SDK_STAT_SEND_SUCC;
    if (item->cb) {
        data = (void *)(head + 1);
        item->cb(hhead.type, data, item->len, NULL, 0, item->stat, item->param);
    }

    sdk_send_mgr_unlock(ctx, WRLOCK);

    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_fail_hdl
 **功    能: 发送失败后的处理
 **输入参数:
 **     ctx: 全局对象
 **     addr: 序列号
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.10 10:23:34 #
 ******************************************************************************/
int sdk_send_fail_hdl(sdk_cntx_t *ctx, void *addr, size_t len)
{
    void *data;
    sdk_send_item_t *item;
    mesg_header_t *head = (mesg_header_t *)addr, hhead;

    MESG_HEAD_NTOH(head, &hhead);

    log_debug(ctx->log, "Send fail! cmd:%d seq:%d", hhead.type, hhead.seq);

    /* > 更新发送状态 */
    item = sdk_send_mgr_query(ctx, hhead.seq, WRLOCK);
    if (NULL == item) {
        assert(0);
        return 0;
    }

    item->stat = SDK_STAT_SEND_FAIL;
    if (item->cb) {
        data = (void *)(head + 1);
        item->cb(hhead.type, data, item->len, NULL, 0, item->stat, item->param);
    }

    sdk_send_mgr_unlock(ctx, WRLOCK);

    /* > 删除失败数据 */
    sdk_send_mgr_delete(ctx, hhead.seq);

    return 0;
}

/******************************************************************************
 **函数名称: sdk_send_data_is_timeout_and_hdl
 **功    能: 发送超时的处理
 **输入参数:
 **     ctx: 全局对象
 **     addr: 将要被发送的数据(头+体)
 **输出参数: NONE
 **返    回: true:已超时 false:未超时
 **实现描述: 
 **注意事项: 此时协议头依然为主机字节序
 **作    者: # Qifeng.zou # 2016.11.10 11:48:21 #
 ******************************************************************************/
bool sdk_send_data_is_timeout_and_hdl(sdk_cntx_t *ctx, void *addr)
{
    void *data;
    sdk_send_item_t *item;
    mesg_header_t *head = (mesg_header_t *)addr;

    item = sdk_send_mgr_query(ctx, head->seq, WRLOCK);
    if (NULL == item) {
        return true;
    }
    else if (time(NULL) >= item->ttl) { /* 判断是否超时 */
        item->stat = SDK_STAT_SEND_TIMEOUT;
        if (item->cb) {
            data = (void *)(head + 1);
            item->cb(head->type, data, head->length, NULL, 0, item->stat, item->param);
        }
        sdk_send_mgr_unlock(ctx, WRLOCK);

        sdk_send_mgr_delete(ctx, head->seq);
        return true;
    }
    item->stat = SDK_STAT_SENDING;
    sdk_send_mgr_unlock(ctx, WRLOCK);

    return false;
}

/******************************************************************************
 **函数名称: sdk_ack_succ_hdl
 **功    能: 应答成功的处理
 **输入参数:
 **     ctx: 全局对象
 **     seq: 序列号
 **输出参数: NONE
 **返    回: true:已处理 false:未处理
 **实现描述: 
 **注意事项: 此时协议头依然为网络字节序
 **作    者: # Qifeng.zou # 2016.11.10 11:48:21 #
 ******************************************************************************/
bool sdk_ack_succ_hdl(sdk_cntx_t *ctx, uint64_t seq, void *ack)
{
    void *data;
    sdk_send_item_t key, *item;
    sdk_send_mgr_t *mgr = &ctx->mgr;
    sdk_cmd_ack_t ack_key, *ack_item;
    mesg_header_t *head = (mesg_header_t *)ack;

    ack_key.ack = head->type;

    ack_item = avl_query(ctx->cmd, &ack_key);
    if (NULL == ack_item) {
        log_error(ctx->log, "Didn't find request command! seq:%lu", seq);
        return false;
    }

    log_debug(ctx->log, "Found request command! seq:%lu head:%d", seq, sizeof(mesg_header_t));

    memset(&key, 0, sizeof(key));

    key.seq = seq;

    pthread_rwlock_wrlock(&mgr->lock);
    item = rbt_query(mgr->tab, (void *)&key);
    if (NULL == item) {
        pthread_rwlock_unlock(&mgr->lock);
        log_debug(ctx->log, "Didn't find seq! seq:%d", seq);
        return false;
    }
    else if (item->cmd == ack_item->req) {
        rbt_delete(mgr->tab, (void *)&key, (void **)&item);
    }
    pthread_rwlock_unlock(&mgr->lock);

    data = (void *)(item->data + sizeof(mesg_header_t));
    item->cb(item->cmd, data, item->len,
            ack + sizeof(mesg_header_t), head->length,
            SDK_STAT_ACK_SUCC, item->param);

    free(item->data);
    free(item);

    return true;
}

#include "comm.h"
#include "lock.h"
#include "redo.h"
#include "syscall.h"
#include "rtmq_mesg.h"
#include "rtmq_proxy.h"

static int rtmq_proxy_creat_send_cmd_fd(rtmq_proxy_t *pxy);
static int rtmq_proxy_creat_work_cmd_fd(rtmq_proxy_t *pxy);

static int rtmq_proxy_cmd_work_chan_init(rtmq_proxy_t *pxy);
static bool rtmq_proxy_conf_isvalid(const rtmq_proxy_conf_t *conf);

/******************************************************************************
 **函数名称: rtmq_proxy_creat_workers
 **功    能: 创建工作线程线程池
 **输入参数:
 **     pxy: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.08.19 #
 ******************************************************************************/
static int rtmq_proxy_creat_workers(rtmq_proxy_t *pxy)
{
    int idx;
    rtmq_worker_t *worker;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    /* > 初始化对象 */
    worker = (rtmq_worker_t *)calloc(conf->work_thd_num, sizeof(rtmq_worker_t));
    if (NULL == worker) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    for (idx=0; idx<conf->work_thd_num; ++idx) {
        if (rtmq_proxy_worker_init(pxy, worker+idx, idx)) {
            log_fatal(pxy->log, "Initialize worker object failed!");
            return RTMQ_ERR;
        }
    }

    /* > 创建线程池 */
    pxy->worktp = thread_pool_init(conf->work_thd_num, NULL, (void *)worker);
    if (NULL == pxy->worktp) {
        log_error(pxy->log, "Initialize thread pool failed!");
        return RTMQ_ERR;
    }

    return RTMQ_OK;
}

/* 创建工作线程通信管道 */
static int rtmq_proxy_creat_work_cmd_fd(rtmq_proxy_t *pxy)
{
    int idx;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    pxy->work_cmd_fd = (pipe_t *)calloc(conf->work_thd_num, sizeof(pipe_t));
    if (NULL == pxy->work_cmd_fd) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_OK;
    }

    for (idx=0; idx<conf->work_thd_num; idx+=1) {
        pipe_creat(&pxy->work_cmd_fd[idx]);
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_tsvr_init_cb
 **功    能: 初始化线程对象
 **输入参数:
 **     pxy: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 根据ip列表新建发送线程.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.07.20 11:28:40 #
 ******************************************************************************/
static int rtmq_proxy_tsvr_init_cb(iplist_item_t *item, rtmq_proxy_t *pxy)
{
    int m, idx;
    rtmq_proxy_conf_t *conf = &pxy->conf;
    rtmq_proxy_tsvr_t *ssvr = thread_pool_get_args(pxy->sendtp);

    for (m=0; m<conf->send_thd_num; ++m) {
        idx = item->idx * conf->send_thd_num + m;
        if (rtmq_proxy_tsvr_init(pxy, ssvr+idx,
                    idx, item->ipaddr, item->port,
                    pxy->sendq[idx % conf->send_thd_num],
                    &pxy->send_cmd_fd[idx % conf->send_thd_num])) {
            log_fatal(pxy->log, "Initialize send thread failed!");
            free(ssvr);
            thread_pool_destroy(pxy->sendtp);
            return RTMQ_ERR;
        }
    }
    return RTMQ_OK;
}

/* 创建发送线程通信管道 */
static int rtmq_proxy_creat_send_cmd_fd(rtmq_proxy_t *pxy)
{
    int idx;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    pxy->send_cmd_fd = (pipe_t *)calloc(conf->send_thd_num, sizeof(pipe_t));
    if (NULL == pxy->send_cmd_fd) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_OK;
    }

    for (idx=0; idx<conf->send_thd_num; idx+=1) {
        pipe_creat(&pxy->send_cmd_fd[idx]);
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_creat_senders
 **功    能: 创建发送线程线程池
 **输入参数:
 **     pxy: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 根据ip列表新建发送线程.
 **注意事项:
 **作    者: # Qifeng.zou # 2015.08.19, 2017.07.20 11:28:40 #
 ******************************************************************************/
static int rtmq_proxy_creat_senders(rtmq_proxy_t *pxy)
{
    int num, total;
    rtmq_proxy_tsvr_t *ssvr;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    num = list_length(pxy->iplist); /* IP数目 */
    total = num * conf->send_thd_num;

    /* > 创建对象 */
    ssvr = (rtmq_proxy_tsvr_t *)calloc(total, sizeof(rtmq_proxy_tsvr_t));
    if (NULL == ssvr) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 创建线程池 */
    pxy->sendtp = thread_pool_init(total, NULL, (void *)ssvr);
    if (NULL == pxy->sendtp) {
        log_error(pxy->log, "Initialize thread pool failed!");
        free(ssvr);
        return RTMQ_ERR;
    }

    /* > 初始化线程 */
    list_trav(pxy->iplist, (trav_cb_t)rtmq_proxy_tsvr_init_cb, pxy);

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_creat_recvq
 **功    能: 创建接收队列
 **输入参数:
 **     pxy: 全局对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.06.04 #
 ******************************************************************************/
static int rtmq_proxy_creat_recvq(rtmq_proxy_t *pxy)
{
    int idx;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    /* > 创建队列对象 */
    pxy->recvq = (queue_t **)calloc(conf->work_thd_num, sizeof(queue_t *));
    if (NULL == pxy->recvq) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 创建接收队列 */
    for (idx=0; idx<conf->work_thd_num; ++idx) {
        pxy->recvq[idx] = queue_creat(conf->recvq.max, conf->recvq.size);
        if (NULL == pxy->recvq[idx]) {
            log_error(pxy->log, "Create recvq failed!");
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_creat_sendq
 **功    能: 创建发送线程的发送队列
 **输入参数:
 **     pxy: 发送对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.01.01 22:32:21 #
 ******************************************************************************/
static int rtmq_proxy_creat_sendq(rtmq_proxy_t *pxy)
{
    int idx;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    /* > 创建队列对象 */
    pxy->sendq = (queue_t **)calloc(conf->send_thd_num, sizeof(queue_t *));
    if (NULL == pxy->sendq) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return RTMQ_ERR;
    }

    /* > 创建发送队列 */
    for (idx=0; idx<conf->send_thd_num; ++idx) {
        pxy->sendq[idx] = queue_creat(conf->sendq.max, conf->sendq.size);
        if (NULL == pxy->sendq[idx]) {
            log_error(pxy->log, "Create send queue failed!");
            return RTMQ_ERR;
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_init
 **功    能: 初始化发送端
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2015.05.19 #
 ******************************************************************************/
rtmq_proxy_t *rtmq_proxy_init(const rtmq_proxy_conf_t *conf, log_cycle_t *log)
{
    rtmq_proxy_t *pxy;

    /* > 判断配置合法性 */
    if (!rtmq_proxy_conf_isvalid(conf)) {
        log_error(log, "Rtmq proxy configuration is invalid!");
        return NULL;
    }

    /* > 创建对象 */
    pxy = (rtmq_proxy_t *)calloc(1, sizeof(rtmq_proxy_t));
    if (NULL == pxy) {
        log_fatal(log, "errmsg:[%d] %s!", errno, strerror(errno));
        return NULL;
    }

    pxy->log = log;

    /* > 加载配置信息 */
    memcpy(&pxy->conf, conf, sizeof(rtmq_proxy_conf_t));

    do {
        /* > 解析IP地址列表 */
        pxy->iplist = iplist_parse(conf->ipaddr);
        if (NULL == pxy->iplist) {
            log_fatal(log, "Parse iplist failed! addr:%s", conf->ipaddr);
            break;
        }

        /* > 创建处理映射表 */
        pxy->reg = avl_creat(NULL, (cmp_cb_t)rtmq_reg_cmp_cb);
        if (NULL == pxy->reg) {
            log_fatal(log, "Create register map failed!");
            break;
        }

        /* > 创建接收队列 */
        if (rtmq_proxy_creat_recvq(pxy)) {
            log_fatal(log, "Create recv-queue failed!");
            break;
        }

        /* > 创建发送队列 */
        if (rtmq_proxy_creat_sendq(pxy)) {
            log_fatal(log, "Create send queue failed!");
            break;
        }

        /* > 创建发送线程通信FD */
        if (rtmq_proxy_creat_send_cmd_fd(pxy)) {
            log_fatal(log, "Create send cmd fd failed!");
            break;
        }

        /* > 创建工作线程通信FD */
        if (rtmq_proxy_creat_work_cmd_fd(pxy)) {
            log_fatal(log, "Create work cmd fd failed!");
            break;
        }

        /* > 创建工作线程池 */
        if (rtmq_proxy_creat_workers(pxy)) {
            log_fatal(pxy->log, "Create work thread pool failed!");
            break;
        }

        /* > 创建发送线程池 */
        if (rtmq_proxy_creat_senders(pxy)) {
            log_fatal(pxy->log, "Create send thread pool failed!");
            break;
        }

        return pxy;
    } while(0);

    free(pxy);
    return NULL;
}

/******************************************************************************
 **函数名称: rtmq_proxy_launch
 **功    能: 启动发送端
 **输入参数:
 **     pxy: 全局信息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 创建工作线程池
 **     2. 创建发送线程池
 **注意事项:
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
int rtmq_proxy_launch(rtmq_proxy_t *pxy)
{
    int idx, m, n, num;
    rtmq_proxy_conf_t *conf = &pxy->conf;

    /* > 注册Worker线程回调 */
    for (idx=0; idx<conf->work_thd_num; ++idx) {
        thread_pool_add_worker(pxy->worktp, rtmq_proxy_worker_routine, pxy);
    }

    /* > 注册Send线程回调 */
    num = list_length(pxy->iplist);
    for (n=0; n<num; ++n) {
        for (m=0; m<conf->send_thd_num; ++m) {
            idx = n*conf->send_thd_num + m;
            thread_pool_add_worker(pxy->sendtp, rtmq_proxy_tsvr_routine, pxy);
        }
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_reg_add
 **功    能: 消息处理的注册接口
 **输入参数:
 **     pxy: 全局对象
 **     type: 扩展消息类型
 **     proc: 回调函数
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **     1. 只能用于注册处理扩展数据类型的处理
 **     2. 不允许重复注册
 **作    者: # Qifeng.zou # 2015.05.19 #
 ******************************************************************************/
int rtmq_proxy_reg_add(rtmq_proxy_t *pxy, int type, rtmq_reg_cb_t proc, void *param)
{
    rtmq_reg_t *item;

    item = (rtmq_reg_t *)calloc(1, sizeof(rtmq_reg_t));
    if (NULL == item) {
        log_error(pxy->log, "errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    item->type = type;
    item->proc = proc;
    item->param = param;

    if (avl_insert(pxy->reg, item)) {
        log_error(pxy->log, "Register callback failed! type:0x%04X!", type);
        free(item);
        return RTMQ_ERR_REPEAT_REG;
    }

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_cmd_send_req
 **功    能: 通知Send服务线程
 **输入参数:
 **     pxy: 上下文信息
 **     idx: 发送服务ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 只要发送的数据小于4096字节, 就"无需"加锁.
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
static int rtmq_proxy_cmd_send_req(rtmq_proxy_t *pxy, int idx)
{
    rtmq_cmd_t cmd;

    cmd.type = RTMQ_CMD_SEND_ALL;

    pipe_write(&pxy->send_cmd_fd[idx], &cmd, sizeof(cmd));

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_async_send
 **功    能: 发送指定数据(对外接口)
 **输入参数:
 **     pxy: 上下文信息
 **     type: 数据类型
 **     nid: 源结点ID
 **     data: 数据地址
 **     size: 数据长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将数据按照约定格式放入队列中
 **注意事项:
 **     1. 只能用于发送自定义数据类型, 而不能用于系统数据类型
 **     2. 不用关注变量num在多线程中的值, 因其不影响安全性
 **作    者: # Qifeng.zou # 2015.01.14 #
 ******************************************************************************/
int rtmq_proxy_async_send(rtmq_proxy_t *pxy, int type, const void *data, size_t size)
{
    int idx;
    void *addr;
    rtmq_header_t *head;
    static uint32_t num = 0; // 无需加锁
    rtmq_proxy_conf_t *conf = &pxy->conf;

    /* > 选择发送队列 */
    idx = (num++) % conf->send_thd_num;

    addr = queue_malloc(pxy->sendq[idx], sizeof(rtmq_header_t)+size);
    if (NULL == addr) {
        log_error(pxy->log, "Alloc from queue failed! size:%d/%d",
                size+sizeof(rtmq_header_t), queue_size(pxy->sendq[idx]));
        return RTMQ_ERR;
    }

    /* > 设置发送数据 */
    head = (rtmq_header_t *)addr;

    head->type = type;
    head->nid = conf->nid;
    head->length = size;
    head->flag = RTMQ_EXP_MESG;
    head->chksum = RTMQ_CHKSUM_VAL;

    memcpy(head+1, data, size);

    log_debug(pxy->log, "rq:%p Head type:0x%04X nid:%d length:%d flag:%d chksum:%d!",
            pxy->sendq[idx]->ring, head->type, head->nid, head->length, head->flag, head->chksum);

    /* > 放入发送队列 */
    if (queue_push(pxy->sendq[idx], addr)) {
        log_error(pxy->log, "Push into shmq failed!");
        queue_dealloc(pxy->sendq[idx], addr);
        return RTMQ_ERR;
    }

    /* > 通知发送线程 */
    rtmq_proxy_cmd_send_req(pxy, idx);

    return RTMQ_OK;
}

/******************************************************************************
 **函数名称: rtmq_proxy_conf_isvalid
 **功    能: 校验配置合法性
 **输入参数:
 **     conf: 配置数据
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述: 逐一检查配置字段的合法性
 **注意事项:
 **作    者: # Qifeng.zou # 2017.11.28 10:39:35 #
 ******************************************************************************/
static bool rtmq_proxy_conf_isvalid(const rtmq_proxy_conf_t *conf)
{
    if ((0 == conf->nid)
        || (0 == conf->gid)
        || (0 == strlen(conf->ipaddr))
        || (0 == conf->send_thd_num)
        || (0 == conf->work_thd_num)
        || (0 == conf->recv_buff_size)
        || ((0 == conf->sendq.max) || (0 == conf->sendq.size))
        || ((0 == conf->recvq.max) || (0 == conf->recvq.size))) {
        return false;
    }
    return true;
}

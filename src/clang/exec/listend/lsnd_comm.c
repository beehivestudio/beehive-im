#include "mesg.h"
#include "redo.h"
#include "listend.h"
#include "cmd_list.h"

#include "mesg.pb-c.h"

static void lsnd_timer_send_user_num(lsnd_cntx_t *ctx);

/******************************************************************************
 **函数名称: lsnd_getopt 
 **功    能: 解析输入参数
 **输入参数: 
 **     argc: 参数个数
 **     argv: 参数列表
 **输出参数:
 **     opt: 参数选项
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **     1. 解析输入参数
 **     2. 验证输入参数
 **注意事项: 
 **     N: 服务名 - 根据服务名可找到配置路径
 **     h: 帮助手册
 **作    者: # Qifeng.zou # 2014.11.15 #
 ******************************************************************************/
int lsnd_getopt(int argc, char **argv, lsnd_opt_t *opt)
{
    int ch;
    const struct option opts[] = {
        {"help",                    no_argument,        NULL, 'h'}
        , {"daemon",                no_argument,        NULL, 'd'}
        , {"log-level",             required_argument,  NULL, 'l'}
        , {"configuration path",    required_argument,  NULL, 'c'}
        , {NULL,                    0,                  NULL, 0}
    };

    memset(opt, 0, sizeof(lsnd_opt_t));

    opt->isdaemon = false;
    opt->log_level = LOG_LEVEL_TRACE;
    opt->conf_path = "../conf/listend.xml";

    /* 1. 解析输入参数 */
    while (-1 != (ch = getopt_long(argc, argv, "l:c:hd", opts, NULL))) {
        switch (ch) {
            case 'c':   /* 配置路径 */
            {
                opt->conf_path = optarg;
                break;
            }
            case 'l':   /* 日志级别 */
            {
                opt->log_level = log_get_level(optarg);
                break;
            }
            case 'd':
            {
                opt->isdaemon = true;
                break;
            }
            case 'h':   /* 显示帮助信息 */
            default:
            {
                return LSND_SHOW_HELP;
            }
        }
    }

    optarg = NULL;
    optind = 1;

    /* 2. 验证输入参数 */
    if (NULL == opt->conf_path) {
        return LSND_SHOW_HELP;
    }

    return 0;
}

/* 显示启动参数帮助信息 */
int lsnd_usage(const char *exec)
{
    printf("\nUsage: %s -c <configuration path> -l <log level> [-h] [-d]\n", exec);
    printf("\t-c: Configuration path\n"
            "\t-l: Log level. range:[error|warn|info|debug|trace]\n"
            "\t-h: Show help\n"
            "\t-d: Run as daemon\n\n");
    return 0;
}

/* 初始化日志模块 */
log_cycle_t *lsnd_init_log(char *fname)
{
    char path[FILE_NAME_MAX_LEN];

    log_get_path(path, sizeof(path), basename(fname));

    return log_init(LOG_LEVEL_ERROR, path);
}

/******************************************************************************
 **函数名称: lsnd_acc_reg_add
 **功    能: 添加下游数据的处理注册回调
 **输入参数:
 **     ctx: 全局信息
 **     type: 数据类型
 **     proc: 处理回调
 **     args: 附加参数
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.17 22:18:53 #
 ******************************************************************************/
int lsnd_acc_reg_add(lsnd_cntx_t *ctx, int type, lsnd_reg_cb_t proc, void *args)
{
    lsnd_reg_t *reg;

    reg = (lsnd_reg_t *)calloc(1, sizeof(lsnd_reg_t));
    if (NULL == reg) {
        return -1;
    }

    reg->type = type;
    reg->proc = proc;
    reg->args = args;

    if (avl_insert(ctx->reg, reg)) {
        free(reg);
        return -1;
    }

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_kick_trav_cb
 **功    能: 编历被踢连接是否到达执行踢除的时间
 **输入参数:
 **     item: 被踢连接
 **     timeout_list: 超时记录链表
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 16:24:05 #
 ******************************************************************************/
static int lsnd_kick_trav_cb(lsnd_kick_item_t *item, list_t *timeout_list)
{
    time_t ctm = time(NULL);

    if (ctm < item->ttl) {
        return 0; /* 未超时 */
    }

    list_rpush(timeout_list, (void *)item->cid);

    return 0;
}

/******************************************************************************
 **函数名称: lsnd_timer_kick_handler
 **功    能: 定时踢除连接
 **输入参数:
 **     ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 16:24:05 #
 ******************************************************************************/
void lsnd_timer_kick_handler(void *_ctx)
{
    void *addr;
    uint64_t cid;
    list_t *timeout_list;
    lsnd_kick_item_t *item, key;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)_ctx;

    log_debug(ctx->log, "Exec timer kick handler!");

    timeout_list = list_creat(NULL);
    if (NULL == timeout_list) {
        log_error(ctx->log, "Initialize kick timeout list failed!");
        return;
    }

    /* > 获取已经的被踢连接CID列表 */
    hash_tab_trav(ctx->kick_list, (trav_cb_t)lsnd_kick_trav_cb, timeout_list, RDLOCK);

    /* > 依次执行踢除连接的操作 */
    while (NULL != (addr = list_lpop(timeout_list))) {
        cid = (uint64_t)addr;

        key.cid = cid;
        item = hash_tab_delete(ctx->kick_list, &key, WRLOCK);
        if (NULL == item) {
            continue;
        }

        free(item);

        acc_async_kick(ctx->access, (uint64_t)cid);

        log_debug(ctx->log, "Async kick cid:%d", cid);
    }

    return;
}

/******************************************************************************
 **函数名称: lsnd_kick_add
 **功    能: 将指定连接加入被踢列表
 **输入参数:
 **     ctx: 全局信息
 **     conn: 被踢连接
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将cid加入被踢列表
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.03 18:05:31 #
 ******************************************************************************/
int lsnd_kick_add(lsnd_cntx_t *ctx, lsnd_conn_extra_t *conn)
{
    lsnd_kick_item_t *item;

    item = calloc(1, sizeof(lsnd_kick_item_t));
    if (NULL == item) {
        log_error(ctx->log, "Alloc memory failed! errmsg:[%d] %s!", errno, strerror(errno));
        return -1;
    }

    conn->stat = CHAT_CONN_STAT_KICK;
    conn->kick_ttl = time(NULL) + LSND_KICK_TTL;

    item->cid = conn->cid;
    item->ttl = conn->kick_ttl;

    hash_tab_insert(ctx->kick_list, item, WRLOCK);

    return 0;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_timer_info_handler
 **功    能: 侦听层信息定时上报
 **输入参数:
 **     _ctx: 全局信息
 **输出参数:
 **返    回: VOID
 **实现描述: 使用PB协议组装接入层上报数据
 **     {
 **         required uint32 network = 1;    // M|网络类型(0:UNKNOWN 1:TCP 2:WS)|数字|<br>
 **         required uint32 nid = 2;        // M|结点ID|数字|<br>
 **         required uint32 opid = 3;       // M|运营商ID|数字|<br>
 **         required string nation = 4;     // M|所属国家|字串|<br>
 **         required string ip = 5;     // M|IP地址|字串|<br>
 **         required uint32 port = 6;       // M|端口号|数字|<br>
 **     }
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.06 23:23:51 #
 ******************************************************************************/
void lsnd_timer_info_handler(void *_ctx)
{
    void *addr;
    unsigned int len;
    mesg_header_t *head;
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)_ctx;
    lsnd_conf_t *conf = &ctx->conf;
    MesgLsndInfo info = MESG_LSND_INFO__INIT;

    /* > 设置上报数据 */
    info.type = LSND_TYPE_TCP;
    info.nid = conf->nid;
    info.opid = conf->operator.id;
    info.nation = conf->operator.nation;
    info.ip = conf->access.ipaddr;
    info.port = conf->access.port;
    info.connections = hash_tab_total(ctx->conn_list);

    log_debug(ctx->log, "Listen info! nid:%d nation:%s opid:%d ip:%s port:%d",
            info.nid, info.nation, info.opid, info.ip, info.port);

    /* > 组装PB协议 */
    len = mesg_lsnd_info__get_packed_size(&info);

    addr = (void *)calloc(1, sizeof(mesg_header_t) + len);
    if (NULL == addr) {
        log_error(ctx->log, "Alloc memory failed! errmsg:[%d] %s!", errno, strerror(errno));
        return;
    }

    head = (mesg_header_t *)addr;

    head->type = CMD_LSND_INFO;
    head->length = len;
    head->chksum = MSG_CHKSUM_VAL;
    head->nid = conf->nid;

    MESG_HEAD_HTON(head, head);

    mesg_lsnd_info__pack(&info, addr + sizeof(mesg_header_t)); /* 组装PB协议 */

    /* > 发送数据 */
    rtmq_proxy_async_send(ctx->frwder, CMD_LSND_INFO, addr, sizeof(mesg_header_t) + len);

    free(addr);
    return;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsnd_offline_notify
 **功    能: 发送下线指令
 **输入参数:
 **     ctx: 全局信息
 **     sid: 会话SID
 **     cid: 连接CID
 **     nid: 节点ID
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2017.05.10 06:45:51 #
 ******************************************************************************/
void lsnd_offline_notify(lsnd_cntx_t *ctx, uint64_t sid, uint64_t cid, uint32_t nid)
{
    void *addr;
    mesg_header_t *head;

    /* > 组装PB协议 */
    addr = (void *)calloc(1, sizeof(mesg_header_t));
    if (NULL == addr) {
        log_error(ctx->log, "Alloc memory failed! errmsg:[%d] %s!", errno, strerror(errno));
        return;
    }

    head = (mesg_header_t *)addr;

    head->type = CMD_OFFLINE;
    head->length = 0;
    head->chksum = MSG_CHKSUM_VAL;
    head->sid = sid;
    head->cid = cid;
    head->nid = nid;

    MESG_HEAD_HTON(head, head);

    /* > 发送数据 */
    rtmq_proxy_async_send(ctx->frwder, CMD_OFFLINE, addr, sizeof(mesg_header_t));

    free(addr);

    return;
}

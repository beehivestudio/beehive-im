/******************************************************************************
 ** Coypright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: lsnd_mesg.c
 ** 版本号: 1.0
 ** 描  述: 侦听相关的消息处理函数的定义
 ** 作  者: # Qifeng.zou # Thu 16 Jul 2015 01:08:20 AM CST #
 ******************************************************************************/

#include "mesg.h"
#include "access.h"
#include "listend.h"
#include "lsnd_mesg.h"

static int chat_callback_creat_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user);
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user);
static int chat_callback_recv_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user, void *in, int len);

/******************************************************************************
 **函数名称: chat_mesg_def_hdl
 **功    能: 消息默认处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 直接将消息转发给上游服务.
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_mesg_def_hdl(unsigned int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发数据 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_online_req_hdl
 **功    能: ONLINE请求处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 在线服务收到该请求收, 将为该连接分配一个SID.
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_online_req_hdl(int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发搜索请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_online_ack_hdl
 **功    能: ONLINE应答处理
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答信息中提取UID, SID等信息, 并构建索引关系.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 23:38:38 #
 ******************************************************************************/
int chat_online_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(ctx->log, &hhead)
    log_debug(ctx->log, "body:%s", head->body);

    return acc_async_send(ctx->access, type, hhead.sid, data, len);
}

/******************************************************************************
 **函数名称: chat_join_req_hdl
 **功    能: JOIN请求处理
 **输入参数:
 **     type: 全局对象
 **     data: 数据内容
 **     length: 数据长度(报头 + 报体)
 **     args: 附加参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述: 请求数据的内存结构: 流水信息 + 消息头 + 消息体
 **注意事项: 需要将协议头转换为网络字节序
 **作    者: # Qifeng.zou # 2016.09.20 22:25:57 #
 ******************************************************************************/
int chat_join_req_hdl(int type, void *data, int length, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data; /* 消息头 */

    log_debug(lsnd->log, "sid:%lu serial:%lu length:%d body:%s!",
            head->sid, head->serial, length, head->body);

    /* > 转换字节序 */
    MESG_HEAD_HTON(head, head);

    /* > 转发搜索请求 */
    return rtmq_proxy_async_send(lsnd->frwder, type, data, length);
}

/******************************************************************************
 **函数名称: chat_join_ack_hdl
 **功    能: JOIN应答处理
 **输入参数:
 **     type: 数据类型
 **     orig: 源结点ID
 **     data: 需要转发的数据
 **     len: 数据长度
 **     args: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: TODO: 从该应答中获取加入聊天室是否成功. 如果成功则构建索引.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 23:40:12 #
 ******************************************************************************/
int chat_join_ack_hdl(int type, int orig, char *data, size_t len, void *args)
{
    lsnd_cntx_t *ctx = (lsnd_cntx_t *)args;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 转化字节序 */
    MESG_HEAD_NTOH(head, &hhead);

    MESG_HEAD_PRINT(ctx->log, &hhead)
    log_debug(ctx->log, "body:%s", head->body);

    return acc_async_send(ctx->access, type, hhead.sid, data, len);
}

/******************************************************************************
 **函数名称: lsnd_acc_callback
 **功    能: ACCESS处理回调
 **输入参数:
 **     acc: ACC上下文
 **     sck: 套接字
 **     reason: 回调的原因
 **     user: 扩展数据
 **     in: 接收数据
 **     len: 接收数据长度
 **     args: 附加数据. 当前为lsnd_cntx_t对象.
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.09.20 22:03:02 #
 ******************************************************************************/
int chat_callback(acc_cntx_t *acc, socket_t *sck, int reason, void *user, void *in, int len, void *args)
{
    lsnd_cntx_t *lsnd = (lsnd_cntx_t *)args;

    switch (reason) {
        case ACC_CALLBACK_CREAT:
            return chat_callback_creat_hdl(lsnd, sck, (chat_conn_user_data_t *)user);
        case ACC_CALLBACK_DESTROY:
            return chat_callback_destroy_hdl(lsnd, sck, user);
        case ACC_CALLBACK_RECEIVE:
            return chat_callback_recv_hdl(lsnd, sck, user, in, len);
        case ACC_CALLBACK_WRITEABLE:
        case ACC_CALLBACK_CLOSED:
        default:
            break;
    }
    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_creat_hdl
 **功    能: 连接创建的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     user: 扩展数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 将新建连接放入CONN_CID_TAB维护起来, 待分配了SID后再转移到CONN_SID_TAB中.
 **作    者: # Qifeng.zou # 2016.09.20 21:30:53 #
 ******************************************************************************/
static int chat_callback_creat_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user)
{
    /* 初始化设置 */
    user->sid = 0;
    user->cid = lsnd_gen_cid(lsnd);
    user->tsi = sck;
    user->create_time = time(NULL);
    user->loc = LSND_DATA_LOC_UNKNOWN;
    user->stat = LSND_CONN_STAT_ESTABLIST;

    /* 加入CID管理表 */
    if (hash_tab_insert(lsnd->conn_cid_tab, (void *)user, WRLOCK)) {
        return -1;
    }

    user->loc = LSND_DATA_LOC_CID_TAB;

    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_destroy_hdl
 **功    能: 连接销毁的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     user: 扩展数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项: 释放user对象内存的所有空间, 但是请勿释放user对象本身.
 **作    者: # Qifeng.zou # 2016.09.20 21:43:13 #
 ******************************************************************************/
static int chat_callback_destroy_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user)
{
    chat_conn_user_data_t key, *item;

    user->stat = LSND_CONN_STAT_CLOSED;

    switch (user->loc) {
        case LSND_DATA_LOC_CID_TAB:
            key.cid = user->cid;
            item = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
            if (item != user) {
                assert(0);
            }
        case LSND_DATA_LOC_SID_TAB:
            key.sid = user->sid;
            item = hash_tab_delete(lsnd->conn_cid_tab, &key, WRLOCK);
            if (item != user) {
                assert(0);
            }
        default:
            assert(0);
    }

    return 0;
}

/******************************************************************************
 **函数名称: chat_callback_recv_hdl
 **功    能: 接收数据的处理
 **输入参数:
 **     lsnd: 全局对象
 **     sck: 套接字
 **     user: 扩展数据
 **     in: 收到的数据
 **     len: 收到数据的长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 
 **注意事项:
 **     1. 暂无需加锁. 原因: 注册表在程序启动时, 就已固定不变.
 **     2. 本函数收到的数据是一条完整的数据, 且其内容网络字节序.
 **作    者: # Qifeng.zou # 2016.09.20 21:44:40 #
 ******************************************************************************/
static int chat_callback_recv_hdl(lsnd_cntx_t *lsnd, socket_t *sck, chat_conn_user_data_t *user, void *in, int len)
{
    lsnd_reg_t *reg, key;
    mesg_header_t *head = (mesg_header_t *)in;

    key.type = ntohl(head->type);

    reg = avl_query(lsnd->reg, &key);
    if (NULL == reg) {
        if (LSND_CONN_STAT_LOGIN != user->stat) {
            log_warn(lsnd->log, "Drop unknown data! type:0x%X", key.type);
            return 0;
        }
        log_warn(lsnd->log, "Forward unknown data! type:0x%X", key.type);
        return chat_mesg_def_hdl(key.type, in, len, (void *)lsnd);
    }

    return reg->proc(reg->type, in, len, reg->args);
}

#include "redo.h"
#include "mesg.h"
#include "access.h"
#include "command.h"
#include "syscall.h"
#include "acc_rsvr.h"

/******************************************************************************
 **函数名称: acc_send_cmd_to_rsvr
 **功    能: 发送分发命令给指定的代理服务
 **输入参数:
 **     ctx: 全局对象
 **     idx: 代理服务的索引
 **输出参数:
 **返    回: >0:成功 <=0:失败
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2015-06-24 23:55:45 #
 ******************************************************************************/
static int acc_send_cmd_to_rsvr(acc_cntx_t *ctx, int rid, int cmd_id)
{
    cmd_data_t cmd;

    cmd.type = cmd_id;

    return pipe_write(&ctx->rsvr_cmd_fd[rid], (void *)&cmd, sizeof(cmd));
}

/******************************************************************************
 **函数名称: acc_async_send
 **功    能: 发送数据(外部接口)
 **输入参数:
 **     ctx: 全局对象
 **     type: 数据类型
 **     cid: 连接ID(Connection ID)
 **     data: 数据内容
 **     len: 数据长度
 **输出参数:
 **返    回: 发送队列的索引
 **实现描述: 将数据放入发送队列
 **注意事项: 
 **     > 发送内容data结构为: 消息头 + 消息体, 且消息头必须为"网络"字节序.
 **作    者: # Qifeng.zou # 2015-06-04 #
 ******************************************************************************/
int acc_async_send(acc_cntx_t *ctx, int type, uint64_t cid, void *data, int len)
{
    int rid; // rsvr id
    ring_t *sendq;
    acc_send_item_t *item;

    /* > 通过cid获取服务ID */
    rid = acc_get_rid_by_cid(ctx, cid);
    if (-1 == rid) {
        log_error(ctx->log, "Get rid by cid failed! cid:%lu", cid);
        return ACC_ERR;
    }

    /* > 准备存储空间 */
    sendq = ctx->sendq[rid];

    item = (void *)calloc(1, sizeof(acc_send_item_t));
    if (NULL == item) {
        log_error(ctx->log, "Alloc memory failed! len:%d", len);
        return ACC_ERR;
    }

    item->data = (void *)calloc(1, len);
    if (NULL == item->data) {
        log_error(ctx->log, "Alloc memory failed! len:%d", len);
        free(item);
        return ACC_ERR;
    }

    item->cid = cid;
    item->len = len;
    memcpy(item->data, data, len);

    /* > 放入发送队列 */
    if (ring_push(sendq, item)) {
        FREE(item->data);
        FREE(item);
        log_error(ctx->log, "Push into ring failed!");
        return ACC_ERR;
    }

    acc_send_cmd_to_rsvr(ctx, rid, CMD_DIST_DATA); /* 发送分发命令 */

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_sck_get_cid
 **功    能: 获取套接字CID.
 **输入参数:
 **     sck: 套接字对象
 **输出参数:
 **返    回: CID对象
 **实现描述: 
 **注意事项: 
 **作    者: # Qifeng.zou # 2019.06.24 12:01:59 #
 ******************************************************************************/
uint64_t acc_sck_get_cid(socket_t *sck)
{
    acc_socket_extra_t *extra = sck->extra;

    return extra->cid;
}

/******************************************************************************
 **函数名称: acc_async_kick
 **功    能: 异步踢人(外部接口)
 **输入参数:
 **     ctx: 全局对象
 **     cid: 连接ID(Connection ID)
 **输出参数:
 **返    回:
 **实现描述:
 **注意事项: 
 **     > 发送内容data结构为: 消息头 + 消息体, 且消息头必须为"网络"字节序.
 **作    者: # Qifeng.zou # 2016.10.01 18:58:51 #
 ******************************************************************************/
int acc_async_kick(acc_cntx_t *ctx, uint64_t cid)
{
    int rid; // rsvr id
    queue_t *kickq;
    acc_kick_req_t *kick;

    /* > 通过cid获取服务ID */
    rid = acc_get_rid_by_cid(ctx, cid);
    if (-1 == rid) {
        log_error(ctx->log, "Get rid by cid failed! cid:%lu", cid);
        return ACC_ERR;
    }

    /* > 放入指定发送队列 */
    kickq = ctx->kickq[rid];

    kick = queue_malloc(kickq, sizeof(acc_kick_req_t));
    if (NULL == kick) {
        log_error(ctx->log, "Alloc data from kickq failed!");
        return ACC_ERR;
    }

    kick->cid = cid;

    queue_push(kickq, kick);

    acc_send_cmd_to_rsvr(ctx, rid, CMD_KICK_CONN); /* 发送分发命令 */

    return ACC_OK;
}

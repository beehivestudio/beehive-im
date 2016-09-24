#include "redo.h"
#include "mesg.h"
#include "access.h"
#include "command.h"
#include "syscall.h"
#include "acc_rsvr.h"

static int acc_cmd_send_dist_req(acc_cntx_t *ctx, int idx);

/******************************************************************************
 **函数名称: acc_async_send
 **功    能: 发送数据(外部接口)
 **输入参数:
 **     ctx: 全局对象
 **     type: 数据类型
 **     cid: 连接ID(Connection ID)
 **     data: 数据内容(必须包含消息头: mesg_header_t)
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
    void *addr;
    ring_t *sendq;
    mesg_header_t *head = (mesg_header_t *)data, hhead;

    /* > 合法性校验 */
    MESG_HEAD_NTOH(head, &hhead);
    if (!MESG_CHKSUM_ISVALID(&hhead)) {
        log_error(ctx->log, "Data format is invalid! cid:%lu", cid);
        return ACC_ERR;
    }

    MESG_HEAD_PRINT(ctx->log, &hhead);

    /* > 通过cid获取服务ID */
    rid = acc_get_rid_by_cid(ctx, cid);
    if (-1 == rid) {
        log_error(ctx->log, "Get rid by cid failed! cid:%lu", cid);
        return ACC_ERR;
    }

    /* > 放入指定发送队列 */
    sendq = ctx->sendq[rid];

    addr = (void *)calloc(1, len);
    if (NULL == addr) {
        log_error(ctx->log, "Alloc memory failed! len:%d", len);
        return ACC_ERR;
    }

    memcpy(addr, data, len);

    if (ring_push(sendq, addr)) { /* 放入队列 */
        FREE(addr);
        log_error(ctx->log, "Push into ring failed!");
        return ACC_ERR;
    }

    acc_cmd_send_dist_req(ctx, rid); /* 发送分发命令 */

    return ACC_OK;
}

/******************************************************************************
 **函数名称: acc_cmd_send_dist_req
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
static int acc_cmd_send_dist_req(acc_cntx_t *ctx, int idx)
{
    cmd_data_t cmd;
    char path[FILE_PATH_MAX_LEN];
    acc_conf_t *conf = ctx->conf;

    cmd.type = CMD_DIST_DATA;
    acc_rsvr_cmd_usck_path(conf, idx, path, sizeof(path));

    return unix_udp_send(ctx->cmd_sck_id, path, (void *)&cmd, sizeof(cmd));
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

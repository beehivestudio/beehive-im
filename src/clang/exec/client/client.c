#include "sdk.h"
#include "redo.h"
#include "cmd_list.h"

/* > 设置配置信息 */
void client_set_conf(sdk_conf_t *conf)
{
    memset(conf, 0, sizeof(sdk_conf_t));

    conf->nid = 0; /* 设备ID: 唯一值 */
    snprintf(conf->path, sizeof(conf->path), "."); /* 工作路径 */
    conf->uid = 10000;                   /* 用户ID */
    conf->sessionid = 10000;             /* 会话ID(备选) */
    conf->terminal = 1;                   /* 终端类型 */
    snprintf(conf->app, sizeof(conf->app), "beehive chat");
    snprintf(conf->version, sizeof(conf->version), "1.0.0");    /* 客户端自身版本号(留做统计用) */

    conf->log_level = LOG_LEVEL_DEBUG;                      /* 日志级别 */
    snprintf(conf->log_path, sizeof(conf->log_path), "client.log");   /* 日志路径(路径+文件名) */
    snprintf(conf->httpsvr, sizeof(conf->httpsvr), "127.0.0.1:8080/client/query");      /* HTTP端(IP+端口/域名) */

    conf->work_thd_num = 1;  /* 工作线程数 */
    conf->recv_buff_size = 1024 * 1024;  /* 接收缓存大小 */
    conf->sendq_len = 1024;  /* 发送队列配置 */
    conf->recvq_len = 1024;  /* 接收队列配置 */

    return;
}

/* > 应答消息PONG的处理 */
int sdk_cmd_pong_handler(int cmd, uint64_t from, char *data, size_t len, void *param)
{
    fprintf(stderr, "Call %s() cmd:%d\n", __func__, cmd);
    return 0;
}

/* > 发送结果回调 */
int sdk_send_cb(uint16_t cmd, const void *orig, size_t size,
        char *ack, size_t ack_len, sdk_send_stat_e stat, void *param)
{
    fprintf(stderr, "Call %s() cmd:%d stat:%d\n", __func__, cmd, stat);
    return 0;
}

int main(int argc, char *argv[])
{
    sdk_conf_t conf;
    sdk_cntx_t *ctx;

    client_set_conf(&conf);

    ctx = sdk_init(&conf);
    if (NULL == ctx) {
        fprintf(stderr, "Initialize sdk failed!\n");
        return -1;
    }

    sdk_register(ctx, CMD_PONG, (sdk_reg_cb_t)sdk_cmd_pong_handler, NULL);

    sdk_launch(ctx);

    while (1) {
        sdk_async_send(ctx, CMD_PING, NULL, 0, 3, (sdk_send_cb_t)sdk_send_cb, NULL);
        Sleep(1);
    }

    return 0;
}

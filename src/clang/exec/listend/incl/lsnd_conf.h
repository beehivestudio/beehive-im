#if !defined(__LSND_CONF_H__)
#define __LSND_CONF_H__

#include "comm.h"
#include "access.h"
#include "rtmq_proxy.h"

/* 侦听配置 */
typedef struct
{
    int nid;                        /* 结点ID */
    char wdir[FILE_PATH_MAX_LEN];   /* 工作路径 */

    struct {
        uint32_t id;                /* 运营商ID */
        char nation[NAME_MAX_LEN];  /* 所属国家 */
    } operator;                     /* 运营商 */

    struct {
        int num;                    /* 队列数 */
        int max;                    /* 队列长度 */
        int size;                   /* 单元大小 */
    } distq;                        /* 分发队列 */
    acc_conf_t access;              /* 帧听配置 */
    rtmq_proxy_conf_t frwder;       /* FRWDER配置 */
} lsnd_conf_t;

int lsnd_load_conf(const char *path, lsnd_conf_t *conf, log_cycle_t *log);

#endif /*__LSND_CONF_H__*/

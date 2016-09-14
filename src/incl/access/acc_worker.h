#if !defined(__ACC_WORKER_H__)
#define __ACC_WORKER_H__

#include "queue.h"
#include "access.h"

/* Worker对象 */
typedef struct
{
    int id;                         /* 对象ID */
    int cmd_sck_id;                 /* 命令套接字 */
    fd_set rdset;                   /* 可读集合 */
    log_cycle_t *log;               /* 日志对象 */
} acc_worker_t;

void *acc_worker_routine(void *_ctx);

int acc_worker_init(acc_cntx_t *ctx, acc_worker_t *worker, int idx);
int acc_worker_destroy(acc_worker_t *worker);

#endif /*__ACC_WORKER_H__*/

#if !defined(__ACC_LISTEN_H__)
#define __ACC_LISTEN_H__

#include "log.h"
#include <stdint.h>

/* 侦听线程 */
typedef struct
{
    int id;                                 /* 线程ID(从0开始计数) */

    log_cycle_t *log;                       /* 日志对象 */
} acc_lsvr_t;

void *acc_lsvr_routine(void *_ctx);

#endif /*__ACC_LISTEN_H__*/

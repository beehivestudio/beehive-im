#if !defined(__TIMER_H__)
#define __TIMER_H__

#include "comm.h"

#define TIMER_CAP    (128)   /* 堆容量 */

/* 管理对象 */
typedef struct
{
    int len;                /* 成员个数 */
    int cap;                /* 最大容量 */
    void **e;               /* 成员列表(数组) */

    pthread_rwlock_t lock;  /* 读写锁 */

    cmp_cb_t cmp;           /* 比较回调 */
} timer_cntx_t;

/* 任务对象 */
typedef struct
{
    int idx;                /* 堆中序号 */
    timer_cntx_t *ctx;      /* 上下文对象 */

    void (*proc)(void *param); /* 定时回调 */
    int start;              /* 开始时间 */
    int interval;           /* 间隔时间 */
    void *param;            /* 附加参数 */

    uint64_t times;         /* 已执行次数 */
    time_t ttl;             /* 下次执行时间 */
} timer_task_t;

timer_cntx_t *timer_cntx_init(void);
int timer_task_init(timer_task_t *task,
        void (*proc)(void *param), int start, int interval, void *param);
int timer_task_add(timer_cntx_t *ctx, timer_task_t *task);
int timer_task_del(timer_cntx_t *ctx, timer_task_t *task);
int timer_task_update(timer_task_t *task,
        void (*proc)(void *param), int start, int interval, void *param);
void *timer_task_routine(void *_ctx);

#endif /*__TIMER_H__*/

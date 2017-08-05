#if !defined(__TIMER_H__)
#define __TIMER_H__

#include "comm.h"
#include "list.h"

#define TIMER_CAP    (128)      /* 堆容量 */

typedef struct
{
    void (*proc)(void *param);  /* 定时回调 */
    void *param;                /* 附加参数 */
} timer_task_item_t;

/* 管理对象 */
typedef struct
{
    int len;                    /* 成员个数 */
    int cap;                    /* 最大容量 */
    void **e;                   /* 成员列表(数组) */

    pthread_rwlock_t lock;      /* 读写锁 */

    cmp_cb_t cmp;               /* 比较回调 */
} timer_cntx_t;

/* 任务对象 */
typedef struct
{
    int idx;                    /* 堆中序号 */
    timer_cntx_t *ctx;          /* 上下文对象 */

    int start;                  /* 开始时间 */
    int interval;               /* 间隔时间 */

    list_t *list;               /* 处理列表(注:指向timer_task_item_t对象) */

    uint64_t times;             /* 已执行次数 */
    time_t ttl;                 /* 下次执行时间 */
} timer_task_t;

timer_cntx_t *timer_cntx_init(void);
timer_task_t *timer_task_init(int start, int interval);
int timer_task_add(timer_task_t *task, void (*proc)(void *param), void *param);
int timer_task_start(timer_cntx_t *ctx, timer_task_t *task);
int timer_task_stop(timer_cntx_t *ctx, timer_task_t *task);
void *timer_task_routine(void *_ctx);

#endif /*__TIMER_H__*/

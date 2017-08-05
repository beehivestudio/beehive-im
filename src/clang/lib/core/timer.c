#include "comm.h"
#include "redo.h"
#include "timer.h"

static int timer_item_cmp_cb(timer_task_t *item1, timer_task_t *item2)
{
    return item1->ttl - item2->ttl;
}

/******************************************************************************
 **函数名称: timer_cntx_init
 **功    能: 初始化定时器
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 上下文对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 16:04:41 #
 ******************************************************************************/
timer_cntx_t *timer_cntx_init(void)
{
    timer_cntx_t *ctx;

    ctx = (timer_cntx_t *)calloc(1, sizeof(timer_cntx_t));
    if (NULL == ctx) {
        return NULL;
    }

    ctx->e = (void **)calloc(TIMER_CAP, sizeof(void *));
    if (NULL == ctx->e) {
        free(ctx);
        return NULL;
    }

    ctx->len = 0;
    ctx->cap = TIMER_CAP;
    ctx->cmp = (cmp_cb_t)timer_item_cmp_cb;
    pthread_rwlock_init(&ctx->lock, NULL);

    return ctx;
}

/******************************************************************************
 **函数名称: timer_task_init
 **功    能: 初始化定时器
 **输入参数: 
 **     task: 定时任务
 **     start: 开始时间
 **     interval: 间隔时间
 **输出参数: NONE
 **返    回: 定时任务
 **实现描述:
 **注意事项:
 **     1. 完成定时任务的初始化之后, 还需要调用timer_task_add()将定时任务加入执行流程.
 **     2. 间隔时间不能为0.
 **作    者: # Qifeng.zou # 2016.12.28 16:04:41 #
 ******************************************************************************/
timer_task_t *timer_task_init(int start, int interval)
{
    timer_task_t *task;
    time_t ctm = time(NULL);

    task = (timer_task_t *)calloc(1, sizeof(timer_task_t));
    if (NULL == task) {
        return NULL;
    }

    task->start = start;        /* 开始时间 */
    task->interval = interval? interval : 1;  /* 间隔时间 */
    task->list = list_creat(NULL);
    if (NULL == task->list) {
        free(task);
        return NULL;
    }

    task->times = 0;            /* 已执行次数 */
    task->ttl = ctm + start;    /* 下次执行时间 */

    return task;
}

/******************************************************************************
 **函数名称: timer_task_add
 **功    能: 添加处理任务
 **输入参数: 
 **     task: 定时任务
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 上下文对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 16:04:41 #
 ******************************************************************************/
int timer_task_add(timer_task_t *task, void (*proc)(void *param), void *param)
{
    timer_task_item_t *item;

    item = (timer_task_item_t *)calloc(1, sizeof(timer_task_item_t));
    if (NULL == item) {
        return -1;
    }

    item->proc = proc;
    item->param = param;

    list_rpush(task->list, item);

    return 0;
}

/******************************************************************************
 **函数名称: timer_task_start
 **功    能: 执行定时任务
 **输入参数:
 **     ctx: 堆对象
 **     data: 数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将新增数据放在堆末尾, 再反向调整堆的结构.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 17:00:42 #
 ******************************************************************************/
int timer_task_start(timer_cntx_t *ctx, timer_task_t *task)
{
    void **e;
    int parent, idx;

    pthread_rwlock_wrlock(&ctx->lock);

    task->ctx = ctx;

    /* > 防止空间不足 申请内存空间 */
    if (ctx->len >= ctx->cap) {
        e = (void **)realloc(ctx->e, (2 * ctx->cap) * sizeof(void *));
        if (NULL == e) {
            pthread_rwlock_unlock(&ctx->lock);
            return -1;
        }

        ctx->e = e;
        ctx->cap = 2 * ctx->cap;
    }

    if (0 == ctx->len) {
        ctx->e[ctx->len++] = task;
        task->idx = 0;
        pthread_rwlock_unlock(&ctx->lock);
        return 0;
    }

    /* > 调整堆的结构 */
    idx = ctx->len;
    parent = idx / 2;

    while ((idx != parent) && (ctx->cmp(ctx->e[parent], task) > 0)) {
        ctx->e[idx] = ctx->e[parent];
        idx = parent;
        parent = idx / 2;
    }

    ctx->e[idx] = (void *)task;
    task->idx = idx;
    ++ctx->len;

    pthread_rwlock_unlock(&ctx->lock);

    return 0;
}

/******************************************************************************
 **函数名称: timer_task_pop
 **功    能: 将任务弹出堆
 **输入参数:
 **     ctx: 堆对象
 **输出参数: NONE
 **返    回: 数据
 **实现描述: 数据出堆后, 再拿末尾数据到堆头调整堆结构.
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 16:04:41 #
 ******************************************************************************/
timer_task_t *timer_task_pop(timer_cntx_t *ctx)
{
    timer_task_t *task, *item;
    int idx, left, right, min;

    pthread_rwlock_wrlock(&ctx->lock);

    if (0 == ctx->len) {
        pthread_rwlock_unlock(&ctx->lock);
        return NULL;
    }

    item = ctx->e[0];

    /* > 调整堆的结构 */
    task = ctx->e[--ctx->len];

    idx = 0;
    left = 1;
    right = 2;
    while (1) {
        if (left >= ctx->len) {
            ctx->e[idx] = task;
            break;
        }

        min = left;
        if (right < ctx->len) {
            min = (ctx->cmp(ctx->e[left], ctx->e[right]) <= 0)? left : right;
        }

        if (ctx->cmp(ctx->e[min], task) >= 0) {
            ctx->e[idx] = task;
            break;
        }

        ctx->e[idx] = ctx->e[min];

        idx = min;
        left = 2 * idx;
        right = 2 * idx + 1;
    }

    pthread_rwlock_unlock(&ctx->lock);

    return item;
}

/******************************************************************************
 **函数名称: timer_task_stop
 **功    能: 停止指定定时任务
 **输入参数:
 **     ctx: 上下文对象
 **     task: 被删的对象
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 将堆最后一个元素填补被删除的元素, 再从被删元素处开始调整堆的结构.
 **注意事项: 外部需要主动释放内存空间, 否则存在内存泄露的可能性
 **作    者: # Qifeng.zou # 2016.12.28 16:04:41 #
 ******************************************************************************/
int timer_task_stop(timer_cntx_t *ctx, timer_task_t *task)
{
    timer_task_t *item;
    int idx, left, right, min;

    pthread_rwlock_wrlock(&ctx->lock);

    if (task->idx >= ctx->len) {
        pthread_rwlock_unlock(&ctx->lock);
        return -1;
    } else if (task != (timer_task_t *)ctx->e[task->idx]) {
        pthread_rwlock_unlock(&ctx->lock);
        return -1;
    }

    idx = task->idx;
    if (0 == idx) {
        left = 1;
        right = 2;
    } else {
        left = 2 * idx;
        right = 2 * idx + 1;
    }

    /* 将堆最后一个元素填补被删除的元素 */
    item = (timer_task_t *)(ctx->e[--ctx->len]);

    /* 再从被删元素处开始调整堆的结构 */
    while (1) {
        if (left >= ctx->len) {
            ctx->e[idx] = item;
            task->idx = idx;
            break;
        }

        min = left;
        if (right < ctx->len) {
            min = (ctx->cmp(ctx->e[left], ctx->e[right]) <= 0)? left : right;
        }

        if (ctx->cmp(ctx->e[min], item) >= 0) {
            ctx->e[idx] = item;
            item->idx = idx;
            break;
        }

        ctx->e[idx] = ctx->e[min];
        ((timer_task_t *)(ctx->e[idx]))->idx = idx;

        idx = min;
        left = 2 * idx;
        right = 2 * idx + 1;
    }

    pthread_rwlock_unlock(&ctx->lock);
    return 0;
}

/******************************************************************************
 **函数名称: timer_task_is_timeout
 **功    能: 是否存在超时任务
 **输入参数:
 **     ctx: 堆对象
 **输出参数: NONE
 **返    回: 数据
 **实现描述: 获取堆头的最近超时时间
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.28 20:06:45 #
 ******************************************************************************/
static bool timer_task_is_timeout(timer_cntx_t *ctx)
{
    int timeout;
    timer_task_t *task;
    time_t ctm = time(NULL);

    pthread_rwlock_rdlock(&ctx->lock);

    if (0 == ctx->len) {
        pthread_rwlock_unlock(&ctx->lock);
        return false;
    }

    task = (timer_task_t *)(ctx->e[0]);
    timeout = (ctm < task->ttl)? task->ttl - ctm : 0;

    pthread_rwlock_unlock(&ctx->lock);

    return (timeout > 0)? false : true;
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 处理回调 */
static int timer_task_proc_cb(timer_task_item_t *item, void *param)
{
    item->proc(item->param);

    return 0;
}

/******************************************************************************
 **函数名称: timer_task_routine
 **功    能: 定时任务处理
 **输入参数:
 **     task: 定时任务
 **输出参数:
 **返    回: VOID
 **实现描述: 睡眠过程中可能有新添加的定时任务, 为防止新任务超时后长时间得不到处
 **          理, 因此, 每次睡眠时间为1秒.
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.12.28 19:46:43 #
 ******************************************************************************/
void *timer_task_routine(void *_ctx)
{
    timer_task_t *task;
    timer_cntx_t *ctx = (timer_cntx_t *)_ctx;

    for (;;) {
        while (timer_task_is_timeout(ctx)) {
            task = timer_task_pop(ctx);
            if (NULL == task) {
                break;
            }

            list_trav(task->list, (trav_cb_t)timer_task_proc_cb, NULL);

            ++task->times;
            task->ttl = time(NULL) + task->interval;

            timer_task_start(ctx, task);
        }
        Sleep(1);
    }

    return NULL;
}

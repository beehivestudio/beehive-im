/******************************************************************************
 ** Copyright(C) 2014-2024 Qiware technology Co., Ltd
 **
 ** 文件名: thread_pool.c
 ** 版本号: 1.0
 ** 描  述: 线程池模块的实现.
 **         通过线程池模块, 可有效的简化多线程编程的处理, 加快开发速度, 同时有
 **         效增强模块的复用性和程序的稳定性。
 ** 作  者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
#include "thread_pool.h"

static void *thread_routine(void *_p);

/******************************************************************************
 **函数名称: thread_pool_init
 **功    能: 初始化线程池
 **输入参数:
 **     num: 线程数目
 **     opt: 选项信息
 **     args: 附加参数信息
 **输出参数:
 **返    回: 线程池
 **实现描述:
 **     1. 分配线程池空间, 并初始化
 **     2. 创建指定数目的线程
 **注意事项:
 **作    者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
thread_pool_t *thread_pool_init(int num, const thread_pool_opt_t *opt, void *args)
{
    int idx;
    thread_pool_t *p;
    thread_pool_opt_t _opt;

    if (NULL == opt) {
        memset(&_opt, 0, sizeof(_opt));
        _opt.pool = (void *)NULL;
        _opt.alloc = (mem_alloc_cb_t)mem_alloc;
        _opt.dealloc = (mem_dealloc_cb_t)mem_dealloc;
        opt = &_opt;
    }

    /* 1. 分配线程池空间, 并初始化 */
    p = (thread_pool_t *)opt->alloc(opt->pool, sizeof(thread_pool_t));
    if (NULL == p) {
        return NULL;
    }

    pthread_mutex_init(&(p->queue_lock), NULL);
    pthread_cond_init(&(p->queue_ready), NULL);
    p->head = NULL;
    p->queue_size = 0;
    p->shutdown = 0;
    p->data = (void *)args;

    p->mem_pool = opt->pool;
    p->alloc = opt->alloc;
    p->dealloc = opt->dealloc;

    p->tid = (pthread_t *)p->alloc(p->mem_pool, num*sizeof(pthread_t));
    if (NULL == p->tid) {
        opt->dealloc(opt->pool, p);
        return NULL;
    }

    /* 2. 创建指定数目的线程 */
    for (idx=0; idx<num; ++idx) {
        if (thread_creat(&p->tid[idx], thread_routine, p)) {
            thread_pool_destroy(p);
            return NULL;
        }
        ++p->num;
    }

    return p;
}

/******************************************************************************
 **函数名称: thread_pool_add_worker
 **功    能: 注册处理任务(回调函数)
 **输入参数:
 **     p: 线程池
 **     process: 回调函数
 **     arg: 回调函数的参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 新建任务结点
 **     2. 将回调函数加入工作队列
 **     3. 唤醒正在等待的线程
 **注意事项:
 **作    者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
int thread_pool_add_worker(thread_pool_t *p, void *(*process)(void *arg), void *arg)
{
    thread_worker_t *worker, *member;

    /* 1. 新建任务节点 */
    worker = (thread_worker_t*)p->alloc(p->mem_pool, sizeof(thread_worker_t));
    if (NULL == worker) {
        return -1;
    }

    worker->process = process;
    worker->arg = arg;
    worker->next = NULL;

    /* 2. 将回调函数加入工作队列 */
    pthread_mutex_lock(&(p->queue_lock));

    member = p->head;
    if (NULL != member) {
        while (NULL != member->next) {
            member = member->next;
        }
        member->next = worker;
    } else {
        p->head = worker;
    }

    p->queue_size++;

    pthread_mutex_unlock(&(p->queue_lock));

    /* 3. 唤醒正在等待的线程 */
    pthread_cond_signal(&(p->queue_ready));

    return 0;
}

/******************************************************************************
 **函数名称: thread_pool_keepalive
 **功    能: 线程保活处理
 **输入参数:
 **     p: 线程池
 **     process: 回调函数
 **     arg: 回调函数的参数
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 判断线程是否正在运行
 **     2. 如果线程已退出, 则重新启动线程
 **注意事项:
 **作    者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
int thread_pool_keepalive(thread_pool_t *p)
{
    int idx;

    for (idx=0; idx<p->num; idx++) {
        if (ESRCH == pthread_kill(p->tid[idx], 0)) {
            if (thread_creat(&p->tid[idx], thread_routine, p) < 0) {
                return -1;
            }
        }
    }

    return 0;
}

/******************************************************************************
 **函数名称: thread_pool_get_tidx
 **功    能: 获取当前线程在线程池中的序列号
 **输入参数:
 **     p: 线程池
 **输出参数:
 **返    回: 线程序列号
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
int thread_pool_get_tidx(thread_pool_t *p)
{
    int idx;
    pthread_t tid = pthread_self();

    for (idx=0; idx<p->num; ++idx) {
        if (p->tid[idx] == tid) {
            return idx;
        }
    }

    return -1;
}

/******************************************************************************
 **函数名称: thread_pool_destroy
 **功    能: 销毁线程池
 **输入参数:
 **     p: 线程池
 **输出参数:
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 设置销毁标志
 **     2. 唤醒所有线程
 **     3. 等待所有线程结束
 **注意事项:
 **作    者: # Qifeng.zou # 2012.12.26 #
 ******************************************************************************/
int thread_pool_destroy(thread_pool_t *p)
{
    int idx;

    if (0 != p->shutdown) {
        return -1;
    }

    /* 1. 设置销毁标志 */
    p->shutdown = 1;

    /* 2. 唤醒所有等待的线程 */
    pthread_cond_broadcast(&(p->queue_ready));

    /* 3. 等待线程结束 */
    for (idx=0; idx<p->num; ++idx) {
        if (ESRCH == pthread_kill(p->tid[idx], 0)) {
            continue;
        }
        pthread_cancel(p->tid[idx]);
    }

    pthread_mutex_destroy(&(p->queue_lock));
    pthread_cond_destroy(&(p->queue_ready));
    p->dealloc(p->mem_pool, p);

    return 0;
}

/******************************************************************************
 **函数名称: thread_routine
 **功    能: 线程运行函数
 **输入参数:
 **     _p: 线程池
 **输出参数:
 **返    回: VOID *
 **实现描述:
 **     判断是否有任务: 如无, 则等待; 如有, 则处理.
 **注意事项:
 **作    者: # Qifeng.zou # 2014.04.18 #
 ******************************************************************************/
static void *thread_routine(void *_p)
{
    thread_worker_t *worker;
    thread_pool_t *p = (thread_pool_t*)_p;

    while (1) {
        pthread_mutex_lock(&(p->queue_lock));
        while ((0 == p->shutdown)
            && (0 == p->queue_size))
        {
            pthread_cond_wait(&(p->queue_ready), &(p->queue_lock));
        }

        if (0 != p->shutdown) {
            pthread_mutex_unlock(&(p->queue_lock));
            pthread_exit(NULL);
        }

        p->queue_size--;
        worker = p->head;
        p->head = worker->next;
        pthread_mutex_unlock(&(p->queue_lock));

        (*(worker->process))(worker->arg);

        pthread_mutex_lock(&(p->queue_lock));
        p->dealloc(p->mem_pool, worker);
        pthread_mutex_unlock(&(p->queue_lock));
    }

    return NULL;
}

/******************************************************************************
 **函数名称: thread_creat
 **功    能: 创建线程
 **输入参数:
 **     process: 线程回调函数
 **     args: 回调函数参数
 **输出参数:
 **     tid: 线程ID
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2014.04.18 #
 ******************************************************************************/
int thread_creat(pthread_t *tid, void *(*process)(void *args), void *args)
{
    pthread_attr_t attr;

    for (;;) {
        /* 属性初始化 */
        if (pthread_attr_init(&attr)) {
            break;
        }

        /* 设置为分离线程 */
        if (pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED)) {
            break;
        }

        /* 设置线程栈大小 */
        if (pthread_attr_setstacksize(&attr, THREAD_ATTR_STACK_SIZE)) {
            break;
        }

        /* 创建线程 */
        if (pthread_create(tid, &attr, process, args)) {
            if (EINTR == errno) {
                pthread_attr_destroy(&attr);
                continue;
            }

            break;
        }

        pthread_attr_destroy(&attr);
        return 0;
    }

    pthread_attr_destroy(&attr);
    return -1;
}

/******************************************************************************
 ** Coypright(C) 2016-2026 Qiware technology Co., Ltd
 **
 ** 文件名: mref.c
 ** 版本号: 1.0
 ** 描  述: 内存引用计数管理
 ** 作  者: # Qifeng.zou # 2016年06月27日 星期一 21时19分37秒 #
 ******************************************************************************/
#include "comm.h"
#include "mref.h"
#include "atomic.h"
#include "hash_tab.h"

#define MREF_SLOT_LEN    (999)

/* 内存引用项 */
typedef struct
{
    void *addr;                     // 内存地址
    uint32_t count;                 // 引用次数

    struct {
        void *pool;                 // 内存池
        mem_dealloc_cb_t dealloc;   // 释放回调
    };
} mref_item_t;

/* 内存应用管理表 */
hash_tab_t *gMemRefTab;

#define GetMemRef() (gMemRefTab)
#define SetMemRef(tab) (gMemRefTab = (tab))

static int mref_add(void *addr, void *pool, mem_dealloc_cb_t dealloc);

/* 哈希回调函数 */
static uint64_t mref_hash_cb(const mref_item_t *item)
{
    return (uint64_t)item->addr;
}

/* 比较回调函数 */
static int64_t mref_cmp_cb(const mref_item_t *item1, const mref_item_t *item2)
{
    return (int64_t)((uint64_t)item1->addr - (uint64_t)item2->addr);
}

/******************************************************************************
 **函数名称: mref_init
 **功    能: 初始化内存引用计数表
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.06.29 14:45:15 #
 ******************************************************************************/
int mref_init(void)
{
    hash_tab_t *tab;

    tab = hash_tab_creat(MREF_SLOT_LEN,
            (hash_cb_t)mref_hash_cb,
            (cmp_cb_t)mref_cmp_cb, NULL);

    SetMemRef(tab);

    return (NULL == tab)? -1 : 0;
}

/******************************************************************************
 **函数名称: mref_alloc
 **功    能: 申请内存空间
 **输入参数:
 **     size: 申请内存大小
 **     pool: 内存池
 **     alloc: 内存分配函数
 **     dealloc: 内存回收函数
 **输出参数: NONE
 **返    回: 内存地址
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.07.04 00:33:34 #
 ******************************************************************************/
void *mref_alloc(size_t size, void *pool, mem_alloc_cb_t alloc, mem_dealloc_cb_t dealloc)
{
    void *addr;

    if (0 == size) {
        return NULL;
    }

    addr = (void *)alloc(pool, size);
    if (NULL == addr) {
        return NULL;
    }

    mref_add(addr, pool, dealloc);

    return addr;
}

/******************************************************************************
 **函数名称: mref_add
 **功    能: 添加新的引用
 **输入参数:
 **     addr: 内存地址
 **     pool: 内存池
 **     dealloc: 内存回收函数
 **输出参数: NONE
 **返    回: 引用次数
 **实现描述:
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.08 #
 ******************************************************************************/
static int mref_add(void *addr, void *pool, mem_dealloc_cb_t dealloc)
{
    int cnt;
    mref_item_t *item, key;
    hash_tab_t *tab = GetMemRef();

AGAIN:
    /* > 查询引用 */
    key.addr = addr;

    item = hash_tab_query(tab, (void *)&key, RDLOCK);
    if (NULL != item) {
        cnt = (int)atomic32_inc(&item->count);
        hash_tab_unlock(tab, &key, RDLOCK);
        return cnt;
    }

    /* > 新增引用 */
    item = (mref_item_t *)calloc(1, sizeof(mref_item_t));
    if (NULL == item) {
        assert(0);
        return -1;
    }

    item->addr = addr;
    cnt = ++item->count;
    item->pool = pool;
    item->dealloc = dealloc;

    if (hash_tab_insert(tab, item, WRLOCK)) {
        free(item);
        goto AGAIN;
    }

    return cnt;
}

/******************************************************************************
 **函数名称: mref_dealloc
 **功    能: 回收内存空间
 **输入参数:
 **     pool: 内存池
 **     addr: 内存地址
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.07.04 00:33:34 #
 ******************************************************************************/
void mref_dealloc(void *pool, void *addr)
{
    mref_dec(addr);
}

/******************************************************************************
 **函数名称: mref_inc
 **功    能: 增加1次引用
 **输入参数:
 **     addr: 内存地址
 **输出参数: NONE
 **返    回: 内存池
 **实现描述:
 **注意事项: 内存addr必须由mref_alloc进行分配.
 **作    者: # Qifeng.zou # 2016.07.06 #
 ******************************************************************************/
int mref_inc(void *addr)
{
    int cnt;
    mref_item_t *item, key;
    hash_tab_t *tab = GetMemRef();

    key.addr = addr;

    item = hash_tab_query(tab, (void *)&key, RDLOCK);
    if (NULL != item) {
        cnt = (int)atomic32_inc(&item->count);
        hash_tab_unlock(tab, &key, RDLOCK);
        return cnt;
    }

    assert(0);
    return -1; // 未创建结点
}

/******************************************************************************
 **函数名称: mref_dec
 **功    能: 减少1次引用
 **输入参数:
 **     addr: 内存地址
 **输出参数: NONE
 **返    回: 内存引用次数
 **实现描述:
 **注意事项:
 **     1. 内存addr是通过系统调用分配的方可使用内存引用
 **     2. 如果引用计数减为0, 则释放该内存空间.
 **作    者: # Qifeng.zou # 2016.06.29 14:53:09 #
 ******************************************************************************/
int mref_dec(void *addr)
{
    int cnt;
    mref_item_t *item, key;
    hash_tab_t *tab = GetMemRef();

    /* > 修改统计计数 */
    key.addr = addr;

    item = hash_tab_query(tab, (void *)&key, RDLOCK);
    if (NULL == item) {
        assert(0);
        return 0; // Didn't find
    }

    cnt = (int)atomic32_dec(&item->count);

    hash_tab_unlock(tab, &key, RDLOCK);

    /* > 是否释放内存 */
    if (0 == cnt) {
        item = hash_tab_query(tab, (void *)&key, WRLOCK);
        if (NULL == item) {
            return 0; // 已被释放
        } else if (0 == item->count) {
            hash_tab_delete(tab, (void *)&key, NONLOCK);
            hash_tab_unlock(tab, &key, WRLOCK);

            item->dealloc(item->pool, item->addr); // 释放被管理的内存
            free(item);
            return 0;
        }
        hash_tab_unlock(tab, &key, WRLOCK);
        return 0;
    }
    return cnt;
}

/******************************************************************************
 **函数名称: mref_check
 **功    能: 内存引用检测
 **输入参数:
 **     addr: 内存地址
 **输出参数: NONE
 **返    回: 引用次数
 **实现描述:
 **注意事项: 
 **作    者: # Qifeng.zou # 2016.09.08 #
 ******************************************************************************/
int mref_check(void *addr)
{
    int cnt;
    mref_item_t *item, key;
    hash_tab_t *tab = GetMemRef();

    /* > 查询引用 */
    key.addr = addr;

    item = (mref_item_t *)hash_tab_query(tab, (void *)&key, RDLOCK);
    if (NULL != item) {
        cnt = item->count;
        hash_tab_unlock(tab, &key, RDLOCK);
        return cnt;
    }

    return 0;
}

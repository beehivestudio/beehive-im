#if !defined(__RTMQ_SUB_H__)
#define __RTMQ_SUB_H__

#include "comm.h"
#include "mesg.h"
#include "vector.h"
#include "hash_tab.h"

/* 订阅连接 */
typedef struct
{
    uint64_t sid;               /* 会话ID */
    int nid;                    /* 订阅结点ID */
} rtmq_sub_node_t;

/* 订阅分组信息 */
typedef struct
{
    uint32_t gid;               /* 分组ID */
    vector_t *nodes;             /* 订阅结点列表(按组管理rtmq_sub_node_t) */
} rtmq_sub_group_t;

/* 订阅列表 */
typedef struct
{
    uint32_t type;              /* 订阅类型 */
    avl_tree_t *groups;         /* 订阅结点列表(按组管理rtmq_sub_group_t) */
} rtmq_sub_list_t;

/* 订阅管理 */
typedef struct
{
} rtmq_sub_mgr_t;

/* 查找订阅列表group中是否存在指定连接 */
static bool rtmq_sub_group_find_sid_cb(rtmq_sub_node_t *node, uint64_t *sid)
{
    return (node->sid == *sid)? true : false;
}

/* 释放订阅结点 */
static void rtmq_sub_node_dealloc(rtmq_sub_node_t *node)
{
    free(node);
}

#endif /*__RTMQ_SUB_H__*/

package chat_tab

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	CT_ROOM_NUM = 999 // 聊天室列表长度
	CT_GRP_NUM  = 999 // 聊天室分组列表长度
	CT_SSN_NUM  = 999 // 聊天室分组列表长度
)

/* 遍历回调 */
type ChatTravProcCb func(obj interface{}, param interface{}) int

/* 会话信息 */
type ChatSession struct {
	sid          uint64          // 会话ID
	rid          uint64          // 聊天室ID
	gid          uint32          // 分组ID
	sync.RWMutex                 // 读写锁
	sub          map[uint32]bool // 订阅列表(true:订阅 false:未订阅)
}

/* SESSION SET信息 */
type ChatSessionSet struct {
	sync.RWMutex                         // 读写锁
	session      map[uint64]*ChatSession // 会话集合[sid]*ChatSession
}

/* 分组信息 */
type ChatGroup struct {
	gid          uint64          // 分组ID
	sid_num      uint64          // 会话数目
	create_tm    int64           // 创建时间
	sync.RWMutex                 // 读写锁
	sid_list     map[uint64]bool // 会话列表[sid]bool
}

/* GROUP SET信息 */
type ChatGroupSet struct {
	sync.RWMutex                       // 读写锁
	group        map[uint32]*ChatGroup // 分组信息[gid]*ChatGroup
}

/* ROOM信息 */
type ChatRoom struct {
	rid       uint64                     // 聊天室ID
	sid_num   uint64                     // 会话数目
	grp_num   uint32                     // 分组数目
	create_tm int64                      // 创建时间
	groups    [CT_GROUP_NUM]ChatGroupSet // 分组信息
}

/* ROOM SET信息 */
type ChatRoomSet struct {
	sync.RWMutex                      // 读写锁
	room         map[uint64]*ChatRoom // ROOM集合[rid]*ChatRoom
}

/* 全局对象 */
type ChatTab struct {
	rooms    [CT_ROOM_NUM]ChatRoomSet   // ROOM信息
	sessions [CT_SSN_NUM]ChatSessionSet // SESSION信息
}

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述: 初始化ctx成员变量
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:46:50 #
 ******************************************************************************/
func Init() *ChatTab {
	ctx := &ChatTab{}

	/* 初始化ROOM SET */
	for idx := 0; idx < CT_ROOM_NUM; idx += 1 {
		rs := ctx.rooms[idx]
		rs.room = make(map[uint64]*ChatRoom)
	}

	/* 初始化SESSION SET */
	for idx := 0; idx < CT_SSN_NUM; idx += 1 {
		ss := ctx.sessions[idx]
		ss.session = make(map[uint64]*ChatSession)
	}

	return ctx
}

/******************************************************************************
 **函数名称: SessionAdd
 **功    能: 会话添加
 **输入参数:
 **     rid: ROOM ID
 **     gid: 分组GID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 将会话SID挂载到聊天室指定分组上.
 **注意事项: 各层级读写锁的操作, 降低锁粒度, 防止死锁.
 **作    者: # Qifeng.zou # 2017.02.22 20:32:20 #
 ******************************************************************************/
func (ctx *ChatTab) SessionAdd(rid uint64, gid uint32, sid uint64) int {
	/* > 加入会话管理表 */
	ok := ctx.session_add(rid, gid, sid)
	if ok {
		return -1 // 异常
	}

ROOM:
	ok = ctx.room_add(rid)
	if ok {
		return -1 // 异常
	}

	room := ctx.room_query(rid, comm.RDLOCK)
	if nil == room {
		goto ROOM
	}

	defer ctx.room_unlock(rid, comm.RDLOCK)

GROUP:
	ok = ctx.group_add(room, gid)
	if ok {
		return -1 // 异常
	}

	group := ctx.group_query(room, gid, comm.RDLOCK)
	if nil == group {
		goto GROUP
	}

	defer ctx.group_unlock(room, gid, comm.RDLOCK)

SIDLIST:
	group.Lock()
	defer group.Unlock()

	ssn, ok = group.sid_list[sid]
	if !ok {
		group.sid_list[sid] = true
		return 0
	}

	return 0
}

/******************************************************************************
 **函数名称: SessionDel
 **功    能: 会话删除
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 20:54:53 #
 ******************************************************************************/
func (ctx *ChatTab) SessionDel(sid uint64) int {
	/* > 判断会话是否存在 */
	ssn, ok := ctx.session[sid]
	if !ok {
		return 0 // 无数据
	}

	delete(ctx.session, sid)

	/* > 清理ROOM信息 */
	room, ok := ctx.room[rid]
	if !ok {
		return 0 // 无数据
	}

	atomic.AddUint64(&room.sid_num, -1)

	group, ok := room.group[gid]
	if !ok {
		return 0 // 无数据
	}

	atomic.AddUint64(&group.sid_num, -1)

	delete(group.sid_list, sid)

	return 0
}

/******************************************************************************
 **函数名称: SubAdd
 **功    能: 订阅添加
 **输入参数:
 **     sid: 会话ID
 **     cmd: 订阅消息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 21:16:08 #
 ******************************************************************************/
func (ctx *ChatTab) SubAdd(sid uint64, cmd uint32) int {
	ctx.ssn_lck.RLock()
	defer ctx.ssn_lck.RUnlock()

	ssn, ok := ctx.session[sid]
	if !ok {
		return -1 // 无数据
	}

	ssn.Lock()
	ssn.sub[cmd] = true
	ssn.Unlock()

	return 0
}

/******************************************************************************
 **函数名称: SubDel
 **功    能: 订阅删除
 **输入参数:
 **     sid: 会话ID
 **     cmd: 订阅消息
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 21:23:25 #
 ******************************************************************************/
func (ctx *ChatTab) SubDel(sid uint64, cmd uint32) int {
	ctx.ssn_lck.RLock()
	defer ctx.ssn_lck.RUnlock()

	ssn, ok := ctx.session[sid]
	if !ok {
		return -1 // 无数据
	}

	ssn.Lock()
	delete(ssn.sub, cmd)
	ssn.Unlock()

	return 0
}

/******************************************************************************
 **函数名称: IsSub
 **功    能: 是否订阅
 **输入参数:
 **     sid: 会话ID
 **     cmd: 订阅消息
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 21:31:37 #
 ******************************************************************************/
func (ctx *ChatTab) IsSub(sid uint64, cmd uint32) bool {
	ctx.ssn_lck.RLock()
	defer ctx.ssn_lck.RUnlock()

	ssn, ok := ctx.session[sid]
	if !ok {
		return false
	}

	ssn.Lock()
	defer ssn.Unlock()

	v, ok := ssn.sub[cmd]
	if !ok {
		return false
	}

	return v
}

/******************************************************************************
 **函数名称: Trav
 **功    能: 遍历聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:52:36 #
 ******************************************************************************/
func (ctx *ChatTab) Trav(rid uint64, gid uint32, proc ChatTravProcCb, param interface{}) int {
	ctx.room_lck.RLock()
	defer ctx.room_lck.RUnlock()

	room, ok := ctx.room[rid]
	if !ok {
		return -1
	} else if 0 == gid {
		return ctx.trav_all_group(room, proc, param)
	}

	room.RLock()
	defer room.RUnlock()

	group, ok := room.group[gid]
	if !ok {
		return -1
	}

	return ctx.trav_group(room, group, proc, param)
}

/******************************************************************************
 **函数名称: Clean
 **功    能: 清理人数为空的聊天室信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.23 22:10:28 #
 ******************************************************************************/
func (ctx *ChatTab) Clean() int {
	list := make(map[uint64]bool)

	/* > 过滤连接数为0的聊天室 */
	ctx.room_lck.RLock()
	for rid, room := range ctx.room {
		if 0 != room.sid_num {
			continue
		}
		list[rid] = true
	}
	ctx.room_lck.RUnlock()

	/* > 清理连接数为0的聊天室 */
	for rid, _ := range list {
		ctx.room_lck.Lock()
		room, ok := ctx.room[rid]
		if !ok {
			ctx.room_lck.Unlock()
			continue
		}
		delete(ctx.room, rid)
		ctx.room_lck.Unlock()

		room.clean(ctx)
	}
}

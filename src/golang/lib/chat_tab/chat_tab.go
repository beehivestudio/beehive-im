package chat_tab

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

type ChatTravProcCb func(data []byte, param interface{}) int

/* 会话信息 */
type ChatSession struct {
	sid          uint64          // 会话ID
	rid          uint64          // 聊天室ID
	gid          uint32          // 分组ID
	sync.RWMutex                 // 读写锁
	sub          map[uint32]bool // 订阅列表(true:订阅 false:未订阅)
}

/* 分组信息 */
type ChatGroup struct {
	gid          uint64          // 分组ID
	sid_num      uint64          // 会话数目
	create_tm    int64           // 创建时间
	sync.RWMutex                 // 读写锁
	sid_list     map[uint64]bool // 会话列表
}

/* ROOM信息 */
type ChatRoom struct {
	rid          uint64               // 聊天室ID
	sid_num      uint64               // 会话数目
	grp_num      uint32               // 分组数目
	create_tm    int64                // 创建时间
	sync.RWMutex                      // 读写锁
	group        map[uint32]ChatGroup // 分组信息
}

/* 全局对象 */
type ChatTab struct {
	room_lck sync.RWMutex           // 读写锁
	room     map[uint64]ChatRoom    // ROOM集合
	ssn_lck  sync.RWMutex           // 读写锁
	session  map[uint64]ChatSession // 会话集合
}

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:46:50 #
 ******************************************************************************/
func Init() *ChatTab {
	ctx := &ChatTab{
		room:    make(map[uint64]ChatRoom),
		session: make(map[uint64]ChatSession),
	}

	return ctx
}

/******************************************************************************
 **函数名称: SessionAdd
 **功    能: 会话添加
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 20:32:20 #
 ******************************************************************************/
func (ctx *ChatTab) SessionAdd(rid uint64, gid uint32, sid uint64) int {
	/* > 判断会话是否存在 */
	ssn, ok := ctx.session[sid]
	if ok {
		if ssn.rid == rid && ssn.gid == gid {
			return 0 // 已存在
		}
		return -1 // 数据不一致
	}

	/* > 添加会话信息 */
	ssn = &ChatSession{
		sid: sid,                   // 会话ID
		rid: rid,                   // 聊天室ID
		gid: gid,                   // 分组ID
		sub: make(map[uint32]bool), // 订阅列表
	}

	ctx.session[idx] = ssn

	/* > 添加ROOM信息 */
	room, ok := ctx.room[rid]
	if !ok {
		room = &ChatRoom{
			rid:       rid,                        // 聊天室ID
			sid_num:   0,                          // 会话数目
			grp_num:   0,                          // 分组数目
			create_tm: time.Now().Unix(),          // 创建时间
			group:     make(map[uint32]ChatGroup), // 分组信息
		}
		atomic.AddUint64(&room.sid_num, 1)
		ctx.room[rid] = room
	}

	group, ok := room.group[gid]
	if !ok {
		group = &ChatGroup{
			gid:       gid,                   // 分组ID
			sid_num:   0,                     // 会话数目
			create_tm: time.Now().Unix(),     // 创建时间
			sid_list:  make(map[uint64]bool), // 会话列表
		}
		atomic.AddUint64(&group.sid_num, 1)
		room.group[gid] = group
	}

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
	ctx.ssn_lck.Rlock()
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
	ctx.ssn_lck.Rlock()
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
	ctx.ssn_lck.Rlock()
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
func (ctx *ChatTab) Trav(rid uint64, gid uint32, ChatTravProcCb proc, param interface{}) int {
	room, ok := ctx.room[rid]
	if !ok {
		return -1
	} else if 0 == gid {
		return ctx.trav_all_group(room, proc, param)
	}

	group, ok := room.group[gid]
	if !ok {
		return -1
	}

	return ctx.trav_group(room, group, proc, param)
}

/******************************************************************************
 **函数名称: TimeoutClean
 **功    能: 清理超时数据
 **输入参数: NONE
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:57:50 #
 ******************************************************************************/
func (ctx *ChatTab) TimeoutClean(rid uint64, gid uint32, ChatTravProcCb proc, param interface{}) {
}

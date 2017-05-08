package chat_tab

import (
	"sync"

	"beehive-im/src/golang/lib/comm"
)

const (
	ROOM_MAX_LEN    = 999 // 聊天室列表长度
	GROUP_MAX_LEN   = 99  // 聊天室分组列表长度
	SESSION_MAX_LEN = 999 // 聊天室分组列表长度
)

/* 遍历回调 */
type ChatTravProcCb func(sid uint64, cid uint64, param interface{}) int

/* 会话信息 */
type ChatSessionItem struct {
	sid          uint64            // 会话SID
	cid          uint64            // 连接CID
	sync.RWMutex                   // 读写锁
	room         map[uint64]uint32 // 聊天室信息map[rid]gid
	sub          map[uint32]bool   // 订阅列表(true:订阅 false:未订阅)
	param        interface{}       // 扩展数据
}

/* SESSION ITEM主键 */
type ChatSessionKey struct {
	sid uint64 // 会话ID
	cid uint64 // 连接ID
}

/* SESSION TAB信息 */
type ChatSessionList struct {
	sync.RWMutex                                     // 读写锁
	session      map[ChatSessionKey]*ChatSessionItem // 会话集合[sid&cid]*ChatSessionItem
}

/* SID->CID信息 */
type ChatSid2CidList struct {
	sync.RWMutex                   // 读写锁
	sid2cid      map[uint64]uint64 // 会话集合[sid]cid
}

/* 分组信息 */
type ChatGroupItem struct {
	gid          uint32                  // 分组ID
	sid_num      int64                   // 会话数目
	create_tm    int64                   // 创建时间
	sync.RWMutex                         // 读写锁
	sid_list     map[ChatSessionKey]bool // 会话列表[sid&cid]bool
}

/* GROUP TAB信息 */
type ChatGroupList struct {
	sync.RWMutex                           // 读写锁
	group        map[uint32]*ChatGroupItem // 分组信息[gid]*ChatGroupItem
}

/* ROOM信息 */
type ChatRoomItem struct {
	rid       uint64                       // 聊天室ID
	sid_num   int64                        // 会话数目
	grp_num   int32                        // 分组数目
	create_tm int64                        // 创建时间
	groups    [GROUP_MAX_LEN]ChatGroupList // 分组信息
}

/* ROOM TAB信息 */
type ChatRoomList struct {
	sync.RWMutex                          // 读写锁
	room         map[uint64]*ChatRoomItem // ROOM集合[rid]*ChatRoomItem
}

/* 全局对象 */
type ChatTab struct {
	rooms    [ROOM_MAX_LEN]ChatRoomList       // ROOM信息
	sessions [SESSION_MAX_LEN]ChatSessionList // SESSION信息
	sid2cids [SESSION_MAX_LEN]ChatSid2CidList // SID->CID映射
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
	for idx := 0; idx < ROOM_MAX_LEN; idx += 1 {
		rs := &ctx.rooms[idx]
		rs.room = make(map[uint64]*ChatRoomItem)
	}

	/* 初始化SESSION信息 */
	for idx := 0; idx < SESSION_MAX_LEN; idx += 1 {
		ss := &ctx.sessions[idx]
		ss.session = make(map[ChatSessionKey]*ChatSessionItem)
	}

	/* 初始化SID->CID映射 */
	for idx := 0; idx < SESSION_MAX_LEN; idx += 1 {
		ss := &ctx.sid2cids[idx]
		ss.sid2cid = make(map[uint64]uint64)
	}

	return ctx
}

/******************************************************************************
 **函数名称: RoomJoin
 **功    能: 加入聊天室
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
func (ctx *ChatTab) RoomJoin(rid uint64, gid uint32, sid uint64, cid uint64) int {
	/* > 加入会话管理表 */
	ret := ctx.session_join_room(rid, gid, sid, cid)
	if 0 != ret {
		return -1 // 异常
	}

ROOM:
	ret = ctx.room_add(rid)
	if 0 != ret {
		return -1 // 异常
	}

	room := ctx.room_query(rid, comm.RDLOCK)
	if nil == room {
		goto ROOM
	}

	defer ctx.room_unlock(rid, comm.RDLOCK)

GROUP:
	ret = room.group_add(gid)
	if 0 != ret {
		return -1 // 异常
	}

	group := room.group_query(gid, comm.RDLOCK)
	if nil == group {
		goto GROUP
	}

	defer room.group_unlock(gid, comm.RDLOCK)

	group.Lock()
	defer group.Unlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	_, ok := group.sid_list[*key]
	if !ok {
		group.sid_list[*key] = true
		return 0
	}

	return 0
}

/******************************************************************************
 **函数名称: RoomQuit
 **功    能: 退出聊天室
 **输入参数:
 **     rid: ROOM ID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 移除会话SID在指定聊天室RID中的记录.
 **注意事项: 各层级读写锁的操作, 降低锁粒度, 防止死锁.
 **作    者: # Qifeng.zou # 2017.03.08 22:02:17 #
 ******************************************************************************/
func (ctx *ChatTab) RoomQuit(rid uint64, sid uint64, cid uint64) int {
	/* > 移除指定聊天室数据 */
	gid, ok := ctx.session_quit_room(rid, sid)
	if !ok {
		return 0 // 无数据
	}

	return ctx.room_del_session(rid, gid, sid, cid)
}

/******************************************************************************
 **函数名称: GetCidBySid
 **功    能: 通过SID获取CID
 **输入参数:
 **     sid: 会话ID
 **输出参数: NONE
 **返    回: 连接CID
 **实现描述: 从sid2cid映射中查询
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.07 07:24:06 #
 ******************************************************************************/
func (ctx *ChatTab) GetCidBySid(sid uint64) uint64 {
	ss := &ctx.sid2cids[sid%SESSION_MAX_LEN]

	ss.RLock()
	defer ss.RUnlock()

	cid, ok := ss.sid2cid[sid]
	if !ok {
		return 0
	}

	return cid
}

/******************************************************************************
 **函数名称: SessionSetCid
 **功    能: 设置SID->CID映射
 **输入参数:
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.07 07:42:07 #
 ******************************************************************************/
func (ctx *ChatTab) SessionSetCid(sid uint64, cid uint64) int {
	ss := &ctx.sid2cids[sid%SESSION_MAX_LEN]

	ss.Lock()
	defer ss.Unlock()

	ss.sid2cid[sid] = cid

	return 0
}

/******************************************************************************
 **函数名称: SessionDel
 **功    能: 会话删除
 **输入参数:
 **     sid: 会话ID
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **     1. 清理会话表中的数据
 **     2. 清理聊天室各层级数据
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 20:54:53 #
 ******************************************************************************/
func (ctx *ChatTab) SessionDel(sid uint64, cid uint64) int {
	/* > 清理会话数据 */
	ssn := ctx.session_del(sid, cid)
	if nil == ssn {
		return 0 // 无数据
	}

	/* > 清理ROOM会话数据 */
	for rid, gid := range ssn.room {
		ctx.room_del_session(rid, gid, sid, cid)
	}

	return 0
}

/******************************************************************************
 **函数名称: SessionSetParam
 **功    能: 设置会话参数
 **输入参数:
 **     sid: 会话SID
 **     param: 扩展数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 20:32:20 #
 ******************************************************************************/
func (ctx *ChatTab) SessionSetParam(sid uint64, cid uint64, param interface{}) int {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	ss.Lock()
	defer ss.Unlock()

	/* > 判断会话是否存在 */
	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if ok {
		return -1 // 已存在
	}

	/* > 添加会话信息 */
	ssn = &ChatSessionItem{
		sid:   sid,                     // 会话ID
		cid:   cid,                     // 连接ID
		room:  make(map[uint64]uint32), // 聊天室信息
		sub:   make(map[uint32]bool),   // 订阅列表
		param: param,                   // 扩展数据
	}

	ss.session[*key] = ssn

	return 0
}

/******************************************************************************
 **函数名称: SessionGetParam
 **功    能: 获取会话参数
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: 扩展数据
 **实现描述:
 **注意事项: 各层级读写锁的操作, 降低锁粒度, 防止死锁.
 **作    者: # Qifeng.zou # 2017.03.07 17:02:35 #
 ******************************************************************************/
func (ctx *ChatTab) SessionGetParam(sid uint64, cid uint64) (param interface{}) {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	ss.RLock()
	defer ss.RUnlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if !ok {
		return nil
	}

	return ssn.param // 已存在
}

/******************************************************************************
 **函数名称: SessionCount
 **功    能: 获取会话总数
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 会话总数
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.21 18:37:55 #
 ******************************************************************************/
func (ctx *ChatTab) SessionCount() uint32 {
	var total uint32

	for idx := 0; idx < SESSION_MAX_LEN; idx += 1 {
		item := &ctx.sessions[idx%SESSION_MAX_LEN]

		item.RLock()
		total += uint32(len(item.session))
		item.RUnlock()
	}
	return total
}

/******************************************************************************
 **函数名称: SessionInRoom
 **功    能: 判断会话是否在聊天室中
 **输入参数:
 **     sid: 会话SID
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 1.分组ID 2.是否存在
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.09 23:24:03 #
 ******************************************************************************/
func (ctx *ChatTab) SessionInRoom(sid uint64, rid uint64) (gid uint32, ok bool) {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	cid := ctx.GetCidBySid(sid)
	if 0 == cid {
		return 0, false // 不存在
	}

	ss.RLock()
	defer ss.RUnlock()

	/* > 判断会话是否存在 */
	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if ok {
		return 0, false // 已存在
	}

	gid, ok = ssn.room[rid]

	return gid, ok
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
 **作    者: # Qifeng.zou # 2017.03.02 10:31:44 #
 ******************************************************************************/
func (ctx *ChatTab) SubAdd(sid uint64, cmd uint32) int {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	cid := ctx.GetCidBySid(sid)
	if 0 == cid {
		return 0 // 不存在
	}

	ss.RLock()
	defer ss.RUnlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if !ok {
		return -1 // 无数据
	}

	ssn.Lock()
	defer ssn.Unlock()

	ssn.sub[cmd] = true

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
 **实现描述: 移除会话对象中sub[cmd]便可.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.22 21:23:25 #
 ******************************************************************************/
func (ctx *ChatTab) SubDel(sid uint64, cmd uint32) int {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	cid := ctx.GetCidBySid(sid)
	if 0 == cid {
		return 0 // 不存在
	}

	ss.RLock()
	defer ss.RUnlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if !ok {
		return 0 // 无数据
	}

	ssn.Lock()
	defer ssn.Unlock()

	delete(ssn.sub, cmd)

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
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	cid := ctx.GetCidBySid(sid)
	if 0 == cid {
		return false // 不存在
	}

	ss.RLock()
	defer ss.RUnlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if !ok {
		return false // 无数据
	}

	ssn.RLock()
	defer ssn.RUnlock()

	v, ok := ssn.sub[cmd]

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
	rs := &ctx.rooms[rid%ROOM_MAX_LEN]

	rs.RLock()
	defer rs.RUnlock()

	room, ok := rs.room[rid]
	if !ok {
		return -1 // 无数据
	} else if 0 == gid {
		return room.group_all_trav(proc, param) // 遍历所有分组
	}

	gs := &room.groups[gid%GROUP_MAX_LEN]

	gs.RLock()
	defer gs.RUnlock()

	group, ok := gs.group[gid]
	if !ok {
		return -1 // 无数据
	}

	return room.group_trav(group, proc, param)
}

/******************************************************************************
 **函数名称: Clean
 **功    能: 清理人数为空的聊天室信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 无需对room对象中的各个数据逐一处理, 其内存会被自动回收.
 **作    者: # Qifeng.zou # 2017.02.23 22:10:28 #
 ******************************************************************************/
func (ctx *ChatTab) Clean() int {
	rlist := make(map[uint64]bool)

	/* > 过滤会话数为0的聊天室 */
	for _, rs := range ctx.rooms {
		rs.RLock()
		for rid, room := range rs.room {
			if 0 != room.sid_num {
				continue
			}
			rlist[rid] = true
		}
		rs.RUnlock()
	}

	/* > 清理会话数为0的聊天室 */
	for rid, _ := range rlist {
		rs := &ctx.rooms[rid%ROOM_MAX_LEN]
		rs.Lock()
		room, ok := rs.room[rid]
		if !ok || 0 != room.sid_num {
			rs.Unlock()
			continue
		}
		delete(rs.room, rid)
		rs.Unlock()
	}
	return 0
}

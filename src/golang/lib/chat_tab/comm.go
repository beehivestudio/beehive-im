package chat_tab

import (
	"sync/atomic"
	"time"

	"beehive-im/src/golang/lib/comm"
)

////////////////////////////////////////////////////////////////////////////////
// 会话操作

/******************************************************************************
 **函数名称: session_join_room
 **功    能: 会话JOIN聊天室
 **输入参数:
 **     rid: ROOM ID
 **     gid: 分组GID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 当会话SID不存在时, 则新增.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 23:37:33 #
 ******************************************************************************/
func (ctx *ChatTab) session_join_room(rid uint64, gid uint32, sid uint64, cid uint64) int {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	key := &ChatSessionKey{sid: sid, cid: cid}

	ss.Lock()
	defer ss.Unlock()

	ssn, ok := ss.session[*key]
	if ok {
		_gid, ok := ssn.room[rid]
		if !ok {
			ssn.room[rid] = gid
			return 0
		} else if _gid == gid {
			return 0 // 已存在
		}
		return -1 // 数据不一致
	}

	/* > 添加会话信息 */
	ssn = &ChatSessionItem{
		sid:  sid,                     // 会话ID
		cid:  cid,                     // 连接ID
		room: make(map[uint64]uint32), // 聊天室信息
		sub:  make(map[uint32]bool),   // 订阅列表
	}

	ssn.room[rid] = gid

	ss.session[*key] = ssn

	return 0
}

/******************************************************************************
 **函数名称: session_quit_room
 **功    能: 会话QUIT聊天室
 **输入参数:
 **     rid: ROOM ID
 **     gid: 分组GID
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 清理会话SID信息中的聊天室RID数据.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.08 22:36:01 #
 ******************************************************************************/
func (ctx *ChatTab) session_quit_room(rid uint64, sid uint64) (gid uint32, ok bool) {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	ss.Lock()
	defer ss.Unlock()

	cid := ctx.GetCidBySid(sid)

	key := &ChatSessionKey{sid: sid, cid: cid}

	/* > 判断会话是否存在 */
	ssn, ok := ss.session[*key]
	if ok {
		gid, ok := ssn.room[rid]
		if !ok {
			delete(ssn.room, rid)
			return gid, true
		}
		return 0, false // 数据不一致
	}
	return 0, false
}

/******************************************************************************
 **函数名称: session_del
 **功    能: 移除会话管理表
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: 会话数据
 **实现描述: 从session[]表中删除sid的会话数据
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.02 10:13:34 #
 ******************************************************************************/
func (ctx *ChatTab) session_del(sid uint64, cid uint64) *ChatSessionItem {
	ss := &ctx.sessions[sid%SESSION_MAX_LEN]

	ss.Lock()
	defer ss.Unlock()

	/* > 判断会话是否存在 */
	key := &ChatSessionKey{sid: sid, cid: cid}

	ssn, ok := ss.session[*key]
	if !ok {
		return nil // 无数据
	}
	delete(ss.session, *key)

	return ssn
}

/******************************************************************************
 **函数名称: sid2cid_del
 **功    能: 移除映射信息
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.11 10:03:30 #
 ******************************************************************************/
func (ctx *ChatTab) sid2cid_del(sid uint64, cid uint64) int {
	sc := &ctx.sid2cids[sid%SESSION_MAX_LEN]

	sc.Lock()
	defer sc.Unlock()

	_cid, ok := sc.sid2cid[sid]
	if !ok {
		return 0 // 无数据
	} else if _cid == cid {
		delete(sc.sid2cid, sid)
	}

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// 会话操作

/******************************************************************************
 **函数名称: trav_list
 **功    能: 遍历会话列表
 **输入参数:
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.10 21:05:08 #
 ******************************************************************************/
func (sl *ChatSessionList) trav_list(proc ChatTravSessionProcCb, param interface{}) {
	sl.RLock()
	defer sl.RUnlock()

	for k, v := range sl.session {
		proc(k.sid, k.cid, v.param, param)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 映射操作

/******************************************************************************
 **函数名称: trav_list
 **功    能: 遍历映射列表
 **输入参数:
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.11 10:07:32 #
 ******************************************************************************/
func (sc *ChatSid2CidList) trav_list(proc ChatTravProcCb, param interface{}) {
	sc.RLock()
	defer sc.RUnlock()

	for sid, cid := range sc.sid2cid {
		proc(sid, cid, param)
	}
}

/******************************************************************************
 **函数名称: query_dirty
 **功    能: 查找脏数据(不一致的数据)
 **输入参数:
 **     ls: 脏数据列表
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.11 10:32:16 #
 ******************************************************************************/
func (sc *ChatSid2CidList) query_dirty(ctx *ChatTab, ls map[uint64]uint64) {
	sc.RLock()
	defer sc.RUnlock()

	for sid, cid := range sc.sid2cid {
		extra := ctx.SessionGetParam(sid, cid)
		if nil != extra {
			continue
		}
		ls[sid] = cid
	}
}

/******************************************************************************
 **函数名称: task_clean_sid2cid
 **功    能: 清理SID->CID映射数据(防止映射始终存在导致内存泄露)
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.11 08:53:21 #
 ******************************************************************************/
func (ctx *ChatTab) task_clean_sid2cid() {
	for {
		list := make(map[uint64]uint64)

		for idx := 0; idx < SESSION_MAX_LEN; idx += 1 {
			sc := &ctx.sid2cids[idx]

			sc.query_dirty(ctx, list)
		}

		for _, sid := range list {
			ctx.sid2cid_del(sid, list[sid])
		}

		time.Sleep(15 * time.Second)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室操作

/******************************************************************************
 **函数名称: room_add
 **功    能: 新建聊天室
 **输入参数:
 **     rid: 聊天室ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 判断聊天室不存在的话, 则新建聊天室.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 22:55:43 #
 ******************************************************************************/
func (ctx *ChatTab) room_add(rid uint64) int {
	rs := &ctx.rooms[rid%ROOM_MAX_LEN]

	rs.Lock()
	defer rs.Unlock()

	room, ok := rs.room[rid]
	if !ok {
		room = &ChatRoomItem{
			rid:       rid,               // 聊天室ID
			sid_num:   0,                 // 会话数目
			grp_num:   0,                 // 分组数目
			create_tm: time.Now().Unix(), // 创建时间
		}

		for idx := uint32(0); idx < GROUP_MAX_LEN; idx += 1 {
			room.groups[idx].group = make(map[uint32]*ChatGroupItem)
		}

		atomic.AddInt64(&room.sid_num, 1)
		rs.room[rid] = room
		return 0
	}

	return 0
}

/******************************************************************************
 **函数名称: room_query
 **功    能: 查找聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     lck: 上锁方式(0:不加锁 1:读锁 2:写锁)
 **输出参数: NONE
 **返    回: 聊天室对象
 **实现描述:
 **注意事项: 如果获取对象失败, 则直接解锁.
 **作    者: # Qifeng.zou # 2017.03.01 22:59:15 #
 ******************************************************************************/
func (ctx *ChatTab) room_query(rid uint64, lck int) *ChatRoomItem {
	rs := &ctx.rooms[rid%ROOM_MAX_LEN]
	switch lck {
	case comm.RDLOCK: // 加读锁
		rs.RLock()
	case comm.WRLOCK: // 加写锁
		rs.Lock()
	}

	room, ok := rs.room[rid]
	if !ok {
		switch lck {
		case comm.RDLOCK: // 加读锁
			rs.RUnlock()
		case comm.WRLOCK: // 加写锁
			rs.Unlock()
		}
		return nil
	}

	return room
}

/******************************************************************************
 **函数名称: room_unlock
 **功    能: 解锁聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     lck: 解锁方式(0:不加锁 1:读锁 2:写锁)
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 23:57:28 #
 ******************************************************************************/
func (ctx *ChatTab) room_unlock(rid uint64, lck int) int {
	rs := &ctx.rooms[rid%ROOM_MAX_LEN]

	switch lck {
	case comm.RDLOCK: // 加读锁
		rs.RUnlock()
	case comm.WRLOCK: // 加写锁
		rs.Unlock()
	}

	return 0
}

/******************************************************************************
 **函数名称: room_trav
 **功    能: 遍历聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     lck: 上锁方式(0:不加锁 1:读锁 2:写锁)
 **输出参数: NONE
 **返    回: 聊天室对象
 **实现描述:
 **注意事项: 如果获取对象失败, 则直接解锁.
 **作    者: # Qifeng.zou # 2017.03.01 22:59:15 #
 ******************************************************************************/
func (rl *ChatRoomList) room_trav(proc ChatTravRoomListProcCb, param interface{}) {
	rl.RLock()
	defer rl.RUnlock()

	for _, item := range rl.room {
		proc(item, param)
	}
}

/******************************************************************************
 **函数名称: room_del_session
 **功    能: 移除聊天室指定会话
 **输入参数:
 **     rid: 聊天室ID
 **     gid: 分组ID
 **     sid: 会话ID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述: 判断聊天室不存在的话, 则新建聊天室.
 **注意事项: 尽量使用读锁, 降低锁冲突.
 **作    者: # Qifeng.zou # 2017.03.02 10:20:10 #
 ******************************************************************************/
func (ctx *ChatTab) room_del_session(rid uint64, gid uint32, sid uint64, cid uint64) int {
	rs := &ctx.rooms[rid%ROOM_MAX_LEN]

	/* > 查找ROOM对象 */
	rs.RLock()
	defer rs.RUnlock()

	room, ok := rs.room[rid]
	if !ok {
		return 0 // 无数据
	}

	/* > 查找GROUP对象 */
	gs := &room.groups[gid%GROUP_MAX_LEN]

	gs.RLock()
	defer gs.RUnlock()

	group, ok := gs.group[gid]
	if !ok {
		return 0 // 无数据
	}

	/* > 清理会话数据 */
	group.Lock()
	defer group.Unlock()

	key := &ChatSessionKey{sid: sid, cid: cid}

	_, ok = group.sid_list[*key]
	if !ok {
		return 0 // 无数据
	}

	delete(group.sid_list, *key)

	atomic.AddInt64(&room.sid_num, -1)  // 人数减1
	atomic.AddInt64(&group.sid_num, -1) // 人数减1

	return 0
}

////////////////////////////////////////////////////////////////////////////////
// 聊天室分组操作

/******************************************************************************
 **函数名称: group_add
 **功    能: 新建分组
 **输入参数:
 **     gid: 分组GID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 22:55:43 #
 ******************************************************************************/
func (room *ChatRoomItem) group_add(gid uint32) int {
	gs := &room.groups[gid%GROUP_MAX_LEN]

	gs.Lock()
	defer gs.Unlock()

	group, ok := gs.group[gid]
	if !ok {
		group = &ChatGroupItem{
			gid:       gid,                           // 分组ID
			sid_num:   0,                             // 会话数目
			create_tm: time.Now().Unix(),             // 创建时间
			sid_list:  make(map[ChatSessionKey]bool), // 会话列表
		}

		atomic.AddInt64(&group.sid_num, 1)

		gs.group[gid] = group
	}

	return 0
}

/******************************************************************************
 **函数名称: group_query
 **功    能: 查找聊天室指定分组
 **输入参数:
 **     gid: 群组ID
 **输出参数: NONE
 **返    回: 群组对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 23:50:25 #
 ******************************************************************************/
func (room *ChatRoomItem) group_query(gid uint32, lck int) *ChatGroupItem {
	gs := &room.groups[gid%GROUP_MAX_LEN]

	switch lck {
	case comm.RDLOCK: // 加读锁
		gs.RLock()
	case comm.WRLOCK: // 加写锁
		gs.Lock()
	}

	group, ok := gs.group[gid]
	if !ok {
		switch lck {
		case comm.RDLOCK: // 加读锁
			gs.RUnlock()
		case comm.WRLOCK: // 加写锁
			gs.Unlock()
		}
		return nil
	}

	return group
}

/******************************************************************************
 **函数名称: group_unlock
 **功    能: 解锁聊天室分组
 **输入参数:
 **     gid: 分组ID
 **     lck: 解锁方式(0:不加锁 1:读锁 2:写锁)
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 17:44:36 #
 ******************************************************************************/
func (room *ChatRoomItem) group_unlock(gid uint32, lck int) int {
	gs := &room.groups[gid%GROUP_MAX_LEN]

	switch lck {
	case comm.RDLOCK: // 加读锁
		gs.RUnlock()
	case comm.WRLOCK: // 加写锁
		gs.Unlock()
	}

	return 0
}

/******************************************************************************
 **函数名称: group_trav
 **功    能: 遍历聊天室指定分组
 **输入参数:
 **     group: 聊天室群组
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.23 20:26:40 #
 ******************************************************************************/
func (room *ChatRoomItem) group_trav(group *ChatGroupItem, proc ChatTravProcCb, param interface{}) int {
	group.RLock()
	defer group.RUnlock()

	for key, _ := range group.sid_list {
		proc(key.sid, key.cid, param)
	}
	return 0
}

/******************************************************************************
 **函数名称: group_all_trav
 **功    能: 遍历聊天室所有分组
 **输入参数:
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.02 13:45:05 #
 ******************************************************************************/
func (room *ChatRoomItem) group_all_trav(proc ChatTravProcCb, param interface{}) int {
	for _, gs := range room.groups {
		gs.RLock()
		for _, group := range gs.group {
			room.group_trav(group, proc, param)
		}
		gs.RUnlock()
	}
	return 0
}

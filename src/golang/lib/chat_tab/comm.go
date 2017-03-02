package chat_tab

import (
	"fmt"
	"strconv"
	"sync/atomic"
)

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: session_add
 **功    能: 加入会话管理表
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
func (ctx *ChatTab) session_add(rid uint64, gid uint32, sid uint64) int {
	ss := ctx.sessions[sid%CT_SSN_NUM]

	/* > 判断会话是否存在 */
	ss.RLock()
	ssn, ok := ss.session[sid]
	if ok {
		if ssn.rid == rid && ssn.gid == gid {
			ss.RUnlock()
			return 0 // 已存在
		}
		ss.RUnlock()
		return -1 // 数据不一致
	}
	ss.RUnlock()

	/* > 添加会话信息 */
	ssn = &ChatSession{
		sid: sid,                   // 会话ID
		rid: rid,                   // 聊天室ID
		gid: gid,                   // 分组ID
		sub: make(map[uint32]bool), // 订阅列表
	}

	ss.Lock()
	ss.session[idx] = ssn
	ss.Unlock()

	return 0
}

/******************************************************************************
 **函数名称: session_del
 **功    能: 移除会话管理表
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 会话数据
 **实现描述: 从session[]表中删除sid的会话数据
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.02 10:13:34 #
 ******************************************************************************/
func (ctx *ChatTab) session_del(sid uint64) *ChatSession {
	ss := ctx.sessions[sid%CT_SSN_NUM]

	ss.Lock()
	defer ss.Unlock()

	/* > 判断会话是否存在 */
	ssn, ok := ss.session[sid]
	if !ok {
		return nil // 无数据
	}
	delete(ss.session, sid)

	return ssn
}

////////////////////////////////////////////////////////////////////////////////

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
	rs := ctx.rooms[rid%CT_ROOM_NUM]

	rs.Lock()
	defer rs.Unlock()

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
func (ctx *ChatTab) room_query(rid uint64, lck int) *ChatRoom {
	rs := ctx.rooms[rid%CT_ROOM_NUM]
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
func (ctx *ChatTab) room_unlock(rid uint64, lck int) *ChatRoom {
	rs := ctx.rooms[rid%CT_ROOM_NUM]

	switch lck {
	case comm.RDLOCK: // 加读锁
		rs.RUnlock()
	case comm.WRLOCK: // 加写锁
		rs.Unlock()
	}

	return 0
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
func (ctx *ChatTab) room_del_session(rid uint64, gid uint32, sid uint64) int {
	rs := ctx.rooms[rid%CT_ROOM_NUM]

	/* > 查找ROOM对象 */
	rs.RLock()
	defer rs.RUnlock()

	room, ok := rs.room[rid]
	if !ok {
		return 0 // 无数据
	}

	/* > 查找GROUP对象 */
	room.RLock()
	defer room.RUnlock()

	gs := room.groups[gid%CT_GRP_NUM]

	group, ok := gs.group[gid]
	if !ok {
		return 0 // 无数据
	}

	/* > 清理会话数据 */
	group.Lock()
	defer group.Unlock()

	_, ok := group.sid_list[sid]
	if ok {
		return 0 // 无数据
	}

	delete(group.sid_list, sid)

	atomic.AddUint64(&room.sid_num, -1)  // 人数减1
	atomic.AddUint64(&group.sid_num, -1) // 人数减1

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: group_add
 **功    能: 新建分组
 **输入参数:
 **     room: 聊天室
 **     gid: 分组GID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 22:55:43 #
 ******************************************************************************/
func (ctx *ChatTab) group_add(room *ChatRoom, gid uint32) int {
	gs := room.groups[gid%CT_GRP_NUM]

	gs.Lock()
	defer gs.Unlock()

	group, ok := gs.group[gid]
	if !ok {
		group = &ChatGroup{
			gid:       gid,                   // 分组ID
			sid_num:   0,                     // 会话数目
			create_tm: time.Now().Unix(),     // 创建时间
			sid_list:  make(map[uint64]bool), // 会话列表
		}

		atomic.AddUint64(&group.sid_num, 1)

		gs.group[gid] = group
	}

	return 0
}

/******************************************************************************
 **函数名称: group_query
 **功    能: 查找聊天室指定分组
 **输入参数:
 **     room: 聊天室
 **     gid: 群组ID
 **输出参数: NONE
 **返    回: 群组对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.01 23:50:25 #
 ******************************************************************************/
func (ctx *ChatTab) group_query(room *ChatRoom, gid uint32, lck int) *ChatGroup {
	gs := room.groups[gid%CT_GRP_NUM]

	switch lck {
	case comm.RDLOCK: // 加读锁
		gs.RLock()
	case comm.WRLOCK: // 加写锁
		gs.Lock()
	}

	group, ok := gs.group[rid]
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
func (ctx *ChatTab) group_trav(group *ChatGroup, proc ChatTravProcCb, param interface{}) int {
	group.RLock()
	defer group.RUnlock()

	for sid, ok := range group.sid_list {
		proc(sid, param)
	}
}

/******************************************************************************
 **函数名称: group_all_trav
 **功    能: 遍历聊天室所有分组
 **输入参数:
 **     room: 聊天室
 **     proc: 处理回调
 **     param: 附加参数
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.02 13:45:05 #
 ******************************************************************************/
func (ctx *ChatTab) group_all_trav(room *ChatRoom, proc ChatTravProcCb, param interface{}) int {
	for _, gs := range room.groups {
		gs.RLock()
		for _, group := range gs.group {
			ctx.group_trav(group, proc, param)
		}
		gs.RUnlock()
	}
}

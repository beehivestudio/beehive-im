package controllers

import ()

////////////////////////////////////////////////////////////////////////////////
// 处理回调的管理

/******************************************************************************
 **函数名称: Register
 **功    能: 注册处理回调
 **输入参数:
 **     cmd: 消息类型
 **     cb: 处理回调
 **     param: 附加数据
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项: 使用读锁
 **作    者: # Qifeng.zou # 2017.02.09 23:50:28 #
 ******************************************************************************/
func (tab *MesgCallBackTab) Register(cmd uint32, cb MesgCallBack, param interface{}) int {
	item := &MesgCallBackItem{
		cmd:   cmd,
		cb:    cb,
		param: param,
	}

	tab.Lock()
	tab.callback[cmd] = item
	tab.Unlock()

	return 0
}

/******************************************************************************
 **函数名称: Query
 **功    能: 查找处理回调
 **输入参数:
 **     cmd: 消息类型
 **输出参数: NONE
 **返    回:
 **     cb: 回调函数
 **     param: 附加数据
 **实现描述:
 **注意事项: 使用读锁
 **作    者: # Qifeng.zou # 2017.02.09 23:50:28 #
 ******************************************************************************/
func (tab *MesgCallBackTab) Query(cmd uint32) (cb MesgCallBack, param interface{}) {
	tab.RLock()
	item, ok := tab.callback[cmd]
	if !ok {
		tab.RUnlock()
		return nil, nil
	}
	cb := item.cb
	param := item.param
	tab.RUnlock()

	return cb, param
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: add_sid_to_cid
 **功    能: 添加会话SID->连接CID映射
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 14:44:21 #
 ******************************************************************************/
func (ctx *LsndCntx) add_sid_to_cid(sid uint64, cid uint64) int {
	tab := ctx.sid2cid.tab[sid%LSND_SID2CID_LEN]
	tab.Lock()
	tab.list[sid] = cid
	tab.Unlock()
	return 0
}

/******************************************************************************
 **函数名称: find_cid_by_sid
 **功    能: 通过会话SID查找连接CID
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 连接CID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 14:31:19 #
 ******************************************************************************/
func (ctx *LsndCntx) find_cid_by_sid(sid uint64) uint64 {
	tab := ctx.sid2cid.tab[sid%LSND_SID2CID_LEN]
	tab.RLock()
	cid, _ := tab.list[sid]
	tab.RUnlock()
	return cid
}

////////////////////////////////////////////////////////////////////////////////
// 连接扩展数据

/******************************************************************************
 **函数名称: SetSid
 **功    能: 设置会话SID
 **输入参数:
 **     sid: 连接SID
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) SetSid(sid uint64) {
	session.Lock()
	defer session.Unlock()

	session.sid = sid
}

/******************************************************************************
 **函数名称: GetSid
 **功    能: 获取会话SID
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 连接CID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) GetSid() uint64 {
	session.RLock()
	defer session.RUnlock()

	return session.sid
}

/******************************************************************************
 **函数名称: GetCid
 **功    能: 获取连接CID
 **输入参数:
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) SetCid(cid uint64) {
	session.Lock()
	defer session.Unlock()

	session.cid = cid
}

/******************************************************************************
 **函数名称: GetCid
 **功    能: 获取连接CID
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 连接CID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) GetCid() uint64 {
	session.RLock()
	defer session.RUnlock()

	return session.cid
}

/******************************************************************************
 **函数名称: SetStatus
 **功    能: 设置连接状态
 **输入参数:
 **     status: 连接状态
 **输出参数: NONE
 **返    回: 连接状态
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) SetStatus(status int) {
	session.Lock()
	defer session.Unlock()

	session.status = status
}

/******************************************************************************
 **函数名称: GetStatus
 **功    能: 获取连接状态
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 连接状态
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) GetStatus() int {
	session.RLock()
	defer session.RUnlock()

	return session.status
}

/******************************************************************************
 **函数名称: IsStatus
 **功    能: 判断连接状态
 **输入参数:
 **     status: 连接状态
 **输出参数: NONE
 **返    回: true:是 false:否
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 20:47:11 #
 ******************************************************************************/
func (session *LsndSessionExtra) IsStatus(status int) bool {
	session.RLock()
	defer session.RUnlock()

	if session.status == status {
		return true
	}
	return false
}

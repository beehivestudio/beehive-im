package controllers

import (
	"time"

	"beehive-im/src/golang/lib/comm"
)

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
		proc:  cb,
		param: param,
	}

	tab.Lock()
	tab.list[cmd] = item
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
	item, ok := tab.list[cmd]
	if !ok {
		tab.RUnlock()
		return nil, nil
	}
	cb = item.proc
	param = item.param
	tab.RUnlock()

	return cb, param
}

////////////////////////////////////////////////////////////////////////////////
// 连接扩展数据操作

/* 设置会话SID */
func (session *LsndSessionExtra) SetSid(sid uint64) {
	session.Lock()
	defer session.Unlock()

	session.sid = sid
}

/* 获取会话SID */
func (session *LsndSessionExtra) GetSid() uint64 {
	session.RLock()
	defer session.RUnlock()

	return session.sid
}

/* 获取连接CID */
func (session *LsndSessionExtra) SetCid(cid uint64) {
	session.Lock()
	defer session.Unlock()

	session.cid = cid
}

/* 获取连接CID */
func (session *LsndSessionExtra) GetCid() uint64 {
	session.RLock()
	defer session.RUnlock()

	return session.cid
}

/* 设置连接状态 */
func (session *LsndSessionExtra) SetStatus(status int) {
	session.Lock()
	defer session.Unlock()

	session.status = status
}

/* 获取连接状态 */
func (session *LsndSessionExtra) GetStatus() int {
	session.RLock()
	defer session.RUnlock()

	return session.status
}

/* 判断连接状态 */
func (session *LsndSessionExtra) IsStatus(status int) bool {
	session.RLock()
	defer session.RUnlock()

	if session.status == status {
		return true
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

/* 加入被踢列表 */
func (ctx *LsndCntx) kick_add(cid uint64) {
	item := &LsndKickItem{
		cid: cid,               // 连接ID
		ttl: time.Now().Unix(), // 被踢时间
	}

	ctx.kick_list <- item
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: closed_notify
 **功    能: 发送下线消息
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 拼接OFFLINE请求, 并转发给上游模块.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.17 17:05:00 #
 ******************************************************************************/
func (ctx *LsndCntx) closed_notify(sid uint64) int {
	/* > 通用协议头 */
	head := &comm.MesgHeader{
		Cmd:    comm.CMD_OFFLINE,
		Flag:   comm.MSG_FLAG_USR,
		Length: 0,
		ChkSum: comm.MSG_CHKSUM_VAL,
		Sid:    sid,
		Nid:    ctx.conf.GetNid(),
	}

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE)

	comm.MesgHeadHton(head, p)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_OFFLINE, p.Buff, uint32(len(p.Buff)))

	return 0
}

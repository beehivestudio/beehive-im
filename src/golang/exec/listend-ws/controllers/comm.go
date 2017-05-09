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
func (conn *LsndConnExtra) SetSid(sid uint64) {
	conn.Lock()
	defer conn.Unlock()

	conn.sid = sid
}

/* 获取会话SID */
func (conn *LsndConnExtra) GetSid() uint64 {
	conn.RLock()
	defer conn.RUnlock()

	return conn.sid
}

/* 获取连接CID */
func (conn *LsndConnExtra) SetCid(cid uint64) {
	conn.Lock()
	defer conn.Unlock()

	conn.cid = cid
}

/* 获取连接CID */
func (conn *LsndConnExtra) GetCid() uint64 {
	conn.RLock()
	defer conn.RUnlock()

	return conn.cid
}

/* 更新消息序列号
 * 注意: 参数seq必须比原有消息序列号大 */
func (conn *LsndConnExtra) UpdateSeq(seq uint64) int {
	conn.Lock()
	defer conn.Unlock()

	if conn.seq < seq {
		conn.seq = seq
		return 0
	}
	return -1
}

/* 设置连接状态 */
func (conn *LsndConnExtra) SetStatus(status int) {
	conn.Lock()
	defer conn.Unlock()

	conn.status = status
}

/* 获取连接状态 */
func (conn *LsndConnExtra) GetStatus() int {
	conn.RLock()
	defer conn.RUnlock()

	return conn.status
}

/* 判断连接状态 */
func (conn *LsndConnExtra) IsStatus(status int) bool {
	conn.RLock()
	defer conn.RUnlock()

	if conn.status == status {
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
 **     cid: 连接CID
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 拼接OFFLINE请求, 并转发给上游模块.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.17 17:05:00 #
 ******************************************************************************/
func (ctx *LsndCntx) closed_notify(sid uint64, cid uint64) int {
	/* > 通用协议头 */
	head := &comm.MesgHeader{
		Cmd:    comm.CMD_OFFLINE,
		Length: 0,
		ChkSum: comm.MSG_CHKSUM_VAL,
		Sid:    sid,
		Cid:    cid,
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

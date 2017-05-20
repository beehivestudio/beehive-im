package controllers

import (
	"sync/atomic"
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
	atomic.StoreUint64(&conn.sid, sid)
}

/* 获取会话SID */
func (conn *LsndConnExtra) GetSid() uint64 {
	return atomic.LoadUint64(&conn.sid)
}

/* 获取连接CID */
func (conn *LsndConnExtra) SetCid(cid uint64) {
	atomic.StoreUint64(&conn.cid, cid)
}

/* 获取连接CID */
func (conn *LsndConnExtra) GetCid() uint64 {
	return atomic.LoadUint64(&conn.cid)
}

/* 更新消息序列号
 * 注意: 参数seq必须比原有消息序列号大 */
func (conn *LsndConnExtra) SetSeq(seq uint64) bool {
AGAIN:
	_seq := atomic.LoadUint64(&conn.seq)
	if _seq < seq {
		ok := atomic.CompareAndSwapUint64(&conn.seq, _seq, seq)
		if !ok {
			goto AGAIN
		}
		return true
	}
	return false
}

/* 获取消息序列号 */
func (conn *LsndConnExtra) GetSeq() uint64 {
	return atomic.LoadUint64(&conn.seq)
}

/* 设置连接状态 */
func (conn *LsndConnExtra) SetStatus(status uint32) {
	atomic.StoreUint32(&conn.status, status)
}

/* 获取连接状态 */
func (conn *LsndConnExtra) GetStatus() uint32 {
	return atomic.LoadUint32(&conn.status)
}

/* 判断连接状态 */
func (conn *LsndConnExtra) IsStatus(status uint32) bool {
	st := atomic.LoadUint32(&conn.status)
	if st == status {
		return true
	}
	return false
}

/* 设置创建时间 */
func (conn *LsndConnExtra) SetCtm(ctm int64) {
	atomic.StoreInt64(&conn.ctm, ctm)
}

/* 获取创建时间 */
func (conn *LsndConnExtra) GetCtm() int64 {
	return conn.ctm
}

/* 设置更新时间 */
func (conn *LsndConnExtra) SetUtm(utm int64) {
	atomic.StoreInt64(&conn.utm, utm)
}

/* 获取更新时间 */
func (conn *LsndConnExtra) GetUtm() int64 {
	return atomic.LoadInt64(&conn.utm)
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
 **函数名称: offline_notify
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
func (ctx *LsndCntx) offline_notify(sid uint64, cid uint64) int {
	/* > 通用协议头 */
	head := &comm.MesgHeader{
		Cmd:    comm.CMD_OFFLINE,
		Length: 0,
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

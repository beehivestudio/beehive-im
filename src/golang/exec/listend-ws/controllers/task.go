package controllers

import (
	"time"

	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/chat_tab"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

const (
	LSND_TASK_KICK_DELAY_SEC = 5 /* 踢延迟时间 */
)

/******************************************************************************
 **函数名称: Task
 **功    能: 启动定时任务
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 启动定时任务
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.09 15:21:36 #
 ******************************************************************************/
func (ctx *LsndCntx) Task() {
	go ctx.task_timer_kick()          /* 定时踢连接 */
	go ctx.task_timer_statistic()     /* 定时统计上报 */
	go ctx.task_timer_clean_timeout() /* 定时清理超时连接 */
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task_timer_kick
 **功    能: 指定定时踢除操作
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 获取kick_list中的数据, 并执行kick操作.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.13 00:51:38 #
 ******************************************************************************/
func (ctx *LsndCntx) task_timer_kick() {
	for {
		item, ok := <-ctx.kick_list
		if !ok {
			ctx.log.Error("Kick list was closed!")
			return
		}

		ctm := time.Now().Unix() - LSND_TASK_KICK_DELAY_SEC
		if item.ttl <= ctm {
			ctx.lws.Kick(item.cid)
			ctx.log.Error("Kick connection! cid:%d ctm:%d ttl:%d", item.cid, ctm, item.ttl)
			continue
		}

		diff := time.Duration(item.ttl - ctm)
		time.Sleep(diff * time.Second)

		ctx.lws.Kick(item.cid) /* 执行踢除操作 */

		ctx.log.Error("Kick connection! cid:%d ctm:%d ttl:%d diff:%d", item.cid, ctm, item.ttl, diff)
	}
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task_timer_statistic
 **功    能: 定时统计侦听层状态
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 调用各维度信息统计接口, 并上报结果
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.14 14:53:13 #
 ******************************************************************************/
func (ctx *LsndCntx) task_timer_statistic() {
	for {
		ctx.gather_base_info() /* 采集侦听层信息, 并上报 */
		ctx.gather_room_stat() /* 采集聊天室信息, 并上报 */
		time.Sleep(5 * time.Second)
	}
}

/******************************************************************************
 **函数名称: gather_base_info
 **功    能: 上报侦听层信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 拼接LSND-INFO报文, 并发送给上游模块.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.14 14:53:13 #
 ******************************************************************************/
func (ctx *LsndCntx) gather_base_info() {
	req := &mesg.MesgLsndInfo{
		Type:        proto.Uint32(comm.LSND_TYPE_WS),       // 网络类型(0:UNKNOWN 1:TCP 2:WS)
		Nid:         proto.Uint32(ctx.conf.GetNid()),       // 结点ID
		Nation:      proto.String(ctx.conf.GetNation()),    // 所属国家
		Opid:        proto.Uint32(ctx.conf.GetOpid()),      // 运营商ID
		Ip:          proto.String(ctx.conf.GetIp()),        // IP地址
		Port:        proto.Uint32(ctx.conf.GetPort()),      // 端口号
		Connections: proto.Uint32(ctx.chat.SessionCount()), // 会话总数
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(req)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return
	}

	length := len(body)
	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head := &comm.MesgHeader{
		Cmd:    comm.CMD_LSND_INFO,  // 消息类型
		Nid:    ctx.conf.GetNid(),   // 结点ID
		Length: uint32(length),      // 消息长度
		ChkSum: comm.MSG_CHKSUM_VAL, // 校验值
	}

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_LSND_INFO, p.Buff, uint32(len(p.Buff)))

	ctx.log.Debug("Send listen report succ!")
}

type ChatTravRoomListProcCb func(item *chat_tab.ChatRoomItem, param interface{}) int

/******************************************************************************
 **函数名称: gather_room_stat
 **功    能: 上报聊天室信息
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 遍历聊天室列表, 取有效信息并上报.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.12 18:08:06 #
 ******************************************************************************/
func (ctx *LsndCntx) gather_room_stat() {
	ctx.chat.TravRoomList(LsndUploadRoomLsnStatCb, ctx)
}

/******************************************************************************
 **函数名称: LsndUploadRoomLsnStatCb
 **功    能: 上报各聊天室在此侦听层的统计
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 拼接ROOM-LSN-STAT报文, 并发送给上游模块.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.12 18:08:06 #
 ******************************************************************************/
func LsndUploadRoomLsnStatCb(item *chat_tab.ChatRoomItem, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	req := &mesg.MesgRoomLsnStat{
		Rid: proto.Uint64(item.GetRid()),         // 聊天室ID
		Nid: proto.Uint32(ctx.conf.GetNid()),     // 聊天室ID
		Num: proto.Uint32(uint32(item.GetNum())), // 聊天室人数
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(req)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head := &comm.MesgHeader{
		Cmd:    comm.CMD_ROOM_LSN_STAT, // 消息类型
		Nid:    ctx.conf.GetNid(),      // 结点ID
		Length: uint32(length),         // 消息长度
		ChkSum: comm.MSG_CHKSUM_VAL,    // 校验值
	}

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_ROOM_LSN_STAT, p.Buff, uint32(len(p.Buff)))

	ctx.log.Debug("Send room user number succ! rid:%d num:%d", item.GetRid(), item.GetNum())

	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: task_timer_clean_timeout
 **功    能: 清理超时连接
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.10 19:58:58 #
 ******************************************************************************/
func (ctx *LsndCntx) task_timer_clean_timeout() {
	for {
		ctx.chat.TravSession(LsndCleanTimeoutHandler, ctx)
		time.Sleep(30 * time.Second)
	}
}

/******************************************************************************
 **函数名称: LsndCleanTimeoutHandler
 **功    能: 判断会话是否超时, 并踢除超时连接.
 **输入参数:
 **     sid: 会话SID
 **     cid: 连接CID
 **     param: 扩展参数
 **输出参数: NONE
 **返    回: 0:正常 !0:异常
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.10 19:58:58 #
 ******************************************************************************/
func LsndCleanTimeoutHandler(sid uint64, cid uint64, extra interface{}, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	conn, ok := extra.(*LsndConnExtra)
	if !ok {
		return -1
	}

	ctm := time.Now().Unix()

	if CONN_STATUS_LOGIN != conn.GetStatus() {
		if (ctm > conn.GetCtm()) && (ctm-conn.GetCtm() > 5) {
			ctx.lws.Kick(conn.GetCid())
		}
	}

	if ctm > conn.GetCtm() && ctm-conn.GetCtm() > 300 {
		ctx.lws.Kick(conn.GetCid())
	}

	return 0
}

package controllers

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: send_data
 **功    能: 下发消息
 **输入参数:
 **     cmd: 命令类型
 **     to: 接收方(会话ID/聊天室ID/群组ID)
 **     nid: 结点ID
 **     seq: 序列号
 **     data: 下发数据
 **     length: 数据长度
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **应答协议:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.22 09:24:00 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) send_data(cmd uint32, to uint64, nid uint32, seq uint64, data []byte, length uint32) int {
	var head comm.MesgHeader

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+int(length))

	head.Cmd = cmd
	head.Sid = to
	head.Nid = nid
	head.Length = length
	head.ChkSum = comm.MSG_CHKSUM_VAL
	head.Serial = seq

	comm.MesgHeadHton(&head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], data)

	/* > 发送协议包 */
	return ctx.frwder.AsyncSend(cmd, p.Buff, uint32(len(p.Buff)))
}

/* 会话属性 */
type SidAttr struct {
	sid uint64 /* 会话SID */
	uid uint64 /* 用户UID */
	nid uint32 /* 结点ID */
}

/******************************************************************************
 **函数名称: get_sid_attr
 **功    能: 获取会话属性
 **输入参数:
 **     sid: 会话ID
 **输出参数: NONE
 **返    回: 会话对应的属性
 **实现描述:
 **应答协议:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.12.25 09:25:00 #
 ******************************************************************************/
func (ctx *MsgSvrCntx) get_sid_attr(sid uint64) *SidAttr {
	rds := ctx.redis.Get()
	defer rds.Close()

	key := fmt.Sprintf(comm.IM_KEY_SID_ATTR, sid)
	vals, _ := redis.Strings(rds.Do("HGET", key, "UID", "NID"))

	uid, _ := strconv.ParseInt(vals[0], 10, 64)
	nid, _ := strconv.ParseInt(vals[1], 10, 32)

	attr := &SidAttr{}

	attr.sid = sid
	attr.uid = uint64(uid)
	attr.nid = uint32(nid)

	return attr
}

package controllers

import (
	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: send_data
 **功    能: 下发消息
 **输入参数:
 **     cmd: 命令类型
 **     sid: 会话SID
 **     cid: 连接CID
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
func (ctx *MsgSvrCntx) send_data(cmd uint32, sid uint64, cid uint64, nid uint32, seq uint64, data []byte, length uint32) int {
	var head comm.MesgHeader

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+int(length))

	head.Cmd = cmd
	head.Sid = sid
	head.Cid = cid
	head.Nid = nid
	head.Length = length
	head.Seq = seq

	comm.MesgHeadHton(&head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], data)

	/* > 发送协议包 */
	return ctx.frwder.AsyncSend(cmd, p.Buff, uint32(len(p.Buff)))
}

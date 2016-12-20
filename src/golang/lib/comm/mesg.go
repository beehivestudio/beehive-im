package comm

import (
	"encoding/binary"
)

type HttpResp struct {
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

const (
	CMD_UNKNOWN = 0 /* 未知消息 */

	/* 请求/应答类消息 */
	CMD_ONLINE_REQ       = 0x0101 /* 上线请求 */
	CMD_ONLINE_ACK       = 0x0102 /* 上线请求应答 */
	CMD_OFFLINE_REQ      = 0x0103 /* 下线请求 */
	CMD_OFFLINE_ACK      = 0x0104 /* 下线请求应答 */
	CMD_JOIN_REQ         = 0x0105 /* 加入聊天室 */
	CMD_JOIN_ACK         = 0x0106 /* 加入聊天室应答 */
	CMD_UNJOIN_REQ       = 0x0107 /* 退出聊天室 */
	CMD_UNJOIN_ACK       = 0x0108 /* 退出聊天室应答 */
	CMD_PING             = 0x0109 /* 客户端心跳 */
	CMD_PONG             = 0x010A /* 客户端心跳应答 */
	CMD_SUB_REQ          = 0x010B /* 订阅请求 */
	CMD_SUB_ACK          = 0x010C /* 订阅应答 */
	CMD_UNSUB_REQ        = 0x010D /* 取消订阅 */
	CMD_UNSUB_ACK        = 0x010E /* 取消订阅应答 */
	CMD_GROUP_MSG        = 0x0110 /* 群聊消息 */
	CMD_GROUP_MSG_ACK    = 0x0111 /* 群聊消息应答 */
	CMD_PRVT_MSG         = 0x0112 /* 私聊消息 */
	CMD_PRVT_MSG_ACK     = 0x0113 /* 私聊消息应答 */
	CMD_BC_MSG           = 0x0114 /* 广播消息 */
	CMD_BC_MSG_ACK       = 0x0115 /* 广播消息应答 */
	CMD_P2P_MSG          = 0x0116 /* 点到点消息(暂时不需要) */
	CMD_P2P_MSG_ACK      = 0x0117 /* 点到点消息应答 */
	CMD_ROOM_MSG         = 0x0118 /* 聊天室消息 */
	CMD_ROOM_MSG_ACK     = 0x0119 /* 聊天室消息应答 */
	CMD_ROOM_BC_MSG      = 0x011A /* 聊天室广播消息 */
	CMD_ROOM_BC_MSG_ACK  = 0x011B /* 聊天室广播消息应答 */
	CMD_EXCEPT_MSG       = 0x011C /* 通用异常消息 */
	CMD_EXCEPT_MSG_ACK   = 0x011D /* 通用异常消息应答 */
	CMD_ROOM_USR_NUM     = 0x011E /* 聊天室人数 */
	CMD_ROOM_USR_NUM_ACK = 0x0120 /* 聊天室人数应答 */
	CMD_SYNC_MSG         = 0x0121 /* 同步消息 */
	CMD_SYNC_MSG_ACK     = 0x0122 /* 同步消息应答 */
	/* 通知类消息 */
	CMD_ONLINE_NTC        = 0x0301 /* 上线通知 */
	CMD_OFFLINE_NTC       = 0x0302 /* 下线通知 */
	CMD_JOIN_NTC          = 0x0303 /* 加入聊天室通知 */
	CMD_QUIT_NTC          = 0x0304 /* 退出聊天室通知 */
	CMD_BAN_ADD_NTC       = 0x0305 /* 禁言通知 */
	CMD_BAN_DEL_NTC       = 0x0306 /* 移除禁言通知 */
	CMD_BLACKLIST_ADD_NTC = 0x0307 /* 加入黑名单通知 */
	CMD_BLACKLIST_DEL_NTC = 0x0308 /* 移除黑名单通知 */
	/* 系统内部消息 */
	CMD_LSN_RPT       = 0x0401 /* 帧听层上报 */
	CMD_LSN_RPT_ACK   = 0x0402 /* 帧听层上报应答 */
	CMD_FRWD_LIST     = 0x0403 /* 转发层上报 */
	CMD_FRWD_LIST_ACK = 0x0404 /* 转发层上报应答 */
	CMD_KICK_REQ      = 0x0405 /* 踢人操作 */
	CMD_KICK_ACK      = 0x0406 /* 踢人应答 */
)

const (
	MSG_FLAG_SYS   = 0          /* 0: 系统数据类型 */
	MSG_FLAG_USR   = 1          /* 1: 自定义数据类型 */
	MSG_CHKSUM_VAL = 0x1ED23CB4 /* 校验码 */
)

var (
	MESG_HEAD_SIZE = binary.Size(MesgHeader{})
)

/* 通用协议头 */
type MesgHeader struct {
	Cmd    uint32 /* 消息类型 */
	Flag   uint32 /* 标识量(0:系统数据类型 1:自定义数据类型) */
	Length uint32 /* 报体长度 */
	ChkSum uint32 /* 校验值 */
	Sid    uint64 /* 会话ID */
	Nid    uint32 /* 结点ID */
	Serial uint64 /* 流水号(注: 全局唯一流水号) */
}

func (head *MesgHeader) GetCmd() uint32 {
	return head.Cmd
}

func (head *MesgHeader) GetFlag() uint32 {
	return head.Flag
}

func (head *MesgHeader) GetLength() uint32 {
	return head.Length
}

func (head *MesgHeader) GetChkSum() uint32 {
	return head.ChkSum
}

func (head *MesgHeader) GetSid() uint64 {
	return head.Sid
}

func (head *MesgHeader) GetNid() uint32 {
	return head.Nid
}

func (head *MesgHeader) GetSerial() uint64 {
	return head.Serial
}

type MesgPacket struct {
	Buff []byte /* 接收数据 */
}

/* "主机->网络"字节序 */
func MesgHeadHton(header *MesgHeader, p *MesgPacket) {
	binary.BigEndian.PutUint32(p.Buff[0:4], header.Cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.Buff[4:8], header.Flag)     /* FLAG */
	binary.BigEndian.PutUint32(p.Buff[8:12], header.Length)  /* LENGTH */
	binary.BigEndian.PutUint32(p.Buff[12:16], header.ChkSum) /* CHKSUM */
	binary.BigEndian.PutUint64(p.Buff[16:24], header.Sid)    /* SID */
	binary.BigEndian.PutUint32(p.Buff[24:28], header.Nid)    /* NID */
	binary.BigEndian.PutUint64(p.Buff[28:36], header.Serial) /* SERIAL */
}

/* "网络->主机"字节序 */
func MesgHeadNtoh(data []byte) *MesgHeader {
	head := &MesgHeader{}

	head.Cmd = binary.BigEndian.Uint32(data[0:4])
	head.Flag = binary.BigEndian.Uint32(data[4:8])
	head.Length = binary.BigEndian.Uint32(data[8:12])
	head.ChkSum = binary.BigEndian.Uint32(data[12:16])
	head.Sid = binary.BigEndian.Uint64(data[16:24])
	head.Nid = binary.BigEndian.Uint32(data[24:28])
	head.Serial = binary.BigEndian.Uint64(data[28:36])

	return head
}

/* 校验头部数据的合法性 */
func MesgHeadIsValid(header *MesgHeader) bool {
	if MSG_CHKSUM_VAL != header.ChkSum ||
		0 == header.Sid || 0 == header.Nid {
		return false
	}
	return true
}

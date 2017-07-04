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

	/* 通用消息 */
	CMD_ONLINE          = 0x0101 /* 上线请求(服务端) */
	CMD_ONLINE_ACK      = 0x0102 /* 上线请求应答(客户端) */
	CMD_OFFLINE         = 0x0103 /* 下线请求(服务端) */
	CMD_OFFLINE_ACK     = 0x0104 /* 下线请求应答(客户端) */
	CMD_PING            = 0x0105 /* 客户端心跳(服务端) */
	CMD_PONG            = 0x0106 /* 客户端心跳应答(客户端) */
	CMD_SUB             = 0x0107 /* 订阅请求(服务端) */
	CMD_SUB_ACK         = 0x0108 /* 订阅应答(客户端) */
	CMD_UNSUB           = 0x0109 /* 取消订阅(服务端) */
	CMD_UNSUB_ACK       = 0x010A /* 取消订阅应答(客户端) */
	CMD_ERROR           = 0x010B /* 通用错误消息 */
	CMD_ERROR_ACK       = 0x010C /* 通用错误消息应答 */
	CMD_SYNC            = 0x010D /* 同步消息 */
	CMD_SYNC_ACK        = 0x010E /* 同步消息应答(客户端) */
	CMD_KICK            = 0x0110 /* 踢人请求 */
	CMD_KICK_ACK        = 0x0111 /* 踢人应答 */
	CMD_ONLINE_NTC      = 0x0151 /* 上线通知 */
	CMD_ONLINE_NTC_ACK  = 0x0152 /* 上线通知应答 */
	CMD_OFFLINE_NTC     = 0x0153 /* 下线通知 */
	CMD_OFFLINE_NTC_ACK = 0x0154 /* 下线通知应答 */

	/* 私聊消息 */
	CMD_CHAT              = 0x0201 /* 私聊消息 */
	CMD_CHAT_ACK          = 0x0202 /* 私聊消息应答 */
	CMD_FRIEND_ADD        = 0x0203 /* 添加好友 */
	CMD_FRIEND_ADD_ACK    = 0x0204 /* 添加好友应答 */
	CMD_FRIEND_DEL        = 0x0205 /* 移除好友 */
	CMD_FRIEND_DEL_ACK    = 0x0206 /* 移除好友应答 */
	CMD_BLACKLIST_ADD     = 0x0207 /* 加入黑名单 */
	CMD_BLACKLIST_ADD_ACK = 0x0208 /* 加入黑名单应答 */
	CMD_BLACKLIST_DEL     = 0x0209 /* 移除黑名单 */
	CMD_BLACKLIST_DEL_ACK = 0x020A /* 移除黑名单应答 */
	CMD_GAG_ADD           = 0x020B /* 设置禁言 */
	CMD_GAG_ADD_ACK       = 0x020C /* 设置禁言应答 */
	CMD_GAG_DEL           = 0x020D /* 移除禁言 */
	CMD_GAG_DEL_ACK       = 0x020E /* 移除禁言应答 */
	CMD_MARK_ADD          = 0x0210 /* 设置备注 */
	CMD_MARK_ADD_ACK      = 0x0211 /* 设置备注应答 */
	CMD_MARK_DEL          = 0x0212 /* 移除备注 */
	CMD_MARK_DEL_ACK      = 0x0213 /* 移除备注应答 */

	/* 群聊消息 */
	CMD_GROUP_CREAT           = 0x0301 /* 创建群组 */
	CMD_GROUP_CREAT_ACK       = 0x0302 /* 创建群组应答 */
	CMD_GROUP_DISMISS         = 0x0303 /* 解散群组 */
	CMD_GROUP_DISMISS_ACK     = 0x0304 /* 解散群组应答 */
	CMD_GROUP_JOIN            = 0x0305 /* 申请入群 */
	CMD_GROUP_JOIN_ACK        = 0x0306 /* 申请入群应答 */
	CMD_GROUP_QUIT            = 0x0307 /* 退群 */
	CMD_GROUP_QUIT_ACK        = 0x0308 /* 退群应答 */
	CMD_GROUP_INVITE          = 0x0309 /* 邀请入群 */
	CMD_GROUP_INVITE_ACK      = 0x030A /* 邀请入群应答 */
	CMD_GROUP_CHAT            = 0x030B /* 群聊消息 */
	CMD_GROUP_CHAT_ACK        = 0x030C /* 群聊消息应答 */
	CMD_GROUP_KICK            = 0x030D /* 群组踢人 */
	CMD_GROUP_KICK_ACK        = 0x030E /* 群组踢人应答 */
	CMD_GROUP_GAG_ADD         = 0x0310 /* 群组禁言 */
	CMD_GROUP_GAG_ADD_ACK     = 0x0311 /* 群组禁言应答 */
	CMD_GROUP_GAG_DEL         = 0x0312 /* 解除群组禁言 */
	CMD_GROUP_GAG_DEL_ACK     = 0x0313 /* 解除群组禁言应答 */
	CMD_GROUP_BL_ADD          = 0x0314 /* 加入群组黑名单 */
	CMD_GROUP_BL_ADD_ACK      = 0x0315 /* 加入群组黑名单应答 */
	CMD_GROUP_BL_DEL          = 0x0316 /* 解除群组黑名单 */
	CMD_GROUP_BL_DEL_ACK      = 0x0317 /* 解除群组黑名单应答 */
	CMD_GROUP_MGR_ADD         = 0x0318 /* 添加群组管理员 */
	CMD_GROUP_MGR_ADD_ACK     = 0x0319 /* 添加群组管理员应答 */
	CMD_GROUP_MGR_DEL         = 0x031A /* 解除群组管理员 */
	CMD_GROUP_MGR_DEL_ACK     = 0x031B /* 解除群组管理员应答 */
	CMD_GROUP_USR_LIST        = 0x031C /* 群组成员列表 */
	CMD_GROUP_USR_LIST_ACK    = 0x031D /* 群组成员列表应答 */
	CMD_GROUP_JOIN_NTC        = 0x0350 /* 入群通知 */
	CMD_GROUP_JOIN_NTC_ACK    = 0x0351 /* 入群通知应答 */
	CMD_GROUP_QUIT_NTC        = 0x0352 /* 退群通知 */
	CMD_GROUP_QUIT_NTC_ACK    = 0x0353 /* 退群通知应答 */
	CMD_GROUP_KICK_NTC        = 0x0354 /* 踢人通知 */
	CMD_GROUP_KICK_NTC_ACK    = 0x0355 /* 踢人通知应答 */
	CMD_GROUP_GAG_ADD_NTC     = 0x0356 /* 禁言通知 */
	CMD_GROUP_GAG_ADD_NTC_ACK = 0x0357 /* 禁言通知应答 */
	CMD_GROUP_GAG_DEL_NTC     = 0x0358 /* 解除禁言通知 */
	CMD_GROUP_GAG_DEL_NTC_ACK = 0x0359 /* 解除禁言通知应答 */
	CMD_GROUP_BL_ADD_NTC      = 0x0360 /* 添加群组黑名单通知 */
	CMD_GROUP_BL_ADD_NTC_ACK  = 0x0361 /* 添加群组黑名单通知应答 */
	CMD_GROUP_BL_DEL_NTC      = 0x0362 /* 解除群组黑名单通知 */
	CMD_GROUP_BL_DEL_NTC_ACK  = 0x0363 /* 解除群组黑名单通知应答 */
	CMD_GROUP_MGR_ADD_NTC     = 0x0364 /* 添加群组管理员通知 */
	CMD_GROUP_MGR_ADD_NTC_ACK = 0x0365 /* 添加群组管理员通知应答 */
	CMD_GROUP_MGR_DEL_NTC     = 0x0366 /* 解除群组管理员通知 */
	CMD_GROUP_MGR_DEL_NTC_ACK = 0x0367 /* 解除群组管理员通知应答 */

	/* 聊天室消息 */
	CMD_ROOM_CREAT        = 0x0401 /* 创建聊天室 */
	CMD_ROOM_CREAT_ACK    = 0x0402 /* 创建聊天室应答 */
	CMD_ROOM_DISMISS      = 0x0403 /* 解散聊天室 */
	CMD_ROOM_DISMISS_ACK  = 0x0404 /* 解散聊天室应答 */
	CMD_ROOM_JOIN         = 0x0405 /* 加入聊天室 */
	CMD_ROOM_JOIN_ACK     = 0x0406 /* 加入聊天室应答 */
	CMD_ROOM_QUIT         = 0x0407 /* 退出聊天室 */
	CMD_ROOM_QUIT_ACK     = 0x0408 /* 退出聊天室应答 */
	CMD_ROOM_KICK         = 0x0409 /* 踢出聊天室 */
	CMD_ROOM_KICK_ACK     = 0x040A /* 踢出聊天室应答 */
	CMD_ROOM_CHAT         = 0x040B /* 聊天室消息 */
	CMD_ROOM_CHAT_ACK     = 0x040C /* 聊天室消息应答 */
	CMD_ROOM_BC           = 0x040D /* 聊天室广播消息 */
	CMD_ROOM_BC_ACK       = 0x040E /* 聊天室广播消息应答 */
	CMD_ROOM_USR_NUM      = 0x0410 /* 聊天室人数 */
	CMD_ROOM_USR_NUM_ACK  = 0x0411 /* 聊天室人数应答 */
	CMD_ROOM_LSN_STAT     = 0x0412 /* 聊天室各侦听层统计 */
	CMD_ROOM_LSN_STAT_ACK = 0x0413 /* 聊天室各侦听层统计应答 */
	CMD_ROOM_JOIN_NTC     = 0x0450 /* 加入聊天室通知 */
	CMD_ROOM_JOIN_NTC_ACK = 0x0451 /* 加入聊天室通知应答 */
	CMD_ROOM_QUIT_NTC     = 0x0452 /* 退出聊天室通知 */
	CMD_ROOM_QUIT_NTC_ACK = 0x0453 /* 退出聊天室通知应答 */
	CMD_ROOM_KICK_NTC     = 0x0454 /* 踢出聊天室通知 */
	CMD_ROOM_KICK_NTC_ACK = 0x0455 /* 踢出聊天室通知应答 */

	/* 推送消息 */
	CMD_BC      = 0x0501 /* 广播消息 */
	CMD_BC_ACK  = 0x0502 /* 广播消息应答 */
	CMD_P2P     = 0x0503 /* 点到点消息(暂时不需要) */
	CMD_P2P_ACK = 0x0504 /* 点到点消息应答(客户端&服务端) */

	/* 系统内部消息 */
	CMD_LSND_INFO     = 0x0601 /* 帧听层信息上报 */
	CMD_LSND_INFO_ACK = 0x0602 /* 帧听层信息上报应答 */
	CMD_FRWD_INFO     = 0x0603 /* 转发层信息上报 */
	CMD_FRWD_INFO_ACK = 0x0604 /* 转发层信息上报应答 */
)

const (
	MSG_CHKSUM_VAL = 0x1ED23CB4 /* 校验码 */
)

var (
	MESG_HEAD_SIZE = binary.Size(MesgHeader{})
)

/* 通用协议头 */
type MesgHeader struct {
	Cmd    uint32 /* 消息类型 */
	Length uint32 /* 报体长度 */
	Sid    uint64 /* 源: 会话ID */
	Cid    uint64 /* 源: 连接ID */
	Nid    uint32 /* 源: 结点ID */
	Seq    uint64 /* 源: 流水号(注: SID全局唯一流水号) */
	Dsid   uint64 /* 目标: 会话ID */
	Dseq   uint64 /* 目标: 流水号(注: DSID全局唯一流水号) */
}

func (head *MesgHeader) SetCmd(cmd uint32) {
	head.Cmd = cmd
}

func (head *MesgHeader) GetCmd() uint32 {
	return head.Cmd
}

func (head *MesgHeader) GetLength() uint32 {
	return head.Length
}

func (head *MesgHeader) SetSid(sid uint64) {
	head.Sid = sid
}

func (head *MesgHeader) GetSid() uint64 {
	return head.Sid
}

func (head *MesgHeader) SetCid(cid uint64) {
	head.Cid = cid
}

func (head *MesgHeader) GetCid() uint64 {
	return head.Cid
}

func (head *MesgHeader) SetNid(nid uint32) {
	head.Nid = nid
}

func (head *MesgHeader) GetNid() uint32 {
	return head.Nid
}

func (head *MesgHeader) GetSeq() uint64 {
	return head.Seq
}

func (head *MesgHeader) GetDsid() uint64 {
	return head.Dsid
}

func (head *MesgHeader) GetDseq() uint64 {
	return head.Dseq
}

type MesgPacket struct {
	Buff []byte /* 接收数据 */
}

/* "主机->网络"字节序 */
func MesgHeadHton(head *MesgHeader, p *MesgPacket) {
	binary.BigEndian.PutUint32(p.Buff[0:4], head.Cmd)    /* CMD */
	binary.BigEndian.PutUint32(p.Buff[4:8], head.Length) /* LENGTH */
	binary.BigEndian.PutUint64(p.Buff[8:16], head.Sid)   /* SID */
	binary.BigEndian.PutUint64(p.Buff[16:24], head.Cid)  /* CID */
	binary.BigEndian.PutUint32(p.Buff[24:28], head.Nid)  /* NID */
	binary.BigEndian.PutUint64(p.Buff[28:36], head.Seq)  /* SEQ */
	binary.BigEndian.PutUint64(p.Buff[36:44], head.Dsid) /* DSID */
	binary.BigEndian.PutUint64(p.Buff[44:52], head.Dseq) /* DSEQ */
}

/* "网络->主机"字节序 */
func MesgHeadNtoh(data []byte) *MesgHeader {
	head := &MesgHeader{}

	head.Cmd = binary.BigEndian.Uint32(data[0:4])
	head.Length = binary.BigEndian.Uint32(data[4:8])
	head.Sid = binary.BigEndian.Uint64(data[8:16])
	head.Cid = binary.BigEndian.Uint64(data[16:24])
	head.Nid = binary.BigEndian.Uint32(data[24:28])
	head.Seq = binary.BigEndian.Uint64(data[28:36])
	head.Dsid = binary.BigEndian.Uint64(data[36:44])
	head.Dseq = binary.BigEndian.Uint64(data[44:52])

	return head
}

/* 校验头部数据的合法性 */
func (head *MesgHeader) IsValid(flag uint32) bool {
	if 0 == head.Nid {
		return false
	} else if 0 != flag && 0 == head.Sid {
		return false
	}
	return true
}

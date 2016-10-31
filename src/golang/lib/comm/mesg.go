package comm

import (
	"encoding/binary"
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
	buff []byte /* 接收数据 */
}

/* "主机->网络"字节序 */
func MesgHeadHton(header *MesgHeader, p *MesgPacket) {
	binary.BigEndian.PutUint32(p.buff[0:4], header.Cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.buff[4:8], header.Flag)     /* FLAG */
	binary.BigEndian.PutUint32(p.buff[8:12], header.Length)  /* LENGTH */
	binary.BigEndian.PutUint32(p.buff[12:16], header.ChkSum) /* CHKSUM */
	binary.BigEndian.PutUint64(p.buff[16:24], header.Sid)    /* SID */
	binary.BigEndian.PutUint32(p.buff[24:28], header.Nid)    /* NID */
	binary.BigEndian.PutUint64(p.buff[28:36], header.Serial) /* SERIAL */
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

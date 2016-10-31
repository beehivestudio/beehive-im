package comm

const (
	MSG_FLAG_SYS   = 0          /* 0: 系统数据类型 */
	MSG_FLAG_USR   = 1          /* 1: 自定义数据类型 */
	MSG_CHKSUM_VAL = 0x1ED23CB4 /* 校验码 */
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

type MesgPacket struct {
	buff []byte /* 接收数据 */
}

/* "主机->网络"字节序 */
func mesg_head_hton(header *MesgHeader, p *MesgPacket) {
	binary.BigEndian.PutUint32(p.buff[0:4], header.Cmd)      /* CMD */
	binary.BigEndian.PutUint32(p.buff[4:8], header.Flag)     /* FLAG */
	binary.BigEndian.PutUint32(p.buff[8:12], header.Length)  /* LENGTH */
	binary.BigEndian.PutUint32(p.buff[12:16], header.ChkSum) /* CHKSUM */
	binary.BigEndian.PutUint64(p.buff[16:24], header.Sid)    /* SID */
	binary.BigEndian.PutUint32(p.buff[24:28], header.Nid)    /* NID */
	binary.BigEndian.PutUint64(p.buff[28:36], header.Serial) /* SERIAL */
}

/* "网络->主机"字节序 */
func mesg_head_ntoh(p *MesgPacket) *MesgHeader {
	head := &MesgHeader{}

	head.Cmd = binary.BigEndian.Uint32(p.buff[0:4])
	head.Flag = binary.BigEndian.Uint32(p.buff[4:8])
	head.Length = binary.BigEndian.Uint32(p.buff[8:12])
	head.ChkSum = binary.BigEndian.Uint32(p.buff[12:16])
	head.Sid = binary.BigEndian.Uint64(p.buff[16:24])
	head.Nid = binary.BigEndian.Uint32(p.buff[24:28])
	head.Serial = binary.BigEndian.Uint64(p.buff[28:36])

	return head
}

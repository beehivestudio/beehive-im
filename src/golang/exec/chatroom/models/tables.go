package models

/* 聊天室数据 */
type RoomChatTabRow struct {
	Rid  uint64 "rid"  // 聊天室ID
	Uid  uint64 "uid"  // 用户UID
	Ctm  int64  "ctm"  // 发送时间
	Data []byte "data" // 原始数据包
}

/* 聊天室黑名单 */
type RoomBlacklistTabRow struct {
	Rid    uint64 "rid"    // 聊天室ID
	Uid    uint64 "uid"    // 用户UID
	Role   uint64 "role"   // 角色
	Status uint8  "status" // 状态(0:正常 1:被踢)
	Ctm    int64  "ctm"    // 设置时间
}

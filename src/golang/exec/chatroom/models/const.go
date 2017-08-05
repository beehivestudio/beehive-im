package models

/* 聊天室角色 */
const (
	ROOM_ROLE_OWNER   = 1 // 聊天室-所有者
	ROOM_ROLE_MANAGER = 2 // 聊天室-管理员
)

/* 聊天室状态 */
const (
	ROOM_STAT_OPEN  = 1 // 聊天室-开启
	ROOM_STAT_CLOSE = 0 // 聊天室-关闭
)

/* 聊天室用户状态 */
const (
	ROOM_USER_STAT_NORMAL = 0 // 正常
	ROOM_USER_STAT_KICK   = 1 // 被踢
)

/* 聊天室数据表 */
const (
	ROOM_TAB_MESG      = "RoomMesg"      // 聊天消息表
	ROOM_TAB_BLACKLIST = "RoomBlacklist" // 黑名单表
)

const (
	ROOM_TTL_SEC = 30 // 聊天室TTL(单位:秒)
)

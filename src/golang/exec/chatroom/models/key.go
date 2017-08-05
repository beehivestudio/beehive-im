package models

/* 聊天室相关KEY定义 */
const (
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	ROOM_KEY_SID_ZSET               = "room:sid:zset"                 //*| ZSET | 会话SID集合 | 成员:SID 分值:TTL |
	ROOM_KEY_UID_ZSET               = "room:uid:zset"                 //| ZSET | 用户UID集合 | 成员:UID 分值:TTL |
	ROOM_KEY_SID_INCR               = "room:sid:incr"                 //*| STRING | 会话SID增量器 | 只增不减 注意:sid不能为0 |
	ROOM_KEY_SID_ATTR               = "room:sid:%d:attr"              //*| HASH | 会话SID属性 | 包含UID/NID |
	ROOM_KEY_UID_TO_SID_SET         = "room:uid:%d:to:sid:set"        //| SET | 用户UID对应的会话SID集合 | SID集合 |
	ROOM_KEY_RID_INCR               = "room:rid:incr"                 //*| STRING | 聊天室RID记录器|
	ROOM_KEY_RID_ZSET               = "room:rid:zset"                 //*| ZSET | 聊天室RID集合 | 成员:RID 分值:TTL |
	ROOM_KEY_RID_ATTR               = "room:rid:%d:attr"              //*| HASH | 聊天室属性信息| STATUS:(0:打开 1:关闭) |
	ROOM_KEY_ROOM_GROUP_CAP_ZSET    = "room:room:group:cap:zset"      //*| ZSET | 聊天室分组容量 | 成员:RID 分值:分组容量 |
	ROOM_KEY_SID_TO_RID_ZSET        = "room:sid:%d:to:rid:zset"       //*| ZSET | 会话SID对应的RID集合 | 成员:RID 分值:GID |
	ROOM_KEY_ROOM_GROUP_USR_NUM     = "room:room:group:usr:num"       //| ZSET | 聊天室分组人数配置 | 成员:RID 分值:USERNUM |
	ROOM_KEY_RID_GID_TO_NUM_ZSET    = "room:rid:%d:to:gid:num:zset"   //*| ZSET | 某聊天室各组人数 | 成员:GID 分值:USERNUM |
	ROOM_KEY_RID_TO_NID_ZSET        = "room:rid:%d:to:nid:zset"       //*| ZSET | 某聊天室->帧听层 | 成员:NID 分值:TTL |
	ROOM_KEY_RID_NID_TO_NUM_ZSET    = "room:rid:%d:nid:to:num:zset"   //*| ZSET | 某聊天室各帧听层人数 | 成员:NID 分值:USERNUM | 由帧听层上报数据获取
	ROOM_KEY_RID_SUB_USR_NUM_ZSET   = "room:rid:sub:usr:num:zset"     //| ZSET | 聊天室人数订阅集合 | 暂无 |
	ROOM_KEY_RID_TO_UID_SID_ZSET    = "room:rid:%d:to:uid:sid:zset"   //| ZSET | 聊天室用户列表 | 成员:"${UID}:${SID}" 分值:TTL |
	ROOM_KEY_RID_TO_SID_ZSET        = "room:rid:%d:to:sid:zset"       //| ZSET | 聊天室SID列表 | 成员:SID 分值:TTL |
	ROOM_KEY_ROOM_MESG_QUEUE        = "room:rid:%d:mesg:queue"        //| LIST | 聊天室消息队列 |
	ROOM_KEY_ROOM_MSGID_INCR        = "room:rid:%d:msgid:incr"        //| STRING | 聊天室消息序列递增记录 |
	ROOM_KEY_ROOM_USR_GAG_SET       = "room:rid:%d:usr:gag:set"       //*| SET | 聊天室用户禁言名单 | 成员:UID |
	ROOM_KEY_ROOM_USR_BLACKLIST_SET = "room:rid:%d:usr:blacklist:set" //*| SET | 聊天室用户黑名单 | 成员:UID |
	ROOM_KEY_ROOM_ROLE_TAB          = "room:rid:%d:role:tab"          //*| HASH | 聊天室管理人员名单 | 成员:UID 值:角色(1:OWNER 2:管理员) |
	ROOM_KEY_ROOM_INFO_TAB          = "room:rid:%d:info:tab"          //*| HASH | 聊天室基本信息管理 |
	ROOM_KEY_ROOM_BC_ZSET           = "room:rid:%d:broadcast:zset"    //| ZSET | 聊天室广播集合 | 成员:消息ID 分值:超时时间 |
	ROOM_KEY_ROOM_BC_HASH           = "room:rid:%d:broadcast:hash"    //| HASH | 聊天室广播内容 | 域:消息ID 值:广播内容 |
)

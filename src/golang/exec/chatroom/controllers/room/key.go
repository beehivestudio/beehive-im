package room

const (
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	//聊天室
	CR_KEY_SID_ZSET               = "cr:sid:zset"                 //*| ZSET | 会话SID集合 | 成员:SID 分值:TTL |
	CR_KEY_UID_ZSET               = "cr:uid:zset"                 //| ZSET | 用户UID集合 | 成员:UID 分值:TTL |
	CR_KEY_SID_INCR               = "cr:sid:incr"                 //*| STRING | 会话SID增量器 | 只增不减 注意:sid不能为0 |
	CR_KEY_SID_ATTR               = "cr:sid:%d:attr"              //*| HASH | 会话SID属性 | 包含UID/NID |
	CR_KEY_UID_TO_SID_SET         = "cr:uid:%d:to:sid:set"        //| SET | 用户UID对应的会话SID集合 | SID集合 |
	CR_KEY_RID_INCR               = "cr:rid:incr"                 //*| STRING | 聊天室RID记录器|
	CR_KEY_RID_ZSET               = "cr:rid:zset"                 //*| ZSET | 聊天室RID集合 | 成员:RID 分值:TTL |
	CR_KEY_RID_ATTR               = "cr:rid:%d:attr"              //*| HASH | 聊天室属性信息| STATUS:(0:打开 1:关闭) |
	CR_KEY_ROOM_GROUP_CAP_ZSET    = "cr:room:group:cap:zset"      //*| ZSET | 聊天室分组容量 | 成员:RID 分值:分组容量 |
	CR_KEY_SID_TO_RID_ZSET        = "cr:sid:%d:to:rid:zset"       //*| ZSET | 会话SID对应的RID集合 | 成员:RID 分值:GID |
	CR_KEY_ROOM_GROUP_USR_NUM     = "cr:room:group:usr:num"       //| ZSET | 聊天室分组人数配置 | 成员:RID 分值:USERNUM |
	CR_KEY_RID_GID_TO_NUM_ZSET    = "cr:rid:%d:to:gid:num:zset"   //*| ZSET | 某聊天室各组人数 | 成员:GID 分值:USERNUM |
	CR_KEY_RID_TO_NID_ZSET        = "cr:rid:%d:to:nid:zset"       //*| ZSET | 某聊天室->帧听层 | 成员:NID 分值:TTL |
	CR_KEY_RID_NID_TO_NUM_ZSET    = "cr:rid:%d:nid:to:num:zset"   //*| ZSET | 某聊天室各帧听层人数 | 成员:NID 分值:USERNUM | 由帧听层上报数据获取
	CR_KEY_RID_SUB_USR_NUM_ZSET   = "cr:rid:sub:usr:num:zset"     //| ZSET | 聊天室人数订阅集合 | 暂无 |
	CR_KEY_RID_TO_UID_SID_ZSET    = "cr:rid:%d:to:uid:sid:zset"   //| ZSET | 聊天室用户列表 | 成员:"${UID}:${SID}" 分值:TTL |
	CR_KEY_RID_TO_SID_ZSET        = "cr:rid:%d:to:sid:zset"       //| ZSET | 聊天室SID列表 | 成员:SID 分值:TTL |
	CR_KEY_ROOM_MESG_QUEUE        = "cr:rid:%d:mesg:queue"        //| LIST | 聊天室消息队列 |
	CR_KEY_ROOM_MSGID_INCR        = "cr:rid:%d:msgid:incr"        //| STRING | 聊天室消息序列递增记录 |
	CR_KEY_ROOM_USR_GAG_SET       = "cr:rid:%d:usr:gag:set"       //*| SET | 聊天室用户禁言名单 | 成员:UID |
	CR_KEY_ROOM_USR_BLACKLIST_SET = "cr:rid:%d:usr:blacklist:set" //*| SET | 聊天室用户黑名单 | 成员:UID |
	CR_KEY_ROOM_ROLE_TAB          = "cr:rid:%d:role:tab"          //*| HASH | 聊天室管理人员名单 | 成员:UID 值:角色(1:OWNER 2:管理员) |
	CR_KEY_ROOM_INFO_TAB          = "cr:rid:%d:info:tab"          //*| HASH | 聊天室基本信息管理 |
	CR_KEY_ROOM_BC_ZSET           = "cr:rid:%d:broadcast:zset"    //| ZSET | 聊天室广播集合 | 成员:消息ID 分值:超时时间 |
	CR_KEY_ROOM_BC_HASH           = "cr:rid:%d:broadcast:hash"    //| HASH | 聊天室广播内容 | 域:消息ID 值:广播内容 |
)

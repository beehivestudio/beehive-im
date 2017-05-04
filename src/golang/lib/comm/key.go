package comm

const (
	IM_FMT_IP_PORT_STR     = "%s:%d"           //| IP+PORT
	CHAT_FMT_UID_SID_STR   = "%d:%d"           // 格式:${UID}:${SID} 说明:主键CHAT_KEY_RID_TO_UID_SID_ZSET的成员
	CHAT_FMT_UID_MSGID_STR = "uid:%d:msgid:%d" //| STRING | UID+MSGID
)

/* 侦听层结点属性 */
const (
	IM_LSND_ATTR_ADDR       = "ATTR"        //| IP地址
	IM_LSND_ATTR_PORT       = "PORT"        //| 侦听PORT
	IM_LSND_ATTR_TYPE       = "TYPE"        //| 侦听层类型(0:未知 1:TCP 2:WS)
	IM_LSND_ATTR_STATUS     = "STATUS"      //| 侦听层状态
	IM_LSND_ATTR_CONNECTION = "CONNECTIONS" //| 在线连接数
)

/* 路由层结点属性 */
const (
	IM_FRWD_ATTR_ADDR     = "ATTR"     //| IP地址
	IM_FRWD_ATTR_BC_PORT  = "BC-PORT"  //| 后置PORT
	IM_FRWD_ATTR_FWD_PORT = "FWD-PORT" //| 前置PORT
)

/* 群组属性 */
const (
	CHAT_GID_ATTR_SWITCH = "SWITCH" //| 开关状态
)

//#IM系统REDIS键值定义列表
const (
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	IM_KEY_SID_ZSET       = "im:sid:zset"          //*| ZSET | 会话SID集合 | 成员:SID 分值:TTL |
	IM_KEY_UID_ZSET       = "im:uid:zset"          //| ZSET | 用户UID集合 | 成员:UID 分值:TTL |
	IM_KEY_SID_INCR       = "im:sid:incr"          //*| STRING | 会话SID增量器 | 只增不减 注意:sid不能为0 |
	IM_KEY_SID_ATTR       = "im:sid:%d:attr"       //*| HASH | 会话SID属性 | 包含UID/NID |
	IM_KEY_UID_TO_SID_SET = "im:uid:%d:to:sid:set" //| SET | 用户UID对应的会话SID集合 | SID集合 |

	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	//私聊
	CHAT_KEY_USR_SEND_MESG_HTAB        = "chat:uid:%d:send:mesg:htab"     //| HTAB | 用户发送的私聊消息 | 字段:消息ID 内容:消息内容 |
	CHAT_KEY_PRIVATE_MESG_TIMEOUT_ZSET = "chat:private:mesg:timeout:zset" //| ZSET | 私聊消息超时管理 | 成员:消息ID 分值:发起时间 |
	CHAT_KEY_USR_OFFLINE_ZSET          = "chat:uid:%d:offline:zset"       //| ZSET | 用户离线数据队列 | 成员:消息ID 分值:发起时间 |
	CHAT_KEY_USR_BLACKLIST_ZSET        = "chat:uid:%d:blacklist:zset"     //| ZSET | 用户黑名单记录 | 成员:用户UID 分值:加入黑名单的时间 |
	CHAT_KEY_USR_GAG_ZSET              = "chat:uid:%d:gag:zset"           //| ZSET | 用户禁言记录 | 成员:用户UID 分值:设置禁言的时间 |
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	//聊天室
	CHAT_KEY_RID_INCR               = "chat:rid:incr"                 //*| STRING | 聊天室RID记录器|
	CHAT_KEY_RID_ZSET               = "chat:rid:zset"                 //*| ZSET | 聊天室RID集合 | 成员:RID 分值:TTL |
	CHAT_KEY_RID_ATTR               = "chat:rid:%d:attr"              //*| HASH | 聊天室属性信息| STATUS:(0:打开 1:关闭) |
	CHAT_KEY_ROOM_GROUP_CAP_ZSET    = "chat:room:group:cap:zset"      //*| ZSET | 聊天室分组容量 | 成员:RID 分值:分组容量 |
	CHAT_KEY_UID_TO_RID             = "chat:uid:%d:to:rid:htab"       //| HASH | 用户UID对应的RID集合 | 成员:RID 分值:GID |
	CHAT_KEY_SID_TO_RID_ZSET        = "chat:sid:%d:to:rid:zset"       //*| ZSET | 会话SID对应的RID集合 | 成员:RID 分值:GID |
	CHAT_KEY_ROOM_GROUP_USR_NUM     = "chat:room:group:usr:num"       //| ZSET | 聊天室分组人数配置 | 成员:RID 分值:USERNUM |
	CHAT_KEY_RID_GID_TO_NUM_ZSET    = "chat:rid:%d:to:gid:num:zset"   //*| ZSET | 某聊天室各组人数 | 成员:GID 分值:USERNUM |
	CHAT_KEY_RID_TO_NID_ZSET        = "chat:rid:%d:to:nid:zset"       //*| ZSET | 某聊天室->帧听层 | 成员:NID 分值:TTL |
	CHAT_KEY_RID_NID_TO_NUM_ZSET    = "chat:rid:%d:nid:to:num:zset"   //*| ZSET | 某聊天室各帧听层人数 | 成员:NID 分值:USERNUM | 由帧听层上报数据获取
	CHAT_KEY_RID_SUB_USR_NUM_ZSET   = "chat:rid:sub:usr:num:zset"     //| ZSET | 聊天室人数订阅集合 | 暂无 |
	CHAT_KEY_RID_TO_UID_SID_ZSET    = "chat:rid:%d:to:uid:sid:zset"   //| ZSET | 聊天室用户列表 | 成员:"${UID}:${SID}" 分值:TTL |
	CHAT_KEY_RID_TO_SID_ZSET        = "chat:rid:%d:to:sid:zset"       //| ZSET | 聊天室SID列表 | 成员:SID 分值:TTL |
	CHAT_KEY_ROOM_MESG_QUEUE        = "chat:rid:%d:mesg:queue"        //| LIST | 聊天室消息队列 |
	CHAT_KEY_ROOM_MSGID_INCR        = "chat:rid:%d:msgid:incr"        //| STRING | 聊天室消息序列递增记录 |
	CHAT_KEY_ROOM_USR_GAG_SET       = "chat:rid:%d:usr:gag:set"       //*| SET | 聊天室用户禁言名单 | 成员:UID |
	CHAT_KEY_ROOM_USR_BLACKLIST_SET = "chat:rid:%d:usr:blacklist:set" //*| SET | 聊天室用户黑名单 | 成员:UID |
	CHAT_KEY_ROOM_ROLE_TAB          = "chat:rid:%d:role:tab"          //*| HASH | 聊天室管理人员名单 | 成员:UID 值:角色(1:OWNER 2:管理员) |
	CHAT_KEY_ROOM_INFO_TAB          = "chat:rid:%d:info:tab"          //*| HASH | 聊天室基本信息管理 |
	CHAT_KEY_ROOM_BC_ZSET           = "chat:rid:%d:broadcast:zset"    //| ZSET | 聊天室广播集合 | 成员:消息ID 分值:超时时间 |
	CHAT_KEY_ROOM_BC_HASH           = "chat:rid:%d:broadcast:hash"    //| HASH | 聊天室广播内容 | 域:消息ID 值:广播内容 |
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	//群聊
	CHAT_KEY_GID_INCR                = "chat:gid:incr"                 //*| STRING | 群组GID记录器|
	CHAT_KEY_GID_ZSET                = "chat:gid:zset"                 //| ZSET | 群ID集合 | 成员:GID 分值:TTL |
	CHAT_KEY_GID_ATTR                = "chat:gid:%d:attr"              //*| HASH |群组属性信息| SWITCH:(0:打开 1:关闭) |
	CHAT_KEY_UID_TO_GID              = "chat:uid:%d:to:gid:htab"       //| HASH | 用户UID对应的GID集合 | 成员:GID 分值:0 |
	CHAT_KEY_GID_TO_NID_ZSET         = "chat:gid:%d:to:nid:zset"       //| ZSET | 某群->帧听层 | 成员:NID 分值:TTL |
	CHAT_KEY_GROUP_CAP_ZSET          = "chat:group:cap:zset"           //*| ZSET | 群组容量 | 成员:GID 分值:容量 |
	CHAT_KEY_GID_TO_UID_ZSET         = "chat:gid:%d:to:uid:zset"       //| ZSET | 某群在线用户列表 | 成员:UID 分值:TTL |
	CHAT_KEY_GID_TO_SID_ZSET         = "chat:gid:%d:to:sid:zset"       //| ZSET | 某群SID列表 | 成员:SID 分值:TTL |
	CHAT_KEY_GROUP_MESG_QUEUE        = "chat:gid:%d:mesg:queue"        //| LIST | 群聊消息队列 |
	CHAT_KEY_GROUP_MSGID_INCR        = "chat:gid:%d:msgid:incr"        //| STRING | 群聊消息序列递增记录 |
	CHAT_KEY_GROUP_USR_GAG_SET       = "chat:gid:%d:usr:gag:set"       //*| SET | 群组用户禁言名单 | 成员:UID |
	CHAT_KEY_GROUP_USR_BLACKLIST_SET = "chat:gid:%d:usr:blacklist:set" //*| SET | 群组用户黑名单 | 成员:UID |
	CHAT_KEY_GROUP_ROLE_TAB          = "chat:gid:%d:role:tab"          //*| HASH | 群组管理人员名单 | 成员:UID 值:角色(1:OWNER 2:管理员) |
	CHAT_KEY_GROUP_INFO_TAB          = "chat:gid:%d:info:tab"          //*| HASH | 群组基本信息管理 |
	//|**宏**|**键值**|**类型**|**描述**|**备注**|
	IM_KEY_LSND_TYPE_ZSET      = "im:lsnd:type:zset"                           //| ZSET | 帧听层"类型"集合 | 成员:"网络类型" 分值:TTL |
	IM_KEY_LSND_NATION_ZSET    = "im:lsnd:type:%d:nation:zset"                 //| ZSET | 某"类型"的帧听层"地区/国家"集合 | 成员:"国家/地区" 分值:TTL |
	IM_KEY_LSND_OP_ZSET        = "im:lsnd:type:%d:nation:%s:op:zset"           //| ZSET | 帧听层"地区/国家"对应的运营商集合 | 成员:运营商名称 分值:TTL |
	IM_KEY_LSND_IP_ZSET        = "im:lsnd:type:%d:nation:%s:op:%d:zset"        //| ZSET | 帧听层"地区/国家"-运营商对应的IP集合 | 成员:IP 分值:TTL |
	IM_KEY_LSND_OP_TO_NID_ZSET = "im:lsnd:type:%d:nation:%s:op:%d:to:nid:zset" //| ZSET | 运营商帧听层NID集合 | 成员:NID 分值:TTL |
	IM_KEY_LSND_NID_ZSET       = "im:lsnd:nid:zset"                            //| ZSET | 帧听层NID集合 | 成员:NID 分值:TTL |
	IM_KEY_LSND_ATTR           = "im:lsnd:nid:%d:attr"                         //| HASH | 帧听层NID->地址 | 键:NID/值:外网IP+端口 |
	IM_KEY_LSND_ADDR_TO_NID    = "im:lsnd:addr:to:nid"                         //| HASH | 转发层地址->NID | 键:内网IP+端口/值:NID |
	IM_KEY_FRWD_NID_ZSET       = "im:frwd:nid:zset"                            //| ZSET | 转发层NID集合 | 成员:NID 分值:TTL |
	IM_KEY_FRWD_ATTR           = "im:frwd:nid:%d:attr"                         //| HASH | 转发层NID->"地址/前置端口/后置端口"等信息
	IM_KEY_PREC_NUM_ZSET       = "im::prec:num:zset"                           //| ZSET | 人数统计精度 | 成员:prec 分值:记录条数 |
	IM_KEY_PREC_USR_MAX_NUM    = "im:usr:statis:prec:%d:max:num"               //| HASH | 某统计精度最大人数 | 键:时间/值:最大人数 |
	IM_KEY_PREC_USR_MIN_NUM    = "im:usr:statis:prec:%d:min:num"               //| HASH | 某统计精度最少人数 | 键:时间/值:最少人数 |
)

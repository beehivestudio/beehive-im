#协议头

---
|**序号**|**字段名**|**字段类型**|**字段长度(字节)**|**字段含义**|**备注**|
|:------:|:---------|:-----------|:-------------:|:-------|:-------|
| 01 | type | uint32_t | 4 |消息类型|命令ID|
| 02 | flag | uint32_t | 4 |标识量|-1：系统数据类型 1:外部数据类型(默认)|
| 03 | length | uint32_t | 4 |报体长度|不包含报头|
| 04 | chksum | uint32_t | 4 |校验值|必须为:0x1ED23CB4|
| 05 | sid | uint64_t | 8 |会话ID|每个连接的会话ID都不一样|
| 06 | nid | uint32_t | 4 |结点ID|外部无需关心|
| 07 | serial | uint64_t | 8 |流水号|暂无|
| 08 | version | uint32_t | 4 |版本号|暂时设置为1|
| 09 | rsv | char[] | 4 |预留字段|暂无|
| 10 | body | char[] | 0 |消息体|各协议报体内容, 紧接在协议头后|

////////////////////////////////////////////////////////////////////////////////
#通用消息

---
命令ID: 0x0101:<br>
命令描述: 上线请求(ONLINE)<br>
协议格式:<br>
>message mesg_online_req<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 sid = 2;        // M|会话ID|数字|<br>
>   required string token = 3;      // M|鉴权TOKEN|字串|<br>
>   required string app = 4;        // M|APP名|字串|<br>
>   required string version = 5;    // M|APP版本|字串|<br>
>   optional uint32 terminal = 6;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
>}


---
命令ID: 0x0102<br>
命令描述: 上线请求应答(ONLINE-ACK)<br>
协议格式:<br>
>message mesg_online_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 sid = 2;        // M|会话ID|数字|<br>
>   required string app = 3;        // M|APP名|字串|<br>
>   required string version = 4;    // M|APP版本|字串|<br>
>   optional uint32 terminal = 5;   // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
>   required uint32 code = 6;       // M|错误码|数字|<br>
>   required string errmsg = 7;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0103<br>
命令描述: 下线请求(OFFLINE)<br>
协议格式: NONE

---
命令ID: 0x0104<br>
命令描述: 下线请求应答(OFFLINE-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0105<br>
命令描述: 客户端心跳(PING)<br>
协议格式: NONE

---
命令ID: 0x0106<br>
命令描述: 客户端心跳应答(PONG)<br>
协议格式: NONE

---
命令ID: 0x0107<br>
命令描述: 订阅请求(SUB)<br>
协议格式:<br>
>message mesg_sub_req<br>
>{<br>
>   optional uint32 sub = 1;        // M|订阅的数据|数字| <br>
>}

---
命令ID: 0x0108<br>
命令描述: 订阅应答(SUB-ACK)<br>
协议格式:<br>
>message mesg_sub_ack<br>
>{<br>
>   required uint32 sub = 1;        // M|订阅的数据|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0109<br>
命令描述: 取消订阅(UNSUB)<br>
协议格式:<br>
>message mesg_unsub_req<br>
>{<br>
>   required uint32 sub = 1;        // M|取消订阅的数据|数字| <br>
>}

---
命令ID: 0x010A<br>
命令描述: 取消订阅应答(UNSUB-ACK)<br>
协议格式:<br>
>message mesg_unsub_ack<br>
>{<br>
>   required uint32 sub = 1;        // M|取消订阅的数据|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x010B<br>
命令描述: 通用异常消息(UNUSUAL)<br>
协议格式:<br>
>message mesg_unusualb_req<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x010C<br>
命令描述: 通用异常消息应答(UNUSUAL-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x010D<br>
命令描述: 同步消息(SYNC)<br>
协议格式: NONE<br>
>message mesg_sync<br>
>{<br>
>   required uint64 uid = 1;       // M|用户ID|数字|<br>
>}

---
命令ID: 0x010E<br>
命令描述: 同步消息应答(SYNC-ACK)<br>
协议格式: NONE<br>
>message mesg_sync_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0110<br>
命令描述: 申请序列号(ALLOC-SEQ)<br>
协议格式:<br>
>message mesg_alloc_seq<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint16 num = 2;        // M|申请序列号个数|数字|<br>
>}

---
命令ID: 0x0111<br>
命令描述: 申请序列号应答(ALLOC-SEQ-ACK)<br>
协议格式: NONE<br>
>message mesg_alloc_seq_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 seq = 2;        // M|序列号起始值|数字|<br>
>   required uint16 num = 3;        // M|分配序列号个数|数字|<br>
>   required uint32 code = 4;       // M|错误码|数字|<br>
>   required string errmsg = 5;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0112<br>
命令描述: 踢连接下线(KICK)<br>
协议格式: <br>
>message mesg_kick_req<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0113<br>
命令描述: 踢连接下线应答(KICK-ACK)<br>
协议格式: NONE<br>

////////////////////////////////////////////////////////////////////////////////
#私聊消息

---
命令ID: 0x0201<br>
命令描述: 私聊消息(CHAT)<br>
协议格式: 透传<br>
>message mesg_chat<br>
>{<br>
>   required uint64 orig = 1;       // M|发送方UID<br>
>   required uint64 dest = 2;       // M|接收方UID<br>
>   required uint32 level = 3;      // M|消息级别<br>
>   required uint64 time = 4;       // M|发送时间<br>
>   required string text = 5;       // M|聊天内容<br>
>   optional bytes data = 6;        // M|透传数据<br>
>}

---
命令ID: 0x0202<br>
命令描述: 私聊消息应答(CHAT-ACK)<br>
协议格式:
>message mesg_chat_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0203<br>
命令描述: 添加好友(FRIEND-ADD)<br>
协议格式:
>message mesg_friend_add<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>   required string mark = 3;       // M|备注信息|字串|<br>
>}

---
命令ID: 0x0204<br>
命令描述: 添加好友应答(FRIEND-ADD-ACK)<br>
协议格式: NONE

---
命令ID: 0x0205<br>
命令描述: 删除好友(FRIEND-DEL)<br>
协议格式:
>message mesg_friend_del<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x0206<br>
命令描述: 删除好友应答(FRIEND-DEL-ACK)<br>
协议格式: NONE

---
命令ID: 0x0207<br>
命令描述: 加入黑名单(BLACKLIST-ADD)<br>
协议格式: <br>
>message mesg_blacklist_add<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x0208<br>
命令描述: 加入黑名单应答(BLACKLIST-ADD-ACK)<br>
协议格式: <br>
>message mesg_blacklist_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0209<br>
命令描述: 移除黑名单(BLACKLIST-DEL)<br>
协议格式:
>message mesg_blacklist_del<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x020A<br>
命令描述: 移除黑名单应答(BLACKLIST-DEL-ACK)<br>
协议格式: <br>
>message mesg_blacklist_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x020B<br>
命令描述: 屏蔽此人(GAG-ADD)<br>
协议格式:
>message mesg_gag_add<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x020C<br>
命令描述: 屏蔽此人应答(GAG-ADD-ACK)<br>
协议格式: <br>
>message mesg_gag_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x020D<br>
命令描述: 取消屏蔽此人(GAG-DEL)<br>
协议格式:
>message mesg_gag_del<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x020E<br>
命令描述: 取消屏蔽此人应答(GAG-DEL-ACK)<br>
协议格式: <br>
>message mesg_gag_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0210<br>
命令描述: 添加备注此人(MARK-ADD)<br>
协议格式:
>message mesg_mark_add<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>   required uint64 mark = 3;       // M|备注名|字串|<br>
>}

---
命令ID: 0x0211<br>
命令描述: 添加备注此人应答(MARK-ADD-ACK)<br>
协议格式: NONE
>message mesg_mark_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0212<br>
命令描述: 添加备注此人(MARK-DEL)<br>
协议格式:
>message mesg_mark_del<br>
>{<br>
>   required uint64 orig = 1;       // M|源用户ID|数字|<br>
>   required uint64 dest = 2;       // M|目标用户ID|数字|<br>
>}

---
命令ID: 0x0213<br>
命令描述: 取消备注此人应答(MARK-DEL-ACK)<br>
协议格式: NONE
>message mesg_mark_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

////////////////////////////////////////////////////////////////////////////////
#群聊消息

---
命令ID: 0x0301<br>
命令描述: 创建群组(GROUP-CREAT)<br>
协议格式: <br>
>message mesg_group_creat<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0302<br>
命令描述: 创建群组应答(GROUP-JOIN-ACK)<br>
协议格式: <br>
>message mesg_group_creat_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0303<br>
命令描述: 解散群组(GROUP-DISMISS)<br>
协议格式: <br>
>message mesg_group_dismiss<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0304<br>
命令描述: 解散群组应答(GROUP-DISMISS-ACK)<br>
协议格式: <br>
>message mesg_group_dismiss_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}



---
命令ID: 0x0305<br>
命令描述: 申请入群(GROUP-JOIN)<br>
协议格式: <br>
>message mesg_group_join<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0306<br>
命令描述: 申请入群应答(GROUP-JOIN-ACK)<br>
协议格式: <br>
>message mesg_group_join_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0307<br>
命令描述: 退群(GROUP-QUIT)<br>
协议格式: <br>
>message mesg_group_quit<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0308<br>
命令描述: 退群应答(GROUP-QUIT-ACK)<br>
协议格式: <br>
>message mesg_group_quit_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0309<br>
命令描述: 邀请入群(GROUP-INVITE)<br>
协议格式: <br>
>message mesg_group_invite<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>   required uint64 to = 3;         // M|被邀请用户ID|数字|<br>
>}

---
命令ID: 0x030A<br>
命令描述: 邀请入群应答(GROUP-INVITE-ACK)<br>
协议格式: <br>
>message mesg_group_invite_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x030B<br>
命令描述: 群聊消息(GROUP-CHAT)<br>
协议格式: <br>
TODO: 协议头中的to为群ID(GID)
>message mesg_group_chat<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|分组ID<br>
>   required uint32 level = 3;      // M|消息级别<br>
>   required uint64 time = 4;       // M|发送时间<br>
>   required string text = 5;       // M|聊天内容<br>
>   optional bytes data = 6;        // M|透传数据<br>
>}

---
命令ID: 0x030C<br>
命令描述: 群聊消息应答(GROUP-CHAT-ACK)<br>
协议格式: <br>
>message mesg_group_chat_ack<br>
>{<br>
>   required uint64 seq = 1;        // M|群消息序列号|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x030D<br>
命令描述: 群组踢人(GROUP-KICK)<br>
协议格式: <br>
>message mesg_group_kick<br>
>{<br>
>   required uint64 uid = 1;        // M|被踢用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x030E<br>
命令描述: 群组踢人应答(GROUP-KICK-ACK)<br>
协议格式: <br>
>message mesg_group_kick_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0310<br>
命令描述: 群组禁言(GROUP-GAG-ADD)<br>
协议格式: <br>
>message mesg_group_gag_add<br>
>{<br>
>   required uint64 uid = 1;        // M|被禁言用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x0311<br>
命令描述: 群组禁言应答(GROUP-GAG-ADD-ACK)<br>
协议格式: <br>
>message mesg_group_gag_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0312<br>
命令描述: 解除群组禁言(GROUP-GAG-DEL)<br>
协议格式: <br>
>message mesg_group_gag_del<br>
>{<br>
>   required uint64 uid = 1;        // M|被解除禁言用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x0313<br>
命令描述: 解除群组禁言应答(GROUP-GAG-DEL-ACK)<br>
协议格式: <br>
>message mesg_group_gag_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0314<br>
命令描述: 加入群组黑名单(GROUP-BL-ADD)<br>
协议格式: <br>
>message mesg_group_bl_add<br>
>{<br>
>   required uint64 uid = 1;        // M|被加入黑名单用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x0315<br>
命令描述: 加入群组黑名单应答(GROUP-BL-ADD-ACK)<br>
协议格式: <br>
>message mesg_group_bl_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0316<br>
命令描述: 移除群组黑名单(GROUP-BL-DEL)<br>
协议格式: <br>
>message mesg_group_bl_del<br>
>{<br>
>   required uint64 uid = 1;        // M|被移除黑名单用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x0317<br>
命令描述: 移除群组黑名单应答(GROUP-BL-DEL-ACK)<br>
协议格式: <br>
>message mesg_group_bl_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0318<br>
命令描述: 添加群组管理员(GROUP-MGR-ADD)<br>
协议格式: <br>
>message mesg_group_mgr_add<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x0319<br>
命令描述: 添加群组管理员应答(GROUP-MGR-ADD-ACK)<br>
协议格式: <br>
>message mesg_group_mgr_add_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x031A<br>
命令描述: 解除群组管理员(GROUP-MGR-DEL)<br>
协议格式: <br>
>message mesg_group_mgr_del<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID<br>
>}

---
命令ID: 0x031B<br>
命令描述: 解除群组管理员应答(GROUP-MGR-DEL-ACK)<br>
协议格式: <br>
>message mesg_group_mgr_del_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x031C<br>
命令描述: 群组成员列表请求(GROUP-USR_LIST_REQ)<br>
协议格式: <br>
>message mesg_group_usr_list_req<br>
>{<br>
>   required uint64 gid = 1;        // M|群组ID<br>
>   required uint32 num = 2;        // M|请求人数|数字|(备注:当num=0时, 表示获取所有人员列表; 当num>0时, 表示获取num个人员列表)
>}

---
命令ID: 0x031B<br>
命令描述: 群组成员列表应答(GROUP-USR_LIST_ACK)<br>
协议格式: <br>
>message mesg_group_usr_list_ack<br>
>{<br>
>   required uint64 gid = 1;        // M|群组ID|数字|<br>
>   required string list = 2;       // M|用户列表|字串|JSON格式<br>
>}

---
命令ID: 0x0350<br>
命令描述: 入群通知(GROUP-JOIN-NTC)<br>
协议格式: <br>
>message mesg_group_join_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0351<br>
命令描述: 退群通知(GROUP-QUIT-NTC)<br>
协议格式: <br>
>message mesg_group_quit_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0352<br>
命令描述: 踢人通知(GROUP-KICK-NTC)<br>
协议格式: <br>
>message mesg_group_kick_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0353<br>
命令描述: 禁言通知(GROUP-GAG-ADD-NTC)<br>
协议格式: <br>
>message mesg_group_gag_add_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0354<br>
命令描述: 解除禁言通知(GROUP-GAG-DEL-NTC)<br>
协议格式: <br>
>message mesg_group_gag_del_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0355<br>
命令描述: 加入黑名单通知(GROUP-BL-ADD-NTC)<br>
协议格式: <br>
>message mesg_group_bl_add_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0356<br>
命令描述: 移除黑名单通知(GROUP-BL-DEL-NTC)<br>
协议格式: <br>
>message mesg_group_bl_del_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0357<br>
命令描述: 添加管理员通知(GROUP-MGR-ADD-NTC)<br>
协议格式: <br>
>message mesg_group_mgr_add_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

---
命令ID: 0x0358<br>
命令描述: 移除管理员通知(GROUP-MGR-DEL-NTC)<br>
协议格式: <br>
>message mesg_group_mgr_del_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 gid = 2;        // M|群组ID|数字|<br>
>}

////////////////////////////////////////////////////////////////////////////////
#聊天室消息

---
命令ID: 0x0401<br>
命令描述: 创建聊天室(ROOM-CREAT)<br>
协议格式:<br>
>message mesg_room_creat<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required string name = 2;       // M|聊天室名称|字串|<br>
>   required string desc = 3;       // M|聊天室描述|字串|<br>
>}

---
命令ID: 0x0402<br>
命令描述: 创建聊天室应答(ROOM-CREAT-ACK)<br>
协议格式:<br>
>message mesg_room_creat_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 code = 3;       // M|错误码|数字|<br>
>   required string errmsg = 4;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0403<br>
命令描述: 解散聊天室(ROOM-DISMISS)<br>
协议格式: NONE
>message mesg_room_dismiss<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>}

---
命令ID: 0x0404<br>
命令描述: 解散聊天室应答(ROOM-DISMISS-ACK)<br>
协议格式: NONE
>message mesg_room_dismiss_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|数字|<br>
>}

---
命令ID: 0x0405<br>
命令描述: 加入聊天室(ROOM-JOIN)<br>
协议格式:<br>
>message mesg_room_join<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>}

---
命令ID: 0x0406<br>
命令描述: 加入聊天室应答(ROOM-JOIN-ACK)<br>
协议格式:<br>
>message mesg_room_join_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 gid = 3;        // M|分组ID|数字|<br>
>   required uint32 code = 4;       // M|错误码|数字|<br>
>   required string errmsg = 5;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0407<br>
命令描述: 退出聊天室(ROOM-QUIT)<br>
协议格式: NONE
>message mesg_room_quit<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>}

---
命令ID: 0x0408<br>
命令描述: 退出聊天室应答(ROOM-QUIT-ACK)<br>
协议格式: NONE
>message mesg_room_quit_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 code = 3;       // M|错误码|数字|<br>
>   required string errmsg = 4;     // M|错误描述|数字|<br>
>}

---
命令ID: 0x0409<br>
命令描述: 踢出聊天室(ROOM-KICK)<br>
协议格式: NONE
>message mesg_room_kick<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 code = 3;       // M|错误码|数字|
>   required string errmsg = 4;     // M|错误描述|数字|
>}

---
命令ID: 0x040A<br>
命令描述: 踢出聊天室应答(ROOM-KICK-ACK)<br>
协议格式: NONE
>message mesg_room_kick_ack<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 code = 3;       // M|错误码|数字|<br>
>   required string errmsg = 4;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x040B<br>
命令描述: 聊天室消息(ROOM-CHAT)<br>
协议格式: <br>
>message mesg_room<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID<br>
>   required uint64 rid = 2;        // M|聊天室ID<br>
>   required uint32 gid = 3;        // M|分组ID<br>
>   required uint32 level = 4;      // M|消息级别<br>
>   required uint64 time = 5;       // M|发送时间<br>
>   required string text = 6;       // M|聊天内容<br>
>   optional bytes data = 7;        // M|透传数据<br>
>}

---
命令ID: 0x040C<br>
命令描述: 聊天室消息应答(ROOM-ACK)<br>
协议格式: NONE<br>
>message mesg_room_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码<br>
>   required string errmsg = 2;     // M|错误描述<br>
>}

---
命令ID: 0x040D<br>
命令描述: 聊天室广播消息(ROOM-BC)<br>
协议格式: <br>
>message mesg_room_bc<br>
>{<br>
>   required uint64 rid = 1;        // M|聊天室ID<br>
>   required uint32 level = 2;      // M|消息级别<br>
>   required uint64 time = 3;       // M|发送时间<br>
>   required uint32 expire = 4;     // M|过期时间<br>
>   required bytes data = 5;        // M|透传数据<br>
>}

---
命令ID: 0x040E<br>
命令描述: 聊天室广播消息应答(ROOM-BC-ACK)<br>
协议格式: NONE<br>
>message mesg_room_bc_ack<br>
>{<br>
>   required uint32 code = 1;       // M|错误码<br>
>   required string errmsg = 2;     // M|错误描述<br>
>}

---
命令ID: 0x0410<br>
命令描述: 聊天室人数(ROOM-USR-NUM)<br>
协议格式:<br>
>message mesg_room_usr_num<br>
>{<br>
>   required uint64 rid = 1;        // M|聊天室ID|数字|<br>
>   required uint32 num = 2;        // M|用户人数|数字|<br>
>}

---
命令ID: 0x0411<br>
命令描述: 聊天室人数应答(ROOM-USR-NUM-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0450<br>
命令描述: 加入聊天室通知(ROOM-JOIN-NTC)<br>
协议格式: <br>
>message mesg_room_join_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0451<br>
命令描述: 退出聊天室通知(ROOM-QUIT-NTC)<br>
协议格式: <br>
>message mesg_room_quit_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0452<br>
命令描述: 踢出聊天室通知(ROOM-KICK-NTC)<br>
协议格式: <br>
>message mesg_room_kick_ntc<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

////////////////////////////////////////////////////////////////////////////////
# 推送消息

---
命令ID: 0x0501<br>
命令描述: 广播消息(BC)<br>
功能描述: 用于给所有人员发送广播消息
协议格式: 透传<br>

---
命令ID: 0x0502<br>
命令描述: 广播消息应答(BC-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0503<br>
命令描述: 点到点消息(P2P)<br>
功能描述: 可用于发送私聊消息、添加/删除好友等点到点的消息<br>
协议格式: 自定义<br>

---
命令ID: 0x0504<br>
命令描述: 点到点消息应答(P2P-ACK)<br>
协议格式: NONE<br>

////////////////////////////////////////////////////////////////////////////////
# 系统内部命令

---
命令ID: 0x0601<br>
命令描述: 帧听层信息上报(LSN-INFO)<br>
协议格式: <br>
>message mesg_lsn_info<br>
>{<br>
>   required uint32 type = 1;       // M|类型(0:Unknown 1:TCP 2:WS)|数字|<br>
>   required uint32 nid = 2;        // M|结点ID|数字|<br>
>   required string nation = 3;     // M|所属国家|字串|<br>
>   required string name = 4;       // M|运营商名称|字串|<br>
>   required string ipaddr = 5;     // M|IP地址|字串|<br>
>   required uint32 port = 6;       // M|端口|数字|<br>
>   required uint32 user_num = 7;   // M|在线人数|数字|<br>
>}

---
命令ID: 0x0602<br>
命令描述: 帧听层上报应答(LSN-INFO-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0603<br>
命令描述: 转发层上报 (FRWD-INFO)<br>
协议格式: <br>
>message mesg_frwd_info<br>
>{<br>
>   required uint32 nid = 1;        // M|结点ID|数字|<br>
>   required string ipaddr = 2;     // M|IP地址|字串|<br>
>   required uint32 forward_port = 3;    // M|前端口号|数字|<br>
>   required uint32 backend_port = 4;    // M|后端口号|数字|<br>
>}

---
命令ID: 0x0604<br>
命令描述: 转发层上报应答(FRWD-INFO-ACK)<br>
协议格式: NONE<br>

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

#命令列表

---
命令ID: 0x0101:<br>
命令描述: 上线请求(ONLINE)<br>
协议格式:<br>
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
命令描述: 加入聊天室(JOIN)<br>
协议格式:<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required string token = 3;      // M|鉴权TOKEN|字串|<br>
>}

---
命令ID: 0x0106<br>
命令描述: 加入聊天室应答(JOIN-ACK)<br>
协议格式:<br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>   required uint32 gid = 3;        // M|分组ID|数字|<br>
>   required uint32 code = 4;       // M|错误码|数字|<br>
>   required string errmsg = 5;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0107<br>
命令描述: 退出聊天室(UNJOIN)<br>
协议格式: NONE
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>   required uint64 rid = 2;        // M|聊天室ID|数字|<br>
>}

---
命令ID: 0x0108<br>
命令描述: 退出聊天室应答(UNJOIN-ACK)<br>
协议格式: NONE

---
命令ID: 0x0109<br>
命令描述: 客户端心跳(PING)<br>
协议格式: NONE

---
命令ID: 0x010A<br>
命令描述: 客户端心跳应答(PONG)<br>
协议格式: NONE

---
命令ID: 0x010B<br>
命令描述: 订阅请求(SUB)<br>
协议格式:<br>
>{<br>
>   optional uint32 sub = 1;        // M|订阅的数据|数字| <br>
>}

---
命令ID: 0x010C<br>
命令描述: 订阅应答(SUB-ACK)<br>
协议格式:<br>
>{<br>
>   required uint32 sub = 1;        // M|订阅的数据|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x010D<br>
命令描述: 取消订阅(UNSUB)<br>
协议格式:<br>
>{<br>
>   required uint32 sub = 1;        // M|取消订阅的数据|数字| <br>
>}

---
命令ID: 0x010E<br>
命令描述: 取消订阅应答(UNSUB-ACK)<br>
协议格式:<br>
>{<br>
>   required uint32 sub = 1;        // M|取消订阅的数据|数字|<br>
>   required uint32 code = 2;       // M|错误码|数字|<br>
>   required string errmsg = 3;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0110<br>
命令描述: 群聊消息(GROUP-MSG)<br>
协议格式: 透传<br>
TODO: 协议头中的to为群ID(GID)
>{<br>
>   required uint64 gid = 1;        // M|分组ID<br>
>   required uint32 level = 2;      // M|消息级别<br>
>   required string text = 3;       // M|聊天内容<br>
>   optional bytes data = 4;        // M|透传数据<br>
>}

---
命令ID: 0x0111<br>
命令描述: 群聊消息应答(GROUP-MSG-ACK)<br>
协议格式: <br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0112<br>
命令描述: 私聊消息(PRVT-MSG)<br>
协议格式: 透传<br>
TODO: 协议头中的to为用户ID(UID)
>{<br>
>   required uint32 level = 1;      // M|消息级别<br>
>   required string text = 2;       // M|聊天内容<br>
>   optional bytes data = 3;        // M|透传数据<br>
>}

---
命令ID: 0x0113<br>
命令描述: 私聊消息应答(PRVG-MSG-ACK)<br>
协议格式:
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0114<br>
命令描述: 广播消息(BC-MSG)<br>
功能描述: 用于给所有人员发送广播消息
协议格式: 透传<br>

---
命令ID: 0x0115<br>
命令描述: 广播消息应答(BC-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0116<br>
命令描述: 点到点消息(P2P-MSG)<br>
功能描述: 可用于发送私聊消息、添加/删除好友等点到点的消息<br>
协议格式: 自定义<br>

---
命令ID: 0x0117<br>
命令描述: 点到点消息应答(P2P-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0118<br>
命令描述: 聊天室消息(ROOM-MSG)<br>
协议格式: <br>
>{<br>
>   required uint64 rid = 1;        // M|聊天室ID<br>
>   required uint32 gid = 2;        // M|分组ID<br>
>   required uint32 level = 3;      // M|消息级别<br>
>   required string text = 4;       // M|聊天内容<br>
>   optional bytes data = 5;        // M|透传数据<br>
>}

---
命令ID: 0x0119<br>
命令描述: 聊天室消息应答(ROOM-MSG-ACK)<br>
协议格式: NONE<br>
>{<br>
>   required uint32 code = 1;       // M|错误码<br>
>   required string errmsg = 2;     // M|错误描述<br>
>}

---
命令ID: 0x011A<br>
命令描述: 聊天室广播消息(ROOM-BC-ACK)<br>
协议格式: 透传<br>

---
命令ID: 0x011B<br>
命令描述: 聊天室广播消息应答(ROOM-BC-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x011C<br>
命令描述: 通用异常消息(UNUSUAL)<br>
协议格式:<br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x011D<br>
命令描述: 通用异常消息应答(UNUSUAL-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x011E<br>
命令描述: 聊天室人数(ROOM-USR-NUM)<br>
协议格式:<br>
>{<br>
>   required uint64 rid = 1;        // M|聊天室ID|数字|<br>
>   required uint32 num = 2;        // M|用户人数|数字|<br>
>}

---
命令ID: 0x0120<br>
命令描述: 聊天室人数应答(ROOM-USR-NUM-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0121<br>
命令描述: 同步消息(SYNC-MSG)<br>
协议格式: NONE<br>

---
命令ID: 0x0122<br>
命令描述: 同步消息应答(SYNC-MSG-ACK)<br>
协议格式: NONE<br>

# 通知类命令

---
命令ID: 0x0301<br>
命令描述: 上线通知(ONLINE-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0302<br>
命令描述: 下线通知(OFFLINE-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0303<br>
命令描述: 加入聊天室通知(JOIN-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0304<br>
命令描述: 退出聊天室通知(UNJOIN-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0305<br>
命令描述: 禁言通知(BAN-ADD-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0306<br>
命令描述: 解除禁言通知(BAN-DEL-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0307<br>
命令描述: 加入黑名单通知(BLACKLIST-ADD-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

---
命令ID: 0x0308<br>
命令描述: 移除黑名单通知(BLACKLIST-DEL-NTC)<br>
协议格式: <br>
>{<br>
>   required uint64 uid = 1;        // M|用户ID|数字|<br>
>}

# 系统内部命令

---
命令ID: 0x0401<br>
命令描述: 帧听层上报(LSN-RPT)<br>
协议格式: <br>
>{<br>
>   required uint32 nid = 1;        // M|结点ID|数字|<br>
>   required uint32 op = 2;         // M|运营商ID|数字|<br>
>   required string ipaddr = 3;     // M|IP地址|字串|<br>
>   required uint32 port = 4;       // M|IP地址|字串|<br>
>}

---
命令ID: 0x0402<br>
命令描述: 帧听层上报应答(LSN-RPT-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0403<br>
命令描述: 转发层上报 (FRWD-RPT)<br>
协议格式: <br>
>{<br>
>   required uint32 nid = 1;        // M|结点ID|数字|<br>
>   required string ipaddr = 2;     // M|IP地址|字串|<br>
>   required uint32 forward_port = 3;    // M|前端口号|数字|<br>
>   required uint32 backend_port = 4;    // M|后端口号|数字|<br>
>}

---
命令ID: 0x0404<br>
命令描述: 转发层上报应答(FRWD-RPT-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0405<br>
命令描述: 踢连接下线(KICK)<br>
协议格式: <br>
>{<br>
>   required uint32 code = 1;       // M|错误码|数字|<br>
>   required string errmsg = 2;     // M|错误描述|字串|<br>
>}

---
命令ID: 0x0406<br>
命令描述: 踢连接下线应答(KICK-ACK)<br>
协议格式: NONE<br>

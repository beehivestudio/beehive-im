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
>   "uid":${uid},               // M|用户ID|数字| <br>
>   "token":"${token}",         // M|鉴权TOKEN|字串|<br>
>   "app":"${app}",             // M|APP名|字串|<br>
>   "version":"${version}",     // M|APP版本|字串|<br>
>   "terminal":${terminal}      // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
>}

---
命令ID: 0x0102<br>
命令描述: 上线请求应答(ONLINE-ACK)<br>
协议格式:<br>
>{<br>
>   "errno":${errno},           // M|错误码|数字|<br>
>   "errmsg":"${errmsg}"        // M|错误描述|字串|<br>
>}

---
命令ID: 0x0103<br>
命令描述: 下线请求(OFFLINE)<br>
协议格式: NONE<br>

---
命令ID: 0x0104<br>
命令描述: 下线请求应答(OFFLINE-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0105<br>
命令描述: 加入聊天室(JOIN)<br>
协议格式:<br>
>{<br>
>   "uid":${uid},               // M|用户ID|数字| <br>
>   "token":"${token}",         // M|鉴权TOKEN|字串|<br>
>   "app":"${app}",             // M|APP名|字串|<br>
>   "version":"${version}",     // M|APP版本|字串|<br>
>   "terminal":${terminal}      // O|终端类型|数字|(0:未知 1:PC 2:TV 3:手机)|<br>
>}

---
命令ID: 0x0106<br>
命令描述: 加入聊天室应答(JOIN-ACK)<br>
协议格式:<br>
>{<br>
>   "errno":${errno},           // M|错误码|数字|<br>
>   "errmsg":"${errmsg}"        // M|错误描述|字串|<br>
>}

---
命令ID: 0x0107<br>
命令描述: 退出聊天室(QUIT)<br>
协议格式: NONE

---
命令ID: 0x0108<br>
命令描述: 退出聊天室应答(QUIT-ACK)<br>
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
>   "sub":${sub}                // M|订阅的数据|数字| <br>
>}

---
命令ID: 0x010C<br>
命令描述: 订阅应答(SUB-ACK)<br>
协议格式:<br>
>{<br>
>   "sub":${sub},               // M|订阅的数据|数字|<br>
>   "errno":${errno},           // M|错误码|数字|<br>
>   "errmsg":"${errmsg}"        // M|错误描述|字串|<br>
>}

---
命令ID: 0x010D<br>
命令描述: 取消订阅(UNSUB)<br>
协议格式:<br>
>{<br>
>   "sub":${sub}                // M|取消订阅的数据|数字| <br>
>}

---
命令ID: 0x010E<br>
命令描述: 取消订阅应答(UNSUB-ACK)<br>
协议格式:<br>
>{<br>
>   "sub":${sub},               // M|取消订阅的数据|数字|<br>
>   "errno":${errno},           // M|错误码|数字|<br>
>   "errmsg":"${errmsg}"        // M|错误描述|字串|<br>
>}

---
命令ID: 0x0110<br>
命令描述: 群聊消息(GROUP-MSG)<br>
协议格式: 透传<br>
TODO: 协议头中的to为群ID(GID)

---
命令ID: 0x0111<br>
命令描述: 群聊消息应答(GROUP-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0112<br>
命令描述: 私聊消息(PRVT-MSG)<br>
协议格式: 透传<br>
TODO: 协议头中的to为用户ID(UID)

---
命令ID: 0x0113<br>
命令描述: 私聊消息应答(PRVG-MSG-ACK)<br>
协议格式: NONE<br>

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
命令描述: 添加好友(ADD-FRIEND)<br>
协议格式:<br>
>{<br>
>   "mark":"${mark}"            // O|备注信息|字串|<br>
>}

---
命令ID: 0x0117<br>
命令描述: 添加好友应答(ADD-FRIEND-ACK)<br>
协议格式: NONE<br>
TODO: 协议头中的to为对方的用户ID(UID)

---
命令ID: 0x0118<br>
命令描述: 回复添加好友(REP-ADD-FRIEND)<br>
协议格式:<br>
>{<br>
>   "errno":${errno},           // M|错误码|数字|(同意/拒绝)|<br>
>   "errmsg":"${errmsg}",       // M|错误描述|字串|<br>
>   "mark":"${mark}"            // O|备注信息|字串|<br>
>}

---
命令ID: 0x0119<br>
命令描述: 回复添加好友应答(REP-ADD-FRIEND-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x011A<br>
命令描述: 删除好友(DEL-FRIEND)<br>
协议格式: NONE<br>
TODO: 协议头中的to为对方的用户ID(UID)

---
命令ID: 0x011B<br>
命令描述: 删除好友应答(UNSUB-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x011C<br>
命令描述: 聊天室消息(ROOM-MSG-ACK)<br>
协议格式: 透传<br>

---
命令ID: 0x011D<br>
命令描述: 聊天室消息应答(ROOM-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x011E<br>
命令描述: 聊天室广播消息(ROOM-BC-ACK)<br>
协议格式: 透传<br>

---
命令ID: 0x0120<br>
命令描述: 聊天室广播消息应答(ROOM-BC-MSG-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0121<br>
命令描述: 通用异常消息(UNUSUAL)<br>
协议格式:<br>
>{<br>
>   "errno":${errno},           // M|错误码|数字|<br>
>   "errmsg":"${errmsg}",       // M|错误描述|字串|<br>
>}

---
命令ID: 0x0122<br>
命令描述: 通用异常消息应答(UNUSUAL-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0123<br>
命令描述: 聊天室人数(ROOM-USR-NUM)<br>
协议格式:<br>
>{<br>
>   "rid":${rid},               // M|聊天室ID|数字|<br>
>   "usrnum":${usrnum}          // M|错误描述|字串|<br>
>}

---
命令ID: 0x0124<br>
命令描述: 聊天室人数应答(ROOM-USR-NUM-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0125<br>
命令描述: 同步消息(SYNC-MSG)<br>
协议格式: NONE<br>

---
命令ID: 0x0126<br>
命令描述: 同步消息应答(SYNC-MSG-ACK)<br>
协议格式: NONE<br>

# 通知类命令

---
命令ID: 0x0301<br>
命令描述: 上线通知(ONLINE-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|用户ID|数字|<br>
>}

---
命令ID: 0x0302<br>
命令描述: 下线通知(OFFLINE-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|用户ID|数字|<br>
>}

---
命令ID: 0x0303<br>
命令描述: 加入聊天室通知(JOIN-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|用户ID|数字|<br>
>}

---
命令ID: 0x0304<br>
命令描述: 退出聊天室通知(QUIT-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|用户ID|数字|<br>
>}

---
命令ID: 0x0305<br>
命令描述: 禁言通知(BAN-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|被禁言用户ID|数字|<br>
>}

---
命令ID: 0x0306<br>
命令描述: 踢人通知(KICK-NOTIFY)<br>
协议格式: <br>
>{<br>
>   "uid":${uid}                // M|被踢用户ID|数字|<br>
>}

# 系统内部命令

---
命令ID: 0x0401<br>
命令描述: 内部心跳(HB)<br>
协议格式: <br>
>{<br>
>   "nid":${nid},               // M|结点ID|数字|<br>
>   "mod":${mod},               // M|模块类型|数字|(1:接入层 2:转发层)<br>
>   "ipaddr":"${ipaddr}",       // M|IP地址|字串|<br>
>   "port":${port}              // M|端口号|数字|<br>
>}

---
命令ID: 0x0402<br>
命令描述: 内部心跳应答(HB-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0403<br>
命令描述: 帧听层上报(LSN-RPT)<br>
协议格式: <br>
>{<br>
>   "nid":${nid},               // M|结点ID|数字|<br>
>   "ipaddr":"${ipaddr}",       // M|IP地址|字串|<br>
>   "port":${port}              // M|端口号|数字|<br>
>}

---
命令ID: 0x0404<br>
命令描述: 帧听层上报应答(LSN-RPT-ACK)<br>
协议格式: NONE<br>

---
命令ID: 0x0405<br>
命令描述: 转发层列表(FRWD-LIST)<br>
协议格式: <br>
>{<br>
>   "len":${len},               // M|结点ID|数字|<br>
>   "list":[
>       {"ipaddr":"${ipaddr}", "port":${port}}, // M|IP+端口|<br>
>       {"ipaddr":"${ipaddr}", "port":${port}}, // M|IP+端口|<br>
>       {"ipaddr":"${ipaddr}", "port":${port}}] // M|IP+端口|<br>
>}

---
命令ID: 0x0406<br>
命令描述: 转发层列表应答(FRWD-LIST-ACK)<br>
协议格式: NONE<br>

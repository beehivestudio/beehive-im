# HTTP接口列表

##数据下发接口
###广播接口
**功能描述**: 用于向全员或某聊天室提交广播消息<br>
**接口类型**: POST<br>
**接口路径**: /chatroom/push?opt=broadcast&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为broadcast.(M)<br>
> rid: 聊天室ID # 当未指定rid时, 则为全员广播消息(O)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###点推接口
**功能描述**: 用于指定聊天室的某人下发消息<br>
**接口类型**: POST<br>
**接口路径**: /chatroom/push?opt=p2p&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为broadcast.(M)<br>
> rid: 聊天室ID # 当未指定rid时, 则为全员广播消息(O)<br>
> uid: 用户ID(M)<br>

**包体内容**: 下发的数据
**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

##配置接口
###踢人接口
**功能描述**: 用于将某人踢出聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=kick&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为kick.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###解除踢人接口
**功能描述**: 用于将某人踢出聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=unkick&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为unkick.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###禁言接口
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=ban&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###解除禁言接口
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=unban&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为unban.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###关闭聊天室接口
**功能描述**: 关闭聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=close&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为close.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###设置聊天室分组人数
**功能描述**: 设置聊天室分组人数<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=group-size&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-size.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是修改默认分组人数; 指明聊天室ID, 则是指明某聊天室的分组人数<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "rid":${rid},        // 聊天室ID(O)<br>
>  "size":${size},      // 分组人数(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###查询聊天室分组人数
**功能描述**: 查询聊天室分组人数<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=group-size&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-size.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是查询默认分组人数; 指明聊天室ID, 则是查询某聊天室的分组人数<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "rid":${rid},        // 聊天室ID(O)<br>
>  "size":${size},      // 分组人数(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###添加在线人数统计
**功能描述**: 添加在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=user-statis-add&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-add.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},   // 错误码(M)<br>
>   "errmsg":"${errmsg}"// 错误描述(M)<br>
>}

###删除在线人数统计
**功能描述**: 删除在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=user-statis-del&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-del.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},   // 错误码(M)<br>
>   "errmsg":"${errmsg}"// 错误描述(M)<br>
>}

###在线人数统计列表
**功能描述**: 在线人数统计列表<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=user-statis-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},   // 错误码(M)<br>
>   "len":${len},       // 列表长度(M)<br>
>   "list":[            // 精度列表(M)<br>
>       {"prec":"{prec}"}, // ${prec}:精度值
>       {"prec":"{prec}"}]
>   "errmsg":"${errmsg}"// 错误描述(M)<br>
>}

###查询在线人数统计
**功能描述**: 查询在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=user-statis&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis.(M)<br>
> prec: 时间精度(M). 如:300s, 600s, 1800s, 3600s(1h), 86400(1d), 1m, 1y<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "prec":"${prec}",       // 时间精度(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 走势列表(M)<br>
>      {"time":"${time}", "max":${max}, "min":${min}}, // ${time}:时间戳 ${max}:峰值 ${min}:底值<br>
>      {"time":"${time}", "max":${max}, "min":${min}},<br>
>      {"time":"${time}", "max":${max}, "min":${min}}]<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###聊天室TOP排行
**功能描述**: 查询各聊天室TOP排行<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=top-list&num=${num}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 排行列表(M)<br>
>      {"rid":${rid}, "total":${total}}, // ${rid}:聊天室ID ${total}:聊天室人数<br>
>      {"rid":${rid}, "total":${total}},<br>
>      {"rid":${rid}, "total":${total}},<br>
>      {"rid":${rid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###查询某聊天室分组列表
**功能描述**: 查询某聊天室分组列表<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=group-list&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-list.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "rid":${rid},           // 聊天室ID(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 分组列表(M)<br>
>      {"gid":${gid}, "total":${total}}, // ${gid}:分组ID ${total}:组人数<br>
>      {"gid":${gid}, "total":${total}},<br>
>      {"gid":${gid}, "total":${total}},<br>
>      {"gid":${gid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###查询人数分布
**功能描述**: 查询人数分布<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=user-dist<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-dist.(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 分组列表(M)<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###某用户在线状态
**功能描述**: 查询某用户在线状态<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=user-online<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-online.(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "uid":"${uid}",         // 用户ID(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 当前正登陆聊天室列表(M)<br>
>      {"rid":${rid}},     // ${rid}:聊天室ID<br>
>      {"rid":${rid}}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

##系统维护接口
###查询侦听层状态
**功能描述**: 查询侦听层状态<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=listen-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 分组列表(M)<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###添加侦听层结点
**功能描述**: 移除侦听层结点<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=listen-add&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-add.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的侦听层结点IP地址.(M)<br>
> port: 将被添加的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###移除侦听层结点
**功能描述**: 移除侦听层结点<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=listen-del&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-del.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的侦听层结点IP地址.(M)<br>
> port: 将被移除的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###查询转发层状态
**功能描述**: 查询转发层状态<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/query?opt=frwder-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "errno":${errno},       // 错误码(M)<br>
>   "len":${len},           // 列表长度(M)<br>
>   "list":[                // 分组列表(M)<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],<br>
>   "errmsg":"${errmsg}"    // 错误描述(M)<br>
>}

###添加转发层结点
**功能描述**: 添加转发层结点<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=frwder-add&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-add.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的转发层结点IP地址.(M)<br>
> port: 将被添加的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

###移除转发层结点
**功能描述**: 移除转发层结点<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=frwder-del&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-del.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的转发层结点IP地址.(M)<br>
> port: 将被移除的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "errno":${errno},    // 错误码(M)<br>
>  "errmsg":"${errmsg}" // 错误描述(M)<br>
>}<br>

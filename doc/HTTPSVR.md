# HTTP接口列表

##1. 登录注册<br>
###1.1 设备注册接口<br>
---
**功能描述**: 设备注册接口<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/register?uid=${uid}&nation=${nation}&city=${city}&town=${town}<br>
**参数描述**:<br>
> uid: 用户ID(M)<br>
> nation: 国家编号(M)<br>
> city: 地市编号(M)<br>
> town: 城镇编号(M)<br>

**返回结果**:<br>
>{<br>
>   "uid":"${uid}",         // 整型 | 用户UID(M)<br>
>   "sid":"${sid}",         // 整型 | 会话SID(M)<br>
>   "nation":"${nation}",   // 整型 | 国家编号(M)<br>
>   "city":"${city}",       // 整型 | 地市编号(M)<br>
>   "town":"${town}",       // 整型 | 城镇编号(M)<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###1.2 获取IPLIST接口<br>
---
**功能描述**: 获取IPLIST接口<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/iplist?type=${type}&uid=${uid}&sid=${sid}&clientip=${clientip}<br>
**参数描述**:<br>
> type: LSN类型(0:Unknown 1:TCP 2:WS)(M)<br>
> uid: 用户ID(M)<br>
> sid: 会话SID(M)<br>
> clientip: 客户端IP(M)<br>

**返回结果**:<br>
>{<br>
>   "uid":${uid},           // 整型 | 用户UID(M)<br>
>   "sid":${sid},           // 整型 | 会话SID(M)<br>
>   "type":${type},         // 整型 | LSN类型(0:UNKNOWN 1:TCP 2:WS)(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "iplist":[              // 数组 | IP列表<br>
>       "${ipaddr}:${port}",<br>
>       "${ipaddr}:${port}",<br>
>       "${ipaddr}:${port}"],<br>
>   "token":"${token}"      // 字串 | 鉴权token(M) # 格式:"uid:${uid}:ttl:${ttl}:sid:${sid}:end"<br>
>   "expire":${expire}      // 整型 | 有效时常(M) # 单位:秒<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##2. 消息推送<br>
###2.1 广播接口<br>
---
**功能描述**: 全员广播消息<br>
**当前状态**: 未实现<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=broadcast<br>
**参数描述**:<br>
> dim: 推送维度, 此时为broadcast.(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###2.2 群组推送<br>
---
**功能描述**: 群组广播消息<br>
**当前状态**: 未实现<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=group&gid=${gid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为group.(M)<br>
> gid: 群组ID(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "gid":${gid},        // 整型 | 群组ID(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###2.3 聊天室推送<br>
---
**功能描述**: 聊天室广播消息<br>
**当前状态**: 待测试<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=room&rid=${rid}&expire=${expire}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为room.(M)<br>
> rid: 聊天室ID(M)<br>
> expire: 过期时间(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "rid":${rid},        // 整型 | 聊天室ID(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###2.4 会话推送<br>
---
**功能描述**: 指定会话SID下发消息<br>
**当前状态**: 未实现<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=sid&sid=${sid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为sid.(M)<br>
> sid: 会话SID(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###2.5 用户推送接口<br>
---
**功能描述**: 指定给某人下发消息<br>
**当前状态**: 未实现<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=uid&uid=${uid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为uid.(M)<br>
> sid: 会话SID(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###2.6 应用推送接口<br>
---
**功能描述**: 指定给应用ID下发消息<br>
**当前状态**: 未实现<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=appid&appid=${appid}&version=${version}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为appid.(M)<br>
> appid: 应用ID(M)<br>
> version: 应用版本号(O)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

##3. 配置操作<br>
###3.1 添加在线人数统计<br>
---
**功能描述**: 添加在线人数统计<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=user-statis&prec=${prec}&num=${num}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为user-statis.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>
> num: 该精度的记录最大数(M).<br>

**返回结果**:<br>
>{<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###3.2 删除在线人数统计<br>
---
**功能描述**: 删除在线人数统计<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=user-statis&prec=${prec}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为user-statis.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>

**返回结果**:<br>
>{<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###3.3 在线人数统计列表<br>
---
**功能描述**: 在线人数统计列表<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=user-statis<br>
**参数描述**:<br>
> action: 操作行为, 此时为list.(M)<br>
> option: 操作选项, 此时为user-statis.(M)<br>

**返回结果**:<br>
>{<br>
>   "len":${len},       // 整型 | 列表长度(M)<br>
>   "list":[            // 数组 | 精度列表(M)<br>
>       {"idx":${idx}, "prec":{prec}, "num":${num}}, // ${idx}:序号 ${prec}:精度值 ${num}:最大记录数<br>
>       {"idx":${idx}, "prec":{prec}, "num":${num}}],<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###3.4 查询在线人数统计<br>
---
**功能描述**: 查询在线人数统计<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=get&option=user-statis&prec=${prec}&num=${num}<br>
**参数描述**:<br>
> action: 操作行为, 此时为get.(M)<br>
> option: 操作选项, 此时为user-statis.(M)<br>
> prec: 时间精度(M). 如:300s, 600s, 1800s, 3600s(1h), 86400(1d), 1m, 1y<br>
> num: 记录条数, 从请求时间往前取${num}条记录.(M)<br>

**返回结果**:<br>
>{<br>
>   "prec":"${prec}",       // 整型 | 时间精度(M)<br>
>   "num":${num},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 走势列表(M)<br>
>      {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}}, // ${time-str}:时间戳 ${num}:在线人数<br>
>      {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}},<br>
>      {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##4. 状态查询<br>
###4.1 某用户在线状态<br>
---
**功能描述**: 查询某用户在线状态<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/query?option=user-status&uid=${uid}<br>
**参数描述**:<br>
> option: 操作选项, 此时为user-status.(M)<br>
> uid: 用户UID.(M)<br>

**返回结果**:<br>
>{<br>
>   "uid":"${uid}",         // 整型 | 用户ID(M)<br>
>   "status":${status},     // 整型 | 当前状态(0:下线 1:在线)(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 当前正登陆聊天室列表(M)<br>
>      {"idx":${idx}, "rid":${rid}},     // ${rid}:聊天室ID<br>
>      {"idx":${idx}, "rid":${rid}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###4.2 某用户SID列表<br>
---
**功能描述**: 查询某用户SID列表<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/query?option=sid-list&uid=${uid}<br>
**参数描述**:<br>
> option: 操作选项, 此时为sid-list.(M)<br>
> uid: 用户UID.(M)<br>

**返回结果**:<br>
>{<br>
>   "uid":"${uid}",         // 整型 | 用户ID(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 当前正登陆设备列表(M)<br>
>      {"idx":${idx}, "sid":${sid}},     // ${sid}:会话ID<br>
>      {"idx":${idx}, "sid":${sid}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##5. 群组接口<br>
###5.1 加入群组黑名单<br>
---
**功能描述**: 将某人加入群组黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=add&option=blacklist&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为blacklist.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.2 移除群组黑名单<br>
---
**功能描述**: 将某人移除群组黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=del&option=blacklist&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为blacklist.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.3 获取群组黑名单<br>
---
**功能描述**: 获取群组用户黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=blacklist&gid=${gid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为get.(M)<br>
> option: 操作选项, 此时为blacklist.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 用户列表(M)<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}}],<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.4 添加群组禁言<br>
---
**功能描述**: 禁止某人在群内发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=add&option=gag&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为gag.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.5 解除群组禁言<br>
---
**功能描述**: 禁止某人在群组发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=del&option=gag&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为gag.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员解除禁言; 否则是解除某人禁言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.6 获取群组禁言<br>
---
**功能描述**: 获取群组禁言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=gag&gid=${gid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为get.(M)<br>
> option: 操作选项, 此时为gag.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 用户列表(M)<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}},<br>
>      {"idx":${idx}, "uid":${uid}}],<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.7 解散群组<br>
---
**功能描述**: 关闭聊天室<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=close&option=group&gid=${gid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为close.(M)<br>
> option: 操作选项, 此时为group.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.8 设置群组"最大容量"限制<br>
---
**功能描述**: 设置群组"最大容量"限制<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=set&option=cap&gid=${gid}&cap=${cap}<br>
**参数描述**:<br>
> action: 操作行为, 此时为set.(M)<br>
> option: 操作选项, 此时为cap.(M)<br>
> gid: 群组ID(M)<br>
> cap: 群组容量(M)<br>

**返回结果**:<br>
>{<br>
>  "gid":${gid},        // 整型 | 群组ID(O)<br>
>  "cap":${cap},        // 整型 | 分组容量(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.9 查询群组"最大容量"限制<br>
---
**功能描述**: 查询群组"最大容量"限制<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=cap&gid=${gid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为get.(M)<br>
> option: 操作选项, 此时为group-cap.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>  "gid":${gid},        // 整型 | 群组ID(O)<br>
>  "cap":${cap},      // 整型 | 群组容量(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.10 群组"人数"TOP排行
---
**功能描述**: 查询各群组TOP排行<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/group/query?option=top-list&num=${num}<br>
**参数描述**:<br>
> option: 操作选项, 此时为top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:整型|群组ID ${total}: 整型|群组人数<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###5.11 群组"消息量"TOP排行<br>
---
**功能描述**: 查询各群组消息量TOP排行<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/group/query?option=mesg-top-list&num=${num}<br>
**参数描述**:<br>
> option: 操作选项, 此时为mesg-top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:整型|群组ID ${total}: 整型|消息数量<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##6. 聊天室接口<br>
###6.1 加入聊天室黑名单<br>
---
**功能描述**: 将某人加入聊天室黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=add&option=blacklist&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为blacklist.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.2 移除聊天室黑名单<br>
---
**功能描述**: 将某人移除聊天室黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=del&option=blacklist&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为blacklist.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.3 聊天室禁言<br>
---
**功能描述**: 禁止某人在聊天室发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=add&option=gag&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为gag.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.4 聊天室解除禁言<br>
---
**功能描述**: 禁止某人在聊天室发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=del&option=gag&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为gag.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.5 关闭聊天室<br>
---
**功能描述**: 关闭聊天室<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=close&option=room&rid=${rid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为close.(M)<br>
> option: 操作选项, 此时为room.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.6 打开聊天室<br>
---
**功能描述**: 打开聊天室<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=open&option=room&rid=${rid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为open.(M)<br>
> option: 操作选项, 此时为room.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.7 设置聊天室分组容量<br>
---
**功能描述**: 设置聊天室分组容量<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=set&option=cap&rid=${rid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为set.(M)<br>
> option: 操作选项, 此时为cap.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是修改默认分组容量; 指明聊天室ID, 则是指明某聊天室的分组容量<br>

**返回结果**:<br>
>{<br>
>  "rid":${rid},        // 整型 | 聊天室ID(O)<br>
>  "cap":${cap},        // 整型 | 分组容量(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###6.8 查询聊天室分组容量<br>
---
**功能描述**: 查询聊天室分组容量<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=get&option=cap&rid=${rid}<br>
**参数描述**:<br>
> action: 操作行为, 此时为get.(M)<br>
> option: 操作选项, 此时为cap.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是查询默认分组容量; 指明聊天室ID, 则是查询某聊天室的分组容量<br>

**返回结果**:<br>
>{<br>
>  "rid":${rid},        // 整型 | 聊天室ID(O)<br>
>  "cap":${cap},        // 整型 | 分组容量(M)<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>


###6.9 聊天室TOP排行<br>
---
**功能描述**: 查询各聊天室TOP排行<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?option=top-list&num=${num}<br>
**参数描述**:<br>
> option: 操作选项, 此时为top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}}, // ${rid}:整型|聊天室ID ${total}: 整型|聊天室人数<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}},<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}},<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###6.10 查询某聊天室分组列表<br>
---
**功能描述**: 查询某聊天室分组列表<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?option=group-list&rid=${rid}<br>
**参数描述**:<br>
> option: 操作选项, 此时为group-list.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>   "rid":${rid},           // 整型 | 聊天室ID(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:分组ID ${total}:组人数<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##7. 系统维护接口<br>
###7.1 查询侦听层状态<br>
---
**功能描述**: 查询侦听层状态<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=listend<br>
**参数描述**:<br>
> action: 操作行为, 此时为list.(M)<br>
> option: 操作选项, 此时为listend.(M)<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###7.2 添加侦听层结点<br>
---
**功能描述**: 移除侦听层结点<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=listend&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为listend.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的侦听层结点IP地址.(M)<br>
> port: 将被添加的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###7.3 移除侦听层结点<br>
---
**功能描述**: 移除侦听层结点<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=listend&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为listend.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的侦听层结点IP地址.(M)<br>
> port: 将被移除的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###7.4 查询转发层状态<br>
---
**功能描述**: 查询转发层状态<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=frwder<br>
**参数描述**:<br>
> action: 操作行为, 此时为list.(M)<br>
> option: 操作选项, 此时为frwder.(M)<br>

**返回结果**:<br>
>{<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}
**补充说明**:<br>
> idx: 整型 | 序列号<br>
> nid: 整型 | 结点ID<br>
> ipaddr: 字串 | IP地址<br>
> port: 整型 | 端口号<br>
> status: 整型 | 状态<br>

###7.5 添加转发层结点<br>
---
**功能描述**: 添加转发层结点<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=frwder&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> action: 操作行为, 此时为add.(M)<br>
> option: 操作选项, 此时为frwder.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的转发层结点IP地址.(M)<br>
> port: 将被添加的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###7.6 移除转发层结点<br>
---
**功能描述**: 移除转发层结点<br>
**当前状态**: 未实现<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=frwder&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> action: 操作行为, 此时为del.(M)<br>
> option: 操作选项, 此时为frwder.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的转发层结点IP地址.(M)<br>
> port: 将被移除的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

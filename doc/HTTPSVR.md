# HTTP接口列表

##1. 推送接口
###1.1 广播接口
---
**功能描述**: 全员广播消息<br>
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

###1.2 群组广播接口
---
**功能描述**: 群组广播消息<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=group&gid=${gid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为group.(M)<br>
> gid: 群组ID(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###1.3 聊天室广播接口
---
**功能描述**: 聊天室广播消息<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=room&rid=${rid}&expire=${expire}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为room.(M)<br>
> rid: 聊天室ID(M)<br>
> expire: 过期时间(M)<br>

**包体内容**: 下发的数据<br>
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###1.4 设备推送接口
---
**功能描述**: 指定给某设备下发消息<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=sid&sid=${sid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为sid.(M)<br>
> sid: 会话SID(M)<br>

**包体内容**: 下发的数据
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###1.5 用户推送接口
---
**功能描述**: 指定给某人下发消息<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=uid&uid=${uid}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为uid.(M)<br>
> sid: 会话SID(M)<br>

**包体内容**: 下发的数据
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###1.6 应用推送接口
---
**功能描述**: 指定给应用ID下发消息<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=appid&appid=${appid}&version=${version}<br>
**参数描述**:<br>
> dim: 推送维度, 此时为appid.(M)<br>
> appid: 应用ID(M)<br>
> version: 应用版本号(O)<br>

**包体内容**: 下发的数据
**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

##2. 配置接口
###2.1 设备注册接口
---
**功能描述**: 设备注册接口<br>
**接口类型**: GET<br>
**接口路径**: /im/register?uid=${uid}&nation=${nation}&city=${city}&town=${town}<br>
**参数描述**:<br>
> uid: 用户ID(M)<br>
> nation: 国家编号(M)<br>
> city: 地市编号(M)<br>
> town: 城镇编号(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "uid":"${uid}",         // 整型 | 用户UID(M)<br>
>   "sid":"${sid}",         // 整型 | 会话SID(M)<br>
>   "nation":"${nation}",   // 整型 | 国家编号(M)<br>
>   "city":"${city}",       // 整型 | 地市编号(M)<br>
>   "town":"${town}",       // 整型 | 城镇编号(M)<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###2.2 获取IPLIST接口
---
**功能描述**: 获取IPLIST接口<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=iplist&uid=${uid}&sid=${sid}&clientip=${clientip}<br>
**参数描述**:<br>
> opt: 固定为iplist(M)<br>
> uid: 用户ID(M)<br>
> sid: 会话SID(M)<br>
> clientip: 客户端IP(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "iplist":[              // 数组 | IP列表<br>
>       "${ipaddr}:${port}",<br>
>       "${ipaddr}:${port}",<br>
>       "${ipaddr}:${port}"],<br>
>   "token":"${token}"      // 字串 | 鉴权token(M) # 格式:"uid:${uid}:ttl:${ttl}:sid:${sid}:end"<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###2.3 添加在线人数统计
---
**功能描述**: 添加在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=user-statis-add&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-add.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>

**返回结果**:<br>
>{<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###2.4 删除在线人数统计
---
**功能描述**: 删除在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=user-statis-del&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-del.(M)<br>
> prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc<br>

**返回结果**:<br>
>{<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###2.5 在线人数统计列表
---
**功能描述**: 在线人数统计列表<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=user-statis-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},     // 整型 | 错误码(M)<br>
>   "len":${len},       // 整型 | 列表长度(M)<br>
>   "list":[            // 数组 | 精度列表(M)<br>
>       {"idx":${idx}, "prec":"{prec}"}, // ${idx}:序号 ${prec}:精度值<br>
>       {"idx":${idx}, "prec":"{prec}"}],<br>
>   "errmsg":"${errmsg}"// 字串 | 错误描述(M)<br>
>}

###2.6 查询在线人数统计
---
**功能描述**: 查询在线人数统计<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=user-statis&prec=${prec}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-statis.(M)<br>
> prec: 时间精度(M). 如:300s, 600s, 1800s, 3600s(1h), 86400(1d), 1m, 1y<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "prec":"${prec}",       // 整型 | 时间精度(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 走势列表(M)<br>
>      {"idx":${idx}, "time":"${time}", "max":${max}, "min":${min}}, // ${time}:时间戳 ${max}:峰值 ${min}:底值<br>
>      {"idx":${idx}, "time":"${time}", "max":${max}, "min":${min}},<br>
>      {"idx":${idx}, "time":"${time}", "max":${max}, "min":${min}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###2.7 查询人数分布
---
**功能描述**: 查询人数分布<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=user-dist<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-dist.(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "total":"${total}"}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###2.8 某用户在线状态
---
**功能描述**: 查询某用户在线状态<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=user-online<br>
**参数描述**:<br>
> opt: 操作选项, 此时为user-online.(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "uid":"${uid}",         // 整型 | 用户ID(M)<br>
>   "status":${status},     // 整型 | 当前状态(0:下线 1:在线)(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 当前正登陆聊天室列表(M)<br>
>      {"idx":${idx}, "rid":${rid}},     // ${rid}:聊天室ID<br>
>      {"idx":${idx}, "rid":${rid}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##3. 群组接口
###3.1 加入群组黑名单
---
**功能描述**: 将某人加入群组黑名单<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=blacklist-add&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为blacklist-add.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.2 移除群组黑名单
---
**功能描述**: 将某人移除群组黑名单<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=blacklist-del&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为blacklist-del.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.3 群组禁言
---
**功能描述**: 禁止某人在群内发言<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=ban-add&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban-add.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.4 群组解除禁言
---
**功能描述**: 禁止某人在群组发言<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=ban-del&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban-del.(M)<br>
> gid: 群组ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员解除禁言; 否则是解除某人禁言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.5 解散群组
---
**功能描述**: 关闭聊天室<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=close&gid=${gid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为close.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.6 设置群组"最大容量"限制
---
**功能描述**: 设置群组"最大容量"限制<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?opt=capacity&gid=${gid}&capacity=${capacity}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为capacity.(M)<br>
> gid: 群组ID(M)<br>
> capacity: 群组容量(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "gid":${gid},        // 整型 | 群组ID(O)<br>
>  "capacity":${capacity},      // 整型 | 分组容量(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.7 查询群组"最大容量"限制
---
**功能描述**: 查询群组"最大容量"限制<br>
**接口类型**: GET<br>
**接口路径**: /im/group/query?opt=capacity&gid=${gid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-capacity.(M)<br>
> gid: 群组ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "gid":${gid},        // 整型 | 群组ID(O)<br>
>  "capacity":${capacity},      // 整型 | 群组容量(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###3.8 群组"人数"TOP排行
---
**功能描述**: 查询各群组TOP排行<br>
**接口类型**: GET<br>
**接口路径**: /im/group/query?opt=top-list&num=${num}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:整型|群组ID ${total}: 整型|群组人数<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###3.9 群组"消息量"TOP排行
---
**功能描述**: 查询各群组消息量TOP排行<br>
**接口类型**: GET<br>
**接口路径**: /im/group/query?opt=mesg-top-list&num=${num}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为mesg-top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:整型|群组ID ${total}: 整型|消息数量<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##4. 聊天室接口
###4.1 加入聊天室黑名单
---
**功能描述**: 将某人加入聊天室黑名单<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=blacklist-add&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为blacklist-add.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.2 移除聊天室黑名单
---
**功能描述**: 将某人移除聊天室黑名单<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=blacklist-del&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为blacklist-del.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.3 聊天室禁言
---
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=ban-add&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban-add.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.4 聊天室解除禁言
---
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=ban-del&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban-del.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.5 聊天室关闭
---
**功能描述**: 关闭聊天室<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=close&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为close.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.6 设置聊天室分组容量
---
**功能描述**: 设置聊天室分组容量<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?opt=group-capacity&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-capacity.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是修改默认分组容量; 指明聊天室ID, 则是指明某聊天室的分组容量<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "rid":${rid},        // 整型 | 聊天室ID(O)<br>
>  "capacity":${capacity},      // 整型 | 分组容量(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###4.7 查询聊天室分组容量
---
**功能描述**: 查询聊天室分组容量<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?opt=group-capacity&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-capacity.(M)<br>
> rid: 聊天室ID(O).当未制定${rid}时, 则是查询默认分组容量; 指明聊天室ID, 则是查询某聊天室的分组容量<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "rid":${rid},        // 整型 | 聊天室ID(O)<br>
>  "capacity":${capacity},      // 整型 | 分组容量(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>


###4.8 聊天室TOP排行
---
**功能描述**: 查询各聊天室TOP排行<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?opt=top-list&num=${num}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为top-list.(M)<br>
> num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 排行列表(M)<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}}, // ${rid}:整型|聊天室ID ${total}: 整型|聊天室人数<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}},<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}},<br>
>      {"idx":${idx}, "rid":${rid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###4.9 查询某聊天室分组列表
---
**功能描述**: 查询某聊天室分组列表<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?opt=group-list&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为group-list.(M)<br>
> rid: 聊天室ID(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "rid":${rid},           // 整型 | 聊天室ID(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:分组ID ${total}:组人数<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}},<br>
>      {"idx":${idx}, "gid":${gid}, "total":${total}}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

##5. 系统维护接口
###5.1 查询侦听层状态
---
**功能描述**: 查询侦听层状态<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=listen-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}

###5.2 添加侦听层结点
---
**功能描述**: 移除侦听层结点<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=listen-add&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-add.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的侦听层结点IP地址.(M)<br>
> port: 将被添加的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.3 移除侦听层结点
---
**功能描述**: 移除侦听层结点<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=listen-del&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为listen-del.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的侦听层结点IP地址.(M)<br>
> port: 将被移除的侦听层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.4 查询转发层状态
---
**功能描述**: 查询转发层状态<br>
**接口类型**: GET<br>
**接口路径**: /im/query?opt=frwder-list<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-list.(M)<br>

**返回结果**:<br>
>{<br>
>   "code":${code},         // 整型 | 错误码(M)<br>
>   "len":${len},           // 整型 | 列表长度(M)<br>
>   "list":[                // 数组 | 分组列表(M)<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},<br>
>      {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],<br>
>   "errmsg":"${errmsg}"    // 字串 | 错误描述(M)<br>
>}
**补充说明**:<br>
> idx: 整型 | 序列号<br>
> nid: 整型 | 结点ID<br>
> ipaddr: 字串 | IP地址<br>
> port: 整型 | 端口号<br>
> status: 整型 | 状态<br>

###5.5 添加转发层结点
---
**功能描述**: 添加转发层结点<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=frwder-add&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-add.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被添加的转发层结点IP地址.(M)<br>
> port: 将被添加的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

###5.6 移除转发层结点
---
**功能描述**: 移除转发层结点<br>
**接口类型**: GET<br>
**接口路径**: /im/config?opt=frwder-del&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为frwder-del.(M)<br>
> nid: 结点ID(M)<br>
> ipaddr: 将被移除的转发层结点IP地址.(M)<br>
> port: 将被移除的转发层结点端口.(M)<br>

**返回结果**:<br>
>{<br>
>  "code":${code},      // 整型 | 错误码(M)<br>
>  "errmsg":"${errmsg}" // 字串 | 错误描述(M)<br>
>}<br>

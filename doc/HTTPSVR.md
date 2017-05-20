# HTTP接口列表

## 1. 登录注册<br>
### 1.1 设备注册接口<br>
---
**功能描述**: 设备注册接口<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/register?uid=${uid}&nation=${nation}&city=${city}&town=${town}<br>
**参数描述**:<br>
```
  uid: 用户ID(M)
  nation: 国家编号(M)
  city: 地市编号(M)
  town: 城镇编号(M)
```
**返回结果**:<br>
```
{
    "uid":"${uid}",         // 整型 | 用户UID(M)
    "sid":"${sid}",         // 整型 | 会话SID(M)
    "nation":"${nation}",   // 整型 | 国家编号(M)
    "city":"${city}",       // 整型 | 地市编号(M)
    "town":"${town}",       // 整型 | 城镇编号(M)
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

### 1.2 获取IPLIST接口<br>
---
**功能描述**: 获取IPLIST接口<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/iplist?type=${type}&uid=${uid}&sid=${sid}&clientip=${clientip}<br>
**参数描述**:<br>
```
  type: LSN类型(0:Unknown 1:TCP 2:WS)(M)
  uid: 用户ID(M)
  sid: 会话SID(M)
  clientip: 客户端IP(M)
```
**返回结果**:<br>
```
{
    "uid":${uid},           // 整型 | 用户UID(M)
    "sid":${sid},           // 整型 | 会话SID(M)
    "type":${type},         // 整型 | LSN类型(0:UNKNOWN 1:TCP 2:WS)(M)
    "len":${len},           // 整型 | 列表长度(M)
    "iplist":[              // 数组 | IP列表
        "${ipaddr}:${port}",
        "${ipaddr}:${port}",
        "${ipaddr}:${port}"],
    "token":"${token}"      // 字串 | 鉴权token(M) # 格式:"uid:${uid}:ttl:${ttl}:sid:${sid}:end"
    "expire":${expire}      // 整型 | 有效时常(M) # 单位:秒
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

## 2. 消息推送<br>
### 2.1 广播接口<br>
---
**功能描述**: 全员广播消息<br>
**当前状态**: 未完成<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=broadcast<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为broadcast.(M)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 2.2 群组推送<br>
---
**功能描述**: 群组广播消息<br>
**当前状态**: 未完成<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=group&gid=${gid}<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为group.(M)
  gid: 群组ID(M)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "gid":${gid},        // 整型 | 群组ID(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 2.3 聊天室推送<br>
---
**功能描述**: 聊天室广播消息<br>
**当前状态**: Ok<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=room&rid=${rid}&expire=${expire}<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为room.(M)
  rid: 聊天室ID(M)
  expire: 过期时间(M)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "rid":${rid},        // 整型 | 聊天室ID(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 2.4 会话推送<br>
---
**功能描述**: 指定会话SID下发消息<br>
**当前状态**: 未完成<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=sid&sid=${sid}<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为sid.(M)
  sid: 会话SID(M)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 2.5 用户推送接口<br>
---
**功能描述**: 指定给某人下发消息<br>
**当前状态**: 未完成<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=uid&uid=${uid}<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为uid.(M)
  sid: 会话SID(M)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 2.6 应用推送接口<br>
---
**功能描述**: 指定给应用ID下发消息<br>
**当前状态**: 未完成<br>
**接口类型**: POST<br>
**接口路径**: /im/push?dim=appid&appid=${appid}&version=${version}<br>
**参数描述**:<br>
```
  dim: 推送维度, 此时为appid.(M)
  appid: 应用ID(M)
  version: 应用版本号(O)
```
**包体内容**: 下发的数据<br>
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

## 3. 配置操作<br>
### 3.1 添加在线人数统计<br>
---
**功能描述**: 添加在线人数统计<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=user-statis&prec=${prec}&num=${num}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为user-statis.(M)
  prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc
  num: 该精度的记录最大数(M).
```
**返回结果**:<br>
```
{
    "code":${code},     // 整型 | 错误码(M)
    "errmsg":"${errmsg}"// 字串 | 错误描述(M)
}
```

### 3.2 删除在线人数统计<br>
---
**功能描述**: 删除在线人数统计<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=user-statis&prec=${prec}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为user-statis.(M)
  prec: 时间精度(M).可以有:300s, 600s, 1800s, 3600s(1h), 86400(1d), etc.
```
**返回结果**:<br>
```
{
    "code":${code},     // 整型 | 错误码(M)
    "errmsg":"${errmsg}"// 字串 | 错误描述(M)
}
```

### 3.3 在线人数统计列表<br>
---
**功能描述**: 在线人数统计列表<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=user-statis<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为list.(M)
  option: 操作选项, 此时为user-statis.(M)
```
**返回结果**:<br>
```
{
    "len":${len},       // 整型 | 列表长度(M)
    "list":[            // 数组 | 精度列表(M)
        {"idx":${idx}, "prec":{prec}, "num":${num}}, // ${idx}:序号 ${prec}:精度值 ${num}:最大记录数
        {"idx":${idx}, "prec":{prec}, "num":${num}}],
    "code":${code},     // 整型 | 错误码(M)
    "errmsg":"${errmsg}"// 字串 | 错误描述(M)
}
```

### 3.4 查询在线人数统计<br>
---
**功能描述**: 查询在线人数统计<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=get&option=user-statis&prec=${prec}&num=${num}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为get.(M)
  option: 操作选项, 此时为user-statis.(M)
  prec: 时间精度(M). 如:300s, 600s, 1800s, 3600s(1h), 86400(1d), 1m, 1y
  num: 记录条数, 从请求时间往前取${num}条记录.(M)
```
**返回结果**:<br>
```
{
    "prec":"${prec}",       // 整型 | 时间精度(M)
    "num":${num},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 走势列表(M)
       {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}}, // ${time-str}:时间戳 ${num}:在线人数
       {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}},
       {"idx":${idx}, "time":${time}, "time-str":"${time-str}", "num":${num}}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

## 4. 状态查询<br>
### 4.1 某用户SID列表<br>
---
**功能描述**: 查询某用户SID列表<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/query?option=sid-list&uid=${uid}<br>
**参数描述**:<br>
```
  option: 操作选项, 此时为sid-list.(M)
  uid: 用户UID.(M)
```
**返回结果**:<br>
```
{
    "uid":"${uid}",         // 整型 | 用户ID(M)
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 当前正登陆设备列表(M)
       {"idx":${idx}, "sid":${sid}},     // ${sid}:会话ID
       {"idx":${idx}, "sid":${sid}}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

## 5. 群组接口<br>
### 5.1 加入群组黑名单<br>
---
**功能描述**: 将某人加入群组黑名单<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=add&option=blacklist&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为blacklist.(M)
  gid: 群组ID(M)
  uid: 用户ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.2 移除群组黑名单<br>
---
**功能描述**: 将某人移除群组黑名单<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=del&option=blacklist&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为blacklist.(M)
  gid: 群组ID(M)
  uid: 用户ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.3 获取群组黑名单<br>
---
**功能描述**: 获取群组用户黑名单<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=blacklist&gid=${gid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为get.(M)
  option: 操作选项, 此时为blacklist.(M)
  gid: 群组ID(M)
```
**返回结果**:<br>
```
{
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 用户列表(M)
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}}],
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.4 添加群组禁言<br>
---
**功能描述**: 禁止某人在群内发言<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=add&option=gag&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
```

  action: 操作行为, 此时为add.(M)<br>
  option: 操作选项, 此时为gag.(M)<br>
  gid: 群组ID(M)<br>
  uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>
```


**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.5 解除群组禁言<br>
---
**功能描述**: 禁止某人在群组发言<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=del&option=gag&gid=${gid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为gag.(M)
  gid: 群组ID(M)
  uid: 用户ID. # 当无uid或uid为0时, 全员解除禁言; 否则是解除某人禁言.
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.6 获取群组禁言<br>
---
**功能描述**: 获取群组禁言<br>
**当前状态**: Ok<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=gag&gid=${gid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为get.(M)
  option: 操作选项, 此时为gag.(M)
  gid: 群组ID(M)
```
**返回结果**:<br>
```
{
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 用户列表(M)
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}},
       {"idx":${idx}, "uid":${uid}}],
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.7 群组开关<br>
---
**功能描述**: 解散群组<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=${action}&option=switch&gid=${gid}<br>
**参数描述**:<br>
```
  action: 操作行为[on:打开 off:关闭](M)
  option: 操作选项, 此时为switch.(M)
  gid: 群组ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.8 设置群组"最大容量"限制<br>
---
**功能描述**: 设置群组"最大容量"限制<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=set&option=cap&gid=${gid}&cap=${cap}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为set.(M)
  option: 操作选项, 此时为cap.(M)
  gid: 群组ID(M)
  cap: 群组容量(M)
```
**返回结果**:<br>
```
{
   "gid":${gid},        // 整型 | 群组ID(O)
   "cap":${cap},        // 整型 | 分组容量(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 5.9 查询群组"最大容量"限制<br>
---
**功能描述**: 查询群组"最大容量"限制<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/group/config?action=get&option=cap&gid=${gid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为get.(M)
  option: 操作选项, 此时为group-cap.(M)
  gid: 群组ID(M)
```
**返回结果**:<br>
```
{
   "gid":${gid},        // 整型 | 群组ID(O)
   "cap":${cap},        // 整型 | 群组容量(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

## 6. 聊天室接口<br>
### 6.1 加入聊天室黑名单<br>
---
**功能描述**: 将某人加入聊天室黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=add&option=blacklist&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为blacklist.(M)
  rid: 聊天室ID(M)
  uid: 用户ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.2 移除聊天室黑名单<br>
---
**功能描述**: 将某人移除聊天室黑名单<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=del&option=blacklist&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为blacklist.(M)
  rid: 聊天室ID(M)
  uid: 用户ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.3 聊天室禁言<br>
---
**功能描述**: 禁止某人在聊天室发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=add&option=gag&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为gag.(M)
  rid: 聊天室ID(M)
  uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.4 聊天室解除禁言<br>
---
**功能描述**: 禁止某人在聊天室发言<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=del&option=gag&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为gag.(M)
  rid: 聊天室ID(M)
  uid: 用户ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.5 关闭聊天室<br>
---
**功能描述**: 关闭聊天室<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=close&option=room&rid=${rid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为close.(M)
  option: 操作选项, 此时为room.(M)
  rid: 聊天室ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.6 打开聊天室<br>
---
**功能描述**: 打开聊天室<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=open&option=room&rid=${rid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为open.(M)
  option: 操作选项, 此时为room.(M)
  rid: 聊天室ID(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.7 设置聊天室分组容量<br>
---
**功能描述**: 设置聊天室分组容量<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=set&option=cap&rid=${rid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为set.(M)
  option: 操作选项, 此时为cap.(M)
  rid: 聊天室ID(O).当未制定${rid}时, 则是修改默认分组容量; 指明聊天室ID, 则是指明某聊天室的分组容量
```
**返回结果**:<br>
```
{
   "rid":${rid},        // 整型 | 聊天室ID(O)
   "cap":${cap},        // 整型 | 分组容量(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.8 查询聊天室分组容量<br>
---
**功能描述**: 查询聊天室分组容量<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/config?action=get&option=cap&rid=${rid}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为get.(M)
  option: 操作选项, 此时为cap.(M)
  rid: 聊天室ID(O).当未制定${rid}时, 则是查询默认分组容量; 指明聊天室ID, 则是查询某聊天室的分组容量
```
**返回结果**:<br>
```
{
   "rid":${rid},        // 整型 | 聊天室ID(O)
   "cap":${cap},        // 整型 | 分组容量(M)
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 6.9 聊天室TOP排行<br>
---
**功能描述**: 查询各聊天室TOP排行<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?option=top-list&num=${num}<br>
**参数描述**:<br>
```
  option: 操作选项, 此时为top-list.(M)
  num: top-${num}排行(O). 如果未设置${num}, 则显示前top-10的排行.
```
**返回结果**:<br>
```
{
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 排行列表(M)
       {"idx":${idx}, "rid":${rid}, "total":${total}}, // ${rid}:整型|聊天室ID ${total}: 整型|聊天室人数
       {"idx":${idx}, "rid":${rid}, "total":${total}},
       {"idx":${idx}, "rid":${rid}, "total":${total}},
       {"idx":${idx}, "rid":${rid}, "total":${total}}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

### 6.10 查询某聊天室分组列表<br>
---
**功能描述**: 查询某聊天室分组列表<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/room/query?option=group-list&rid=${rid}<br>
**参数描述**:<br>
```
  option: 操作选项, 此时为group-list.(M)
  rid: 聊天室ID(M)
```
**返回结果**:<br>
```
{
    "rid":${rid},           // 整型 | 聊天室ID(M)
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 分组列表(M)
       {"idx":${idx}, "gid":${gid}, "total":${total}}, // ${gid}:分组ID ${total}:组人数
       {"idx":${idx}, "gid":${gid}, "total":${total}},
       {"idx":${idx}, "gid":${gid}, "total":${total}},
       {"idx":${idx}, "gid":${gid}, "total":${total}}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

## 7. 系统维护接口<br>
### 7.1 查询侦听层状态<br>
---
**功能描述**: 查询侦听层状态<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=listend<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为list.(M)
  option: 操作选项, 此时为listend.(M)
```
**返回结果**:<br>
```
{
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 分组列表(M)
       {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},
       {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},
       {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"},
       {"idx":${idx}, "nid":${nid}, "type":${type}, "ipaddr":"{ipaddr}", "status":${status}, "total":"${total}"}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```

### 7.2 添加侦听层结点<br>
---
**功能描述**: 移除侦听层结点<br>
**当前状态**: 未完成<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=listend&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为listend.(M)
  nid: 结点ID(M)
  ipaddr: 将被添加的侦听层结点IP地址.(M)
  port: 将被添加的侦听层结点端口.(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 7.3 移除侦听层结点<br>
---
**功能描述**: 移除侦听层结点<br>
**当前状态**: 未完成<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=listend&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为listend.(M)
  nid: 结点ID(M)
  ipaddr: 将被移除的侦听层结点IP地址.(M)
  port: 将被移除的侦听层结点端口.(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 7.4 查询转发层状态<br>
---
**功能描述**: 查询转发层状态<br>
**当前状态**: 待测试<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=list&option=frwder<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为list.(M)
  option: 操作选项, 此时为frwder.(M)
```
**返回结果**:<br>
```
{
    "len":${len},           // 整型 | 列表长度(M)
    "list":[                // 数组 | 分组列表(M)
       {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},
       {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},
       {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"},
       {"idx":${idx}, "nid":${nid}, "ipaddr":"{ipaddr}", "port":${port}, "status":"${status}"}],
    "code":${code},         // 整型 | 错误码(M)
    "errmsg":"${errmsg}"    // 字串 | 错误描述(M)
}
```
**补充说明**:<br>
  idx: 整型 | 序列号<br>
  nid: 整型 | 结点ID<br>
  ipaddr: 字串 | IP地址<br>
  port: 整型 | 端口号<br>
  status: 整型 | 状态<br>

### 7.5 添加转发层结点<br>
---
**功能描述**: 添加转发层结点<br>
**当前状态**: 未完成<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=add&option=frwder&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为add.(M)
  option: 操作选项, 此时为frwder.(M)
  nid: 结点ID(M)
  ipaddr: 将被添加的转发层结点IP地址.(M)
  port: 将被添加的转发层结点端口.(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

### 7.6 移除转发层结点<br>
---
**功能描述**: 移除转发层结点<br>
**当前状态**: 未完成<br>
**接口类型**: GET<br>
**接口路径**: /im/config?action=del&option=frwder&nid=${nid}&ipaddr=${ipaddr}&port=${port}<br>
**参数描述**:<br>
```
  action: 操作行为, 此时为del.(M)
  option: 操作选项, 此时为frwder.(M)
  nid: 结点ID(M)
  ipaddr: 将被移除的转发层结点IP地址.(M)
  port: 将被移除的转发层结点端口.(M)
```
**返回结果**:<br>
```
{
   "code":${code},      // 整型 | 错误码(M)
   "errmsg":"${errmsg}" // 字串 | 错误描述(M)
}
```

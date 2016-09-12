# 接口列表

##广播接口
**功能描述**: 用于向全员或某聊天室提交广播消息<br>
**接口类型**: POST<br>
**接口路径**: /chatroom/push?opt=broadcast&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为broadcast.(M)<br>
> rid: 聊天室ID # 当未指定rid时, 则为全员广播消息(O)<br>
**包体内容**: 下发的数据

##点推接口
**功能描述**: 用于指定聊天室的某人下发消息<br>
**接口类型**: POST<br>
**接口路径**: /chatroom/push?opt=p2p&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为broadcast.(M)<br>
> rid: 聊天室ID # 当未指定rid时, 则为全员广播消息(O)<br>
> uid: 用户ID(M)<br>
**包体内容**: 下发的数据

##踢人接口
**功能描述**: 用于将某人踢出聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=kick&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为kick.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

##解除踢人接口
**功能描述**: 用于将某人踢出聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=unkick&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为unkick.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

##禁言接口
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=ban&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为ban.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID. # 当无uid或uid为0时, 全员禁言; 否则是禁止某人发言.<br>

##解除禁言接口
**功能描述**: 禁止某人在聊天室发言<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=unban&rid=${rid}&uid=${uid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为unban.(M)<br>
> rid: 聊天室ID(M)<br>
> uid: 用户ID(M)<br>

##关闭聊天室接口
**功能描述**: 关闭聊天室<br>
**接口类型**: GET<br>
**接口路径**: /chatroom/config?opt=close&rid=${rid}<br>
**参数描述**:<br>
> opt: 操作选项, 此时为close.(M)<br>
> rid: 聊天室ID(M)<br>

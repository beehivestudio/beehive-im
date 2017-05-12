# 错误码说明

## 错误返回值格式
>{<br>
>  "errno" : 10001,<br>
>  "errmsg" : "System error"<br>
>}

## 错误码说明

20503

|服务级错误(1为系统级错误) | 服务模块代码 | 具体错误码 |
|:------------------------:|:------------:|:----------:|
| 2 | 05 | 03 |

## 错误码对照表

系统级错误码

| 错误码 | 错误信息 | 详细描述 |
|:------:|:---------|:---------|
| 10001 | System error | 系统错误 |
| 10002 | Service unavailable | 服务暂停 |
| 10003 | Remote service error | 远程服务错误 |
| 10004 | IP limit | IP限制不能请求该资源 |
| 10005 | Permission denied, need a high level appkey | 该资源需要appkey拥有授权 |
| 10006 | Source paramter (appkey) is missing | 缺少source (appkey) 参数 |
| 10007 | Unsupport mediatype (%s) | 不支持的MediaType (%s) |
| 10008 | Param error, see doc for more info | 参数错误, 请参考API文档 |
| 10009 | Too many pending tasks, system is busy | 任务过多, 系统繁忙 |
| 10010 | Job expired | 任务超时 |
| 10011 | RPC error | RPC错误 |
| 10012 | Illegal request | 非法请求 |
| 10013 | Invalid user | 不合法的用户 |
| 10014 | Insufficient app permissions | 应用的接口访问权限受限 |
| 10016 | Miss required parameter (%s) , see doc for more info | 缺失必选参数 (%s), 请参考API文档 |
| 10017 | Parameter (%s)'s value invalid, expect (%s) , but get (%s) , see doc for more info | 参数值非法, 需为 (%s), 实际为 (%s), 请参考API文档 |
| 10018 | Request body length over limit | 请求长度超过限制 |
| 10020 | Request api not found | 接口不存在 |
| 10021 | HTTP method is not suported for this request | 请求的HTTP METHOD不支持, 请检查是否选择了正确的POST/GET方式 |
| 10022 | IP requests out of rate limit | IP请求频次超过上限 |
| 10023 | User requests out of rate limit | 用户请求频次超过上限 |
| 10024 | User requests for (%s) out of rate limit | 用户请求特殊接口 (%s) 频次超过上限 |

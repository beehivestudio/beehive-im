package comm

const (
	OK = 0 // 正常
)

/*
 *错误码说明
 *20503
 *|服务级错误(1为系统级错误) | 服务模块代码 | 具体错误码 |
 *|:------------------------:|:------------:|:----------:|
 *| 2 | 05 | 03 |
 */
// 系统级错误码
const (
	ERR_SYS_SYSTEM                    = 10001 // System error | 系统错误 |
	ERR_SYS_SVC_UNAVAILABLE           = 10002 // Service unavailable | 服务暂停 |
	ERR_SYS_REMOTE_SERVICE            = 10003 // Remote service error | 远程服务错误 |
	ERR_SYS_IP_LIMIT                  = 10004 // IP limit | IP限制不能请求该资源 |
	ERR_SYS_PERM_DENIED               = 10005 // Permission denied, need a high level appkey | 该资源需要appkey拥有授权 |
	ERR_SYS_MISS_PARAM                = 10006 // Source paramter (appkey) is missing | 缺少source (appkey) 参数 |
	ERR_SYS_UNSUPPORT                 = 10007 // Unsupport mediatype (%s) | 不支持的MediaType (%s) |
	ERR_SYS_TOO_BUSY                  = 10009 // Too many pending tasks, system is busy | 任务过多, 系统繁忙 |
	ERR_SYS_JOB_EXIRED                = 10010 // Job expired | 任务超时 |
	ERR_SYS_RPC                       = 10011 // RPC error | RPC错误 |
	ERR_SYS_ILLEGAL_REQUEST           = 10012 // Illegal request | 非法请求 |
	ERR_SYS_INVALID_USER              = 10013 // Invalid user | 不合法的用户 |
	ERR_SYS_INVALID_PARAM             = 10017 // Parameter (%s)'s value invalid, expect (%s) , but get (%s) , see doc for more info | 参数值非法, 需为 (%s), 实际为 (%s), 请参考API文档 |
	ERR_SYS_BODY_OVER_LIMIT           = 10018 // Request body length over limit | 请求长度超过限制 |
	ERR_SYS_REQ_API_NOT_FOUND         = 10020 // Request api not found | 接口不存在 |
	ERR_SYS_HTTP_NOT_SUPPORT          = 10021 // HTTP method is not suported for this request | 请求的HTTP METHOD不支持, 请检查是否选择了正确的POST/GET方式 |
	ERR_SYS_IP_REQ_OUT_OF_RATE_LIMIT  = 10022 // IP requests out of rate limit | IP请求频次超过上限 |
	ERR_SYS_USR_REQ_OUT_OF_RATE_LIMIT = 10023 // User requests out of rate limit | 用户请求频次超过上限 |
	ERR_SYS_DB                        = 10024 // Database error | 数据库错误 |
)

// 业务级错误码
const (
	ERR_SVR_ONLINE_REQ     = 20001 // Online request isn't right! | ONLINE请求有误 |
	ERR_SVR_OFFLINE_REQ    = 20002 // Offline request isn't right! | OFFLINE请求有误 |
	ERR_SVR_JOIN_REQ       = 20003 // Join request isn't right! | JOIN请求有误 |
	ERR_SVR_UNJOIN_REQ     = 20004 // Unjoin request isn't right! | UNJOIN请求有误 |
	ERR_SVR_PARSE_PARAM    = 20005 // Parse paramter | 解析参数错误 |
	ERR_SVR_MISS_PARAM     = 20006 // Miss paramter | 缺失参数 |
	ERR_SVR_INVALID_PARAM  = 20007 // Invalid paramter | 非法参数 |
	ERR_SVR_AUTH_FAIL      = 20008 // Auth failed | 鉴权失败 |
	ERR_SVR_DATA_COLLISION = 20009 // Data collision | 数据冲突 |
	ERR_SVR_HEAD_INVALID   = 20010 // Head invalid | 头部不合法 |
	ERR_SVR_BODY_INVALID   = 20011 // Body invalid| 报体不合法 |
	ERR_SVR_CHECK_FAIL     = 20012 // Check invalid| 校验失败 |
	ERR_SVR_SEQ_EXHAUSTION = 20013 // Seqence exhaustion | 序列号耗尽 |
)

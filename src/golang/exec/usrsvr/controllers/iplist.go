package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/crypt"
)

type UsrSvrIplistCtrl struct {
	BaseController
}

func (this *UsrSvrIplistCtrl) Iplist() {
	ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	case "iplist": // 获取IP列表
		this.query_iplist(ctx)
		return
	}

	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s.", option))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/* 注册参数 */
type UsrSvrIpListParam struct {
	network  int    // 网络类型(0:Unknown 1:TCP 2:WS)
	uid      uint64 // 用户ID
	sid      uint64 // 会话ID
	clientip string // 客户端IP
}

func (this *UsrSvrIplistCtrl) query_iplist(ctx *UsrSvrCntx) {
	/* > 提取注册参数 */
	param, err := this.iplist_parse_param(ctx)
	if nil != err {
		ctx.log.Error("Parse param failed! uid:%d sid:%d clientip:%s",
			param.uid, param.sid, param.clientip)
		this.Error(comm.ERR_SVR_PARSE_PARAM, err.Error())
		return
	}

	ctx.log.Debug("Param list. uid:%d sid:%d clientip:%s", param.uid, param.sid, param.clientip)

	/* > 获取IP列表 */
	this.iplist_handler(ctx, param)

	return
}

/******************************************************************************
 **函数名称: iplist_parse_param
 **功    能: 解析参数
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回:
 **     param: 注册参数
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 23:17:52 #
 ******************************************************************************/
func (this *UsrSvrIplistCtrl) iplist_parse_param(ctx *UsrSvrCntx) (*UsrSvrIpListParam, error) {
	var param UsrSvrIpListParam

	/* > 提取注册参数 */
	param.network, _ = this.GetInt("network")

	id, _ := this.GetInt64("uid")
	param.uid = uint64(id)

	id, _ = this.GetInt64("sid")
	param.sid = uint64(id)

	param.clientip = this.GetString("clientip")

	/* > 校验参数合法性 */
	if 0 == param.network || 0 == param.uid ||
		0 == param.sid || "" == param.clientip {
		ctx.log.Error("Paramter is invalid. uid:%d sid:%d clientip:%s",
			param.uid, param.sid, param.clientip)
		return &param, errors.New("Paramter is invalid!")
	}

	ctx.log.Debug("Get ip list param. uid:%d sid:%d clientip:%s",
		param.uid, param.sid, param.clientip)

	return &param, nil
}

/* IP列表应答 */
type UsrSvrIpListRsp struct {
	Uid     uint64   `json:"uid"`     // 用户ID
	Sid     uint64   `json:"sid"`     // 会话ID
	Network int      `json:"network"` // 网络类型(0:Unknown 1:TCP 2:WS)
	Token   string   `json:"token"`   // 鉴权TOKEN
	Expire  int      `json:"expire"`  // 过期时长
	Len     int      `json:"len"`     // 列表长度
	List    []string `json:"list"`    // IP列表
	Code    int      `json:"code"`    // 错误码
	ErrMsg  string   `json:"errmsg"`  // 错误描述
}

/******************************************************************************
 **函数名称: iplist_handler
 **功    能: 获取IP列表
 **输入参数:
 **     ctx: 上下文
 **     param: URL参数
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.24 17:00:07 #
 ******************************************************************************/
func (this *UsrSvrIplistCtrl) iplist_handler(ctx *UsrSvrCntx, param *UsrSvrIpListParam) {
	iplist := this.iplist_get(ctx, param.network, param.clientip)
	if nil == iplist {
		ctx.log.Error("Get ip list failed!")
		this.Error(comm.ERR_SYS_SYSTEM, "Get ip list failed!")
		return
	}

	this.success(param, iplist)

	return
}

/******************************************************************************
 **函数名称: success
 **功    能: 应答处理成功
 **输入参数:
 **     param: 注册参数
 **     iplist: IP列表
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.25 23:13:02 #
 ******************************************************************************/
func (this *UsrSvrIplistCtrl) success(param *UsrSvrIpListParam, iplist []string) {
	var resp UsrSvrIpListRsp

	resp.Uid = param.uid
	resp.Sid = param.sid
	resp.Network = param.network
	resp.Token = this.iplist_token(param)
	resp.Expire = comm.TIME_DAY
	resp.Len = len(iplist)
	resp.List = iplist
	resp.Code = 0
	resp.ErrMsg = "OK"

	this.Data["json"] = &resp
	this.ServeJSON()
}

/******************************************************************************
 **函数名称: iplist_token
 **功    能: 生成TOKEN字串
 **输入参数:
 **输出参数: NONE
 **返    回:
 **     token: 加密TOKEN
 **     expire: 过期时间
 **实现描述:
 **注意事项:
 **     TOKEN的格式"uid:${uid}:ttl:${ttl}:sid:${sid}:end"
 **     uid: 用户ID
 **     ttl: 该token的最大生命时间
 **     sid: 会话UID
 **作    者: # Qifeng.zou # 2016.11.25 23:54:27 #
 ******************************************************************************/
func (this *UsrSvrIplistCtrl) iplist_token(param *UsrSvrIpListParam) string {
	ctx := GetUsrSvrCtx()
	ttl := time.Now().Unix() + comm.TIME_DAY

	/* > 原始TOKEN */
	token := fmt.Sprintf("uid:%d:ttl:%d:sid:%d:end", param.uid, ttl, param.sid)

	/* > 加密TOKEN */
	cry := crypt.CreateEncodeCtx(ctx.conf.Cipher)

	return crypt.Encode(cry, token)
}

/******************************************************************************
 **函数名称: iplist_get
 **功    能: 获取IP列表
 **输入参数:
 **     ctx: 上下文
 **     network: 网络类型(0:Unknown 1:TCP 2:WS)
 **     clientip: 客户端IP
 **输出参数: NONE
 **返    回: IP列表
 **实现描述: 首先根据网络类型, 再根据运营商类型筛选.
 **注意事项: 加读锁
 **作    者: # Qifeng.zou # 2016.11.27 07:42:54 #
 ******************************************************************************/
func (this *UsrSvrIplistCtrl) iplist_get(ctx *UsrSvrCntx, network int, clientip string) []string {
	ctx.listend.RLock()
	defer ctx.listend.RUnlock()

	listend, ok := ctx.listend.network[network]
	if !ok || 0 == network {
		return nil
	}

	item := ctx.ipdict.Query(clientip)
	if nil == item {
		listend.RLock()
		defer listend.RUnlock()
		return listend.get_default(ctx)
	}

	listend.RLock()
	defer listend.RUnlock()

	/* > 获取国家/地区下辖的运营商列表 */
	operators, ok := listend.list[item.GetNation()]
	if nil == operators || !ok {
		return listend.get_default(ctx)
	}

	/* > 获取运营商下辖的侦听层列表 */
	list, ok := operators[item.GetOperator()]
	if nil == list || !ok {
		return listend.get_default(ctx)
	}

	items := make([]string, 0)
	items = append(items, list[rand.Intn(len(list))])

	return items
}

/******************************************************************************
 **函数名称: get_default
 **功    能: 获取默认IP列表
 **输入参数:
 **     ctx: 上下文
 **     network: 网络类型(0:Unknown 1:TCP 2:WS)
 **输出参数: NONE
 **返    回: IP列表
 **实现描述:
 **注意事项: 外部已经加读锁
 **作    者: # Qifeng.zou # 2016.11.27 19:33:49 #
 ******************************************************************************/
func (listend *UsrSvrLsndList) get_default(ctx *UsrSvrCntx) []string {
	var ok bool

	/* > 获取"默认"国家/地区下辖的运营商列表 */
	operators, ok := listend.list["CN"]
	if nil == operators || 0 == len(operators) || !ok {
		ctx.log.Error("Get default iplist by nation failed!")
		return nil
	}

	for k, v := range operators {
		ctx.log.Debug("k:%s v:%s", k, v)
	}

	/* > 获取"默认"运营商下辖的侦听层列表 */
	list, ok := operators["中国电信"]
	if nil == list || 0 == len(list) || !ok {
		ctx.log.Error("Get default iplist by operator failed!")
		return nil
	}

	items := make([]string, 0)
	items = append(items, list[rand.Intn(len(list))])

	return items
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
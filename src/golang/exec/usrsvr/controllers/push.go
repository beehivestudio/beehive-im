package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

/* 推送接口 */
type UsrSvrPushCtrl struct {
	BaseController
}

func (this *UsrSvrPushCtrl) Push() {
	ctx := GetUsrSvrCtx()

	opt := this.GetString("dim")
	switch opt {
	case "room": // 聊天室广播
		this.room_broadcast(ctx)
		return
	}

	errmsg := fmt.Sprintf("Unsupport this option:%s.", opt)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
}

////////////////////////////////////////////////////////////////////////////////

/* 聊天室广播参数 */
type UsrSvrRoomBcParam struct {
	rid    uint64 // 聊天室ID
	expire uint32 // 超时时间
}

func (this *UsrSvrPushCtrl) room_broadcast(ctx *UsrSvrCntx) {
	/* > 提取广播参数 */
	param, err := this.room_broadcast_param_parse(ctx)
	if nil != err {
		ctx.log.Error("Parse room broadcast param failed! rid:%d", param.rid)
		this.Error(comm.ERR_SVR_PARSE_PARAM, err.Error())
		return
	}

	/* > 聊天室广播处理 */
	code, err := this.room_broadcast_handler(ctx, param)
	if nil != err {
		this.Error(code, err.Error())
		return
	}

	this.room_broadcast_success(param)

	return
}

/******************************************************************************
 **函数名称: room_broadcast_param_parse
 **功    能: 解析参数
 **输入参数:
 **     ctx: 上下文
 **输出参数: NONE
 **返    回:
 **     param: 注册参数
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 11:10:59 #
 ******************************************************************************/
func (this *UsrSvrPushCtrl) room_broadcast_param_parse(
	ctx *UsrSvrCntx) (param *UsrSvrRoomBcParam, err error) {
	param = &UsrSvrRoomBcParam{}

	/* > 提取注册参数 */
	rid, _ := this.GetInt64("rid")
	param.rid = uint64(rid)

	expire, _ := this.GetInt32("expire")
	param.expire = uint32(expire)

	/* > 校验参数合法性 */
	if 0 == param.rid {
		ctx.log.Error("Paramter is invalid. rid:%d", param.rid)
		return param, errors.New("Paramter is invalid!")
	}

	ctx.log.Debug("Get room broadcast param. rid:%d", param.rid)

	return param, nil
}

/* 聊天室广播应答 */
type UsrSvrRoomBcRsp struct {
	Rid    uint64 `json:"rid"`    // 聊天室ID
	Code   int    `json:"code"`   // 错误码
	ErrMsg string `json:"errmsg"` // 错误描述
}

/******************************************************************************
 **函数名称: room_broadcast_handler
 **功    能: 聊天室广播请求的处理
 **输入参数:
 **     ctx: 上下文
 **     param: URL参数
 **输出参数:
 **返    回:
 **     code: 错误码
 **     err: 错误信息
 **实现描述:
 **     {
 **         required uint64 rid = 1;        // M|聊天室ID
 **         required uint32 level = 2;      // M|消息级别
 **         required uint64 time = 3;       // M|发送时间
 **         required uint32 expire = 4;     // M|过期时间
 **         required bytes data = 5;        // M|透传数据
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 12:27:53 #
 ******************************************************************************/
func (this *UsrSvrPushCtrl) room_broadcast_handler(
	ctx *UsrSvrCntx, param *UsrSvrRoomBcParam) (code int, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	ctm := time.Now().Unix()
	ttl := ctm + int64(param.expire)
	data := this.Ctx.Input.RequestBody

	/* > 申请聊天室消息序列号  */
	key := fmt.Sprintf(comm.CHAT_KEY_ROOM_MSGID_INCR, param.rid)

	msgid_str, err := redis.String(rds.Do("ZINCRBY", key, 1))
	if nil != err {
		ctx.log.Error("Get room msgid failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	msgid, _ := strconv.ParseInt(msgid_str, 10, 64)

	/* > 生成PB数据 */
	rsp := &mesg.MesgRoomBc{
		Rid:    proto.Uint64(param.rid),    // 聊天室ID
		Level:  proto.Uint32(0),            // 优先级别
		Time:   proto.Uint64(uint64(ctm)),  // 发送时间
		Expire: proto.Uint32(param.expire), // 超时时间
		Data:   []byte(data),               // 透传内容
	}

	body, err := proto.Marshal(rsp)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	/* > 放入聊天室广播集合 */
	key = fmt.Sprintf(comm.CHAT_KEY_ROOM_BC_HASH, param.rid)
	pl.Send("HSETNX", key, msgid, body)

	key = fmt.Sprintf(comm.CHAT_KEY_ROOM_BC_ZSET, param.rid)
	pl.Send("ZADD", key, ttl, msgid)

	/* > 申请内存空间 */
	length := len(body)
	p := &comm.MesgPacket{}

	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	/* > 获取侦听层集合 */
	nid_list, err := redis.Ints(rds.Do("ZRANGEBYSCORE", comm.IM_KEY_LSN_NID_ZSET, ctm, "+inf"))
	if nil != err {
		ctx.log.Error("Get listen nid failed! errmsg:%s", err.Error())
		return comm.ERR_SYS_SYSTEM, err
	}

	num := len(nid_list)

	for idx := 0; idx < num; idx += 1 {
		head := &comm.MesgHeader{}

		head.Cmd = comm.CMD_ROOM_BC
		head.Sid = param.rid // 会话ID改为聊天室ID
		head.Length = uint32(length)
		head.Nid = uint32(nid_list[idx])
		head.ChkSum = comm.MSG_CHKSUM_VAL

		comm.MesgHeadHton(head, p)
		copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

		/* > 发送协议包 */
		ctx.frwder.AsyncSend(comm.CMD_ROOM_BC, p.Buff, uint32(len(p.Buff)))
	}

	return 0, nil
}

/******************************************************************************
 **函数名称: room_broadcast_success
 **功    能: 聊天室广播处理成功
 **输入参数:
 **     param: 请求参数
 **输出参数:
 **返    回: NONE
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.14 16:00:42 #
 ******************************************************************************/
func (this *UsrSvrPushCtrl) room_broadcast_success(param *UsrSvrRoomBcParam) {
	var resp UsrSvrRoomBcRsp

	resp.Rid = param.rid
	resp.Code = 0
	resp.ErrMsg = "OK"

	this.Data["json"] = &resp
	this.ServeJSON()
}

////////////////////////////////////////////////////////////////////////////////

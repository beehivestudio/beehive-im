package controllers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/chat"
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

// 群聊处理

////////////////////////////////////////////////////////////////////////////////
/* 创建群组 */

/******************************************************************************
 **函数名称: group_creat_parse
 **功    能: 解析GROUP-CREAT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:32:20 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) group_creat_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgGroupCreat, code uint32, err error) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)
	if !head.IsValid(1) {
		errmsg := "Header of group-create failed!"
		ctx.log.Error("Header is invalid! cmd:0x%04X nid:%d chksum:0x%08X",
			head.GetCmd(), head.GetNid(), head.GetChkSum())
		return nil, nil, comm.ERR_SVR_HEAD_INVALID, errors.New(errmsg)
	}

	/* > 解析PB协议 */
	req = &mesg.MesgGroupCreat{}
	err = proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal group-creat request failed! errmsg:%s", err.Error())
		return head, nil, comm.ERR_SVR_HEAD_INVALID, err
	}

	return head, req, 0, nil
}

/******************************************************************************
 **函数名称: group_creat_failed
 **功    能: 发送GROUP-CREAT应答(异常)
 **输入参数:
 **     head: 协议头
 **     req: GROUP-CREAT请求
 **     code: 错误码
 **     errmsg: 错误描述
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1;   // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:33:42 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) group_creat_failed(
	head *comm.MesgHeader, req *mesg.MesgGroupCreat, code uint32, err error) int {
	if nil == head {
		return -1
	}

	/* > 设置协议体 */
	ack := &mesg.MesgGroupCreatAck{
		Code:   proto.Uint32(code),
		Errmsg: proto.String(err.Error()),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_GROUP_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GROUP_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: group_creat_ack
 **功    能: 发送GROUP-CREAT应答
 **输入参数:
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **应答协议:
 **     {
 **         required uint32 code = 1;   // M|错误码|数字|
 **         required string errmsg = 2; // M|错误描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:28:40 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) group_creat_ack(
	head *comm.MesgHeader, req *mesg.MesgGroupCreat, rid uint64) int {
	/* > 设置协议体 */
	ack := &mesg.MesgGroupCreatAck{
		Code:   proto.Uint32(0),
		Errmsg: proto.String("Ok"),
	}

	/* 生成PB数据 */
	body, err := proto.Marshal(ack)
	if nil != err {
		ctx.log.Error("Marshal protobuf failed! errmsg:%s", err.Error())
		return -1
	}

	length := len(body)

	/* > 拼接协议包 */
	p := &comm.MesgPacket{}
	p.Buff = make([]byte, comm.MESG_HEAD_SIZE+length)

	head.Cmd = comm.CMD_GROUP_CREAT_ACK
	head.Length = uint32(length)

	comm.MesgHeadHton(head, p)
	copy(p.Buff[comm.MESG_HEAD_SIZE:], body)

	/* > 发送协议包 */
	ctx.frwder.AsyncSend(comm.CMD_GROUP_CREAT_ACK, p.Buff, uint32(len(p.Buff)))

	return 0
}

/******************************************************************************
 **函数名称: alloc_gid
 **功    能: 申请群组ID
 **输入参数: NONE
 **输出参数: NONE
 **返    回:
 **     gid: 群组ID
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:24:23 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) alloc_gid() (rid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	/* > 申请群组ID */
	gid_str, err := redis.String(rds.Do("INCR", comm.CHAT_KEY_GID_INCR))
	if nil != err {
		return 0, err
	}

	gid_int, _ := strconv.ParseInt(gid_str, 10, 64)

	return uint64(gid_int), nil
}

/******************************************************************************
 **函数名称: group_creat_handler
 **功    能: GROUP-CREAT处理
 **输入参数:
 **     head: 协议头
 **     req: GROUP-CREAT请求
 **输出参数: NONE
 **返    回:
 **     gid: 群组ID
 **     err: 错误描述
 **实现描述:
 **注意事项: 已验证了GROUP-CREAT请求的合法性
 **作    者: # Qifeng.zou # 2017.01.19 22:22:50 #
 ******************************************************************************/
func (ctx *UsrSvrCntx) group_creat_handler(
	head *comm.MesgHeader, req *mesg.MesgGroupCreat) (gid uint64, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 分配群组ID */
	gid, err = ctx.alloc_gid()
	if nil != err {
		ctx.log.Error("Alloc gid failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 设置群组所有者 */
	key := fmt.Sprintf(comm.CHAT_KEY_GROUP_ROLE_TAB, gid)

	ok, err := redis.Int(rds.Do("HSETNX", key, req.GetUid(), chat.GROUP_ROLE_OWNER))
	if nil != err {
		ctx.log.Error("Set room owner failed! uid:%d errmsg:%s",
			req.GetUid(), err.Error())
		return 0, err
	} else if 0 == ok {
		ctx.log.Error("Set room owner failed! errmsg:%s", err.Error())
		return 0, err
	}

	/* > 设置群组信息 */
	key = fmt.Sprintf(comm.CHAT_KEY_GROUP_INFO_TAB, gid)

	pl.Send("HMSET", key, "NAME", req.GetName(), "DESC", req.GetDesc())

	return gid, nil
}

/******************************************************************************
 **函数名称: UsrSvrGroupCreatHandler
 **功    能: 创建群组
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **请求协议:
 **     {
 **        required uint64 uid = 1;    // M|用户ID|数字|
 **        required string name = 2;   // M|群组名称|字串|
 **        required string desc = 3;   // M|群组描述|字串|
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2017.01.19 22:21:48 #
 ******************************************************************************/
func UsrSvrGroupCreatHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*UsrSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv join request! cmd:0x%04X nid:%d length:%d", cmd, nid, length)

	/* > 解析创建请求 */
	head, req, code, err := ctx.group_creat_parse(data)
	if nil == req {
		ctx.log.Error("Parse room-creat request failed!")
		ctx.group_creat_failed(head, req, code, err)
		return -1
	}

	/* > 创建群组处理 */
	rid, err := ctx.group_creat_handler(head, req)
	if nil != err {
		ctx.log.Error("Group creat handler failed!")
		ctx.group_creat_failed(head, req, comm.ERR_SYS_SYSTEM, err)
		return -1
	}

	/* > 发送应答 */
	ctx.group_creat_ack(head, req, rid)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
/* 解散群组 */
func UsrSvrGroupDismissHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 申请入群 */
func UsrSvrGroupJoinHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 退群 */
func UsrSvrGroupQuitHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 邀请入群 */
func UsrSvrGroupInviteHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 群组踢人 */
func UsrSvrGroupKickHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 群组禁言 */
func UsrSvrGroupGagAddHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 解除群组禁言 */
func UsrSvrGroupGagDelHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 加入群组黑名单 */
func UsrSvrGroupBlacklistAddHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 移除群组黑名单 */
func UsrSvrGroupBlacklistDelHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 添加群组管理员 */
func UsrSvrGroupMgrAddHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 移除群组管理员 */
func UsrSvrGroupMgrDelHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

/* 群组成员列表 */
func UsrSvrGroupUsrListHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	return 0
}

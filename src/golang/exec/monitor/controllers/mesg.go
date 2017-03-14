package controllers

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"

	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/mesg"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: lsn_rpt_isvalid
 **功    能: 判断LSN-RPT是否合法
 **输入参数:
 **     req: HB请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 08:38:48 #
 ******************************************************************************/
func (ctx *MonSvrCntx) lsn_rpt_isvalid(req *mesg.MesgLsnRpt) bool {
	if 0 == req.GetNid() ||
		0 == req.GetPort() ||
		0 == len(req.GetNation()) ||
		0 == len(req.GetName()) ||
		0 == len(req.GetIpaddr()) {
		return false
	}
	return true
}

/******************************************************************************
 **函数名称: lsn_rpt_parse
 **功    能: 解析LSN-PRT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 08:14:10 #
 ******************************************************************************/
func (ctx *MonSvrCntx) lsn_rpt_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgLsnRpt) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgLsnRpt{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal lsn-rpt failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */
	if !ctx.lsn_rpt_isvalid(req) {
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: lsn_rpt_has_conflict
 **功    能: 判断数据是否冲突
 **输入参数:
 **     req: 帧听层上报消息
 **输出参数: NONE
 **返    回: true:存在冲突 false:不存在冲突
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 10:34:00 #
 ******************************************************************************/
func (ctx *MonSvrCntx) lsn_rpt_has_conflict(req *mesg.MesgLsnRpt) (has bool, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	addr := fmt.Sprintf("%s:%d", req.GetIpaddr(), req.GetPort())
	ok, err := redis.Bool(rds.Do("HEXISTS", comm.IM_KEY_LSN_ADDR_TO_NID, addr))
	if nil != err {
		ctx.log.Error("Exec hexists failed! err:%s", err.Error())
		return false, err
	} else if true == ok {
		nid, err := redis.Int(rds.Do("HGET", comm.IM_KEY_LSN_ADDR_TO_NID, addr))
		if nil != err {
			ctx.log.Error("Exec hget failed! err:%s", err.Error())
			return false, err
		} else if uint32(nid) != req.GetNid() {
			ctx.log.Error("Node id conflict! nid:%d/%d", nid, req.GetNid())
			return true, nil
		}
	}

	ok, err = redis.Bool(rds.Do("HEXISTS", comm.IM_KEY_LSN_NID_TO_ADDR, req.GetNid()))
	if nil != err {
		ctx.log.Error("Exec hexists failed! err:%s", err.Error())
		return
	} else if true == ok {
		_addr, err := redis.String(rds.Do("HGET", comm.IM_KEY_LSN_NID_TO_ADDR, req.GetNid()))
		if nil != err {
			ctx.log.Error("Exec hget failed! err:%s", err.Error())
			return false, err
		} else if _addr != addr {
			ctx.log.Error("Node id conflict! addr:%s/%s", addr, _addr)
			return true, nil
		}
	}

	return false, nil
}

/******************************************************************************
 **函数名称: lsn_rpt_handler
 **功    能: LSN-RPT处理
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 08:41:18 #
 ******************************************************************************/
func (ctx *MonSvrCntx) lsn_rpt_handler(head *comm.MesgHeader, req *mesg.MesgLsnRpt) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 判断数据是否冲突 */
	has, err := ctx.lsn_rpt_has_conflict(req)
	if nil != err {
		ctx.log.Error("Something was wrong! errmsg:%s", err.Error())
		return
	} else if true == has {
		ctx.log.Error("Data has conflict!")
		return
	}

	ttl := time.Now().Unix() + comm.CHAT_OP_TTL

	addr := fmt.Sprintf("%s:%d", req.GetIpaddr(), req.GetPort())
	pl.Send("HSETNX", comm.IM_KEY_LSN_NID_TO_ADDR, req.GetNid(), addr)
	pl.Send("HSETNX", comm.IM_KEY_LSN_ADDR_TO_NID, addr, req.GetNid())

	/* 网络类型结合 */
	pl.Send("ZADD", comm.IM_KEY_LSND_NETWORK_ZSET, ttl, req.GetNetwork())

	/* 国家集合 */
	key := fmt.Sprintf(comm.IM_KEY_LSND_NATION_ZSET, req.GetNetwork())
	pl.Send("ZADD", key, ttl, req.GetNation())

	/* 国家 -> 运营商列表 */
	key = fmt.Sprintf(comm.IM_KEY_LSND_OP_ZSET, req.GetNetwork(), req.GetNation())
	pl.Send("ZADD", key, ttl, req.GetName())

	/* 国家+运营商 -> 结点列表 */
	key = fmt.Sprintf(comm.IM_KEY_LSND_OP_TO_NID_ZSET, req.GetNetwork(), req.GetNation(), req.GetName())
	pl.Send("ZADD", key, ttl, req.GetNid())

	/* 国家+运营商 -> 侦听层IP列表 */
	key = fmt.Sprintf(comm.IM_KEY_LSND_IP_ZSET, req.GetNetwork(), req.GetNation(), req.GetName())
	val := fmt.Sprintf("%s:%d", req.GetIpaddr(), req.GetPort())
	pl.Send("ZADD", key, ttl, val)

	return
}

/******************************************************************************
 **函数名称: MonLsnRptHandler
 **功    能: 帧听层上报
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **协议格式:
 **     {
 **        required uint64 nid = 1;    // M|结点ID|数字|<br>
 **        required string nation = 2; // M|所属国家|字串|<br>
 **        required string name = 3;   // M|运营商名称|字串|<br>
 **        required string ipaddr = 4; // M|IP地址|字串|<br>
 **        required uint32 port = 5;   // M|端口号|数字|<br>
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 06:32:03 #
 ******************************************************************************/
func MonLsnRptHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MonSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv lsn-rpt request!")

	/* 1. > 解析LSN-RPT请求 */
	head, req := ctx.lsn_rpt_parse(data)
	if nil == head || nil == req {
		ctx.log.Error("Parse lsn-rpt failed!")
		return -1
	}

	/* 2. > LSN-RPT请求处理 */
	ctx.lsn_rpt_handler(head, req)

	return 0
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: frwd_rpt_isvalid
 **功    能: 判断LSN-RPT是否合法
 **输入参数:
 **     req: HB请求
 **输出参数: NONE
 **返    回: true:合法 false:非法
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 11:05:03 #
 ******************************************************************************/
func (ctx *MonSvrCntx) frwd_rpt_isvalid(req *mesg.MesgFrwdRpt) bool {
	if 0 == req.GetForwardPort() ||
		0 == req.GetBackendPort() ||
		0 == len(req.GetIpaddr()) {
		return false
	}
	return true
}

/******************************************************************************
 **函数名称: frwd_rpt_parse
 **功    能: 解析LSN-PRT请求
 **输入参数:
 **     data: 接收的数据
 **输出参数: NONE
 **返    回:
 **     head: 通用协议头
 **     req: 协议体内容
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 11:04:57 #
 ******************************************************************************/
func (ctx *MonSvrCntx) frwd_rpt_parse(data []byte) (
	head *comm.MesgHeader, req *mesg.MesgFrwdRpt) {
	/* > 字节序转换 */
	head = comm.MesgHeadNtoh(data)

	/* > 解析PB协议 */
	req = &mesg.MesgFrwdRpt{}
	err := proto.Unmarshal(data[comm.MESG_HEAD_SIZE:], req)
	if nil != err {
		ctx.log.Error("Unmarshal lsn-rpt failed! errmsg:%s", err.Error())
		return nil, nil
	}

	/* > 校验协议合法性 */
	if !ctx.frwd_rpt_isvalid(req) {
		return nil, nil
	}

	return head, req
}

/******************************************************************************
 **函数名称: frwd_rpt_has_conflict
 **功    能: 判断数据是否冲突
 **输入参数:
 **     req: 帧听层上报消息
 **输出参数: NONE
 **返    回: true:存在冲突 false:不存在冲突
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 11:04:49 #
 ******************************************************************************/
func (ctx *MonSvrCntx) frwd_rpt_has_conflict(req *mesg.MesgFrwdRpt) (has bool, err error) {
	rds := ctx.redis.Get()
	defer rds.Close()

	addr := fmt.Sprintf("%s:%d:%d", req.GetIpaddr(), req.GetForwardPort(), req.GetBackendPort())
	ok, err := redis.Bool(rds.Do("HEXISTS", comm.IM_KEY_FRWD_ADDR_TO_NID, addr))
	if nil != err {
		ctx.log.Error("Exec hexists failed! err:%s", err.Error())
		return false, err
	} else if true == ok {
		nid, err := redis.Int(rds.Do("HGET", comm.IM_KEY_FRWD_ADDR_TO_NID, addr))
		if nil != err {
			ctx.log.Error("Exec hget failed! err:%s", err.Error())
			return false, err
		} else if uint32(nid) != req.GetNid() {
			ctx.log.Error("Node id conflict! nid:%d/%d", nid, req.GetNid())
			return true, nil
		}
	}

	ok, err = redis.Bool(rds.Do("HEXISTS", comm.IM_KEY_LSN_NID_TO_ADDR, req.GetNid()))
	if nil != err {
		ctx.log.Error("Exec hexists failed! err:%s", err.Error())
		return
	} else if true == ok {
		_addr, err := redis.String(rds.Do("HGET", comm.IM_KEY_LSN_NID_TO_ADDR, req.GetNid()))
		if nil != err {
			ctx.log.Error("Exec hget failed! err:%s", err.Error())
			return false, err
		} else if _addr != addr {
			ctx.log.Error("Node id conflict! addr:%s/%s", addr, _addr)
			return true, nil
		}
	}

	return false, nil
}

/******************************************************************************
 **函数名称: frwd_rpt_handler
 **功    能: FRWD-RPT处理
 **输入参数:
 **     head: 协议头
 **     req: 上线请求
 **输出参数: NONE
 **返    回: 异常信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 11:04:42 #
 ******************************************************************************/
func (ctx *MonSvrCntx) frwd_rpt_handler(head *comm.MesgHeader, req *mesg.MesgFrwdRpt) {
	pl := ctx.redis.Get()
	defer func() {
		pl.Do("")
		pl.Close()
	}()

	/* > 判断数据是否冲突 */
	has, err := ctx.frwd_rpt_has_conflict(req)
	if nil != err {
		ctx.log.Error("Something was wrong! errmsg:%s", err.Error())
		return
	} else if true == has {
		ctx.log.Error("Data has conflict!")
		return
	}

	addr := fmt.Sprintf("%s:%d:%d", req.GetIpaddr(), req.GetForwardPort(), req.GetBackendPort())
	pl.Send("HSETNX", comm.IM_KEY_FRWD_NID_TO_ADDR, req.GetNid(), addr)
	pl.Send("HSETNX", comm.IM_KEY_FRWD_ADDR_TO_NID, addr, req.GetNid())

	ttl := time.Now().Unix() + comm.CHAT_NID_TTL
	pl.Send("ZADD", comm.IM_KEY_FRWD_NID_ZSET, ttl, req.GetNid())

	return
}

/******************************************************************************
 **函数名称: MonFrwdRptHandler
 **功    能: 转发层上报
 **输入参数:
 **     cmd: 消息类型
 **     nid: 结点ID
 **     data: 收到数据
 **     length: 数据长度
 **     param: 附加参数
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **协议格式:
 **     {
 **         required uint64 nid = 1;        // M|结点ID|数字|<br>
 **         required string ipaddr = 2;     // M|IP地址|字串|<br>
 **         required uint32 forward_port = 3;    // M|前端口号|数字|<br>
 **         required uint32 backend_port = 4;    // M|后端口号|数字|<br>
 **     }
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 11:04:36 #
 ******************************************************************************/
func MonFrwdRptHandler(cmd uint32, nid uint32, data []byte, length uint32, param interface{}) int {
	ctx, ok := param.(*MonSvrCntx)
	if !ok {
		return -1
	}

	ctx.log.Debug("Recv frwd-rpt request!")

	/* 1. > 解析FRWD-RPT请求 */
	head, req := ctx.frwd_rpt_parse(data)
	if nil == head || nil == req {
		ctx.log.Error("Parse frwd-rpt failed!")
		return -1
	}

	/* 2. > LSN-RPT请求处理 */
	ctx.frwd_rpt_handler(head, req)

	return 0
}

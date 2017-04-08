package controllers

import (
	"git.apache.org/thrift.git/lib/go/thrift"

	"beehive-im/src/golang/lib/im"
	"beehive-im/src/golang/lib/mesg/seqsvr"
)

type SeqSvrThrift struct {
	ctx *SeqSvrCntx
}

/******************************************************************************
 **函数名称: AllocSid
 **功    能: 申请会话SID
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 会话SID
 **实现描述: 从MYSQL中申请会话SID
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.31 22:48:00 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocSid() (int64, error) {
	ctx := this.ctx

	sid, err := im.AllocSid(ctx.mysql)
	if nil != err {
		ctx.log.Error("Alloc sid failed! errmsg:%s", err.Error())
		return 0, err
	}

	ctx.log.Debug("Alloc sid success! sid:%d", sid)

	return int64(sid), nil
}

/******************************************************************************
 **函数名称: AllocSeq
 **功    能: 申请序列号
 **输入参数:
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: 会话SID
 **实现描述: 从MYSQL中申请序列号
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.08 00:02:24 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocSeq(uid int64) (int64, error) {
	ctx := this.ctx

	ctx.log.Debug("Alloc sequence success! uid:%d", uid)

	return int64(0), nil
}

/******************************************************************************
 **函数名称: launch_thrift
 **功    能: 启动Thrift服务
 **输入参数:
 **     addr: 侦听地址(IP+端口)
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 侦听指定端口, 并启动服务.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.31 22:48:00 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) launch_thrift(addr string) {
	transport := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocol := thrift.NewTBinaryProtocolFactoryDefault()
	//protocol := thrift.NewTCompactProtocolFactory()

	socket, err := thrift.NewTServerSocket(addr)
	if nil != err {
		ctx.log.Error("Listen addr [%s] failed! errmsg:%s", addr, err.Error())
		return
	}

	handler := &SeqSvrThrift{ctx: ctx}
	processor := seqsvr.NewSeqSvrThriftProcessor(handler)

	server := thrift.NewTSimpleServer4(processor, socket, transport, protocol)

	ctx.log.Debug("Thrift server in: %s", addr)

	server.Serve()
}

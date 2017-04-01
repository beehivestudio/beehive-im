package controllers

import (
	"fmt"

	"git.apache.org/thrift.git/lib/go/thrift"

	"beehive-im/src/golang/lib/mesg/seqsvr"
)

type SeqSvrThrift struct {
	ctx *SeqSvrCntx
}

func (this *SeqSvrThrift) AllocSid() (sid uint64, err error) {
	ctx := this.ctx

	ctx.log.Debug("Call AllocSid()")

	return 12123123, nil
}

/******************************************************************************
 **函数名称: launch_thrift
 **功    能: 启动Thrift服务
 **输入参数:
 **     port: 侦听端口
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 侦听指定端口, 并启动服务.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.31 22:48:00 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) launch_thrift(port uint16) {
	transport := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocol := thrift.NewTBinaryProtocolFactoryDefault()
	//protocol := thrift.NewTCompactProtocolFactory()

	addr := fmt.Sprintf("*:%d", port)

	socket, err := thrift.NewTServerSocket(addr)
	if nil != err {
		ctx.log.Error("Listen addr [%s] failed! errmsg:%s", addr, err.Error())
		return
	}

	handler := &SeqSvrThrift{}
	processor := seqsvr.NewSeqSvrThriftProcessor(handler)

	server := thrift.NewTSimpleServer4(processor, socket, transport, protocol)

	ctx.log.Debug("Thrift server in: %s", addr)

	server.Serve()
}

package controller

import (
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/lws"
)

/******************************************************************************
 **函数名称: lsnd_conn_recv
 **功    能: 接收数据处理
 **输入参数:
 **     client: 客户端对象
 **     reason: 回调原因
 **     user: 自定义数据
 **     in: 收到的数据
 **     length: 收到数据的长度
 **     param: 扩展参数
 **输出参数:
 **返    回: 0:正常 !0:异常
 **实现描述:
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.04 15:21:43 #
 ******************************************************************************/
func (ctx *LsndCntx) lsnd_conn_recv(client *socket_t, reason int,
	user interface{}, in []byte, length int, param interface{}) int {
	return 0
}

/******************************************************************************
 **函数名称: lsnd_conn_send
 **功    能: 发送数据处理
 **输入参数:
 **     client: 客户端对象
 **     reason: 回调原因
 **     user: 自定义数据
 **     in: 收到的数据
 **     length: 收到数据的长度
 **     param: 扩展参数
 **输出参数:
 **返    回: 0:正常 !0:异常
 **实现描述:
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.04 15:21:43 #
 ******************************************************************************/
func (ctx *LsndCntx) lsnd_conn_send(client *socket_t, reason int,
	user interface{}, in []byte, length int, param interface{}) int {
}

/******************************************************************************
 **函数名称: lsnd_conn_destroy
 **功    能: 销毁CONN对象
 **输入参数:
 **     client: 客户端对象
 **     reason: 回调原因
 **     user: 自定义数据
 **     in: 收到的数据
 **     length: 收到数据的长度
 **     param: 扩展参数
 **输出参数:
 **返    回: 0:正常 !0:异常
 **实现描述:
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.04 15:21:43 #
 ******************************************************************************/
func (ctx *LsndCntx) lsnd_conn_destroy(client *socket_t, reason int,
	user interface{}, in []byte, length int, param interface{}) int {
	return 0
}

/******************************************************************************
 **函数名称: LsndLwsCallBack
 **功    能: LWS处理回调
 **输入参数:
 **     ctx: LWS对象
 **     client: 客户端对象
 **     reason: 回调原因
 **     user: 自定义数据
 **     in: 收到的数据
 **     length: 收到数据的长度
 **     param: 扩展参数
 **输出参数:
 **返    回: 0:正常 !0:异常
 **实现描述:
 **注意事项: 返回!0值将导致连接断开
 **作    者: # Qifeng.zou # 2017.03.04 00:16:09 #
 ******************************************************************************/
func LsndLwsCallBack(l *lws.LwsCntx, client *socket_t, reason int,
	user interface{}, in []byte, length int, param interface{}) int {
	ctx, ok := param.(*LsndCntx)
	if !ok {
		return -1
	}

	switch reason {
	case lws.LWS_CALLBACK_REASON_CREAT:
		return ctx.lsnd_conn_init(client, user, in, length, param)
	case lws.LWS_CALLBACK_REASON_RECV:
		return ctx.lsnd_conn_recv(client, user, in, length, param)
	case lws.LWS_CALLBACK_REASON_SEND:
		return ctx.lsnd_conn_send(client, user, in, length, param)
	case lws.LWS_CALLBACK_REASON_CLOSE:
		return ctx.lsnd_conn_destroy(client, user, in, length, param)
	}
	return 0
}

/******************************************************************************
 **函数名称: LsndGetMesgBodyLen
 **功    能: 获取报体长度
 **输入参数:
 **     head: 报头(注: 网络字节序)
 **输出参数:
 **返    回: 报体长度
 **实现描述:
 **注意事项: 进行网络字节序的转换
 **作    者: # Qifeng.zou # 2017.03.04 00:16:09 #
 ******************************************************************************/
func LsndGetMesgBodyLen(nhead *comm.MesgHeader) uint32 {
	hhead := &comm.MesgHeader{
		Length: head.Length,
	}

	comm.MesgHeadNtoh(hhead)

	return hhead.GetLength()
}

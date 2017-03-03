package controller

import (
	"beehive-im/src/golang/lib/comm"
	"beehive-im/src/golang/lib/lws"
)

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
func LsndLwsCallBack(lws *LwsCntx, client *socket_t, reason int,
	user interface{}, in []byte, length int, param interface{}) int {
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

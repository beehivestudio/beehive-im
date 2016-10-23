package rtmq

/* 工作协程的处理 */
func rtmq_proxy_work_routine(pxy *RtmqProxyCntx, idx int) {
	var item RtmqRegItem

	recvq := pxy.recv[idx]

	for data := range recvq {
		/* 已为本机字节序 */
		header := data.(*RtmqHeader)

		/* 获取CMD对应的注册项 */
		if item, ok := pxy.reg[header.cmd]; !ok {
			continue
		}

		/* 调用注册处理函数 */
		item.Proc(header.cmd, header.nid, data[RTMQ_HEAD_SIZE:], header.length, item.Param)
	}
}

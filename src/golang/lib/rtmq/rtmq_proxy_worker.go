package rtmq

/* 工作协程的处理 */
func rtmq_proxy_work_routine(pxy *RtmqProxyCntx, idx int) {
	svr := pxy.server[idx]
	recv_chan := svr.recv_chan

	for p := range recv_chan {
		header := rtmq_head_ntoh(p)

		/* 获取CMD对应的注册项 */
		item, ok := pxy.reg[header.cmd]
		if !ok {
			continue
		}

		/* 调用注册处理函数 */
		item.Proc(header.cmd, header.nid, p.buff[RTMQ_HEAD_SIZE:], header.length, item.Param)
	}
}

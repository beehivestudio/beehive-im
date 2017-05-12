package lws

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

/******************************************************************************
 **函数名称: ConnPoolAdd
 **功    能: 加入连接池操作
 **输入参数:
 **     cid: 连接CID
 **     client: 客户端对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.07 15:45:03 #
 ******************************************************************************/
func (ctx *LwsCntx) ConnPoolAdd(cid uint64, client *Client) int {
	pool := ctx.pool[cid%LWS_CONN_POOL_LEN]

	pool.Lock()
	defer pool.Unlock()

	pool.list[cid] = client
	return 0
}

/******************************************************************************
 **函数名称: ConnPoolDel
 **功    能: 移除连接池操作
 **输入参数:
 **     cid: 连接CID
 **     client: 客户端对象
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.07 15:45:03 #
 ******************************************************************************/
func (ctx *LwsCntx) ConnPoolDel(cid uint64) int {
	pool := ctx.pool[cid%LWS_CONN_POOL_LEN]

	pool.Lock()
	defer pool.Unlock()

	client, ok := pool.list[cid]
	if !ok {
		return 0
	}

	delete(pool.list, cid)
	if !client.iskick {
		close(client.sendq)
		client.iskick = true
	}
	return 0
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: conn_handler
 **功    能: 连接请求处理
 **输入参数:
 **     ctx: 全局对象
 **     w: 写对象
 **     r: HTTP请求
 **输出参数: NONE
 **返    回: 错误码
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.06 23:19:44 #
 ******************************************************************************/
func conn_handler(ctx *LwsCntx, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if nil != err {
		ctx.log.Error("Upgrade websocket failed! errmsg:%s", err.Error())
		return
	}

	client := &Client{
		ctx:    ctx,                                  /* 全局对象 */
		cid:    atomic.AddUint64(&ctx.cid, 1),        /* 连接ID */
		conn:   conn,                                 /* 连接对象 */
		sendq:  make(chan []byte, ctx.conf.SendqMax), /* 发送队列 */
		iskick: false,                                /* 是否被踢 */
	}

	/* 创建连接的回调 */
	ret := ctx.protocol.Callback(client.ctx, client,
		LWS_CALLBACK_REASON_CREAT, nil, 0, ctx.protocol.Param)
	if 0 != ret {
		ctx.log.Error("Create socket failed!")
		return
	}

	ctx.ConnPoolAdd(client.cid, client)

	go client.send_routine() // 发送协程
	client.recv_routine()    // 接收协程
}

/******************************************************************************
 **函数名称: recv_routine
 **功    能: 接收协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述: 使用ws.ReadMessage()接收数据
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.07 22:30:37 #
 ******************************************************************************/
func (client *Client) recv_routine() {
	ctx := client.ctx
	defer func() {
		if err := recover(); nil != err {
			//log.Error("Recv routine crashed! errmsg:%s", err)
		}
		/* 关闭连接的处理 */
		client.conn.Close()
		ctx.protocol.Callback(ctx, client,
			LWS_CALLBACK_REASON_CLOSE, nil, 0, ctx.protocol.Param)
		ctx.ConnPoolDel(client.cid)
	}()

	for {
		/* 读取一条完整的数据(协议头+协议体) */
		_, data, err := client.conn.ReadMessage()
		if nil != err {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				ctx.log.Error("Connections was closed! errmsg:%s", err.Error())
			}
			break
		}

		//length := ctx.protocol.GetPacketBodyLenCb(data) // 获取报体长度

		/* 调用回调函数(注意:返回非0值将导致连接被关闭) */
		if 0 != ctx.protocol.Callback(ctx, client,
			LWS_CALLBACK_REASON_RECV, data, len(data), ctx.protocol.Param) {
			break
		}
	}
}

/******************************************************************************
 **函数名称: send_routine
 **功    能: 发送协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **     1. 将发送队列中的数据发送给客户端
 **     2. 定时发送PING请求给客户端
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.07 22:57:38 #
 ******************************************************************************/
func (client *Client) send_routine() {
	ticker := time.NewTicker(LWS_PING_PERIOD_SEC)
	defer func() {
		if err := recover(); nil != err {
			//log.Error("Recv routine crashed! errmsg:%s", err)
		}
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case data, ok := <-client.sendq: // 等待下发数据
			client.conn.SetWriteDeadline(time.Now().Add(LWS_WRITE_WAIT_SEC))
			if !ok {
				// The ctx closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if nil != err {
				return
			}
			w.Write(data) // 发送数据

			/* 下发队列所有数据 */
			n := len(client.sendq)
			for i := 0; i < n; i++ {
				w.Write(<-client.sendq) // 发送数据
			}

			if err := w.Close(); nil != err {
				return
			}
		case <-ticker.C: // 定时下发WS的PING请求
			client.conn.SetWriteDeadline(time.Now().Add(LWS_WRITE_WAIT_SEC))
			if err := client.conn.WriteMessage(websocket.PingMessage, []byte{}); nil != err {
				return
			}
		}
	}
}

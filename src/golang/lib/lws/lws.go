package lws

import (
	"github.com/gorilla/websocket"
)

/* 常量定义 */
const (
	LWS_WRITE_WAIT_SEC  = 10 * time.Second             /* 最大发送阻塞时间 */
	LWS_PONG_WAIT_SEC   = 60 * time.Second             /* 接收PONG的间隔时间 */
	LWS_PING_PERIOD_SEC = (LWS_PONG_WAIT_SEC * 9) / 10 /* 发送PING的间隔时间 */
	maxMessageSize      = 8192                         /* 一次接收最大消息长度 */
)

/* 回调原因 */
const (
	LWS_CALLBACK_REASON_CREAT = 1 /* 创建连接 */
	LWS_CALLBACK_REASON_RECV  = 2 /* 接收数据 */
	LWS_CALLBACK_REASON_WRITE = 3 /* 发送数据 */
	LWS_CALLBACK_REASON_CLOSE = 4 /* 关闭连接 */
)

type lws_callback_t func(ctx *LwsCntx, client *socket_t,
	reason int, user interface{}, in interface{}, length int, param interface{}) int
type lws_get_packet_body_size_cb func(void *head) uint32

/* 帧听协议 */
type Protocol struct {
	callback             lws_callback_t              /* 处理回调 */
	per_packet_head_size uint32                      /* 每个包的报头长度 */
	get_packet_body_size lws_get_packet_body_size_cb /* 每个包的报体长度 */
	param                interface{}                 /* 附加参数 */
}

/* 连接池对象 */
type ConnPool struct {
	sync.RWMutex                 // 读写锁
	clients      map[int]*Client // 连接管理池
}

/* 全局对象 */
type LwsCntx struct {
	conf       *Conf        // 配置参数
	protocol   *Protocol    // 注册协议
	log        *Beego.Logs  // 日志对象
	cid        uint64       // 连接序列号(原子递增)
	conn_pool  ConnPool     // 连接池
	unregister chan *Client // Unregister requests from clients.
}

/* 配置对象 */
type Conf struct {
	addr string // IP+端口
}

/******************************************************************************
 **函数名称: Init
 **功    能: 程序初始化
 **输入参数:
 **     conf: 配置信息
 **     log: 日志对象
 **输出参数: NONE
 **返    回: WS上下文
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.06 22:44:04 #
 ******************************************************************************/
func Init(conf *Conf, log *Beego.Logs) *LwsCntx {
	ctx := &LwsCntx{
		cid:  0,
		log:  log,
		conf: Conf,
	}

	ctx.conn_pool.clients = make(map[int]*Client)

	return ctx
}

/******************************************************************************
 **函数名称: Register
 **功    能: 注册路径
 **输入参数:
 **     path: URL路径
 **输出参数: NONE
 **返    回: 错误码
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.07 08:36:08 #
 ******************************************************************************/
func (ctx *LwsCntx) Register(path string) int {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		conn_handler(ctx, w, r)
	})

	return 0
}

/******************************************************************************
 **函数名称: run
 **功    能: 工作协程
 **输入参数: NONE
 **输出参数: NONE
 **返    回: VOID
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.07 22:50:25 #
 ******************************************************************************/
func (ctx *LwsCntx) run() {
	for {
		select {
		case client := <-ctx.unregister:
			if _, ok := ctx.clients[client]; ok {
				delete(ctx.clients, client)
				close(client.sendq)
			}
		}
	}
}

/******************************************************************************
 **函数名称: Launch
 **功    能: 启动程序
 **输入参数:
 **     protocol: 协议内容
 **输出参数: NONE
 **返    回: 错误码
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.05 20:25:27 #
 ******************************************************************************/
func (ctx *LwsCntx) Launch(protocol *Protocol) int {
	ctx.protocol = protocol

	go ctx.run()

	/* 侦听指定端口 */
	err := http.ListenAndServe(ctx.conf.addr, nil)
	if nil != err {
		log.Fatal("ListenAndServe: ", err)
	}
	return 0
}

/******************************************************************************
 **函数名称: AsyncSend
 **功    能: 异步发送数据
 **输入参数:
 **     cid: 连接ID
 **     data: 发送的数据
 **输出参数: NONE
 **返    回: 错误码
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.06 23:05:54 #
 ******************************************************************************/
func (ctx *LwsCntx) AsyncSend(cid uint64, data []byte) int {
	ctx.conn_pool.RLock()
	client, ok := ctx.conn_pool.clients[cid]
	if !ok {
		ctx.conn_pool.RUnlock()
		return -1
	}
	client.sendq <- data
	ctx.conn_pool.RUnlock()
	return 0
}

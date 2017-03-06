package lws

import (
	"github.com/gorilla/websocket"
)

/* 常量定义 */
const (
	LWS_WRITE_WAIT_SEC  = 10 * time.Second             /* 最大发送阻塞时间 */
	LWS_PONG_WAIT_SEC   = 60 * time.Second             /* 接收PONG的间隔时间 */
	LWS_PING_PERIOD_SEC = (LWS_PONG_WAIT_SEC * 9) / 10 /* 发送PING的间隔时间 */
)

/* 回调原因 */
const (
	LWS_CALLBACK_REASON_CREAT = 1 /* 创建连接 */
	LWS_CALLBACK_REASON_RECV  = 2 /* 接收数据 */
	LWS_CALLBACK_REASON_SEND  = 3 /* 发送数据 */
	LWS_CALLBACK_REASON_CLOSE = 4 /* 关闭连接 */
)

type LwsCallback func(ctx *LwsCntx, client *Client,
	reason int, data interface{}, length int, param interface{}) int
type LwsGetPacketBodyLenFunc func(void *head) uint32

/* 帧听协议 */
type Protocol struct {
	Callback           LwsCallback             /* 处理回调 */
	PerPacketHeadSize  uint32                  /* 每个包的报头长度 */
	GetPacketBodyLenCb LwsGetPacketBodyLenFunc /* 每个包的报体长度 */
	Param              interface{}             /* 附加参数 */
}

/* 连接池对象 */
type ConnPool struct {
	sync.RWMutex                 // 读写锁
	list         map[int]*Client // 连接管理池
}

/* 全局对象 */
type LwsCntx struct {
	conf       *Conf        // 配置参数
	protocol   *Protocol    // 注册协议
	log        *Beego.Logs  // 日志对象
	cid        uint64       // 连接序列号(原子递增)
	conn       ConnPool     // 连接池
	unregister chan *Client // Unregister requests from clients.
}

/* 配置对象 */
type Conf struct {
	Ip       string // IP地址
	Port     uint32 // 端口号
	Max      uint32 // 最大连接数
	Timeout  uint32 // 连接超时时间
	SendqMax uint32 // 发送队列长度
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

	ctx.conn.list = make(map[int]*Client)

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
			if _, ok := ctx.conn.list[client]; ok {
				delete(ctx.conn.list, client)
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
	addr := fmt.Sprintf("%s:%d", ctx.conf.Ip, ctx.conf.Port)

	err := http.ListenAndServe(addr, nil)
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
	ctx.conn.RLock()
	client, ok := ctx.conn.list[cid]
	if !ok {
		ctx.conn.RUnlock()
		return -1
	}
	client.sendq <- data
	ctx.conn.RUnlock()
	return 0
}

/******************************************************************************
 **函数名称: Kick
 **功    能: 踢除指定连接
 **输入参数:
 **     cid: 连接ID
 **输出参数: NONE
 **返    回: 错误码
 **实现描述: 关闭发送队列
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.06 21:46:19 #
 ******************************************************************************/
func (ctx *LwsCntx) Kick(cid uint64) int {
	ctx.conn.RLock()
	client, ok := ctx.conn.list[cid]
	if !ok {
		ctx.conn.RUnlock()
		return -1
	}
	close(client.sendq)
	return 0
}

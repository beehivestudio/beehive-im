package lws

import (
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/* 客户端对象 */
type Client struct {
	ctx    *LwsCntx        /* 全局对象 */
	cid    uint64          /* 连接ID */
	conn   *websocket.Conn /* WS连接对象 */
	sendq  chan []byte     /* 发送队列 */
	iskick bool            /* 是否被踢 */
	user   interface{}     /* 附加数据 */
}

/* 获取连接ID */
func (c *Client) GetCid() uint64 {
	return c.cid
}

/* 获取用户数据 */
func (c *Client) GetUserData() interface{} {
	return c.user
}

/* 设置用户数据 */
func (c *Client) SetUserData(user interface{}) {
	c.user = user
}

package lws

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"src/golang/lib/comm"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the ctx.
type Client struct {
	ctx   *Hub            // 全局对象
	cid   uint64          // 连接ID
	conn  *websocket.Conn /* WS连接对象 */
	sendq chan []byte     // 发送队列
	user  interface{}     // 附加数据
}

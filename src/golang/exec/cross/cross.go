package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	Head = 4
	str  = `<?xml version="1.0" encoding="UTF-8"?>
            <!DOCTYPE cross-domain-policy SYSTEM "http://www.macromedia.com/xml/dtds/cross-domain-policy.dtd">
            <cross-domain-policy>
            <allow-access-from domain="*" to-ports="*"/>
            </cross-domain-policy>`
)

var (
	ClientMap map[int]net.Conn = make(map[int]net.Conn)
)

func main() {
	fmt.Println(os.Args[0])
	ip_port := ":843"
	fmt.Println("Flash AS 策略服务运行中...\n",
		"自动回应SocketXml端口(即843端口)的crossdomain.xml请求\n",
		"如需指定ip和端口，可以在程序启动时指定参数，格式如下\ngameserve",
		"\n当前正在侦听", ip_port, "\n请不要关闭此窗口...")
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip_port)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	clientIndex := 0

	for {
		clientIndex++
		conn, err := listener.Accept()
		if nil != err {
			continue
		}
		go handleClient(conn, clientIndex)
	}
}

func handleClient(conn net.Conn, index int) {
	ClientMap[index] = conn

	fmt.Println("")
	fmt.Println("=======================")
	fmt.Println("新用户连接, 来自: ", conn.RemoteAddr(), "index: ", index)
	fc := func() {
		time.Sleep(time.Second) //给客户端1秒的响应的时间，否则客户端有可能读不到数据就提前Close了
		conn.Close()
		delete(ClientMap, index)
		fmt.Println("移除序号为: ", index, "的客户端，断开客户端的连接")
		fmt.Println("=======================")
	}
	defer fc()
	sendFirstMsg(conn)
}

func sendFirstMsg(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	writer.WriteString(str)
	writer.Flush()
	fmt.Println("已经回应策略文件：crossdomain.xml")
}

func checkError(err error) {
	if nil != err {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

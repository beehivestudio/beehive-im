package main

import (
	"filter"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"

	"beehive-im/src/golang/lib/thrift_pool"
)

const (
	HOST = "127.0.0.1"
	PORT = "9093"
)

var rIndex int
var wg sync.WaitGroup

func testAA(p *thrift_pool.Pool, total int) {
	rIndex++
	conn, _ := p.Get()
	if conn == nil {
		fmt.Println("Get nil conn", rIndex)
	} else {
		for i := 0; i < total; i++ {
			content := "3d打印枪支图纸" + strconv.Itoa(i+1)
			msgtest := filter.Reqmsg{"filter,sensitive", content}
			conn.(*filter.FilterThriftClient).Test(&msgtest)
			//r1, _ := conn.(*filter.FilterThriftClient).Test(&msgtest)
			//fmt.Println("GOClient Call->Test", r1)
		}
	}

	wg.Done()
	if conn != nil {
		p.Put(conn, false)
	}
	//fmt.Println("GOClient Call->Test", total, p)
}
func printAct(p *thrift_pool.Pool) {
	fmt.Println("pool MaxActive", p.ActiveCount())
}

func main() {
	startTime := currentTimeMillis()

	thriftPool = &thrift_pool.Pool{
		Dial: func() (interface{}, error) {
			addr := net.JoinHostPort(HOST, PORT)
			sock, err := thrift.NewTSocket(addr) // client端不设置超时
			if err != nil {
				//fmt.Fprintln("thrift.NewTSocketTimeout(%s) error(%v)", addr, err)
				return nil, err
			}
			tF := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
			pF := thrift.NewTBinaryProtocolFactoryDefault()
			client := filter.NewFilterThriftClientFactory(tF.GetTransport(sock), pF)
			if err = client.Transport.Open(); err != nil {
				//fmt.Fprintln("client.Transport.Open() error(%v)", err)
				fmt.Println("OPen erro")
				return nil, err
			}
			fmt.Println(client)
			return client, nil
		},
		Close: func(v interface{}) error {
			v.(*filter.FilterThriftClient).Transport.Close()
			return nil
		},
		MaxActive:   4,
		MaxIdle:     100,
		IdleTimeout: 5000000,
	}

	// conn, _ := thriftPool.Get()
	// thriftPool.Put(conn, false)

	// conn2, _ := thriftPool.Get()
	// thriftPool.Put(conn2, false)

	total := 12
	wg.Add(8)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	time.Sleep(5 * 1e8)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	go testAA(thriftPool, total)
	t := time.Tick(time.Second)

	go func() {
		for {
			select {
			case <-t:
				printAct(thriftPool)
			}
		}
	}()

	wg.Wait()
	thriftPool.Release()
	endTime := currentTimeMillis()
	fmt.Printf("cnt:%d 本次调用用时: %d-%d=%d毫秒\n", total, endTime, startTime, (endTime - startTime))

}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

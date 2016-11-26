package main

import (
	"fmt"

	"chat/src/golang/lib/comm"
)

func main() {
	dict, err := comm.LoadIpDict("./ipdict.txt")
	if nil != err {
		fmt.Printf(err.Error())
	}

	item := dict.Query("223.255.37.129")
	if nil == item {
		fmt.Printf("Didn't find!\n")
		return
	}
	fmt.Printf("find! country:%s operator:%s\n", item.Country, item.Operator)
}

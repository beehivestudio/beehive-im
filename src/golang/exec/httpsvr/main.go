package main

import (
	"github.com/astaxie/beego"

	_ "chat/src/golang/exec/httpsvr/routers"
)

func main() {
	beego.Run(":50000")
}

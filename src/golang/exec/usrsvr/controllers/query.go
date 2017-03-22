package controllers

import (
	"fmt"

	"beehive-im/src/golang/lib/comm"
)

type UsrSvrQueryCtrl struct {
	BaseController
}

func (this *UsrSvrQueryCtrl) Query() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	}
	this.Error(comm.ERR_SVR_INVALID_PARAM, fmt.Sprintf("Unsupport this option:%s", option))
}

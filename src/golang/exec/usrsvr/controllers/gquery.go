package controllers

import (
	"fmt"

	"beehive-im/src/golang/lib/comm"
)

type UsrSvrGroupQueryCtrl struct {
	BaseController
}

func (this *UsrSvrGroupQueryCtrl) Query() {
	//ctx := GetUsrSvrCtx()

	option := this.GetString("option")
	switch option {
	}

	errmsg := fmt.Sprintf("Unsupport this option:%s", option)

	this.Error(comm.ERR_SVR_INVALID_PARAM, errmsg)
}

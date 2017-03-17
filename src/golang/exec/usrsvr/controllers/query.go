package controllers

import (
	"beehive-im/src/golang/lib/comm"
)

type UsrSvrQueryCtrl struct {
	BaseController
}

func (this *UsrSvrQueryCtrl) Query() {
	this.Error(comm.ERR_SVR_INVALID_PARAM, "Unsupport this option!")
}

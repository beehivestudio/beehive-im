package controllers

import (
	"errors"
)

/******************************************************************************
 **函数名称: AllocSid
 **功    能: 申请会话SID
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 会话SID
 **实现描述: 从MYSQL中申请会话SID
 **注意事项:
 **作    者: # Qifeng.zou # 2017.03.31 22:48:00 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocSid() (int64, error) {
	ctx := this.ctx

	sid, err := ctx.alloc_sid()
	if nil != err {
		ctx.log.Error("Alloc sid failed! errmsg:%s", err.Error())
		return 0, err
	}

	ctx.log.Debug("Alloc sid success! sid:%d", sid)

	return int64(sid), nil
}

/******************************************************************************
 **函数名称: alloc_sid
 **功    能: 申请会话SID
 **输入参数:
 **     db: 数据库
 **输出参数: NONE
 **返    回:
 **     sid: 会话SID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.02 10:48:40 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_sid() (sid uint64, err error) {
	tx, err := ctx.mysql.Begin()
	if nil != err {
		return 0, err
	}

	defer tx.Commit()

	rows, err := tx.Query("SELECT sid from IM_SID_GEN_TAB WHERE type=0 FOR UPDATE")
	if nil != err {
		rows.Close()
		return 0, err
	} else if rows.Next() {
		err = rows.Scan(&sid)
		if nil != err {
			rows.Close()
			return 0, err
		}
		rows.Close()
		_, err := tx.Exec("UPDATE IM_SID_GEN_TAB SET sid=sid+1 WHERE type=0")
		if nil != err {
			return 0, err
		}
		return sid, nil
	}

	rows.Close()
	return 0, errors.New("Alloc sid failed!")
}

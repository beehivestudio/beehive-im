package controllers

import ()

/******************************************************************************
 **函数名称: AllocRoomId
 **功    能: 申请聊天室ID
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 会话SID
 **实现描述: 从MYSQL中申请聊天室ID
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.29 08:35:04 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocRoomId() (int64, error) {
	ctx := this.ctx

	rid, err := ctx.alloc_rid()
	if nil != err {
		ctx.log.Error("Alloc rid failed! errmsg:%s", err.Error())
		return 0, err
	}

	ctx.log.Debug("Alloc rid success! rid:%d", rid)

	return int64(rid), nil
}

/******************************************************************************
 **函数名称: alloc_rid
 **功    能: 申请聊天室ID
 **输入参数:
 **     db: 数据库
 **输出参数: NONE
 **返    回:
 **     rid: 聊天室ID
 **     err: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.29 08:36:13 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_rid() (rid uint64, err error) {
	tx, err := ctx.mysql.Begin()
	if nil != err {
		return 0, err
	}

	defer tx.Commit()

AGAIN:
	rows, err := tx.Query("SELECT rid from IM_RID_GEN_TAB WHERE id=0 FOR UPDATE")
	if nil != err {
		rows.Close()
		return 0, err
	} else if rows.Next() {
		err = rows.Scan(&rid)
		if nil != err {
			rows.Close()
			return 0, err
		}
		rows.Close()
		_, err := tx.Exec("UPDATE IM_RID_GEN_TAB SET rid=rid+1 WHERE id=0")
		if nil != err {
			return 0, err
		}
		return rid, nil
	}

	rows.Close()

	/* > 新增RID生成器 */
	_, err = tx.Exec("INSERT INTO IM_RID_GEN_TAB(id) VALUES(0)")
	if nil != err {
		ctx.log.Error("Add rid gen failed! errmsg:%s", err.Error())
		return 0, err
	}

	goto AGAIN
}

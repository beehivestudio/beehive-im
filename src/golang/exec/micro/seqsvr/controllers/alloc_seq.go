package controllers

/******************************************************************************
 **函数名称: AllocSeq
 **功    能: 申请序列号
 **输入参数:
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: 会话SID
 **实现描述: 从MYSQL中申请序列号
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.08 00:02:24 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocSeq(uid int64) (int64, error) {
	ctx := this.ctx

	ctx.log.Debug("Alloc sequence success! uid:%d", uid)

	return int64(0), nil
}

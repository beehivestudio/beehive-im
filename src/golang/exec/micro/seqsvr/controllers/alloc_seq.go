package controllers

import (
	"errors"
	"fmt"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: AllocSeq
 **功    能: 申请序列号
 **输入参数:
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: 序列号
 **实现描述: 每10万个UID为同一个段(SECTION), 同一段的UID共享最大序列号的增长.
 **          如: A~...为同一个段的用户, 则他们的序列号均处在[MIN, MAX]区间, 各自增长.
 **           -----------------------------------------------
 **          |  A  |  B  |  C  |  D  |  E  |  F  |  G  | ... |
 **           -----------------------------------------------
 **          | 501 | 302 | 300 | 521 | 792 | 888 | 912 | ... |
 **           -----------------------------------------------
 **          |                MIN:300 MAX:1000               |
 **           -----------------------------------------------
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.08 00:02:24 #
 ******************************************************************************/
func (this *SeqSvrThrift) AllocSeq(uid int64) (int64, error) {
	ctx := this.ctx

	id := uint64(uid) / comm.SECTION_UID_NUM

	err := ctx.section_add(id)
	if nil != err {
		ctx.log.Error("Add section [%d] failed! errmsg:%s", id, err.Error())
		return 0, err
	}

	seq, err := ctx.alloc_seq(id, uint64(uid))
	if nil != err {
		ctx.log.Error("Alloc seq failed! uid:%d errmsg:%s", uid, err.Error())
		return 0, err
	}

	return int64(seq), nil
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: section_add
 **功    能: 新增SECTION
 **输入参数:
 **     id: SECTION编号
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.11 23:45:32 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) section_add(id uint64) error {
	list := ctx.section[id/SECTION_LIST_LEN]
AGAIN:
	list.RLock()
	item, ok := list.items[id]
	if !ok {
		list.RUnlock()
		min, max, err := ctx.alloc_seq_from_db(id)
		if nil != err {
			return err
		}

		list.Lock()
		item, ok = list.items[id]
		if !ok {
			section := &SectionItem{min: min, max: max}
			for idx := 0; idx < USER_LIST_LEN; idx += 1 {
				section.ulist[idx].items = make(map[uint64]*UserItem)
			}
			list.items[id] = section
		} else { // 无需更新min值
			item.Lock()
			if item.max < max {
				item.max = max
			}
			item.Unlock()
		}
		list.Unlock()
		goto AGAIN
	}
	defer list.RUnlock()

	return nil
}

/******************************************************************************
 **函数名称: alloc_seq_from_db
 **功    能: 从DB中申请序列号
 **输入参数:
 **     id: 段编号
 **输出参数: NONE
 **返    回: 最小&最大序列号
 **实现描述:
 **注意事项: 返回值为0时表示系统异常
 **作    者: # Qifeng.zou # 2017.04.11 23:54:14 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_seq_from_db(id uint64) (min uint64, max uint64, err error) {
	var seq uint64

	tx, err := ctx.mysql.Begin()
	if nil != err {
		return 0, 0, err
	}

	defer tx.Commit()

	rows, err := tx.Query("SELECT seq from IM_SEQ_GEN_TAB WHERE id=? FOR UPDATE", id)
	if nil != err {
		rows.Close()
		return 0, 0, err
	} else if rows.Next() {
		err = rows.Scan(&seq)
		if nil != err {
			rows.Close()
			return 0, 0, err
		}
		rows.Close()
		_, err := tx.Exec("UPDATE IM_SEQ_GEN_TAB SET seq=seq+1000 WHERE id=?", id)
		if nil != err {
			return 0, 0, err
		}
		return seq, seq + 1000, nil
	}

	rows.Close()
	return 0, 0, errors.New("Alloc seq failed!")
}

/******************************************************************************
 **函数名称: alloc_seq
 **功    能: 申请序列号
 **输入参数:
 **     id: 段编号
 **     uid: 用户UID
 **输出参数: NONE
 **返    回: 序列号+错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.12 23:18:51 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_seq(id uint64, uid uint64) (seq uint64, err error) {
	has_update := false

AGAIN:
	seq, code, err := ctx.alloc_seq_from_section(id, uint64(uid))
	if comm.ERR_SVR_SEQ_EXHAUSTION == code {
		if has_update {
			return 0, errors.New("Update sequence failed!")
		}
		ctx.update_seq(id, seq)
		has_update = true
		goto AGAIN
	} else if nil != err {
		ctx.log.Error("Alloc seq failed! uid:%d errmsg:%s", uid, err.Error())
		return 0, err
	}
	return 0, nil
}

/******************************************************************************
 **函数名称: alloc_seq_from_section
 **功    能: 申请序列号
 **输入参数:
 **     id: 段编号
 **     uid: 用户UID
 **输出参数: NONE
 **返    回:
 **     seq: 序列号
 **     code: 错误码
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.12 23:18:51 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_seq_from_section(id uint64, uid uint64) (seq uint64, code int, err error) {
	slist := &ctx.section[id/SECTION_LIST_LEN]

	slist.RLock()
	defer slist.RUnlock()

	section, ok := slist.items[id]
	if !ok {
		ctx.log.Error("Get section failed! uid:%d id:%d", uid, id)
		return 0, comm.ERR_SVR_SEQ_EXHAUSTION, errors.New("Get section failed!")
	}

	section.RLock()
	defer section.RUnlock()

	ulist := section.ulist[uid/USER_LIST_LEN]

USER:
	ulist.RLock()
	user, ok := ulist.items[uid]
	if !ok {
		ulist.RUnlock()
		ulist.Lock()
		_, ok = ulist.items[uid]
		if !ok {
			ulist.items[uid] = &UserItem{uid: uid, seq: section.min}
		}
		ulist.Unlock()
		goto USER
	}
	defer ulist.RUnlock()

	user.Lock()
	defer user.Unlock()

	seq = user.seq + 1
	if seq > section.max {
		return seq, comm.ERR_SVR_SEQ_EXHAUSTION, errors.New("Sequence exhaustion!")
	}
	user.seq += 1

	ctx.log.Debug("Alloc sequence success! uid:%d section:%d", uid, section)

	return seq, 0, nil
}

/******************************************************************************
 **函数名称: update_seq
 **功    能: 更新序列号
 **输入参数:
 **     id: 段编号
 **     seq: 当前需要申请的序列号
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.11 23:54:14 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) update_seq(id uint64, seq uint64) error {
	list := ctx.section[id/SECTION_LIST_LEN]

	list.RLock()
	defer list.RUnlock()
	item, ok := list.items[id]
	if !ok {
		return errors.New(fmt.Sprintf("Didn't find section! id:%d", id))
	}

	item.Lock()
	defer item.Unlock()

AGAIN:
	if item.max < seq {
		_, max, err := ctx.alloc_seq_from_db(id)
		if nil != err {
			return err
		}
		if item.max < max {
			item.max = max
		}
		goto AGAIN
	}

	return nil
}

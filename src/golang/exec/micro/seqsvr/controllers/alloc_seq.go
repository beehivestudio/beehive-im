package controllers

import (
	"errors"
	"fmt"

	"beehive-im/src/golang/lib/comm"
)

/******************************************************************************
 **函数名称: GetSessionSeq
 **功    能: 获取会话SID对应的最新序列号
 **输入参数:
 **     sid: 会话SID
 **输出参数: NONE
 **返    回: 序列号
 **实现描述: 每10万个SID为同一个段(SECTION), 同一段的SID共享最大序列号的增长.
 **          如: A~...为同一个段的SID, 则他们的序列号均处在[MIN, MAX]区间, 各自增长.
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
func (this *SeqSvrThrift) GetSessionSeq(sid int64) (int64, error) {
	ctx := this.ctx

	secid := uint64(sid) / comm.SECTION_SID_NUM

	err := ctx.section_add(secid)
	if nil != err {
		ctx.log.Error("Add section [%d] failed! errmsg:%s", secid, err.Error())
		return 0, err
	}

	seq, err := ctx.query_seq_by_sid(secid, uint64(sid))
	if nil != err {
		ctx.log.Error("Query seq by sid [%d] failed! errmsg:%s", sid, err.Error())
		return 0, err
	}

	return int64(seq), nil
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: section_add
 **功    能: 新增SECTION
 **输入参数:
 **     secid: SECTION编号
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项: 为了支持分布式部署, 当更新各段max值时, 也需同时更新缓存section的min值.
 **作    者: # Qifeng.zou # 2017.04.11 23:45:32 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) section_add(secid uint64) error {
	section := ctx.ctrl.section[secid/SECTION_LIST_LEN]
AGAIN:
	section.RLock()
	item, ok := section.item[secid]
	if !ok {
		section.RUnlock()
		min, max, err := ctx.alloc_seq_from_db(secid)
		if nil != err {
			return err
		}

		section.Lock()
		item, ok = section.item[secid]
		if !ok {
			item := &SectionItem{min: min, max: max}
			section.item[secid] = item
		} else {
			item.Lock()
			if item.max < max {
				if min != item.max {
					item.min = min
				}
				item.max = max
			}
			item.Unlock()
		}
		section.Unlock()
		goto AGAIN
	}
	defer section.RUnlock()

	return nil
}

/******************************************************************************
 **函数名称: section_find
 **功    能: 查找SECTION
 **输入参数:
 **     secid: SECTION编号
 **输出参数: NONE
 **返    回: 最小&最大序列号
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.14 00:18:03 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) section_find(secid uint64) (min uint64, max uint64, err error) {
	section := ctx.ctrl.section[secid/SECTION_LIST_LEN]

	section.RLock()
	item, ok := section.item[secid]
	if !ok {
		section.RUnlock()

		/* 从DB中申请序列号 */
		min, max, err := ctx.alloc_seq_from_db(secid)
		if nil != err {
			return 0, 0, err
		}

		section.Lock()
		item, ok = section.item[secid]
		if !ok {
			section.item[secid] = &SectionItem{min: min, max: max}
		} else { // 无需更新min值
			item.Lock()
			if item.max < max {
				item.max = max
			}
			item.Unlock()
		}
		section.Unlock()
		return min, max, nil
	}
	defer section.RUnlock()

	item.RLock()
	defer item.RUnlock()

	min = item.min
	max = item.max

	return min, max, nil
}

////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: load_seq_from_db
 **功    能: 从DB中加载序列号
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述: 遍历SEQ生成表所有段secid.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.15 11:21:34 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) load_seq_from_db() (err error) {
	var secid uint64

	/* > 查询消息序列 */
	rows, err := ctx.mysql.Query("SELECT secid from IM_SEQ_GEN_TAB")
	if nil != err {
		rows.Close()
		ctx.log.Error("Query SEQ [%d] failed! errmsg:%s", secid, err.Error())
		return err
	}

	/* > 遍历查询结果 */
	for rows.Next() {
		err = rows.Scan(&secid)
		if nil != err {
			rows.Close()
			ctx.log.Error("Scan query failed! secid:%d errmsg:%s", secid, err.Error())
			return err
		}
		ctx.section_add(secid)
	}

	rows.Close()

	return nil
}

/******************************************************************************
 **函数名称: alloc_seq_from_db
 **功    能: 从DB中申请序列号
 **输入参数:
 **     secid: 段编号
 **输出参数: NONE
 **返    回: 最小&最大序列号
 **实现描述: 当被查找的段secid不存在时, 则新建一条记录.
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.11 23:54:14 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) alloc_seq_from_db(secid uint64) (min uint64, max uint64, err error) {
	var seq uint64

	tx, err := ctx.mysql.Begin()
	if nil != err {
		ctx.log.Error("Create transaction failed! secid:%d errmsg:%s", secid, err.Error())
		return 0, 0, err
	}

	defer tx.Commit()

AGAIN:
	/* > 查询消息序列号 */
	rows, err := tx.Query("SELECT seq from IM_SEQ_GEN_TAB WHERE secid=? FOR UPDATE", secid)
	if nil != err {
		rows.Close()
		ctx.log.Error("Query SEQ [%d] failed! errmsg:%s", secid, err.Error())
		return 0, 0, err
	} else if rows.Next() {
		err = rows.Scan(&seq)
		if nil != err {
			rows.Close()
			ctx.log.Error("Scan query failed! secid:%d errmsg:%s", secid, err.Error())
			return 0, 0, err
		}
		rows.Close()
		/* > 更新消息序列号 */
		_, err := tx.Exec("UPDATE IM_SEQ_GEN_TAB SET seq=seq+1000 WHERE secid=?", secid)
		if nil != err {
			ctx.log.Error("Update SEQ [%d] failed! errmsg:%s", secid, err.Error())
			return 0, 0, err
		}
		return seq, seq + 1000, nil
	}

	rows.Close()

	/* > 新增消息序列号 */
	_, err = tx.Exec("INSERT INTO IM_SEQ_GEN_TAB(secid, seq) VALUES(?, 1)", secid)
	if nil != err {
		ctx.log.Error("Add SEQ [%d] failed! errmsg:%s", secid, err.Error())
		return 0, 0, err
	}

	goto AGAIN
}

/******************************************************************************
 **函数名称: query_seq_by_sid
 **功    能: 查询SID最新序列号
 **输入参数:
 **     secid: 段编号
 **     sid: 用户UID
 **输出参数: NONE
 **返    回:
 **     seq: 序列号
 **     err: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.12 23:18:51 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) query_seq_by_sid(secid uint64, sid uint64) (seq uint64, err error) {
	session := ctx.ctrl.session[sid/USER_LIST_LEN]

SESSION:
	session.RLock()
	item, ok := session.item[sid]
	if !ok {
		session.RUnlock()
		min, max, err := ctx.section_find(secid)
		if nil != err {
			return 0, err
		}
		session.Lock()
		item, ok = session.item[sid]
		if !ok {
			session.item[sid] = &SessionItem{sid: sid, seq: min, max: max}
		} else {
			item.Lock()
			if item.max < max {
				item.seq = min
				item.max = max
			}
			item.Unlock()
		}
		session.Unlock()
		goto SESSION
	}
	defer session.RUnlock()

	item.RLock()
	defer item.RUnlock()

	return item.seq, nil
}

/******************************************************************************
 **函数名称: update_seq
 **功    能: 更新序列号
 **输入参数:
 **     secid: 段编号
 **     seq: 当前需要申请的序列号
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.11 23:54:14 #
 ******************************************************************************/
func (ctx *SeqSvrCntx) update_seq(secid uint64, seq uint64) error {
	section := ctx.ctrl.section[secid/SECTION_LIST_LEN]

	section.RLock()
	defer section.RUnlock()
	item, ok := section.item[secid]
	if !ok {
		return errors.New(fmt.Sprintf("Didn't find item! secid:%d", secid))
	}

	item.Lock()
	defer item.Unlock()

AGAIN:
	if item.max < seq {
		min, max, err := ctx.alloc_seq_from_db(secid)
		if nil != err {
			return err
		}
		if item.max < max {
			item.min = min // 更新最小值
			item.max = max // 更新最大值
		}
		goto AGAIN
	}

	return nil
}

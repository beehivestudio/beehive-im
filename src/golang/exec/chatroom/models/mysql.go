package models

import (
	"database/sql"
	"fmt"
	"time"

	"beehive-im/src/golang/lib/mesg"
)

/******************************************************************************
 **函数名称: RoomAdd
 **功    能: 添加聊天室
 **输入参数:
 **     rid: 聊天室ID
 **     req: 创建请求
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.05.29 20:37:40 #
 ******************************************************************************/
func RoomAdd(db *sql.DB, rid uint64, req *mesg.MesgRoomCreat) error {
	/* > 准备SQL语句 */
	sql := fmt.Sprintf(`
    INSERT INTO
        CHAT_ROOM_INFO_TAB(
            rid, name, status, description,
            create_time, update_time, owner)
    VALUES(?, ?, ?, ?, ?, ?, ?)`)

	stmt, err := db.Prepare(sql)
	if nil != err {
		return err
	}

	defer stmt.Close()

	/* > 执行SQL语句 */
	_, err = stmt.Exec(rid, req.GetName(), ROOM_STAT_OPEN,
		req.GetDesc(), time.Now().Unix(), time.Now().Unix(), req.GetUid())
	if nil != err {
		return err
	}

	return nil
}

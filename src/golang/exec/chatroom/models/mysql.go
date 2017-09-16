package models

import (
	"database/sql"
	"fmt"
	"time"

	"beehive-im/src/golang/lib/dbase"
	"beehive-im/src/golang/lib/mesg"
)

type RoomDbObj struct {
	mysql *sql.DB
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化
 **输入参数:
 **     addr: 地址
 **     pwd: 密码
 **输出参数: NONE
 **返    回: 错误描述
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.09.16 09:55:43 #
 ******************************************************************************/
func (db *RoomDbObj) Init(usr string, pwd string, addr string, dbname string) (err error) {
	auth := dbase.MySqlAuthStr(usr, pwd, addr, dbname)

	db.mysql, err = sql.Open("mysql", auth)
	if nil != err {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

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
func (db *RoomDbObj) RoomAdd(rid uint64, req *mesg.MesgRoomCreat) error {
	/* > 准备SQL语句 */
	sql := fmt.Sprintf(`
    INSERT INTO
        CHAT_ROOM_INFO_TAB(
            rid, name, status, description,
            create_time, update_time, owner)
    VALUES(?, ?, ?, ?, ?, ?, ?)`)

	stmt, err := db.mysql.Prepare(sql)
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

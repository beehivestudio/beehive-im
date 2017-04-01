package dbase

import (
	"fmt"
)

/******************************************************************************
 **函数名称: MySqlAuthStr
 **功    能: 获取MYSQL鉴权字串
 **输入参数:
 **     usr: 用户名
 **     pwd: 登录密码
 **     addr: IP+端口(格式:127.0.0.1:1250)
 **     dbname: 数据库名
 **输出参数: NONE
 **返    回: MYSQL鉴权字串
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.04.01 12:03:37 #
 ******************************************************************************/
func MySqlAuthStr(usr string, pwd string, addr string, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", usr, pwd, addr, dbname)
}

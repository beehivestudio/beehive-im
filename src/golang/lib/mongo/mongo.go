package mongo

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

/* 连接池对象 */
type Pool struct {
	session *mgo.Session /* 连接对象 */
}

/******************************************************************************
 **函数名称: CreatePool
 **功    能: 创建连接池
 **输入参数:
 **     addr: ${IP}:${PORT}
 **     passwd: 登录密码
 **输出参数: NONE
 **返    回: 连接池对象
 **实现描述:
 **注意事项: 连接子串的格式如下:
 **         "mongodb://${user}:${pwd}@${host:port},${host:port},.../${dbname}?${options}"
 **         示例如下:
 **         "mongodb://10.110.98.193:26408/admin?maxPoolSize=1000"
 **         "mongodb://10.110.98.193:26408,10.110.98.196:26408/admin?maxPoolSize=1000"
 **         "mongodb://admin:ZGY3ZWVkZDIxNTc@10.110.98.193:26408/admin?maxPoolSize=1000"
 **作    者: # Qifeng.zou # 2017.06.12 23:28:27 #
 ******************************************************************************/
func CreatePool(addr string, usr string, pwd string, dbname string, options string, timeout time.Duration) (*Pool, error) {
	str := "mongodb://"
	if 0 != len(usr) && 0 != len(pwd) {
		str += fmt.Sprintf("%s:%s@", usr, pwd)
	}
	str += fmt.Sprintf("%s", addr)
	if 0 != len(dbname) {
		str += fmt.Sprintf("/%s", dbname)
	}

	if 0 != len(options) {
		str += fmt.Sprintf("?%s", options)
	}

	session, err := mgo.DialWithTimeout(str, timeout)
	if nil != err {
		return nil, err
	}

	return &Pool{session: session}, nil
}

/******************************************************************************
 **函数名称: Get
 **功    能: 获取连接
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 连接对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.12 23:28:27 #
 ******************************************************************************/
func (pool *Pool) Get() *mgo.Session {
	return pool.session.Clone()
}

/******************************************************************************
 **函数名称: Exec
 **功    能: 执行MONGO操作
 **输入参数:
 **     db: 数据库名
 **     cn: collection名
 **     cb: 处理回调
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.12 23:28:27 #
 ******************************************************************************/
func (pool *Pool) Exec(db string, cn string, cb func(*mgo.Collection) error) error {
	session := pool.Get()
	defer session.Close()

	c := session.DB(db).C(cn)

	return cb(c)
}

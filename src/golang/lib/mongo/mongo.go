package mongo

import (
	"labix.org/v2/mgo"
)

/* 连接池对象 */
type Pool struct {
	session *mgo.Session /* 连接对象 */
}

/******************************************************************************
 **函数名称: CreatePool
 **功    能: 创建连接池
 **输入参数:
 **     addr: IP地址
 **     passwd: 登录密码
 **输出参数: NONE
 **返    回: 连接池对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.12 23:28:27 #
 ******************************************************************************/
func CreatePool(addr string, passwd string) (*Pool, error) {
	session, err := mgo.Dial(addr)
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
 **函数名称: Collection
 **功    能: 回收连接
 **输入参数:
 **     db: 数据库名
 **     c: collection名
 **     cb: 处理回调
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.06.12 23:28:27 #
 ******************************************************************************/
func (pool *Pool) Exec(db string, c string, cb func(*mgo.Collection) error) error {
	session := pool.Get()
	defer session.Close()

	coll := session.DB(db).C(c)

	return cb(coll)
}

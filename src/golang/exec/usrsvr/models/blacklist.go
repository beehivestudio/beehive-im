package models

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"

	"beehive-im/src/golang/lib/mongo"
)

/******************************************************************************
 **函数名称: DbBlacklistAdd
 **功    能: 添加黑名单(同步到数据库...)
 **输入参数:
 **     mongo: 配置信息
 **     dbname: 配置信息
 **     uid: 配置信息
 **     bl_uid: 配置信息
 **输出参数: NONE
 **返    回: 错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.08.12 13:47:44 #
 ******************************************************************************/
func DbBlacklistAdd(mongo *mongo.Pool, dbname string, uid uint64, bl_uid uint64) error {
	blacklist := &BlacklistTabRow{
		Uid:  uid,               // 用户ID
		Buid: bl_uid,            // 被加入黑名单的用户ID
		Ctm:  time.Now().Unix(), // 设置时间
	}

	cb := func(c *mgo.Collection) (err error) {
		c.Insert(blacklist)
		return err
	}

	return mongo.Exec(dbname, TAB_BLACKLIST, cb)
}

/******************************************************************************
 **函数名称: RdsBlacklistAdd
 **功    能: 添加黑名单(同步到Redis缓存...)
 **输入参数:
 **     redis: REDIS对象
 **     uid: 用户UID
 **     bl_uid: 配置信息
 **输出参数: NONE
 **返    回: 错误码+错误信息
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.08.15 18:05:49 #
 ******************************************************************************/
func RdsBlacklistAdd(redis *redis.Pool, uid uint64, bl_uid uint64) (code uint32, err error) {
	rds := redis.Get()
	defer rds.Close()

	ctm := time.Now().Unix()

	/* > 加入用户黑名单(缓存) */
	key := fmt.Sprintf(comm.CHAT_KEY_USR_BLACKLIST_TAB, uid)

	_, err = rds.Do("HSET", key, bl_uid, ctm)
	if nil != err {
		return comm.ERR_SYS_SYSTEM, err
	}

	return comm.OK, err
}

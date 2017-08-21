package models

/* 数据库表名定义 */
const (
	TAB_BLACKLIST = "BlackList"
)

/* 用户黑名单 */
type BlacklistTabRow struct {
	Uid  uint64 "uid"  // 用户ID
	Buid uint64 "buid" // 被加入UID黑名单列表的用户UID
	Ctm  int64  "ctm"  // 设置时间
}

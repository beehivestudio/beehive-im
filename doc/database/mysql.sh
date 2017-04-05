mysql -uroot -p111111 -e "
drop database if exists testdb;
create database testdb;
use testdb;

# 创建会话SID生成表
CREATE TABLE IF NOT EXISTS IM_SID_GEN_TAB(
    id bigint(20) NOT NULL AUTO_INCREMENT,
    type int(20) NOT NULL default 0,
    sid bigint(20) NOT NULL default 1,
    UNIQUE(id),
    PRIMARY KEY(type)
    );
quit"

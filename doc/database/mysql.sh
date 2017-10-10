#!/bin/sh

USER=root
PASSWD=111111
DBNAME=testdb

mysql -u$USER -p$PASSWD -e "
DROP DATABASE IF EXISTS $DBNAME;
CREATE DATABASE $DBNAME;
USE $DBNAME;

# 创建会话SID生成表
CREATE TABLE IF NOT EXISTS IM_SID_GEN_TAB(
    type tinyint NOT NULL DEFAULT 0 COMMENT '会话类型',
    sid bigint NOT NULL DEFAULT 1 COMMENT '会话ID',

    PRIMARY KEY(type)
    );

# 创建序号SEQ生成表
CREATE TABLE IF NOT EXISTS IM_SEQ_GEN_TAB(
    id bigint NOT NULL DEFAULT 0 COMMENT '段编号',
    seq bigint NOT NULL DEFAULT 100 COMMENT '序列号最新值',

    PRIMARY KEY(id)
    );

################################################################################
## 聊天室

# 房间信息表
DROP TABLE IF EXISTS CHAT_ROOM_INFO_TAB;

CREATE TABLE IF NOT EXISTS CHAT_ROOM_INFO_TAB(
    rid bigint NOT NULL AUTO_INCREMENT COMMENT '房间ID[主键]',
    name varchar(64) NOT NULL COMMENT '房间名称',
    type tinyint NOT NULL DEFAULT 0 COMMENT '房间类型',
    level tinyint NOT NULL DEFAULT 0 COMMENT '房间级别',
    owner bigint NOT NULL COMMENT '房主UID',
    status tinyint NOT NULL DEFAULT 0 COMMENT '房间状态',
    image varchar(1024) NOT NULL COMMENT '房间封面',
    description varchar(256) NOT NULL COMMENT '房间描述',
    create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
    update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',

    PRIMARY KEY(rid),
    INDEX(owner)
    );

# 创建聊天室RID生成表
CREATE TABLE IF NOT EXISTS IM_RID_GEN_TAB(
    id tinyint NOT NULL DEFAULT 0 COMMENT '编号 -- 无实际意义',
    rid bigint NOT NULL DEFAULT 1 COMMENT '聊天室ID',

    PRIMARY KEY(id)
    );
quit"

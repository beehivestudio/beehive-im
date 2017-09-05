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
CREATE TABLE IF NOT EXISTS CHAT_ROOM_INFO_TAB(
    rid bigint NOT NULL AUTO_INCREMENT COMMENT '房间ID[主键]',
    name varchar(64) NOT NULL COMMENT '房间名称',
    type tinyint NOT NULL DEFAULT 0 COMMENT '房间类型',
    level tinyint NOT NULL DEFAULT 0 COMMENT '房间级别',
    owner bigint NOT NULL COMMENT '房主UID',
    status tinyint NOT NULL DEFAULT 0 COMMENT '房间状态',
    image varchar(1024) NOT NULL COMMENT '房间封面',
    desc varchar(256) NOT NULL COMMENT '房间描述',
    create_time bigint NOT NULL DEFAULT unix_timestamp() COMMENT '创建时间',
    update_time bigint NOT NULL DEFAULT unix_timestamp() COMMENT '更新时间',

    PRIMARY KEY(rid)
    INDEX(owner)
    );

# 创建聊天室RID生成表
CREATE TABLE IF NOT EXISTS IM_RID_GEN_TAB(
    id tinyint NOT NULL DEFAULT 0 COMMENT '编号 -- 无实际意义',
    rid bigint NOT NULL DEFAULT 1 COMMENT '聊天室ID',

    PRIMARY KEY(id)
    );


quit"

idx=0
for (( idx=0; idx<256; idx+=1 ))
do
    mysql -u$USER -p$PASSWD -e "
    use $DBNAME;

    # 创建USER信息表
    DROP TABLE IF EXISTS IM_USER_TAB_$idx;

    CREATE TABLE IF NOT EXISTS IM_USER_TAB_$idx(
        uid bigint NOT NULL DEFAULT 0 COMMENT '用户UID',
        name varchar(32) NOT NULL COMMENT '真实姓名',
        nickname varchar(32) COMMENT '昵称',
        gender tinyint NOT NULL DEFAULT 0 COMMENT '性别',
        passwd varchar(64) NOT NULL COMMENT '登陆密码(MD5)',
        image varchar(1024) COMMENT '头像',
        level tinyint NOT NULL DEFAULT 0 COMMENT '用户级别',
        status tinyint NOT NULL DEFAULT 0 COMMENT '用户状态(0:正常 1:禁言 2:黑名单 3:锁定)',
        create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
        update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
        role tinyint NOT NULL DEFAULT 0 COMMENT '用户角色',
        tel varchar(32) COMMENT '电话',
        email varchar(32) COMMENT '邮箱',
        identify varchar(32) COMMENT '身份证号',
        nation int COMMENT '国家ID',
        city int COMMENT '地市ID',
        town int COMMENT '城镇ID',
        addr varchar(128) COMMENT '详细地址',

        PRIMARY KEY(uid),
        UNIQUE(tel),
        UNIQUE(email),
        UNIQUE(identify)
        );

    # 创建ROOM信息表
    DROP TABLE IF EXISTS CHAT_ROOM_INFO_$idx;

    CREATE TABLE IF NOT EXISTS CHAT_ROOM_INFO_$idx(
        rid bigint NOT NULL DEFAULT 0 COMMENT '聊天室ID',
        name varchar(32) NOT NULL COMMENT '房间名称',
        image varchar(1024) COMMENT '房间图片',
        status tinyint NOT NULL DEFAULT 0 COMMENT '房间状态(0:正常 1:关闭)',
        user_num int NOT NULL DEFAULT 0 COMMENT '房间人数',
        create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
        update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
        type tinyint NOT NULL DEFAULT 0 COMMENT '房间类型',
        description varchar(256) NOT NULL COMMENT '房间描述',
        owner bigint NOT NULL DEFAULT 0 COMMENT '房主UID',
        level tinyint NOT NULL DEFAULT 0 COMMENT '房间级别',

        PRIMARY KEY(rid),
        INDEX(owner)
        );

    # 创建GROUP信息表
    CREATE TABLE IF NOT EXISTS CHAT_GROUP_INFO_$idx(
        gid bigint NOT NULL DEFAULT 0 COMMENT '群组ID',
        name varchar(32) NOT NULL COMMENT '群组名称',
        type tinyint NOT NULL DEFAULT 0 COMMENT '群组类型',
        user_num int NOT NULL DEFAULT 0 COMMENT '群组人数',
        image varchar(1024) COMMENT '群组图片',
        create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
        update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
        owner bigint NOT NULL DEFAULT 0 COMMENT '群组UID',
        level tinyint NOT NULL DEFAULT 0 COMMENT '群组级别',
        description varchar(256) NOT NULL COMMENT '群组描述',
        status tinyint NOT NULL DEFAULT 0 COMMENT '群组状态',

        PRIMARY KEY(gid),
        INDEX(owner)
        );

    # 创建ROOM用户表
    CREATE TABLE IF NOT EXISTS CHAT_ROOM_USER_INFO_$idx(
        uid bigint NOT NULL DEFAULT 0 COMMENT '房主UID',
        rid bigint NOT NULL DEFAULT 0 COMMENT '房间RID',
        create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
        update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
        status tinyint NOT NULL DEFAULT 0 COMMENT '用户在房间状态(0:正常 1:黑名单 2:禁言)',
        role tinyint NOT NULL DEFAULT 0 COMMENT '用户在房间角色',

        PRIMARY KEY(uid, rid)
        );

    # 创建GROUP用户表
    CREATE TABLE IF NOT EXISTS CHAT_GROUP_USER_INFO_$idx(
        uid bigint NOT NULL DEFAULT 0 COMMENT '群组ID',
        gid bigint NOT NULL DEFAULT 0 COMMENT '群组ID',
        create_time bigint NOT NULL DEFAULT 0 COMMENT '创建时间',
        update_time bigint NOT NULL DEFAULT 0 COMMENT '更新时间',
        status tinyint NOT NULL DEFAULT 0 COMMENT '用户在群组状态(0:正常 1:黑名单 2:禁言)',
        role tinyint NOT NULL DEFAULT 0 COMMENT '用户在群组角色',

        PRIMARY KEY(uid, gid)
        );
    quit"
done

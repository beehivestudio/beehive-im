#!/bin/sh

USER=root
PASSWD=111111
DBNAME=testdb

mysql -u$USER -p$PASSWD -e "
drop database if exists testdb;
create database testdb;
use $DBNAME;

# 创建会话SID生成表
CREATE TABLE IF NOT EXISTS IM_SID_GEN_TAB(
    id bigint NOT NULL AUTO_INCREMENT,
    type tinyint NOT NULL default 0,
    sid bigint NOT NULL default 1,

    UNIQUE(id),
    PRIMARY KEY(type)
    );

# 创建序号SEQ生成表
CREATE TABLE IF NOT EXISTS IM_SEQ_GEN_TAB(
    id bigint NOT NULL default 0,      # 段编号
    seq bigint NOT NULL default 100,   # 序列号最新值

    PRIMARY KEY(id)
    );
quit"

idx=0
for (( idx=0; idx<256; idx+=1 ))
do
    mysql -u$USER -p$PASSWD -e "
    use $DBNAME;

    # 创建USER信息表
    CREATE TABLE IF NOT EXISTS IM_USER_TAB-$idx(
        uid bigint NOT NULL default 0,                  # 用户ID
        name varchar(32) NOT NULL,                      # 真实姓名
        nickname varchar(32),                           # 昵称
        gender tinyint NOT NULL default 0,              # 性别
        passwd varchar(64) NOT NULL,                    # 登陆密码(MD5)
        image varchar(1024),                            # 头像
        level tinyint NOT NULL default 0,               # 用户级别
        status tinyint NOT NULL default 0,              # 用户状态(0:正常 1:禁言 2:黑名单 3:锁定)
        create_time timestamp NOT NULL default 0,       # 创建时间
        update_time timestamp NOT NULL default 0,       # 更新时间
        role tinyint NOT NULL default 0,                # 用户角色
        tel varchar(32),                                # 电话
        email varchar(32),                              # 邮箱
        identify varchar(32),                           # 身份证号
        nation int,                                     # 国家ID
        city int,                                       # 地市ID
        town int,                                       # 城镇ID
        addr varchar(128),                              # 详细地址

        PRIMARY KEY(uid),
        UNIQUE(tel),
        UNIQUE(email),
        UNIQUE KEY(identify)
        );

    # 创建ROOM信息表
    CREATE TABLE IF NOT EXISTS CHAT_ROOM_INFO-$idx(
        rid bigint NOT NULL default 0,                  # 聊天室ID
        name varchar(32) NOT NULL,                      # 房间名称
        image varchar(1024),                            # 房间图片
        status tinyint NOT NULL default 0,              # 房间状态(0:正常 1:禁言 2:黑名单 3:锁定)
        user_num int NOT NULL default 0,                # 房间人数
        create_time timestamp NOT NULL default 0,       # 创建时间
        update_time timestamp NOT NULL default 0,       # 更新时间
        type tinyint NOT NULL default 0,                # 房间类型
        desc varchar(256) NOT NULL,                     # 房间描述
        owner bigint NOT NULL default 0,                # 房间UID
        level tinyint NOT NULL default 0,               # 房间级别

        PRIMARY KEY(rid)
        );

    # 创建GROUP信息表
    CREATE TABLE IF NOT EXISTS CHAT_GROUP_INFO-$idx(
        gid bigint NOT NULL default 0,                  # 群组ID
        name varchar(32) NOT NULL,                      # 群组名称
        type tinyint NOT NULL default 0,                # 群组类型
        user_num int NOT NULL default 0,                # 群组人数
        image varchar(1024),                            # 群组图片
        create_time timestamp NOT NULL default 0,       # 创建时间
        update_time timestamp NOT NULL default 0,       # 更新时间
        owner bigint NOT NULL default 0,                # 群组UID
        level tinyint NOT NULL default 0,               # 群组级别
        desc varchar(256) NOT NULL,                     # 群组描述
        status tinyint NOT NULL default 0,              # 群组状态

        PRIMARY KEY(gid)
        );

    # 创建ROOM用户表
    CREATE TABLE IF NOT EXISTS CHAT_ROOM_USER_INFO-$idx(
        uid bigint NOT NULL default 0,                  # 房间ID
        rid bigint NOT NULL default 0,                  # 房间ID
        create_time timestamp NOT NULL default 0,       # 创建时间
        update_time timestamp NOT NULL default 0,       # 更新时间
        status tinyint NOT NULL default 0,              # 用户在房间状态(0:正常 1:黑名单 2:禁言)
        role tinyint NOT NULL default 0,                # 用户在房间角色

        PRIMARY KEY(uid, rid)
        );

    # 创建GROUP用户表
    CREATE TABLE IF NOT EXISTS CHAT_GROUP_USER_INFO-$idx(
        uid bigint NOT NULL default 0,                  # 群组ID
        gid bigint NOT NULL default 0,                  # 群组ID
        create_time timestamp NOT NULL default 0,       # 创建时间
        update_time timestamp NOT NULL default 0,       # 更新时间
        status tinyint NOT NULL default 0,              # 用户在群组状态(0:正常 1:黑名单 2:禁言)
        role tinyint NOT NULL default 0,                # 用户在群组角色

        PRIMARY KEY(uid, gid)
        );
    quit"
done

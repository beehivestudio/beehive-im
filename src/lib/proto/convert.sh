#!/bin/sh

mkdir -p ../../incl/proto

protoc-c --c_out=. mesg_room.proto
mv mesg_room.pb-c.h ../../incl/proto

protoc-c --c_out=. mesg_online.proto
mv mesg_online.pb-c.h ../../incl/proto

protoc-c --c_out=. mesg_online_ack.proto
mv mesg_online_ack.pb-c.h ../../incl/proto

protoc-c --c_out=. mesg_join_ack.proto
mv mesg_join_ack.pb-c.h ../../incl/proto

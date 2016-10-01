#!/bin/sh

mkdir -p ../../incl/proto

protoc-c --c_out=. chat_room_mesg.proto
mv chat_room_mesg.pb-c.h ../../incl/proto

protoc-c --c_out=. chat_online_mesg.proto
mv chat_online_mesg.pb-c.h ../../incl/proto

protoc-c --c_out=. chat_online_ack_mesg.proto
mv chat_online_ack_mesg.pb-c.h ../../incl/proto

protoc-c --c_out=. chat_join_ack_mesg.proto
mv chat_join_ack_mesg.pb-c.h ../../incl/proto

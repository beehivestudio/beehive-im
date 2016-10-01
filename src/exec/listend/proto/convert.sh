#!/bin/sh

protoc-c --c_out=. chat_room_mesg.proto
protoc-c --c_out=. chat_online_mesg.proto
protoc-c --c_out=. chat_online_ack_mesg.proto
protoc-c --c_out=. chat_join_ack_mesg.proto

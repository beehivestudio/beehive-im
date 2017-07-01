#!/bin/sh

rm -fr ../log/*

redis-server ../conf/redis.conf &

./frwder.v.1.1 -c ../conf/frwder.xml -d -l debug
./seqsvr.v.1.1 &
./msgsvr.v.1.1 &
./tasker.v.1.1 &
./usrsvr.v.1.1 &
./monitor.v.1.1 &
./listend.v.1.1 -c ../conf/listend.xml -d -l debug
./listend-ws.v.1.1 &

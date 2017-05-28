#!/bin/sh

redis-server ../conf/redis.conf &

./frwder.v.1.1 -c ../conf/frwder.xml -d -l error
./seqsvr.v.1.1 &
./msgsvr.v.1.1 &
./tasker.v.1.1 &
./usrsvr.v.1.1 &
./monitor.v.1.1 &
./listend.v.1.1 -c ../conf/listend.xml -d -l error
./listend-ws.v.1.1 -c ../conf/websocket.xml -d -l error

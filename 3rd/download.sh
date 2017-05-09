#!/bin/sh

root="$GOPATH/src"

LIBS[0]="github.com/gorilla/websocket"
LIBS[1]="github.com/astaxie/beego/logs"
LIBS[2]="github.com/golang/protobuf/proto"
LIBS[3]="github.com/garyburd/redigo/redis"

echo ${LIBS[*]}

for item in ${LIBS[*]}
do
    if [ ! -e "$root/$item" ]; then
        echo "go get $item"
        go get $item
    else
        echo "$root/$item exists!"
    fi
done

#!/bin/bash

function random()
{
    min=$1;
    max=`expr $2 - $1`;
    num=$(date +%s%N);
    ret=`expr $num % $max + $min`;
    echo $ret
}

idx=0
ip_list=(
127.0.0.1)
ip_num=1
rid=$1
uid=96439818
begin=`date +%H%M%S.%N`

while true
do
    addtime=`date +%s`
    idx=`expr $idx + 1`
    str=`date +%H%M%S.%N`;
    len=$(random 1 2);
    ch=`expr $idx % 10`
    data="额...$idx"
    echo "len: "$len
    for (( i=0; i<$len; i++ ))
    do
        data=$data"x"
    done
    pos=`expr $pos + 1`
    pos=`expr $pos % 10`
    uid=`expr $uid + 1`
    mesg="{\"roomId\":\"$rid\",\"type\":1,\"message\":\"$data\",\"color\":\"\",\"font\":\"m\",\"position\":\"$pos\",\"addtime\":$addtime,\"ip\":\"10.58.100.157\",\"service\":{\"forhost\":false},\"user\":{\"uid\":\"$uid\",\"nickname\":\"186xxxxxx2324_222\",\"icon\":\"http:\/\/i2.letvimg.com\/lc01_user\/201510\/26\/17\/56\/14458533794163_50_50.jpg\",\"vip\":0,\"role\":2},\"msgid\":\"569e051ff54cbf8d6c8b4568\", \"showtime\":0}";
    mod=`expr $idx % $ip_num`
    gid=0
    echo $gid
    if [ $mod -gt 0 ]; then
        curl "http://${ip_list[$mod]}:8000/im/push?dim=room&rid=$rid&expire=$expire" -d "$mesg"
    else
        curl "http://${ip_list[$mod]}:8000/im/push?dim=room&rid=$rid&expire=$expire" -d "$mesg"
    fi

    echo $mesg
    sleep 1
done



#!/bin/sh

LIST="room/mesg_room.proto"
LIST=$LIST" online/mesg_online.proto"
LIST=$LIST" online_ack/mesg_online_ack.proto"
LIST=$LIST" join/mesg_join.proto"
LIST=$LIST" join_ack/mesg_join_ack.proto"
LIST=$LIST" unjoin/mesg_unjoin.proto"

for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc --go_out=. ${ITEM};
        INCL=`echo ${ITEM} | awk -F. '{print $1}'`;
    fi
done

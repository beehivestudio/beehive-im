#!/bin/sh

LIST="mesg_room.proto"
LIST=$LIST" mesg_online.proto"
LIST=$LIST" mesg_join.proto"
LIST=$LIST" mesg_unjoin.proto"
LIST=$LIST" mesg_hb.proto"

for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc --go_out=. ${ITEM};
        INCL=`echo ${ITEM} | awk -F. '{print $1}'`;
    fi
done

#!/bin/sh

LIST="mesg.proto"

for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc --go_out=. ${ITEM};
    fi
done

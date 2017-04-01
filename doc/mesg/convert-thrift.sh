#!/bin/sh

LIST="mesg.thrift"

GOLANG_MESG_DIR="../../src/golang/lib/mesg"
for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        thrift  -r --gen go ${ITEM}
        cp -fr ./gen-go/* ${GOLANG_MESG_DIR}
    fi
done
rm -fr gen-go

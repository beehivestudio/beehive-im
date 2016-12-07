#!/bin/sh

INCL_DIR=../../incl/mesg
mkdir -p $INCL_DIR

LIST="mesg.proto"

for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc-c --c_out=. ${ITEM};

        INCL=`echo ${ITEM} | awk -F. '{print $1}'`;

        echo "mv ${INCL}.pb-c.h ${INCL_DIR}";
        mv "${INCL}.pb-c.h" ${INCL_DIR};
    fi
done

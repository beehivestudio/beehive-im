#!/bin/sh

LIST="mesg.proto"

################################################################################
GOLANG_MESG_DIR=../../src/golang/lib/mesg
for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc --go_out=. ${ITEM};
        mv "mesg.pb.go" ${GOLANG_MESG_DIR};
    fi
done

################################################################################
CLANG_SRC_DIR=../../src/clang/lib/mesg
CLANG_INCL_DIR=../../src/clang/incl/mesg
mkdir -p $CLANG_SRC_DIR
mkdir -p $CLANG_INCL_DIR

for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc-c --c_out=. ${ITEM};

        echo "mv mesg.pb-c.h ${CLANG_INCL_DIR}";
        mv "mesg.pb-c.c" ${CLANG_SRC_DIR};
        mv "mesg.pb-c.h" ${CLANG_INCL_DIR};
    fi
done

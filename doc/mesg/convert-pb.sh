#!/bin/sh

LIST="mesg.proto"

################################################################################
GOLANG_MESG_DIR=../../src/golang/lib/mesg
for ITEM in $LIST;
do
    echo ${ITEM}
    if [ -e ${ITEM} ]; then
        protoc --go_out=. ${ITEM};
        cp -fr "mesg.pb.go" ${GOLANG_MESG_DIR};
        rm -fr "mesg.pb.go";
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

        echo "cp -fr mesg.pb-c.h ${CLANG_INCL_DIR}";
        cp -fr "mesg.pb-c.c" ${CLANG_SRC_DIR};
        rm -fr "mesg.pb-c.c";
        cp -fr "mesg.pb-c.h" ${CLANG_INCL_DIR};
        rm -fr "mesg.pb-c.h";
    fi
done

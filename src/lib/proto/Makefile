###############################################################################
## Copyright(C) 2016-2025 Letv technology Co., Ltd
##
## 文件名: Makefile
## 版本号: 1.0
## 描  述: 核心模块的编译
## 作  者: # Qifeng.zou # 2016.09.20 19:40:44 #
###############################################################################
include $(PROJ)/make/build.mak

INCLUDE = -I. \
		  -I$(PROJ)/src/incl \
		  -I$(PROJ)/src/incl/proto
LIBS_PATH = -L$(PROJ)/lib -L$(PROJ)/../clang/lib
# 静态链接库
STATIC_LIB_LIST = libcore.a
LIBS = $(call func_find_static_link_lib,$(STATIC_LIB_PATH),$(STATIC_LIB_LIST))
LIBS += -lpthread -lm -dl -lprotobuf-c
LIBS += $(SHARED_LIB)

SRC_LIST = chat_room_mesg.pb-c.c \
			chat_online_mesg.pb-c.c \
			chat_online_ack_mesg.pb-c.c \
			chat_join_ack_mesg.pb-c.c

OBJS = $(subst .c,.o, $(SRC_LIST))
HEADS = $(call func_get_dep_head_list, $(SRC_LIST))

TARGET = libchat_mesg_proto.a

.PHONY: all clean

all: $(TARGET)
$(TARGET): $(OBJS)
	@$(AR) $(AFLAGS) -s $@ $(OBJS)
	@echo "AR $@"
	@mv $@ $(PROJ_LIB)

$(OBJS): %.o : %.c $(HEADS)
	@$(CC) $(CFLAGS) -c $< -o $@ $(INCLUDE)
	@echo "CC $(PWD)/$<"

clean:
	@rm -fr *.o $(PROJ_LIB)/$(TARGET)
	@echo "rm -fr *.o $(PROJ_LIB)/$(TARGET)"

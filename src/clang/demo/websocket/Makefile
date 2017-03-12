###############################################################################
## Copyright(C) 2015-2025 Letv technology Co., Ltd
##
## 文件名: Makefile
## 版本号: 1.0
## 描  述: 接入层模块
## 作  者: # Qifeng.zou # 2015.11.03 14:52:05 #
###############################################################################
include $(PROJ)/make/build.mak

INCLUDE = -I. \
			-I./incl \
			-I$(PROJ)/src/clang/incl \
			-I$(PROJ)/../cctrl/src/incl \
			-I$(PROJ)/src/clang/incl/mesg \
			-I/usr/local/include
LIBS_PATH = -L$(PROJ)/lib

# 注: 静态库请放在动态库之前
STATIC_LIB_LIST = libev.a libcore.a libmesg.a
LIBS = $(call func_find_static_link_lib,$(STATIC_LIB_PATH),$(STATIC_LIB_LIST))
LIBS += -lpthread -lm -ldl -lwebsockets -lprotobuf-c

SRC_LIST = websocket-client.c

OBJS = $(subst .c,.o, $(SRC_LIST))
HEADS = $(call func_get_dep_head_list, $(SRC_LIST))

TARGET = websocket-client

.PHONY: all clean

all: $(TARGET)
$(TARGET): $(OBJS)
	@$(CC) $(CFLAGS) -o $@ $(OBJS) $(INCLUDE) $(LIBS_PATH) $(LIBS)
	@echo "CC $@"
	@mv $@ $(PROJ_BIN)/$@
	@echo "$@ is OK!"

$(OBJS): %.o : %.c $(HEADS)
	@$(CC) $(CFLAGS) -c $< -o $@ $(INCLUDE)
	@echo "CC $(PWD)/$<"

clean:
	@rm -fr *.o $(PROJ_BIN)/$(TARGET)
	@echo "rm -fr *.o $(PROJ_BIN)/$(TARGET)"

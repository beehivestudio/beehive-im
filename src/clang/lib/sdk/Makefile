###############################################################################
## Copyright(C) 2013-2014 Qiware technology Co., Ltd
##
## 文件名: Makefile
## 版本号: 1.0
## 描  述: 客户端SDK模块的编译
## 作  者: # Qifeng.zou # 2016.11.13 16:25:25 #
###############################################################################
include $(PROJ)/make/build.mak

INCLUDE = -I. \
		  -I$(PROJ)/src/clang/incl \
		  -I$(PROJ)/src/clang/incl/sdk \
		  -I$(PROJ)/src/clang/incl/mesg \
		  -I$(PROJ)/../cctrl/src/incl 
INCLUDE += $(GLOBAL_INCLUDE)
LIBS_PATH = -L$(PROJ)/lib

# 注: 静态库请放在动态库之前
STATIC_LIB_LIST = libcore.a
LIBS = $(call func_find_static_link_lib,$(STATIC_LIB_PATH),$(STATIC_LIB_LIST))
LIBS += -lpthread -lm -dl -lcjson
LIBS += $(SHARED_LIB)

SRC_LIST = sdk.c \
		   sdk_comm.c \
		   sdk_ssvr.c \
		   sdk_mesg.c \
		   sdk_worker.c

OBJS = $(subst .c,.o, $(SRC_LIST))
HEADS = $(call func_get_dep_head_list, $(SRC_LIST))

TARGET = libsdk.so
TARGET_A = libsdk.a

.PHONY: all clean

all: $(TARGET) $(TARGET_A)
$(TARGET): $(OBJS)
	@$(CC) $(LFLAGS) -o $@ $(OBJS) $(INCLUDE) $(LIBS_PATH) $(LIBS)
	@echo "CC $@"
	@mv $@ $(PROJ_LIB)
	@echo "$@ is OK!"

$(TARGET_A): $(OBJS)
	@$(AR) $(AFLAGS) -s $@ $(OBJS)
	@echo "AR $@"
	@mv $@ $(PROJ_LIB)
	@echo "$@ is OK!"

$(OBJS): %.o : %.c $(HEADS)
	@$(CC) $(CFLAGS) -c $< -o $@ $(INCLUDE)
	@echo "CC $(PWD)/$<"

clean:
	@rm -fr *.o
	@echo "rm -fr *.o"

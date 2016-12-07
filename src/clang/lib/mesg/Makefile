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
		  -I$(PROJ)/src/clang/incl \
		  -I$(PROJ)/src/clang/incl/mesg
LIBS_PATH = -L$(PROJ)/lib
LIBS = -lpthread -lm -dl -lprotobuf-c
LIBS += $(SHARED_LIB)

SRC_LIST = mesg.pb-c.c

OBJS = $(subst .c,.o, $(SRC_LIST))
HEADS = $(call func_get_dep_head_list, $(SRC_LIST))

TARGET = libmesg.a

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

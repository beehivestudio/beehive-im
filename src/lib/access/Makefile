###############################################################################
## Coypright(C) 2014-2024 Qiware technology Co., Ltd
##
## 文件名: Makefile
## 版本号: 1.0
## 描  述: 帧听模块
## 作  者: # Qifeng.zou # 2014.11.15 #
###############################################################################
include $(PROJ)/make/build.mak

INCLUDE = -I. \
			-I$(PROJ)/src/incl \
			-I$(PROJ)/src/incl/access
INCLUDE += $(GLOBAL_INCLUDE)
LIBS_PATH = -L$(PROJ)/lib
LIBS = -lpthread -lcore -lutils
LIBS += $(SHARED_LIB)

SRC_LIST = access.c \
			acc_comm.c \
			acc_lsn.c \
			acc_rsvr.c

OBJS = $(subst .c,.o, $(SRC_LIST))
HEADS = $(call func_get_dep_head_list, $(SRC_LIST))

TARGET = libaccess.so
TARGET_A = libaccess.a

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
	@rm -fr *.o $(PROJ_LIB)/$(TARGET) $(PROJ_LIB)/$(TARGET_A)
	@echo "rm -fr *.o $(PROJ_LIB)/$(TARGET) $(PROJ_LIB)/$(TARGET_A)"

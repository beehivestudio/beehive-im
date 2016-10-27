###############################################################################
## Copyright(C) 2014-2024 Qiware technology Co., Ltd
##
## 功    能: 遍历编译目录，并执行指定的操作
##			1. 编译操作
##			2. 删除操作
##			3. 重新编译 
## 注意事项: 
##			1. 当需要增加编译目录时, 请将目录加入变量DIR中, 不用修改该文件其他数据!
## 			2. 如果只想编译指定目录的代码, 则可执行命令:
## 				Make DIR=指定目录 如: Make DIR=src/lib/core
## 作    者: # Qifeng.zou # 2014.08.28 #
###############################################################################
include ./make/func.mak

export VERSION=v.1.1 # 版本号

# 根目录
export PROJ = ${PWD}
export PROJ_3RD = ${PROJ}/3rd
export PROJ_BIN = ${PROJ}/bin
export PROJ_LIB = ${PROJ}/lib
export PROJ_LOG = ${PROJ}/log
export PROJ_CONF = ${PROJ}/conf
export GCC_LOG = ${PROJ_LOG}/gcc.log

# 编译目录(注：编译按顺序执行　注意库之间的依赖关系)
CLANG_LIB_DIR = "src/clang/lib"
DIR += "$(CLANG_LIB_DIR)/chat"
DIR += "$(CLANG_LIB_DIR)/mesg"

CLANG_EXEC_DIR = "src/clang/exec"
DIR += "$(CLANG_EXEC_DIR)/frwder"
DIR += "$(CLANG_EXEC_DIR)/listend"

GOLANG_EXEC_DIR = "src/golang/exec"
DIR += "$(GOLANG_EXEC_DIR)/cross"
DIR += "$(GOLANG_EXEC_DIR)/olsvr"

# 获取系统配置
CPU_CORES = $(call func_cpu_cores)

.PHONY: all clean rebuild help

# 1. 编译操作
all:
	$(call func_mkdir)
	@for ITEM in ${DIR}; \
	do \
		echo $${ITEM}; \
		clang=`echo $${ITEM} | grep 'clang' | wc -l`; \
		golang=`echo $${ITEM} | grep 'golang' | wc -l`; \
		if [ $${clang} -eq 1 ]; then \
			if [ -e $${ITEM}/Makefile ]; then \
				cd $${ITEM}; \
				#make -j$(CPU_CORES) 2>&1 | tee -a ${GCC_LOG}; \
				make -j$(CPU_CORES) 2>&1 || exit; \
				cd ${PROJ}; \
			else \
				echo "File [$${ITEM}/Makefile] isn't exist!"; exit; \
			fi \
		elif [ $${golang} -eq 1 ]; then \
			if [ -e $${ITEM} ]; then \
				echo "cd $${ITEM}"; \
				cd $${ITEM}; \
				go build -gcflags "-N -l"; \
				EXEC=`basename \`pwd\``; \
				mv $${EXEC} $${PROJ_BIN}/$${EXEC}.${VERSION}; \
				cd ${PROJ}; \
			else \
				echo "Path [$${ITEM}] isn't exist!"; exit; \
			fi \
		fi \
	done

# 2. 清除操作
clean:
	@for ITEM in ${DIR}; \
	do \
		if [ -e $${ITEM}/Makefile ]; then \
			cd $${ITEM}; \
			make clean; \
			cd ${PROJ}; \
		fi \
	done

# 3. 重新编译 
rebuild: clean all

# 4. 显示帮助
help:
	@cat make/help.mak


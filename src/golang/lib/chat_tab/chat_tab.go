package chat_tab

import (
	"fmt"
	"strconv"
)

/* 会话信息 */
type ChatSession struct {
	sid uint64          // 会话ID
	rid uint64          // 聊天室ID
	gid uint32          // 分组ID
	sub map[uint32]bool // 订阅列表(true:订阅 false:未订阅)
}

/* 分组信息 */
type ChatGroup struct {
	gid       uint64          // 分组ID
	sid_num   uint64          // 会话数目
	create_tm int64           // 创建时间
	sid_list  map[uint64]bool // 会话列表
}

/* ROOM信息 */
type ChatRoom struct {
	rid       uint64               // 聊天室ID
	sid_num   uint64               // 会话数目
	grp_num   uint32               // 分组数目
	create_tm int64                // 创建时间
	group     map[uint32]ChatGroup // 分组信息
}

/* 全局对象 */
type ChatTab struct {
	room    map[uint64]ChatRoom    // ROOM集合
	session map[uint64]ChatSession // 会话集合
}

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 全局对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:46:50 #
 ******************************************************************************/
func Init() *ChatTab {
	ctx := &ChatTab{
		room:    make(map[uint64]ChatRoom),
		session: make(map[uint64]ChatSession),
	}

	return ctx
}

/******************************************************************************
 **函数名称: SessionAdd
 **功    能: 会话添加
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:50:17 #
 ******************************************************************************/
func (ctx *ChatTab) SessionAdd(rid uint64, gid uint32, sid uint64) int {
	return 0
}

/******************************************************************************
 **函数名称: SessionDel
 **功    能: 会话删除
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:50:17 #
 ******************************************************************************/
func (ctx *ChatTab) SessionDel(sid uint64) int {
	return 0
}

/******************************************************************************
 **函数名称: SubAdd
 **功    能: 订阅添加
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:50:17 #
 ******************************************************************************/
func (ctx *ChatTab) SubAdd(rid uint64, gid uint32, sid uint64) int {
	return 0
}

/******************************************************************************
 **函数名称: SubDel
 **功    能: 订阅删除
 **输入参数: NONE
 **输出参数: NONE
 **返    回: 0:成功 !0:失败
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:52:36 #
 ******************************************************************************/
func (ctx *ChatTab) SubDel(sid uint64) int {
	return 0
}

/******************************************************************************
 **函数名称: IsSub
 **功    能: 是否订阅
 **输入参数: NONE
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:52:36 #
 ******************************************************************************/
func (ctx *ChatTab) SubDel(sid uint64, cmd uint32) bool {
	return false
}

/******************************************************************************
 **函数名称: Trav
 **功    能: 遍历聊天室
 **输入参数: NONE
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:52:36 #
 ******************************************************************************/
func (ctx *ChatTab) Trav(rid uint64, gid uint32, ChatTravProc proc, param interface{}) {
}

/******************************************************************************
 **函数名称: TimeoutClean
 **功    能: 清理超时数据
 **输入参数: NONE
 **输出参数: NONE
 **返    回: true:订阅 false:未订阅
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2017.02.20 23:57:50 #
 ******************************************************************************/
func (ctx *ChatTab) TimeoutClean(rid uint64, gid uint32, ChatTravProc proc, param interface{}) {
}

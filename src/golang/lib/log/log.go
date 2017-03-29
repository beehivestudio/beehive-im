package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
)

/* 日志配置 */
type Conf struct {
	Level int    // 日志级别
	Path  string // 日志目录
}

/******************************************************************************
 **函数名称: Init
 **功    能: 初始化日志
 **输入参数:
 **     level: 日志级别
 **     path: 日志路径
 **     fname: 日志文件名
 **输出参数: NONE
 **返    回: 日志对象
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 21:44:14 #
 ******************************************************************************/
func Init(level int, path string, fname string) *logs.BeeLogger {
	log := logs.NewLogger(20000)

	/* > 创建日志路径 */
	err := os.Mkdir(path, 0755)
	if nil != err && false == os.IsExist(err) {
		log.Emergency(err.Error())
		return nil
	}

	log_path := fmt.Sprintf(`{"filename":"%s/%s"}`, path, fname)

	log.SetLogger("file", log_path)
	log.SetLevel(level)

	return log
}

/******************************************************************************
 **函数名称: GetLevel
 **功    能: 日志级别
 **输入参数:
 **     level: 日志级别(字串)
 **输出参数: NONE
 **返    回: 日志级别
 **实现描述:
 **注意事项:
 **作    者: # Qifeng.zou # 2016.11.04 21:43:27 #
 ******************************************************************************/
func GetLevel(level string) int {
	lower := strings.ToLower(level)

	switch lower {
	case "emerg":
		return logs.LevelEmergency
	case "alert":
		return logs.LevelAlert
	case "crit":
		return logs.LevelCritical
	case "error":
		return logs.LevelError
	case "warn":
		return logs.LevelWarning
	case "notice":
		return logs.LevelNotice
	case "info":
		return logs.LevelInformational
	case "debug":
		return logs.LevelDebug
	}
	return logs.LevelDebug
}

package util

import (
	"os"

	"github.com/astaxie/beego/logs"
)

// FileLogs 打印日志
var ConsoleLogs *logs.BeeLogger

// FileLogs 操作流程日志
var FileLogs *logs.BeeLogger

// ErrFileLogs 错误日志
var ErrFileLogs *logs.BeeLogger

var CrashFileLogs *logs.BeeLogger

func init() {
	// 判断logs文件夹是否存在
	isExists, _ := PathExists("./logs")
	if !isExists {
		// 创建目录
		os.Mkdir("./logs", os.ModePerm)
	}

	ConsoleLogs = logs.NewLogger()
	ConsoleLogs.SetLogger("console")

	FileLogs = logs.NewLogger()
	FileLogs.EnableFuncCallDepth(true)
	FileLogs.SetLogger("file", `{"filename":"logs/FTClane.log","level":7,"maxlines":0,"maxsize":5000000,"daily":true,"maxdays":30}`)
	FileLogs.SetLogger(logs.AdapterConsole)

	ErrFileLogs = logs.NewLogger()
	ErrFileLogs.EnableFuncCallDepth(true)
	ErrFileLogs.SetLogger("file", `{"filename":"logs/errInfo.log"}`)

	CrashFileLogs = logs.NewLogger()
	CrashFileLogs.EnableFuncCallDepth(true)
	CrashFileLogs.SetLogger("file", `{"filename":"logs/crashInfo.log"}`)
}

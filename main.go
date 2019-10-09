package main

import (
	"FTC/config"
	"FTC/device"
	"FTC/models"
	"FTC/tollmode"
	"FTC/util"
)

func main() {
	util.FileLogs.Info("自由流系统启动中.................................")

	config.InitConfig("./conf/config.conf")
	device.InitObjs()
	tollmode.InitObjs()

	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_OPEN:
		go tollmode.POpenModeObj.GoRun()
	case models.TOLLMODE_ENTRY:
		go tollmode.PEntryModeObj.GoRun()
	case models.TOLLMODE_FLAG:
		go tollmode.PFlagModeObj.GoRun()
	case models.TOLLMODE_EXIT:
		go tollmode.PExitModeObj.GoRun()
	}

	//文件传输
	go device.GoFlowProc()
	go device.GoAutoDevState()

	for {
		util.MySleep_s(60)
	}
	//beego.Run()
}

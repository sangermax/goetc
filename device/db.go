package device

import (
	"encoding/json"

	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//数据服务
func Flow2Txt(info models.MTransData) {
	var exitdata models.ReqExitFlowData
	filenm := info.FlowNo

	exitdata.FlowNo = info.FlowNo
	exitdata.Vehclass = info.Vehclass
	exitdata.Vehplate = info.Vehplate
	exitdata.ExitNetwork = info.ExitNetwork
	exitdata.ExitStationid = info.ExitStationid
	exitdata.ExitLandid = info.ExitLandid
	exitdata.ExitOperator = info.ExitOperator
	exitdata.ExitShiftdate = info.ExitShiftdate
	exitdata.ExitShift = info.ExitShift
	exitdata.Paytype = info.Paytype
	exitdata.Paymethod = info.Paymethod
	exitdata.Payid = info.Payid
	exitdata.Transtime = info.Transtime
	exitdata.Paytime = info.Paytime
	exitdata.Toll = info.Toll
	exitdata.Goodsno = info.Goodsno
	exitdata.Tradeno = info.Tradeno
	exitdata.PrintFlag = info.PrintFlag
	exitdata.Psamtermid = info.Psamtermid
	exitdata.Psamtradno = info.Psamtradno
	exitdata.Cardtradno = info.Cardtradno
	exitdata.CardId = info.CardId
	exitdata.Tac = info.Tac

	exitdata.EntryNetwork = info.EntryNetwork
	exitdata.EntryStationId = info.EntryStationId
	exitdata.EntryLaneId = info.EntryLaneId
	exitdata.EntryOperator = info.EntryOperator
	exitdata.EntryShift = info.EntryShift
	exitdata.EntryTime = info.EntryTime
	exitdata.FlagStationid = info.FlagStationid
	exitdata.FlagTime = info.FlagTime

	exitdata.RegVehclass = info.RegVehclass
	exitdata.RegVehplate = info.RegVehplate
	exitdata.AntennaID = info.AntennaID
	exitdata.TID = info.TID
	exitdata.VehColor = info.VehColor
	exitdata.TransFrom = info.TransFrom
	exitdata.StationMode = info.StationMode

	switch info.TransMemo {
	case models.TRANSMEMO_FREEFLOWPASS, models.TRANSMEMO_HISTORY:
		return
	default:
		if config.ConfigData["tollmode"].(int) == models.TOLLMODE_EXIT &&
			config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
			exitdata.TransMemo = models.TRANSMEMO_FREEFLOW
		} else {
			exitdata.TransMemo = models.TRANSMEMO_INIT
		}
	}

	b, err := json.Marshal(exitdata)
	if err != nil {
		util.FileLogs.Info("Flow2Txt failed,%s,%s.\r\n", filenm, err.Error())
		return
	}

	util.WriteFile(models.FLOWDIR+filenm, b)
	util.FileLogs.Info("save Flow2Txt suc(%s-%s).\r\n", filenm, exitdata.Vehplate)
}

func Img2Txt(info models.MTransData) {
	//没有识别到
	if info.CarfaceFlag != models.TRANSMATCH_YES {
		return
	}

	filenm := info.FlowNo
	var imgdata models.ReqExitImgData
	imgdata.RstVehPlateInfo = info.RstVehPlateInfo
	imgdata.AddsInfo.FlowNo = info.FlowNo
	imgdata.AddsInfo.ExitNetwork = info.ExitNetwork
	imgdata.AddsInfo.ExitStationid = info.ExitStationid
	imgdata.AddsInfo.ExitLandid = info.ExitLandid
	imgdata.AddsInfo.ExitOperator = info.ExitOperator
	imgdata.AddsInfo.ExitShiftdate = info.ExitShiftdate
	imgdata.AddsInfo.ExitShift = info.ExitShift
	imgdata.AddsInfo.Transtime = info.Transtime

	b, err := json.Marshal(imgdata)
	if err != nil {
		util.FileLogs.Info("Img2Txt failed,%s,%s.\r\n", filenm, err.Error())
		return
	}

	util.WriteFile(models.IMGDIR+filenm, b)
	util.FileLogs.Info("save Img2Txt suc(%s).\r\n", filenm)
}

func init() {
	util.CreateFileDir(models.FLOWDIR)
	util.CreateFileDir(models.IMGDIR)

	util.CreateFileDir(models.FAIL_FLOWDIR)
	util.CreateFileDir(models.FAIL_IMGDIR)
}

func procflow(flowtype int, flowpath string) bool {
	bret := true
	//读取本地缓存流水

	l := util.GetFilelist(flowpath)
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == nil {
			continue
		}

		strfile := e.Value.(string)

		//读文件 发送
		buf, err := util.ReadFile(strfile)
		if err != nil {
			//fmt.Printf("procflow ReadFile err:%s.\r\n", err.Error())
			util.MySleep_s(5)
			continue
		}

		switch flowtype {
		case models.FLOW_EXIT:
			{
				var exitflow models.ReqExitFlowData
				if err := json.Unmarshal(buf, &exitflow); err != nil {
					continue
				}

				if PPCloudObj != nil {
					bret = PPCloudObj.FuncReqExitflow(exitflow)
				}
			}

		case models.FLOW_EXITIMG:
			{
				var exitimg models.ReqExitImgData
				if json.Unmarshal(buf, &exitimg) != nil {
					continue
				}

				if PPCloudObj != nil {
					bret = PPCloudObj.FuncReqExitimg(exitimg)
				}
			}
		}
	}

	return bret
}

func GoFlowProc() {
	for {
		procflow(models.FLOW_EXIT, models.FLOWDIR)
		procflow(models.FLOW_EXITIMG, models.IMGDIR)

		util.MySleep_s(5)
	}
}

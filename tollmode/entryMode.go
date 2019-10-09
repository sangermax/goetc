package tollmode

import (
	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//EntryManage 入口式
type EntryManage struct {
}

func (p *EntryManage) createByCoil() {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_ENTRY)

	transdata.EntryStationId = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.EntryLaneId = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.EntryNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.EntryOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.EntryTime = transdata.Transtime
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.EntryShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_COIL
	transdata.CoilFlag = models.TRANSMATCH_YES
	transdata.TransState = models.TRANSSTATE_INIT

	GTransList.PushBack(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

func (p *EntryManage) createByPlateReg(info models.RstVehRecognizeInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_ENTRY)

	transdata.EntryStationId = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.EntryLaneId = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.EntryNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.EntryOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.EntryTime = transdata.Transtime
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.EntryShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_PLATE
	transdata.CarfaceFlag = models.TRANSMATCH_YES
	transdata.RegVehplate = info.VehicleInfo.CarBaseInfo.Lpn
	transdata.RegVehclass = info.VehicleInfo.CarBaseInfo.Vehtype
	transdata.RstVehPlateInfo = info
	transdata.Vehplate = transdata.RegVehplate
	transdata.Vehclass = transdata.RegVehclass

	transdata.TransState = models.TRANSSTATE_INIT

	GTransList.PushFront(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

func (p *EntryManage) createByEP(info models.EPResultReadInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_ENTRY)

	transdata.EntryStationId = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.EntryLaneId = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.EntryNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.EntryOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.EntryTime = transdata.Transtime
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.EntryShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_EP
	transdata.EPFlag = models.TRANSMATCH_YES
	transdata.TransFrom = models.TRANSFROM_EP
	transdata.TID = info.TID
	transdata.AntennaID = util.ConvertI2S(info.AntennaID)
	/*
		transdata.CardId = info.ReadDataInfo.CardNumber
		transdata.Vehplate = util.Unicode2Utf8(info.ReadDataInfo.LicenseCode + info.ReadDataInfo.PlateNumber)
		transdata.VehColor = util.GetEPVehColor(info.ReadDataInfo.Color)
		transdata.Vehclass = util.GetEPVehClass(info.ReadDataInfo.ApprovedLoad, info.ReadDataInfo.VehicleType)
	*/
	transdata.TransState = models.TRANSSTATE_INIT

	GTransList.PushBack(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s,%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.TID, transdata.Vehplate)
}

func (p *EntryManage) EPResultFreeflowProc(info models.EPResultReadInfo) {
	if chkEPInCurrentList(info) {
		return
	}

	r1, _ := chkEPInHistoryList(info)
	if r1 {
		return
	}

	GTransListLock.Lock()
	p.createByEP(info)
	GTransListLock.Unlock()
}

func (p *EntryManage) ETCResultFreeflowProc(info models.EPResultReadInfo) {
	if chkEPInCurrentList(info) {
		return
	}

	r1, _ := chkEPInHistoryList(info)
	if r1 {
		return
	}

	GTransListLock.Lock()
	p.createByEP(info)
	GTransListLock.Unlock()
}

func (p *EntryManage) EPResultProc(info models.EPResultReadInfo) {
	if chkEPInCurrentList(info) {
		return
	}

	bShow := false
	bMatch := false
	bInHistory := false
	r1, r2 := chkEPInHistoryList(info)
	if r1 && r2 != nil {
		bInHistory = true
		r2.TransMemo = models.TRANSMEMO_HISTORY
	}

	epVehplate := "" //util.Unicode2Utf8(info.ReadDataInfo.LicenseCode + info.ReadDataInfo.PlateNumber)

	GTransListLock.Lock()

	//先通过车牌检索
	i := 0
	for e := GTransList.Front(); e != nil; e = e.Next() {
		i += 1
		ev := e.Value.(*models.MTransData)
		if ev.EPFlag != models.TRANSMATCH_YES {
			//通过车牌检索一遍
			if epVehplate == ev.Vehplate {
				bMatch = true
				if bInHistory {
					*ev = *r2
					if i == 1 {
						bShow = true
					}
				}

				ev.TID = info.TID
				ev.AntennaID = util.ConvertI2S(info.AntennaID)
				/*
					ev.CardId = info.ReadDataInfo.CardNumber
					ev.Vehplate = util.Unicode2Utf8(info.ReadDataInfo.LicenseCode + info.ReadDataInfo.PlateNumber)
					ev.VehColor = util.GetEPVehColor(info.ReadDataInfo.Color)
					ev.Vehclass = util.GetEPVehClass(info.ReadDataInfo.ApprovedLoad, info.ReadDataInfo.VehicleType)
				*/
				ev.EPFlag = models.TRANSMATCH_YES
				if !ChkTransSuc(*ev) {
					ev.TransFlow = models.TRANSFLOW_INIT
				}

				util.FileLogs.Info("%s,%s 检测到EP信号，通过车牌匹配到第%d笔交易", ev.TID, ev.Vehplate, i)
			}
		}
	}

	if !bMatch {
		//再考虑车检匹配
		i = 0
		for e := GTransList.Front(); e != nil; e = e.Next() {
			i += 1
			ev := e.Value.(*models.MTransData)
			if ev.EPFlag != models.TRANSMATCH_YES && !ChkTransSuc(*ev) {
				bMatch = true
				if bInHistory {
					*ev = *r2
					if i == 1 {
						bShow = true
					}
				}

				ev.TID = info.TID
				ev.AntennaID = util.ConvertI2S(info.AntennaID)
				/*
					ev.CardId = info.ReadDataInfo.CardNumber
					ev.Vehplate = util.Unicode2Utf8(info.ReadDataInfo.LicenseCode + info.ReadDataInfo.PlateNumber)
					ev.VehColor = util.GetEPVehColor(info.ReadDataInfo.Color)
					ev.Vehclass = util.GetEPVehClass(info.ReadDataInfo.ApprovedLoad, info.ReadDataInfo.VehicleType)
				*/
				ev.EPFlag = models.TRANSMATCH_YES
				if !ChkTransSuc(*ev) {
					ev.TransFlow = models.TRANSFLOW_INIT
				}

				util.FileLogs.Info("%s,%s 检测到EP信号，车检匹配到第%d笔交易", ev.TID, ev.Vehplate, i)
			}

		}
	}

	if !bMatch {
		if bInHistory {
			GTransList.PushBack(r2)
		} else {
			p.createByEP(info)
		}
	}

	GTransListLock.Unlock()

	if bShow {
		ShowNextFeedisp()
	}
}

//目前只考虑电子车牌
func (p *EntryManage) GoRun() {
	//defer util.PanicHandler()
	util.FileLogs.Info("启动入口模式")

	for {
		//车牌付
		p.PlatePayFlowProc()
		//etc付
		p.ETCPayFlowProc()

		//成功入库
		FuncInsertFlowProc()

		util.MySleep_ms(10)
	}

}

//电子车牌或车脸识别处理流程
func (p *EntryManage) PlatePayFlowProc() {
	GTransListLock.Lock()
	for i, e := 0, GTransList.Front(); e != nil; i, e = i+1, e.Next() {
		ev := e.Value.(*models.MTransData)

		if ChkTransSuc(*ev) || (ev.TransFlow != models.TRANSFLOW_INIT) {
			continue
		}

		//车牌付或电子车牌付判断
		if ev.TransFlow == models.TRANSFLOW_INIT {
			if CheckPlatePay(ev) {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransState = models.TRANSSTATE_SUC
				ev.Paytime = util.GetNow(false)
				ev.PayFinishtime = util.GetNow(false)
				if i == 0 { //显示首笔
					FuncFeedispShowfee(ev.Vehplate, "0")
					//首笔成功则抬栏杆
					FuncCtlLanganProc(models.COIL_OPEN)
				}

				if ev.EPFlag == models.TRANSMATCH_YES {
					ev.TransFrom = models.TRANSFROM_EP
				} else {
					ev.TransFrom = models.TRANSFROM_PLATE
				}

			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransState = models.TRANSSTATE_FAIL
				ev.TransFailState = models.TRANSFAIL_NOTSUPPORTPLATEPAY
				if i == 0 { //显示首笔
					UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
				}
			}
		}
	}

	GTransListLock.Unlock()

}

//etc卡片付流程
func (p *EntryManage) ETCPayFlowProc() {

}

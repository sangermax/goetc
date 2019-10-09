package tollmode

import (
	"FTC/config"
	"FTC/device"
	"FTC/models"
	"FTC/util"
	"os"
)

func NewExitManage() *ExitManage {
	return &ExitManage{}
}

//ExitManage 出口式
type ExitManage struct {
	TransManualFlow  int //自助操作流程
	TransProcState   int
	TransProcTimeout int   //单位 ms
	TransLastTimeMs  int64 //单位 ms

	FstTransdata     models.MTransData
	selfWorkStation  byte
	equipWorkStation byte

	bManualDeviceState bool
}

func (p *ExitManage) createByCoil() {
	//创建
	transdata := new(models.MTransData)

	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_EXIT)

	transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.ExitShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_COIL
	transdata.CoilFlag = models.TRANSMATCH_YES
	transdata.TransState = models.TRANSSTATE_INIT

	GTransList.PushBack(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

// 测试模式下，可以卡片直接创建交易，正式模式下，禁止卡片直接创建交易
func (p *ExitManage) createByCard() {
	//如果没有或第一笔是成功，则创建
	bCreate := false
	pTrans1 := GetTransHeadLock()
	if pTrans1 == nil {
		bCreate = true
	}

	if !bCreate && ChkTransSuc(*pTrans1) {
		if models.GDebugxl &&
			(device.PReaderObj.Cardtypeinfo.Cardid == pTrans1.CardId || device.PReaderObj.Cardtypeinfo.Cardid == pTrans1.ProCardId) {
			//util.FileLogs.Info("该卡已交易成功，在当前队列中")
		} else {
			bCreate = true
		}
	}
	ReaseTransUnLock()

	if bCreate {
		transdata := new(models.MTransData)

		transdata.StationMode = util.ConvertI2S(models.TOLLMODE_EXIT)

		transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
		transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
		transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
		transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

		transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
		transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
		transdata.ExitShift = util.GetShiftId(transdata.Transtime)

		transdata.CreateBy = models.CREATEBY_CARD
		transdata.TransState = models.TRANSSTATE_INIT

		GTransList.PushBack(transdata)
		util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)

		p.FstTransdata = *transdata
		p.TransManualFlow = models.TRANSFLOW_PUTCARD
	}
}

func (p *ExitManage) createByPlateReg(info models.RstVehRecognizeInfo) {
	//创建
	transdata := new(models.MTransData)

	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_EXIT)

	transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.ExitShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_PLATE
	transdata.CarfaceFlag = models.TRANSMATCH_YES
	transdata.RegVehplate = info.VehicleInfo.CarBaseInfo.Lpn
	transdata.RegVehclass = info.VehicleInfo.CarBaseInfo.Vehtype
	transdata.RstVehPlateInfo = info
	transdata.Vehplate = transdata.RegVehplate
	transdata.Vehclass = transdata.RegVehclass

	GTransList.PushFront(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

func (p *ExitManage) createByEP(info models.EPResultReadInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_EXIT)

	transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.ExitShift = util.GetShiftId(transdata.Transtime)

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

func (p *ExitManage) EPResultFreeflowProc(info models.EPResultReadInfo) {
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

func (p *ExitManage) EPResultProc(info models.EPResultReadInfo) {
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
				if !ChkTransSuc(*ev) && PExitModeObj.ChkSelfEquipBusy() {
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
				//首笔，自助机在工作，则不匹配
				if i == 1 && PExitModeObj.ChkSelfEquipBusy() {
					continue
				}

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
			bShow = true
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

func (p *ExitManage) GoManualProcRun() {
	for {
		if p.TransProcState == models.TRANSPROCSTATE_INIT {
			util.MySleep_ms(1000)
			continue
		}

		if p.TransManualFlow == models.TRANSFLOW_PUTCARD ||
			p.TransManualFlow == models.TRANSFLOW_SCAN ||
			p.TransManualFlow == models.TRANSFLOW_TICKET {
			nowtm := util.GetTimeStampMs()
			if nowtm > p.TransLastTimeMs+int64(p.TransProcTimeout) {
				if p.TransManualFlow == models.TRANSFLOW_TICKET {
					p.TransProcState = models.TRANSPROCSTATE_END
				} else {
					p.TransProcState = models.TRANSPROCSTATE_INIT
				}
			}
		}

		util.MySleep_ms(1000)
	}
}

func (p *ExitManage) GoManualDeviceJudge() {
	p.bManualDeviceState = true

	for {
		bflag := true
		str := ""

		if device.PEquipmentObj != nil {
			if device.PEquipmentObj.State != models.DEVSTATE_OK {
				bflag = false
				str += "自助缴费机故障;"
			}
		}

		if device.PPrinterObj != nil {
			if device.PPrinterObj.DnState != models.DEVSTATE_OK {
				bflag = false
				str += "打印机故障;"
			}
		}

		if device.PScanObj != nil {
			if device.PScanObj.DnState != models.DEVSTATE_OK {
				bflag = false
				str += "扫码器故障;"
			}
		}

		if device.PReaderObj != nil {
			if device.PReaderObj.StateGrp.DnReaderState1 != models.DEVSTATE_OK {
				bflag = false
				str += "读写器故障;"
			}
		}

		if device.PEquipmentObj != nil && p.bManualDeviceState != bflag {
			if bflag {
				//device.PEquipmentObj.Package75(models.SHOW1)
				device.PEquipmentObj.Package74(device.ShowTip16(models.SELFDEFAULTTIP, models.ColorGreen))
			} else {
				device.PEquipmentObj.Package74(device.ShowTip16(str, models.ColorRed))
			}
		}
		p.bManualDeviceState = bflag

		util.MySleep_ms(5000)
	}
}

func (p *ExitManage) GoRun() {
	//defer util.PanicHandler()
	util.FileLogs.Info("启动出口模式:%d", config.ConfigData["bFreeflow"].(int))

	if config.ConfigData["bFreeflow"].(int) != models.STATIONMODE_FREEFLOWEXIT {
		p.bManualDeviceState = true
		p.TransManualFlow = models.TRANSFLOW_INIT
		p.TransProcState = models.TRANSPROCSTATE_INIT
		p.TransProcTimeout = models.TIMEOUT_DEFAULT
		p.TransLastTimeMs = 0
		go p.GoManualProcRun()
		go p.GoManualDeviceJudge()
	}

	go p.goFlowfreeFeedisp()
	/*
		if models.GDebugxl {

			util.MySleep_s(10)
			var addsinfo models.ExitImgAddsData
				_, _, regplateinfo := device.TestPlate(addsinfo)
				p.createByPlateReg(regplateinfo)

		}
	*/
	for {
		HistoryListTimeoutProc()

		//车牌付
		p.PlatePayFlowProc()
		//etc付
		p.ETCPayFlowProc()
		//自助缴费机支付
		if device.PEquipmentObj != nil && device.PEquipmentObj.IsConn() && p.bManualDeviceState {
			p.SelfEquipFlowProc()
		}

		//成功入库
		FuncInsertFlowProc()

		util.MySleep_ms(10)
	}
}

func (p *ExitManage) goFlowfreeFeedisp() {
	if device.PDevFeedispObj == nil {
		return
	}

	if config.ConfigData["bFreeflow"].(int) != models.STATIONMODE_FREEFLOWEXIT {
		return
	}

	bDefault := false
	line1 := ""
	line2 := ""
	line3 := ""
	for {
		if !device.PDevFeedispObj.IsGrpcConn() {
			//util.FileLogs.Info("费显断连，抛弃")
			util.MySleep_s(3)
			continue
		}

		bflag := false
		l1 := ""
		l2 := ""
		l3 := ""

		GTransListLock.Lock()
		/* for i, e := 0, GTransList.Front(); e != nil && i < 3; i, e = i+1, e.Next() {
			bflag = true
			ev := e.Value.(*models.MTransData)

			if ChkTransSuc(*ev) {
				switch i {
				case 0:
					l1 = ev.Vehplate
				case 1:
					l2 = ev.Vehplate
				case 2:
					l3 = ev.Vehplate
				}
			}
		}
		*/

		if e := GTransList.Front(); e != nil {
			ev := e.Value.(*models.MTransData)
			if ChkTransSuc(*ev) {
				bflag = true
				l1 = ev.Vehplate
				l2 = util.GetFeeShow(ev.Toll)
				l3 = "预缴费成功"
			}
		}
		GTransListLock.Unlock()

		if bflag {
			if l1 != line1 || l2 != line2 || l3 != line3 || bDefault {
				line1 = l1
				line2 = l2
				line3 = l3
				FuncFeedispShowTip(l1, l2, l3)
			}
			bDefault = false

		} else {
			if !bDefault {
				ShowDefaultFeedisp()
			}

			bDefault = true
		}

		util.MySleep_s(3)
	}
}

//电子车牌或车脸识别处理流程
func (p *ExitManage) PlatePayFlowProc() {
	//如果是自由流车道，费显显示为轮播模式，失败不显示
	bFeeDisp := true
	if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
		bFeeDisp = false
	}

	GTransListLock.Lock()
	for i, e := 0, GTransList.Front(); e != nil; i, e = i+1, e.Next() {
		ev := e.Value.(*models.MTransData)

		if ChkTransSuc(*ev) || (ev.TransFlow != models.TRANSFLOW_INIT) {
			continue
		}

		if ev.EPFlag == models.TRANSMATCH_YES || ev.CarfaceFlag == models.TRANSMATCH_YES {
		} else {
			continue
		}

		//车牌付或电子车牌付判断
		if ev.TransFlow == models.TRANSFLOW_INIT {
			util.FileLogs.Info("车牌付，正在处理第%d笔,%s,%s", i+1, ev.TID, ev.Vehplate)

			//查看是出口，查看是否在自由流缴费成功
			if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEPLUSEXIT &&
				(ev.EPFlag == models.TRANSMATCH_YES || ev.CarfaceFlag == models.TRANSMATCH_YES) {
				rlt, rltdata := ChkFreeflowSuc(ev.Vehplate, ev.Vehclass, ev.TID)
				if rlt && rltdata.Result == "1" {
					util.FileLogs.Info("自由流缴费成功:%s,%s,%s,%s", ev.TID, rltdata.Vehplate, rltdata.Vehclass, rltdata.Toll)
					ev.TransFlow = models.TRANSFLOW_FINISH
					ev.TransState = models.TRANSSTATE_SUC
					ev.TransMemo = models.TRANSMEMO_FREEFLOWPASS
					ev.Toll = rltdata.Toll
					ev.Vehclass = rltdata.Vehclass
					ev.Vehplate = rltdata.Vehplate
					ev.PayFinishtime = util.GetNow(false)
					Append2Histroylist(ev)

					NotifyFreeflow(ev.Vehplate, ev.Vehclass, ev.TID)
					if i == 0 { //显示首笔
						FuncCtlLanganProc(models.COIL_OPEN)
						if bFeeDisp {
							FuncFeedispShowTip(ev.Vehplate, util.GetFeeShow(ev.Toll), "已完成缴费")
						}
					}

					continue
				}

			}

			if CheckPlatePay(ev) {
				ev.TransFlow = models.TRANSFLOW_CARPATH
			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransFailState = models.TRANSFAIL_NOTSUPPORTPLATEPAY
				if i == 0 && bFeeDisp { //显示首笔
					UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
				}
			}
		}

		//请求路径信息
		if ev.TransFlow == models.TRANSFLOW_CARPATH {
			if FuncCarpath(ev) {
				ev.TransFlow = models.TRANSFLOW_FEE
			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransFailState = models.TRANSFAIL_PATHLOST

				if i == 0 && bFeeDisp { //显示首笔
					UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
				}
			}
		}

		//计费
		if ev.TransFlow == models.TRANSFLOW_FEE {
			if FuncFEE(ev) {
				ev.TransFlow = models.TRANSFLOW_PAY
			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransFailState = models.TRANSFAIL_FEECALC

				if i == 0 && bFeeDisp { //显示首笔
					UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
				}
			}
		}

		if ev.TransFlow == models.TRANSFLOW_PAY {
			if FuncPay(ev) {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransState = models.TRANSSTATE_SUC
				ev.Paytype = models.PAYTYPE_PLATE

				if i == 0 { //显示首笔
					//首笔成功则抬栏杆
					FuncCtlLanganProc(models.COIL_OPEN)
					if bFeeDisp {
						if ev.CarfaceFlag == models.TRANSMATCH_YES {
							FuncFeedispShowCarfacefee(ev.Vehplate, ev.Toll, ev.Vehclass, ev.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand)
						} else {
							FuncFeedispShowfee(ev.Vehplate, ev.Toll)
						}
					}
				}

			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransFailState = models.TRANSFAIL_FEEPAY
				if i == 0 && bFeeDisp { //显示首笔
					UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
				}
			}
		}

	}

	GTransListLock.Unlock()

}

//etc卡片付流程
func (p *ExitManage) ETCPayFlowProc() {

}

//自助机正在处理交易
func (p *ExitManage) ChkSelfEquipBusy() bool {
	if p.TransManualFlow == models.TRANSFLOW_INIT ||
		p.TransManualFlow == models.TRANSFLOW_PUTCARD ||
		p.TransManualFlow == models.TRANSFLOW_END {
		return false
	}

	return true
}

//自助缴费机支付，此流程耗时，
func (p *ExitManage) SelfEquipFlowProc() {
	bContinue := true

	pTrans1 := GetTransHeadLock()
	if pTrans1 == nil {
		bContinue = false
	}

	if bContinue && ChkTransSuc(*pTrans1) && pTrans1.Paytype == models.PAYTYPE_SCAN && p.TransManualFlow == models.TRANSFLOW_TICKET {

	} else if bContinue && ChkTransSuc(*pTrans1) && p.TransManualFlow != models.TRANSFLOW_INIT && pTrans1.FlowNo == p.FstTransdata.FlowNo {
		//如果是车牌付或etc交易成功，更新到当前
		bContinue = false
		p.FstTransdata = *pTrans1
	} else if bContinue && ChkTransSuc(*pTrans1) && p.TransManualFlow != models.TRANSFLOW_TICKET {
		bContinue = false
	}

	//如果自助已处理但是失败，则不再处理，转其它车道
	if bContinue && pTrans1.TransState == models.TRANSSTATE_FAIL {
		bContinue = false
	}

	//如果没有车型，暂不处理
	if bContinue && len(pTrans1.Vehclass) <= 0 {
		bContinue = false
	}

	if bContinue && p.TransManualFlow == models.TRANSFLOW_INIT {
		p.FstTransdata = *pTrans1
	}

	if bContinue && pTrans1.FlowNo != p.FstTransdata.FlowNo {
		util.FileLogs.Info("SelfEquipFlowProc 当前交易发生变更，抛弃此次交易…….\r\n")
		bContinue = false
	}

	if bContinue && pTrans1.TransFailState != models.TRANSFAIL_NONE {
		switch pTrans1.TransFailState {
		case models.TRANSFAIL_FEEPAY:
			util.FileLogs.Info("SelfEquipFlowProc 支付失败，转自助机扫码支付.\r\n")

			//由于自助缴费机流程问题，伸扫码臂必须在伸卡臂之后
			p.selfWorkStation, p.equipWorkStation = util.GetSE(util.ConvertS2I(pTrans1.Vehclass))
			device.PEquipmentObj.FuncCReaderMove(p.equipWorkStation, false)
			p.TransManualFlow = models.TRANSFLOW_SCAN
		}

		pTrans1.TransFailState = models.TRANSFAIL_NONE
	}
	ReaseTransUnLock()

	//如果交易已在处理中，需要中止交易
	if !bContinue {
		if p.TransManualFlow != models.TRANSFLOW_INIT {
			util.FileLogs.Info("SelfEquipFlowProc 中止交易…….\r\n")
			p.FuncEndProc()

			//恢复费显内容
			ShowNextFeedisp()
		}

		return
	}

	p.selfWorkStation, p.equipWorkStation = util.GetSE(util.ConvertS2I(p.FstTransdata.Vehclass))

	switch p.TransManualFlow {
	case models.TRANSFLOW_INIT:
		{
			p.TransManualFlow = models.TRANSFLOW_PUTCARD
		}
	case models.TRANSFLOW_PUTCARD:
		{
			p.FuncPutCardProc()
		}
	case models.TRANSFLOW_READCARD:
		{
			//device..Package75(models.SHOW2)
			//device.PEquipmentObj.Package74(device.ShowTip16("正在读卡", models.ColorGreen))

			p.FuncReadCardProc()
		}
	case models.TRANSFLOW_FEE:
		{
			//device.PEquipmentObj.Package74(device.ShowTip16("正在计费", models.ColorGreen))
			p.FuncFeeProc()
		}
	case models.TRANSFLOW_SCAN:
		{
			p.FuncScanProc()
		}
	case models.TRANSFLOW_PAY:
		{
			p.FuncPayProc()
		}
	case models.TRANSFLOW_TICKET:
		{
			p.FuncTicketProc()
		}
	case models.TRANSFLOW_OUTCARD:
		{
			p.FuncOutCardProc()
		}
	}

	if p.TransManualFlow == models.TRANSFLOW_END {
		p.FuncEndProc()
	}
}

func (p *ExitManage) FuncUpdateManualProcState(procState int) {
	p.TransProcState = procState
}

//放卡处理流程
func (p *ExitManage) FuncPutCardProc() {
	switch p.TransProcState {
	case models.TRANSPROCSTATE_INIT:
		//自助机伸臂动作
		util.FileLogs.Info("缴费机伸出卡机，等待放卡…….\r\n")
		device.PEquipmentObj.FuncCReaderMoveLinked(p.equipWorkStation)

		p.TransProcTimeout = models.TIMEOUT_PUTCARD
		p.TransLastTimeMs = util.GetTimeStampMs()
		p.TransProcState = models.TRANSPROCSTATE_PROCING

		FuncFeedispShowTip(p.FstTransdata.Vehplate, "请放通行卡", "")

	case models.TRANSPROCSTATE_PROCING:
		//等待放卡
	case models.TRANSPROCSTATE_END:
		if device.PReaderObj != nil {
			//关闭费显报警
			FuncFeedispAlarmOff()

			p.FstTransdata.ProCardId = device.PReaderObj.Cardtypeinfo.Cardid
			p.FstTransdata.CardId = device.PReaderObj.Cardtypeinfo.Cardid
			p.FstTransdata.Cardtype = device.PReaderObj.Cardtypeinfo.Cardtype
			p.FstTransdata.Psamtermid = device.PReaderObj.Cardtypeinfo.PsamTermId
		}
		//读卡
		p.TransManualFlow = models.TRANSFLOW_READCARD
		p.TransProcState = models.TRANSPROCSTATE_INIT
	}
}

//读卡处理流程
func (p *ExitManage) FuncReadCardProc() {
	bContinue := false

	//暂没判断流通状态=======================
	switch p.FstTransdata.Cardtype {
	case models.MifarePro, models.MifareProX:
		{
			var einfo models.ReaderETCReadReqData
			einfo.FileID = 0x0015
			einfo.Length = 50
			r1, r2, r3 := device.PReaderObj.ReaderETCReadReqData(einfo)
			if models.GRPCRESULT_OK == r1 && r3.Result == 0 {
				p.FstTransdata.StrFile0015 = util.ConvertByte2Hexstring(r3.Data, false)

				p4, p2 := util.ETCParse0015(r3.Data)
				if p2 != nil {
					util.FileLogs.Info("Parse0015 err:%s.", r2)
					bContinue = false
				} else {
					p.FstTransdata.File0015 = p4
					bContinue = true
				}
			} else {
				util.FileLogs.Info("FuncReadCardProc0015 err:%s.", r2)
				bContinue = false
			}

			if bContinue {
				einfo.FileID = 0x0019
				einfo.Length = 43
				r1, r2, r4 := device.PReaderObj.ReaderETCReadReqData(einfo)
				if models.GRPCRESULT_OK == r1 && r4.Result == 0 {
					p.FstTransdata.StrFile0019 = util.ConvertByte2Hexstring(r4.Data, false)

					p5, p2 := util.ETCParse0019(r4.Data)
					if p2 != nil {
						util.FileLogs.Info("Parse0019 err:%s.", r2)
						bContinue = false
					} else {
						p.FstTransdata.File0019 = p5
						bContinue = true
					}
				} else {
					util.FileLogs.Info("FuncReadCardProc0015 err:%s.", r2)
					bContinue = false
				}
			}

			if bContinue {
				var balanceinfo models.ReaderETCBalanceReqData
				r1, r2, r5 := device.PReaderObj.ReaderETCBalanceReqData(balanceinfo)
				if models.GRPCRESULT_OK == r1 && r5.Result == 0 {
					p.FstTransdata.BeforeBalance = r5.Balance
					bContinue = true
				} else {
					util.FileLogs.Info("FuncReadCardProc0015 balance:%s.", r2)
					bContinue = false
				}
			}

			if bContinue {
				p.FstTransdata.EntryNetwork = p.FstTransdata.File0019.InStationNetWork
				p.FstTransdata.EntryStationId = p.FstTransdata.File0019.InStation
				p.FstTransdata.EntryLaneId = p.FstTransdata.File0019.InLane
				p.FstTransdata.EntryOperator = p.FstTransdata.File0019.InOperator
				p.FstTransdata.EntryShift = util.ConvertI2S(p.FstTransdata.File0019.InBanci)
				p.FstTransdata.EntryTime = p.FstTransdata.File0019.InTime
				p.FstTransdata.FlagStationid = p.FstTransdata.File0019.FlagSta

				p.FstTransdata.CardId = p.FstTransdata.File0015.CardId
				p.FstTransdata.Vehplate = p.FstTransdata.File0015.VehPlate
				p.FstTransdata.Vehclass = util.ConvertI2S(p.FstTransdata.File0015.VehClass)
				p.FstTransdata.VehColor = util.ConvertI2S(p.FstTransdata.File0015.VehColor)
				p.TransManualFlow = models.TRANSFLOW_FEE
			} else {
				p.TransManualFlow = models.TRANSFLOW_END

				FuncFeedispShowErr(p.FstTransdata.Vehplate, "读卡信息失败", "请转其它车道")
			}
		}
		break
	}
}

//计费
func (p *ExitManage) FuncFeeProc() {
	ev := GetTransHeadLock()
	if ev != nil {
		if ev.FlowNo == p.FstTransdata.FlowNo {
			if FuncFEE(ev) {
				if false { //(p.FstTransdata.Cardtype == models.MifarePro || p.FstTransdata.Cardtype == models.MifareProX) &&
					//(p.FstTransdata.BeforeBalance >= uint(util.ConvertS2I(p.FstTransdata.Toll))) {

					p.FstTransdata.Toll = ev.Toll
					p.FstTransdata.Paytype = models.PAYTYPE_ETC
					p.TransManualFlow = models.TRANSFLOW_PAY
					p.TransProcState = models.TRANSPROCSTATE_INIT
				} else {
					p.FstTransdata.Toll = ev.Toll
					p.FstTransdata.Paytype = models.PAYTYPE_SCAN
					p.TransManualFlow = models.TRANSFLOW_SCAN
					p.TransProcState = models.TRANSPROCSTATE_INIT
				}

			} else {
				p.TransManualFlow = models.TRANSFLOW_END
				p.FstTransdata.TransState = models.TRANSSTATE_FAIL

				FuncFeedispShowErr(p.FstTransdata.Vehplate, "计费失败", "请转其它车道")
			}
		} else {
			util.FileLogs.Info("FuncFeeProc 交易发生变更，放弃此次操作.\r\n")
			p.TransManualFlow = models.TRANSFLOW_END
			p.FstTransdata.TransState = models.TRANSSTATE_FAIL

			FuncFeedispShowErr(p.FstTransdata.Vehplate, "交易异常", "请转其它车道")
		}
	} else {
		util.FileLogs.Info("FuncFeeProc 交易无，放弃此次操作.\r\n")
		p.FuncChkOutCardProc()
	}

	ReaseTransUnLock()
}

func (p *ExitManage) FuncScanProc() {
	if device.PScanObj == nil {
		util.FileLogs.Info("系统模式错误，不支持扫码")

		//自助缴费机提示扫码服务异常，不支持扫码服务
		p.FstTransdata.TransState = models.TRANSSTATE_FAIL
		p.TransManualFlow = models.TRANSFLOW_END

		return
	}

	switch p.TransProcState {
	case models.TRANSPROCSTATE_INIT:
		util.FileLogs.Info("缴费机伸出扫码机，等待扫码…….\r\n")
		device.PEquipmentObj.FuncCScanMoveLinked(p.equipWorkStation)

		s1 := util.GetFeeShow(p.FstTransdata.Toll)
		s2 := p.FstTransdata.Vehplate + ";" + s1 + ";;" + "请出示付款码;"
		device.PEquipmentObj.Package74(device.ShowTip16(s2, models.ColorGreen))
		FuncFeedispShowTip(p.FstTransdata.Vehplate, s1, "请出示付款码")

		var askScan models.ReqScanOpenData
		askScan.WorkStation = util.ConvertI2S(int(p.selfWorkStation))
		askScan.ScanSpanTm = util.ConvertI2S(config.ConfigData["scanKeepTime"].(int))
		if err := device.PScanObj.FuncReqScanOpen(askScan); err != nil {
			util.FileLogs.Info("发送扫码指令失败:%s", err)

			//自助缴费机提示扫码服务异常，不支持扫码服务
			p.FstTransdata.TransState = models.TRANSSTATE_FAIL
			p.TransManualFlow = models.TRANSFLOW_END

			FuncFeedispShowErr(p.FstTransdata.Vehplate, "扫码设备故障", "请转其它车道")
			return
		}

		p.TransProcState = models.TRANSPROCSTATE_PROCING
		p.TransProcTimeout = models.TIMEOUT_SCAN
		p.TransLastTimeMs = util.GetTimeStampMs()

	case models.TRANSPROCSTATE_PROCING:
	case models.TRANSPROCSTATE_END:
		//关闭费显报警
		FuncFeedispAlarmOff()

		//信号量返回处理
		rltCode, rltDes, rltScanData := device.PScanObj.FuncProcScanOpen()
		if rltCode == models.GRPCRESULT_OK {
			util.FileLogs.Info("扫码成功:%s", rltScanData.ScanBar)

			p.FstTransdata.Paytype = models.PAYTYPE_SCAN
			p.FstTransdata.Payid = rltScanData.ScanBar

			p.TransManualFlow = models.TRANSFLOW_PAY
			p.TransProcState = models.TRANSPROCSTATE_INIT
		} else {
			util.FileLogs.Info("扫码失败,error:%s", rltDes)
		}
	}
}

func (p *ExitManage) PackageExitEtc0019() []byte {
	var tmp0019 models.F0019Info

	tmp0019.InStationNetWork = ""
	tmp0019.InStation = ""
	tmp0019.InLane = ""
	tmp0019.InTime = ""
	tmp0019.VehClass = util.ConvertS2I(p.FstTransdata.Vehclass)
	tmp0019.FlowState = 4
	tmp0019.OutStationNetWork = p.FstTransdata.ExitNetwork
	tmp0019.OutStation = p.FstTransdata.ExitStationid
	tmp0019.InOperator = ""
	tmp0019.InBanci = 0
	tmp0019.VehPlate = p.FstTransdata.Vehplate

	return util.ETCPackage0019(tmp0019)
}

func (p *ExitManage) FuncETCPay(ev *models.MTransData) {
	var einfo models.ReaderETCPayReqData
	einfo.Money = util.ConvertS2I(p.FstTransdata.Toll)
	einfo.Data = make([]byte, 43)
	einfo.Paytime = util.ConvertTime_2(p.FstTransdata.Transtime)

	//回写0019文件
	copy(einfo.Data, p.PackageExitEtc0019())

	r1, r2, r3 := device.PReaderObj.ReaderETCPayReqData(einfo)
	if models.GRPCRESULT_OK == r1 && r3.Result == 0 {
		p.FstTransdata.Psamtradno = util.ConvertByte2Hexstring(r3.TermTradNo, false)
		p.FstTransdata.Cardtradno = util.ConvertByte2Hexstring(r3.TradNo, false)
		p.FstTransdata.Tac = util.ConvertByte2Hexstring(r3.Tac, false)
		p.FstTransdata.Paytime = util.GetNow(false) //自助缴费支付时间
		p.FstTransdata.PayFinishtime = util.GetNow(false)

		p.FstTransdata.TransState = models.TRANSSTATE_SUC
		p.FstTransdata.TransFrom = models.TRANSFROM_ETC

		//支付成功，值写回队列
		*ev = p.FstTransdata
	} else {
		FuncFeedispShowErr(p.FstTransdata.Vehplate, "卡片扣款失败", "请转其它车道")

		util.FileLogs.Info("FuncETCPay err:%s.", r2)
		p.TransManualFlow = models.TRANSFLOW_END
		p.FstTransdata.TransState = models.TRANSSTATE_FAIL
		os.Exit(-1)
	}
}

//支付
func (p *ExitManage) FuncPayProc() {
	ev := GetTransHeadLock()
	if ev != nil {
		if ev.FlowNo == p.FstTransdata.FlowNo {

			if p.FstTransdata.Paytype == models.PAYTYPE_ETC {
				p.FuncETCPay(ev)
			} else {
				ev.Paytype = p.FstTransdata.Paytype
				ev.Payid = p.FstTransdata.Payid
				ev.Vehplate = p.FstTransdata.Vehplate

				if FuncPay(ev) {
					p.FstTransdata.TransState = models.TRANSSTATE_SUC
					p.FstTransdata.Paytime = util.GetNow(false)
					p.FstTransdata.PayFinishtime = util.GetNow(false)
					p.FstTransdata.TransFrom = models.TRANSFROM_EQUIPMENT

					//支付成功，值写回队列
					*ev = p.FstTransdata
				} else {
					p.TransManualFlow = models.TRANSFLOW_END
					p.FstTransdata.TransState = models.TRANSSTATE_FAIL

					FuncFeedispShowErr(p.FstTransdata.Vehplate, "扣款失败", "请转其它车道")
				}
			}

		} else {
			util.FileLogs.Info("FuncPayProc 交易无，放弃此次操作.\r\n")
			p.TransManualFlow = models.TRANSFLOW_END
			p.FstTransdata.TransState = models.TRANSSTATE_FAIL

			FuncFeedispShowErr(p.FstTransdata.Vehplate, "交易异常", "请转其它车道")
		}
	} else {
		p.FuncChkOutCardProc()
	}

	ReaseTransUnLock()

	if ChkTransSuc(p.FstTransdata) {
		util.FileLogs.Info("FuncPayProc 支付成功")
		device.PEquipmentObj.Package73("支付成功")

		device.PEquipmentObj.FuncCReaderKeep(p.equipWorkStation, true)
		device.PEquipmentObj.FuncCScanKeep(p.equipWorkStation, true)
		device.PEquipmentObj.Package74(device.ShowTip16(";支付成功;;请按键取票", models.ColorGreen))

		FuncFeedispShowTip(p.FstTransdata.Vehplate, "支付成功", "请按键取票")

		if p.FstTransdata.Paytype == models.PAYTYPE_ETC {
			p.TransManualFlow = models.TRANSFLOW_END
		} else {
			p.TransManualFlow = models.TRANSFLOW_TICKET
			p.TransProcState = models.TRANSPROCSTATE_INIT
		}

		FuncCtlLanganProc(models.COIL_OPEN)
	}
}

func (p *ExitManage) FuncTicketProc() {
	if device.PPrinterObj == nil {
		return
	}

	switch p.TransProcState {
	case models.TRANSPROCSTATE_INIT:
		//打票
		ticketinfo := PackagePrinter(p.FstTransdata)
		ticketinfo.WorkStation = util.ConvertI2S(int(p.selfWorkStation))
		ticketinfo.LineNums = util.ConvertI2S(len(ticketinfo.PrintRsds))
		device.PPrinterObj.FuncReqPrintTicket(ticketinfo)

		p.TransProcState = models.TRANSPROCSTATE_PROCING
		p.TransProcTimeout = models.TIMEOUT_TICKET
		p.TransLastTimeMs = util.GetTimeStampMs()

		device.PEquipmentObj.Package73("正在出票，请稍等")
		device.PEquipmentObj.Package74(device.ShowTip16(";正在出票 请稍候", models.ColorGreen))

	case models.TRANSPROCSTATE_PROCING: //打票到出票延时

	case models.TRANSPROCSTATE_END:
		device.PEquipmentObj.FuncCTicketMoveLinked(p.equipWorkStation)

		p.TransManualFlow = models.TRANSFLOW_END
		p.TransProcState = models.TRANSPROCSTATE_INIT
	}
}

func (p *ExitManage) FuncOutCardProc() {

}

func (p *ExitManage) FuncEndProc() {
	//如果结束，则读卡器恢复初始状态
	if device.PReaderObj != nil {
		var info models.ReaderClosecardReqData
		device.PReaderObj.ReaderClosecardReqData(info)
	}

	if ChkTransSuc(p.FstTransdata) {
		device.PEquipmentObj.Recovery()
	} else {
		ev := GetTransHeadLock()
		if ev != nil {
			if ev.FlowNo == p.FstTransdata.FlowNo {
				ev.TransState = p.FstTransdata.TransState
				util.FileLogs.Info("自助缴费处理失败，中止交易，转其它车道处理")

				device.PEquipmentObj.Package73("交易失败，请转其它车道")
				device.PEquipmentObj.Package74(device.ShowTip16(";交易失败;;请按报警键转人工", models.ColorRed))
			}
		}
		ReaseTransUnLock()
	}

	p.TransManualFlow = models.TRANSFLOW_INIT
	p.TransProcState = models.TRANSPROCSTATE_INIT
	p.TransProcTimeout = models.TIMEOUT_DEFAULT
	p.TransLastTimeMs = 0
	p.FstTransdata = models.MTransData{}
}

func (p *ExitManage) FuncChkOutCardProc() {
	p.TransManualFlow = models.TRANSFLOW_END

	/*
		if p.FstTransdata.HasCard == models.CARD_YES {
			p.TransManualFlow = models.TRANSFLOW_OUTCARD
		} else {
			p.TransManualFlow = models.TRANSFLOW_END
		}
	*/
}

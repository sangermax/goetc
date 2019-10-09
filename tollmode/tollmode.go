package tollmode

import (
	"container/list"
	"fmt"
	"sync"

	"FTC/config"
	"FTC/device"
	"FTC/models"
	"FTC/util"
)

var (
	POpenModeObj  *OpenManage
	PEntryModeObj *EntryManage
	PFlagModeObj  *FlagManage
	PExitModeObj  *ExitManage

	//当前交易队列
	GTransListLock *sync.Mutex
	GTransList     *list.List

	//ETC缓存队列
	GETCFrameListLock *sync.Mutex
	GETCFrameList     *list.List

	//成功已驶离的交易队列
	GTransHistoryList *list.List
)

func InitObjs() {
	GTransListLock = new(sync.Mutex)
	GTransList = list.New()
	GTransHistoryList = list.New()

	GETCFrameListLock = new(sync.Mutex)
	GETCFrameList = list.New()

	POpenModeObj = new(OpenManage)
	PEntryModeObj = new(EntryManage)
	PFlagModeObj = new(FlagManage)
	PExitModeObj = new(ExitManage)

	ShowNextFeedisp()
	go GoChanRun()
	go GoChkTransTimeout()

	go goEtcFrameProc()
}

func CoilInSignalFunc() {
	GTransListLock.Lock()
	defer GTransListLock.Unlock()

	isize := GTransList.Len()
	if isize <= 0 {
		switch config.ConfigData["tollmode"].(int) {
		case models.TOLLMODE_OPEN:
			POpenModeObj.createByCoil()
		case models.TOLLMODE_ENTRY:
			PEntryModeObj.createByCoil()
		case models.TOLLMODE_FLAG:
			PFlagModeObj.createByCoil()
		case models.TOLLMODE_EXIT:
			PExitModeObj.createByCoil()

		}
	} else {
		util.FileLogs.Info("检测到车检信号，匹配到第一笔交易")

		//匹配
		first := GTransList.Front().Value.(*models.MTransData)
		first.CoilFlag = models.TRANSMATCH_YES
	}

}

func HistoryListTimeoutProc() {
	timeoutsecs := config.ConfigData["historyTimeout"].(int)

	for e := GTransHistoryList.Front(); e != nil; {
		e1 := e
		e = e.Next()

		ev := e1.Value.(*models.MTransData)
		if util.ChkTimeOut(ev.PayFinishtime, timeoutsecs) {
			util.FileLogs.Info("%s 超时从历史队列中清除", ev.Vehplate)
			GTransHistoryList.Remove(e1)
		}
	}
}

func chkEPInHistoryList(info models.EPResultReadInfo) (bool, *models.MTransData) {
	if models.CHECK_NO == config.ConfigData["bCheckHistoryList"].(int) {
		return false, nil
	}

	for e := GTransHistoryList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(*models.MTransData)

		if ev.EPFlag == models.TRANSMATCH_YES && ev.TID == info.TID {
			util.FileLogs.Info("%s,%s 检测到EP信号，在历史交易中", ev.TID, ev.Vehplate)
			return true, ev
		}
	}

	return false, nil
}

func chkETCInHistoryList(sobuid string) (bool, *models.MTransData) {
	if models.CHECK_NO == config.ConfigData["bCheckHistoryList"].(int) {
		return false, nil
	}

	for e := GTransHistoryList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(*models.MTransData)

		if ev.OBUID == sobuid {
			util.FileLogs.Info("%s,%s 检测到ETC信号，在历史交易中", ev.OBUID, ev.Vehplate)
			return true, ev
		}
	}

	return false, nil
}

func chkEPInCurrentList(info models.EPResultReadInfo) bool {
	if models.CHECK_NO == config.ConfigData["bCheckCurrentList"].(int) {
		return false
	}

	bfind := false

	GTransListLock.Lock()

	i := 0
	for e := GTransList.Front(); e != nil; e = e.Next() {
		i += 1
		ev := e.Value.(*models.MTransData)

		if ev.EPFlag == models.TRANSMATCH_YES && ev.TID == info.TID {
			util.FileLogs.Info("%s 检测到EP信号，已在当前交易中:%d", ev.Vehplate, i)
			bfind = true
			break
		}

	}

	GTransListLock.Unlock()

	if bfind {
		ShowNextFeedisp()
	}
	return bfind
}

func EPSignalFunc() {
	device.PEPAntObj.EPReadListLock.Lock()
	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_ENTRY:
		if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				PEntryModeObj.EPResultFreeflowProc(e.Value.(models.EPResultReadInfo))
			}
		} else {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				PEntryModeObj.EPResultProc(e.Value.(models.EPResultReadInfo))
			}
		}
	case models.TOLLMODE_FLAG:
		for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
			PFlagModeObj.EPResultFreeflowProc(e.Value.(models.EPResultReadInfo))
		}
	case models.TOLLMODE_OPEN:
		if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				POpenModeObj.EPResultFreeflowProc(e.Value.(models.EPResultReadInfo))
			}
		} else {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				POpenModeObj.EPResultProc(e.Value.(models.EPResultReadInfo))
			}
		}
	case models.TOLLMODE_EXIT:
		if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				PExitModeObj.EPResultFreeflowProc(e.Value.(models.EPResultReadInfo))
			}
		} else {
			for e := device.PEPAntObj.EPReadList.Front(); e != nil; e = e.Next() {
				PExitModeObj.EPResultProc(e.Value.(models.EPResultReadInfo))
			}
		}
	}

	device.PEPAntObj.EPReadList.Init()
	device.PEPAntObj.EPReadListLock.Unlock()
}

func Append2Histroylist(transdata *models.MTransData) {
	GTransHistoryList.PushBack(transdata)

	maxsize := config.ConfigData["historySize"].(int)
	isize := GTransHistoryList.Len()

	if isize > maxsize {
		//删除第一笔
		e := GTransHistoryList.Front()
		GTransHistoryList.Remove(e)
	}
}

func GoChkTransTimeout() {
	for {
		ChkTransTimeout()

		//TestChkTransTimeout()

		util.MySleep_s(1)
	}
}

//该协程不要阻塞
func GoChanRun() {
	//defer util.PanicHandler()
	util.FileLogs.Info("启动EP线程")

	for {
		select {
		case <-device.ChanETC:
			util.FileLogs.Info("收到ETC信号.\n")
			ETCSignalFunc()
		case <-device.ChanEP:
			//util.FileLogs.Info("收到EP电子车牌信号.\n")
			EPSignalFunc()
		case <-device.ChanPlate:
			util.FileLogs.Info("收到车脸识别结果信号.\n")
			CarfaceSignalFunc()
		case <-device.ChanVehIn:
			util.FileLogs.Info("收到车辆驶入车检信号.\n")
			CoilInSignalFunc()
		case <-device.ChanVehOut:
			util.FileLogs.Info("收到车辆驶离车检信号.\n")
			CoilOutSignalFunc()

		case <-device.ChanScan:
			util.FileLogs.Info("收到扫码信号.\n")
			ScanSignalFunc()

		case <-device.ChanCard:
			util.FileLogs.Info("收到有卡信号.\n")
			if config.ConfigData["tollmode"].(int) == models.TOLLMODE_EXIT {
				PExitModeObj.createByCard()
			}

			CardSignalFunc()
		}
	}
}

func ScanSignalFunc() {
	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_OPEN:
		POpenModeObj.FuncUpdateManualProcState(models.TRANSPROCSTATE_END)
	case models.TOLLMODE_EXIT:
		PExitModeObj.FuncUpdateManualProcState(models.TRANSPROCSTATE_END)
	}

}

func CardSignalFunc() {
	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_OPEN:
		POpenModeObj.FuncUpdateManualProcState(models.TRANSPROCSTATE_END)
	case models.TOLLMODE_EXIT:
		PExitModeObj.FuncUpdateManualProcState(models.TRANSPROCSTATE_END)
	}

}

func UpdateFeeErrDisp(plate string, failstate int) {
	switch failstate {
	case models.TRANSFAIL_NONE:
		FuncFeedispAlarmOff()

	case models.TRANSFAIL_NOTSUPPORTPLATEPAY:
		FuncFeedispShowErr(plate, "非车牌付车辆", "请转自助服务")

	case models.TRANSFAIL_PATHLOST:
		FuncFeedispShowErr(plate, "行驶路径丢失", "请转其它车道")

	case models.TRANSFAIL_FEECALC:
		FuncFeedispShowErr(plate, "计费失败", "请转其它车道")

	case models.TRANSFAIL_FEEPAY:
		FuncFeedispShowErr(plate, "扣款失败", "请转自助服务")
	}
}

// 车辆驶离信号
func CoilOutSignalFunc() {
	GTransListLock.Lock()
	if e := GTransList.Front(); e != nil {
		ev := e.Value.(*models.MTransData)
		if ChkTransSuc(*ev) {
			GTransList.Remove(e)
		}
	}
	GTransListLock.Unlock()

	ShowNextFeedisp()
}

// 车脸识别结果信号
func CarfaceSignalFunc() {
	if device.PPlateObj == nil {
		return
	}

	rltCode, rltDes, rltPlate := device.PPlateObj.FuncCarFaceRstProc()
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("CarfaceSignalFunc 失败,err：%s.\n", rltDes)
		return
	}

	//车牌识别结果匹配到第一辆车
	bCreate := true
	GTransListLock.Lock()
	defer GTransListLock.Unlock()

	e := GTransList.Front()
	//有交易则匹配
	if e != nil {
		ev := e.Value.(*models.MTransData)
		srp := rltPlate.VehicleInfo.CarBaseInfo.Lpn
		srv := rltPlate.VehicleInfo.CarBaseInfo.Vehtype

		c := util.ComparePlate(ev.Vehplate, srp, config.ConfigData["comparePlateLen"].(int))
		//第一笔已经交易成功
		if ChkTransSuc(*ev) {
			if c {
				bCreate = false
				util.FileLogs.Info("第一笔交易已经成功，抛弃车脸识别结果:%s,%s.\n", srp, srv)

				return

			} else {
				//如果车牌不一致，是否需要创建？？？？？？？？？？？？
				bCreate = true
			}
		} else {
			if c {
				bCreate = false
				util.FileLogs.Info("收到车脸识别结果，匹配到第一笔交易:%s,%s.\n", srp, srv)
			} else {
				//如果车牌不一致，查看第一笔交易是否有车牌，如果无车牌，则匹配，如果有车牌，则新创建?????
				if len(ev.Vehclass) <= 0 {
					bCreate = false
					util.FileLogs.Info("收到车脸识别结果，匹配到第一笔交易2:%s,%s.\n", srp, srv)
				} else {
					bCreate = true
				}

			}
		}

		if !bCreate {
			ev.RegVehplate = srp
			ev.RegVehclass = srv
			ev.CarfaceFlag = models.TRANSMATCH_YES
			ev.RstVehPlateInfo = rltPlate

			if len(ev.Vehclass) <= 0 {
				ev.Vehplate = srp
				ev.Vehclass = srv
			}

			//第一笔交易没成功，收到车脸识别结果后 重新走车牌付判断流程
			if ChkTransEnd(*ev) {
				ev.TransFlow = models.TRANSFLOW_INIT
			}
		}
	}

	if bCreate { // 没有交易，则新创建
		util.FileLogs.Info("收到车脸识别结果，创建新交易")

		switch config.ConfigData["tollmode"].(int) {
		case models.TOLLMODE_OPEN:
			POpenModeObj.createByPlateReg(rltPlate)
		case models.TOLLMODE_ENTRY:
			PEntryModeObj.createByPlateReg(rltPlate)
		case models.TOLLMODE_FLAG:
			PFlagModeObj.createByPlateReg(rltPlate)
		case models.TOLLMODE_EXIT:
			PExitModeObj.createByPlateReg(rltPlate)

		}
	}
}

//车脸付或电子车牌付校验
func CheckPlatePay(curtrans *models.MTransData) bool {
	if CheckCarfacePlatePay(*curtrans) || CheckEPPlatePay(*curtrans) {
		util.FileLogs.Info("%s,%s 车牌付用户判断请求.\r\n", curtrans.TID, curtrans.Vehplate)
		var askPlatepayChk models.ReqCheckPlatepayData
		askPlatepayChk.Vehclass = curtrans.Vehclass
		askPlatepayChk.Vehplate = curtrans.Vehplate
		askPlatepayChk.TID = curtrans.TID

		rltCode, rltDes, rltPlatePayChk := device.PPCloudObj.FuncReqPlatePayCheck(askPlatepayChk)
		if rltCode != models.GRPCRESULT_OK {
			util.FileLogs.Info("%s,%s 车牌付校验失败:%s,%s", curtrans.TID, curtrans.Vehplate, rltCode, rltDes)
			return false
		} else {
			iRet := util.ConvertS2I(rltPlatePayChk.CheckResult)
			util.FileLogs.Info("%s,%s 车辆付校验结果:%s.\r\n", curtrans.TID, curtrans.Vehplate, util.GetPlatePayChkDes(iRet))
			if iRet == models.PLATEPAY_YES {
				curtrans.Vehclass = rltPlatePayChk.Vehclass
				curtrans.Vehplate = rltPlatePayChk.Vehplate
				return true
			}
		}
	}

	return false
}

//车脸付判断
func CheckCarfacePlatePay(curtrans models.MTransData) bool {
	if models.CHECK_YES == config.ConfigData["bPlatePay"].(int) &&
		!ChkTransSuc(curtrans) &&
		curtrans.CarfaceFlag == models.TRANSMATCH_YES {
		return true
	}

	return false
}

//电子车牌付判断
func CheckEPPlatePay(curtrans models.MTransData) bool {
	if !ChkTransSuc(curtrans) && curtrans.EPFlag == models.TRANSMATCH_YES {
		return true
	}

	return false
}

//交易支付成功
func ChkTransSuc(curtrans models.MTransData) bool {
	if curtrans.TransState == models.TRANSSTATE_SUC {
		return true
	}

	return false
}

//交易完全结束
func ChkTransFinish(curtrans models.MTransData) bool {
	if curtrans.TransState == models.TRANSSTATE_SUC && curtrans.TransFlow == models.TRANSFLOW_FINISH {
		return true
	}

	return false
}

func ChkTransEnd(curtrans models.MTransData) bool {
	if ChkTransSuc(curtrans) {
		return false
	}

	if curtrans.TransFlow == models.TRANSFLOW_END {
		return true
	}

	return false
}

func FuncCarpath(curtrans *models.MTransData) bool {
	util.FileLogs.Info("%s,%s 车辆通行信息请求……", curtrans.TID, curtrans.Vehplate)
	var askCarpath models.ReqCarpathData
	askCarpath.Vehclass = curtrans.Vehclass
	askCarpath.Vehplate = curtrans.Vehplate

	rltCode, rltDes, rltCarpath := device.PPCloudObj.FuncReqCarpath(askCarpath)
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("%s,%s 获取车辆通行信息失败:%s,%s", curtrans.TID, curtrans.Vehplate, rltCode, rltDes)
		return false
	} else {
		util.FileLogs.Info("%s,%s 获取车辆通行信息成功.\r\n", curtrans.TID, curtrans.Vehplate)
		curtrans.EntryNetwork = rltCarpath.EntryNetwork
		curtrans.EntryStationId = rltCarpath.EntryStationId
		curtrans.EntryLaneId = rltCarpath.EntryLaneId
		curtrans.EntryOperator = rltCarpath.EntryOperator
		curtrans.EntryShift = rltCarpath.EntryShift
		curtrans.EntryTime = rltCarpath.EntryTime
		curtrans.FlagStationid = rltCarpath.FlagStationid
		curtrans.HasCard = rltCarpath.HasCard

		return true
	}

}

func FuncFEE(curtrans *models.MTransData) bool {
	util.FileLogs.Info("%s,%s 车辆计费信息请求……", curtrans.TID, curtrans.Vehplate)
	/*	if models.GDebugxl {
			curtrans.Toll = "1"
			return true
		}
	*/
	var askfee models.ReqCalcFeeData
	askfee.Vehplate = curtrans.Vehplate
	askfee.Vehclass = curtrans.Vehclass
	askfee.Exitstaid = curtrans.ExitLandid
	askfee.Entrystaid = curtrans.EntryStationId
	askfee.Flagstaid = curtrans.FlagStationid
	rltCode, rltDes, rltFeeData := device.PPCloudObj.FuncReqFee(askfee)
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("%s,%s 计费失败:%s,%s", curtrans.TID, curtrans.Vehplate, rltCode, rltDes)
		return false
	} else {
		util.FileLogs.Info("%s,%s 计费成功:%s", curtrans.TID, curtrans.Vehplate, rltFeeData.Toll)
		curtrans.Toll = rltFeeData.Toll
		return true
	}
}

func FuncPay(curtrans *models.MTransData) bool {
	util.FileLogs.Info("%s,%s 车辆支付信息请求……", curtrans.TID, curtrans.Vehplate)
	//支付
	var askPay models.ReqPayData
	askPay.Operatorid = curtrans.ExitOperator
	askPay.Shiftid = curtrans.ExitShift
	askPay.TransDate = curtrans.ExitShiftdate
	askPay.Paytype = curtrans.Paytype
	askPay.PayCode = curtrans.Payid
	askPay.Vehplate = curtrans.Vehplate
	askPay.Toll = curtrans.Toll
	askPay.Transtime = curtrans.Transtime

	rltCode, rltDes, rltPayData := device.PPCloudObj.FuncReqPay(askPay) //
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("%s,%s 支付失败:%s,%s", curtrans.TID, curtrans.Vehplate, rltCode, rltDes)

		return false
	} else {
		util.FileLogs.Info("%s,%s 支付成功", curtrans.TID, curtrans.Vehplate)
		curtrans.Paytype = rltPayData.Paytype
		curtrans.Paymethod = rltPayData.Paymethod
		curtrans.Goodsno = rltPayData.Goodsno
		curtrans.Tradeno = rltPayData.Tradeno

		curtrans.PayFinishtime = util.GetNow(false)
		curtrans.Paytime = rltPayData.Paytime

		return true
	}
}

func GetTransHeadLock() *models.MTransData {
	GTransListLock.Lock()
	e := GTransList.Front()
	if e == nil {
		return nil
	}
	return e.Value.(*models.MTransData)
}

func ReaseTransUnLock() {
	GTransListLock.Unlock()
}

func PackagePrinter(curtrans models.MTransData) models.ReqPrinterTicketData {
	ticketinfo := models.ReqPrinterTicketData{}

	var pc1 models.PrinterContent
	pc1.Aligyntype = util.ConvertI2S(models.AlignCenter)
	pc1.Fontsize = util.ConvertI2S(models.SizeTimes)
	pc1.Content = fmt.Sprintf("%s", util.GetVehclassDes(util.ConvertS2I(curtrans.Vehclass)))
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	pc1.Aligyntype = util.ConvertI2S(models.AlignCenter)
	pc1.Fontsize = util.ConvertI2S(models.SizeTimes)
	iToll := util.ConvertS2I(curtrans.Toll)
	pc1.Content = fmt.Sprintf("%.2f元", (float32(iToll)+float32(0.005))/100.00)
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
	pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
	pc1.Content = fmt.Sprintf("时间:%s  工号:%s %s", curtrans.Transtime, curtrans.ExitOperator, curtrans.ExitLandid)
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	if len(curtrans.Goodsno) > 0 {
		pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
		pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
		pc1.Content = fmt.Sprintf("支付方式:%s  交易序号:%s", util.GetPayMethodDes(curtrans.Paymethod), curtrans.Goodsno)
		ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)
	}

	if len(curtrans.Payid) > 0 {
		pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
		pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
		pc1.Content = fmt.Sprintf("二维码:%s", curtrans.Payid)
		ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)
	}

	return ticketinfo
}

//交易来源
func FuncGetTransFrom(curtrans *models.MTransData) string {
	return curtrans.TransFrom
}

func UpdateFstTrans(curtrans models.MTransData) {
	b := false
	ev := GetTransHeadLock()
	if ev != nil {
		if ev.FlowNo == curtrans.FlowNo {
			ev = &curtrans
			b = true
		}
	}

	ReaseTransUnLock()

	if !b {
		util.FileLogs.Info("更新交易成功异常")
	}
}

func FuncInsertFlowProc() {
	GTransListLock.Lock()
	for e := GTransList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(*models.MTransData)

		if !ChkTransFinish(*ev) && ChkTransSuc(*ev) {
			util.FileLogs.Info("InsertFlow2Db :%s.", ev.FlowNo)

			ev.TransFlow = models.TRANSFLOW_FINISH

			//写文本
			device.Flow2Txt(*ev)
			//图片写文件，保存
			device.Img2Txt(*ev)
			//写历史队列
			Append2Histroylist(ev)
		}
	}

	GTransListLock.Unlock()
}

func FuncFlagStaTransDel() {
	GTransListLock.Lock()
	GTransList.Init()
	GTransListLock.Unlock()
}

func FuncFeedispShowfee(vehplate, toll string) {
	l1 := ""
	l2 := ""
	l3 := ""

	l1 = vehplate

	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_ENTRY, models.TOLLMODE_FLAG:
		l2 = "交易成功"
		l3 = fmt.Sprintf("请快速驶离")
	case models.TOLLMODE_OPEN, models.TOLLMODE_EXIT:
		l2 = util.GetFeeShow(toll)
		l3 = "支付成功"

	}
	FuncFeedispShowTip(l1, l2, l3)
}

func FuncFeedispShowCarfacefee(vehplate, toll, vehclass, brand string) {
	l1 := ""
	l2 := ""
	l3 := ""

	l1 = vehplate

	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_ENTRY, models.TOLLMODE_FLAG:
		l2 = "交易成功"
		l3 = fmt.Sprintf("请快速驶离")
	case models.TOLLMODE_OPEN, models.TOLLMODE_EXIT:
		l2 = util.GetFeeShow(toll)
		l3 = brand + " " + util.GetVehclassDes(util.ConvertS2I(vehclass))

	}
	FuncFeedispShowTip(l1, l2, l3)
}

func FuncFeedispShowTip(l1, l2, l3 string) {
	if device.PDevFeedispObj == nil {
		return
	}

	var req models.FeedispShowData

	req.Line1 = l1
	req.Line2 = l2
	req.Line3 = l3
	req.Color1 = util.Convertb2s(models.ColorGreen)
	req.Color2 = util.Convertb2s(models.ColorGreen)
	req.Color3 = util.Convertb2s(models.ColorGreen)
	device.PDevFeedispObj.FuncFeedispShow(req)

	FuncFeedispAlarmOff()
}

func FuncFeedispAlarmOff() {
	if device.PDevFeedispObj == nil {
		return
	}

	var reqAlarm models.FeedispAlarmData
	reqAlarm.AlarmValue = util.Convertb2s(models.Alarmoff)
	reqAlarm.AlarmTm = util.ConvertI2S(0)
	device.PDevFeedispObj.FuncFeedispAlarm(reqAlarm)
}

func FuncFeedispShowErr(l1, l2, l3 string) {
	if device.PDevFeedispObj == nil {
		return
	}

	var req models.FeedispShowData
	var reqAlarm models.FeedispAlarmData

	req.Line1 = l1
	req.Line2 = l2
	req.Line3 = l3
	req.Color1 = util.Convertb2s(models.ColorRed)
	req.Color2 = util.Convertb2s(models.ColorRed)
	req.Color3 = util.Convertb2s(models.ColorRed)
	device.PDevFeedispObj.FuncFeedispShow(req)

	reqAlarm.AlarmValue = util.Convertb2s(models.Alarmon)
	reqAlarm.AlarmTm = util.ConvertI2S(models.Alarmtimeshort)
	device.PDevFeedispObj.FuncFeedispAlarm(reqAlarm)
}

func FuncCtlLanganProc(val int) {
	//抬杆
	if device.PIOObj != nil {
		device.PIOObj.FuncLanganProc(val)

		switch val {
		case models.COIL_OPEN:
			util.FileLogs.Info("控制抬杆")
		case models.COIL_CLOSE:
			util.FileLogs.Info("控制落杆")
		}

	}
}

//交易超过10s还停留，则给予提示信息
//判断超时是否删除交易
func GetDelFlagByStationMode() bool {
	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_FLAG:
		return true
	case models.TOLLMODE_EXIT, models.TOLLMODE_ENTRY, models.TOLLMODE_OPEN:
		if models.STATIONMODE_FREEFLOWEXIT == config.ConfigData["bFreeflow"].(int) {
			return true
		}

	}

	return false
}

func ChkTransTimeout() {
	bDel := false
	iTimeout := 10

	if !GetDelFlagByStationMode() {
		return
	}

	GTransListLock.Lock()

	if e := GTransList.Front(); e != nil {
		ev := e.Value.(*models.MTransData)
		if util.ChkTimeOut(ev.Transtime, iTimeout) {
			//如果是etc交易，b2，b3，b4都正确，且还未结束，则等待
			if ev.ETCFlag == models.TRANSMATCH_YES &&
				!ChkTransSuc(*ev) &&
				ev.TransFlow != models.TRANSFLOW_END &&
				ev.ErrcodeB2 == models.FRAMERLT_OK &&
				ev.ErrcodeB3 == models.FRAMERLT_OK &&
				ev.ErrcodeB4 == models.FRAMERLT_OK {
				util.FileLogs.Info("%s ETC交易仍在处理中，等待", ev.OBUID)
			} else {
				util.FileLogs.Info("%s,%s,%s:%d ChkTransTimeout 超时删除", ev.Vehplate, ev.TID, ev.OBUID, ev.TransState)
				GTransList.Remove(e)
				bDel = true
			}
		}

	}
	GTransListLock.Unlock()

	if bDel {
		ShowNextFeedisp()
	}
}

func TestChkTransTimeout() {
	if !models.GDebugxl {
		return
	}

	bDel := false
	iTimeout := 60

	GTransListLock.Lock()

	if e := GTransList.Front(); e != nil {
		ev := e.Value.(*models.MTransData)
		if util.ChkTimeOut(ev.Transtime, iTimeout) {
			util.FileLogs.Info("%s:%d TestChkTransTimeout 超时删除", ev.Vehplate, ev.TransState)
			GTransList.Remove(e)
			bDel = true
		}

	}
	GTransListLock.Unlock()

	if bDel {
		ShowNextFeedisp()
	}
}

func ShowNextFeedisp() {
	if config.ConfigData["bFreeflow"].(int) == models.STATIONMODE_FREEFLOWEXIT {
		return
	}

	GTransListLock.Lock()
	defer GTransListLock.Unlock()

	//如果下一笔是失败，则落杆
	if e := GTransList.Front(); e != nil {
		ev := e.Value.(*models.MTransData)
		if !ChkTransSuc(*ev) {
			FuncCtlLanganProc(models.COIL_CLOSE)
			UpdateFeeErrDisp(ev.Vehplate, ev.TransFailState)
		} else {
			FuncCtlLanganProc(models.COIL_OPEN)

			switch ev.TransMemo {
			case models.TRANSMEMO_FREEFLOWPASS:
				l2 := util.GetFeeShow(ev.Toll)
				FuncFeedispShowTip(ev.Vehplate, l2, "已完成缴费")
			case models.TRANSMEMO_HISTORY:
				FuncFeedispShowTip(ev.Vehplate, "已完成缴费", "请快速驶离")
			default:
				if ev.CarfaceFlag == models.TRANSMATCH_YES && ev.Paytype == models.PAYTYPE_PLATE {
					FuncFeedispShowCarfacefee(ev.Vehplate, ev.Toll, ev.Vehclass, ev.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand)
				} else {
					FuncFeedispShowfee(ev.Vehplate, ev.Toll)
				}
			}
		}
	}

	if GTransList.Len() == 0 {
		FuncCtlLanganProc(models.COIL_CLOSE)
		ShowDefaultFeedisp()

		if device.PEquipmentObj != nil {
			device.PEquipmentObj.Package74(device.ShowTip16(models.SELFDEFAULTTIP, models.ColorGreen))
		}
	}
}

func ShowDefaultFeedisp() {
	line1 := config.ConfigData["feeDispLine1"].(string)
	line2 := config.ConfigData["feeDispLine2"].(string)
	line3 := config.ConfigData["feeDispLine3"].(string)
	FuncFeedispShowTip(line1, line2, line3)
}

func ChkFreeflowSuc(vehplate string, vehclass string, strTid string) (bool, models.RstChkFreeflowData) {
	util.FileLogs.Info("%s,%s ChkFreeflowSuc 自由流结果问询", strTid, vehplate)

	rst := models.RstChkFreeflowData{}

	var req models.ReqChkFreeflowData
	req.Vehplate = vehplate
	req.Vehclass = vehclass
	req.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	req.Exitstaid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	req.TID = strTid

	rltCode, rltDes, rltData := device.PPCloudObj.FuncChkFreeflow(req)
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("自由流问询失败:%s,%s", rltCode, rltDes)
		return false, rst
	} else {
		util.FileLogs.Info("自由流问询成功:%s", rltData.Result)
		return true, rltData
	}

	return false, rst
}

func NotifyFreeflow(vehplate string, vehclass string, strTid string) bool {
	util.FileLogs.Info("%s,%s NotifyFreeflow 自由流通行通知", strTid, vehplate)

	var req models.ReqNotifyFreeflowData
	req.Vehplate = vehplate
	req.Vehclass = vehclass
	req.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	req.Exitstaid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	req.TID = strTid

	rltCode, rltDes, rltData := device.PPCloudObj.FuncNotifyFreeflow(req)
	if rltCode != models.GRPCRESULT_OK {
		util.FileLogs.Info("自由流通知失败:%s,%s", rltCode, rltDes)
		return false
	} else {
		util.FileLogs.Info("自由流通知成功:%s", rltData.Result)
		return true
	}

	return true
}

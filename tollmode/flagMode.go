package tollmode

import (
	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//FlagManage 标识站
type FlagManage struct {
}

func (p *FlagManage) createByCoil() {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_FLAG)

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.FlagTime = transdata.Transtime
	transdata.FlagStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))

	transdata.CreateBy = models.CREATEBY_COIL
	transdata.CoilFlag = models.TRANSMATCH_YES
	transdata.TransState = models.TRANSSTATE_INIT

	GTransList.PushBack(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

func (p *FlagManage) createByPlateReg(info models.RstVehRecognizeInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_FLAG)

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.FlagTime = transdata.Transtime
	transdata.FlagStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))

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

func (p *FlagManage) createByEP(info models.EPResultReadInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_FLAG)

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.FlagTime = transdata.Transtime
	transdata.FlagStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))

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

func (p *FlagManage) EPResultFreeflowProc(info models.EPResultReadInfo) {
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

//目前只考虑电子车牌
func (p *FlagManage) GoRun() {
	//defer util.PanicHandler()
	util.FileLogs.Info("启动标识站模式")

	for {
		//车牌付
		p.PlatePayFlowProc()
		//etc付
		p.ETCPayFlowProc()

		//成功入库
		FuncInsertFlowProc()

		//不论成功与否，交易队列清空
		FuncFlagStaTransDel()

		util.MySleep_ms(10)
	}
}

//电子车牌或车脸识别处理流程
func (p *FlagManage) PlatePayFlowProc() {
	GTransListLock.Lock()
	for e := GTransList.Front(); e != nil; e = e.Next() {
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

				if ev.EPFlag == models.TRANSMATCH_YES {
					ev.TransFrom = models.TRANSFROM_EP
				} else {
					ev.TransFrom = models.TRANSFROM_PLATE
				}

			} else {
				ev.TransFlow = models.TRANSFLOW_END
				ev.TransState = models.TRANSSTATE_FAIL
				ev.TransFailState = models.TRANSFAIL_NOTSUPPORTPLATEPAY
			}
		}
	}

	GTransListLock.Unlock()

}

//etc卡片付流程
func (p *FlagManage) ETCPayFlowProc() {

}

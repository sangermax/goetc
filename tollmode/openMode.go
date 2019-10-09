package tollmode

import (
	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//OpenMode 开放式
type OpenManage struct {
	TransManualFlow  int //自助操作流程
	TransProcState   int
	TransProcTimeout int   //单位 ms
	TransLastTimeMs  int64 //单位 ms

	FstTransdata     models.MTransData
	selfWorkStation  byte
	equipWorkStation byte
}

//交易创建分开，每种模式赋值有出入
func (p *OpenManage) createByCoil() {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_OPEN)

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

	//开放道，出入口信息一致
	transdata.EntryNetwork = transdata.ExitNetwork
	transdata.EntryStationId = transdata.ExitStationid
	transdata.EntryLaneId = transdata.ExitLandid
	transdata.EntryOperator = transdata.ExitOperator
	transdata.EntryShift = transdata.ExitShift
	transdata.EntryTime = transdata.Transtime

	GTransList.PushBack(transdata)
	util.FileLogs.Info("%s 创建第%d笔交易:%s", util.GetCreateByDes(transdata.CreateBy), GTransList.Len(), transdata.Vehplate)
}

func (p *OpenManage) createByPlateReg(info models.RstVehRecognizeInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_OPEN)

	transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.ExitShift = util.GetShiftId(transdata.Transtime)

	//开放道，出入口信息一致
	transdata.EntryNetwork = transdata.ExitNetwork
	transdata.EntryStationId = transdata.ExitStationid
	transdata.EntryLaneId = transdata.ExitLandid
	transdata.EntryOperator = transdata.ExitOperator
	transdata.EntryShift = transdata.ExitShift
	transdata.EntryTime = transdata.Transtime

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

func (p *OpenManage) createByEP(info models.EPResultReadInfo) {
	//创建
	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_OPEN)

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

	//开放道，出入口信息一致
	transdata.EntryNetwork = transdata.ExitNetwork
	transdata.EntryStationId = transdata.ExitStationid
	transdata.EntryLaneId = transdata.ExitLandid
	transdata.EntryOperator = transdata.ExitOperator
	transdata.EntryShift = transdata.ExitShift
	transdata.EntryTime = transdata.Transtime

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

func (p *OpenManage) EPResultFreeflowProc(info models.EPResultReadInfo) {
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

func (p *OpenManage) EPResultProc(info models.EPResultReadInfo) {
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

func (p *OpenManage) GoRun() {
	//defer util.PanicHandler()
	util.FileLogs.Info("启动开放模式")

	for {

		util.MySleep_s(30)
	}

}

func (p *OpenManage) FuncUpdateManualProcState(procState int) {
	p.TransProcState = procState
}

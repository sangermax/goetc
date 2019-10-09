package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/util"
)

var (
	ChanVehIn   chan bool //车辆进入信号
	ChanVehOut  chan bool //车辆离开信号
	ChanPlate   chan bool //车牌结果信号
	ChanCard    chan bool //检测到卡片信号
	ChanEP      chan bool //电子车牌信号
	ChanCarpath chan bool //车辆行驶信息
	ChanETC     chan bool //ETC信号

	ChanFee            chan bool //计费结果信号
	ChanPayCheck       chan bool
	ChanPay            chan bool //支付结果信号
	ChanScan           chan bool //扫码结果信号
	ChanChkFreeflow    chan bool //自由流结果信号
	ChanNotifyFreeflow chan bool //自由流过车通知

	ChanExitflow chan bool //出口流水
	ChanExitimg  chan bool //图片流水

	PIOObj         *DevIO
	PPlateObj      *DevPlate
	PReaderObj     *DevReader
	PPrinterObj    *DevPrinter
	PScanObj       *DevScan
	PEquipmentObj  *DevSelfEquipment
	PPCloudObj     *DevCloud
	PEPAntObj      *DevEPAnt
	PETCAntObj     *DevETCAnt
	PDevFeedispObj *DevFeedisp
)

func InitObjs() {
	ChanVehIn = make(chan bool, 1)
	ChanVehOut = make(chan bool, 1)
	ChanPlate = make(chan bool, 1)
	ChanCard = make(chan bool, 1)
	ChanFee = make(chan bool, 1)
	ChanPayCheck = make(chan bool, 1)
	ChanPay = make(chan bool, 1)
	ChanScan = make(chan bool, 1)
	ChanExitflow = make(chan bool, 1)
	ChanExitimg = make(chan bool, 1)
	ChanEP = make(chan bool, 1)
	ChanETC = make(chan bool, 1)
	ChanCarpath = make(chan bool, 1)
	ChanChkFreeflow = make(chan bool, 1)
	ChanNotifyFreeflow = make(chan bool, 1)

	PPCloudObj = new(DevCloud)
	PEPAntObj = new(DevEPAnt)
	PETCAntObj = new(DevETCAnt)

	PEPAntObj.InitDevAnt()
	PPCloudObj.InitDevCloud()
	PETCAntObj.InitDevAnt()

	/*
		PDevFeedispObj = new(DevFeedisp)
		if config.ConfigData["tollmode"].(int) != models.TOLLMODE_FLAG {
			PDevFeedispObj.InitDevFeedisp()

			var req models.FeedispShowData
			req.Line1 = "自助缴费车道"
			req.Line2 = "  系统启动  "
			req.Line3 = ""
			req.Color1 = util.Convertb2s(models.ColorYellow)
			req.Color2 = util.Convertb2s(models.ColorYellow)
			req.Color3 = util.Convertb2s(models.ColorYellow)
			PDevFeedispObj.FuncFeedispShow(req)
		}

		switch config.ConfigData["tollmode"].(int) {
		case models.TOLLMODE_OPEN, models.TOLLMODE_EXIT:
			{
				//自由流车道除外
				if config.ConfigData["bFreeflow"].(int) != models.STATIONMODE_FREEFLOWEXIT {
					PIOObj = new(DevIO)
					PPlateObj = new(DevPlate)
					PReaderObj = new(DevReader)
					PPrinterObj = new(DevPrinter)
					PScanObj = new(DevScan)
					PEquipmentObj = new(DevSelfEquipment)

					PEquipmentObj.InitDev()
					PPrinterObj.InitDevPrinter()
					PScanObj.InitDevScan()
					PIOObj.InitDevIO()

					PPlateObj.InitDevPlate()
					PReaderObj.InitDevReader()
					//为了设备归位，主要是自助设备
					util.MySleep_s(5)
				}

			}

		}
	*/
}

func GetGrpcDes(devtype int) string {
	switch devtype {
	case models.DEVGRPCTYPE_CLOUD:
		return "GRPC云路由"
	case models.DEVGRPCTYPE_SCAN:
		return "GRPC扫码"
	case models.DEVGRPCTYPE_PRINTER:
		return "GRPC打印机"
	case models.DEVGRPCTYPE_IO:
		return "GRPCIO"
	case models.DEVGRPCTYPE_PLATE:
		return "GRPC车牌识别"
	case models.DEVGRPCTYPE_FEEDISP:
		return "GRPC费显"
	case models.DEVGRPCTYPE_EP:
		return "GRPC电子车牌天线"
	case models.DEVGRPCTYPE_READER:
		return "GRPC读写器"
	case models.DEVGRPCTYPE_ETC:
		return "ETC天线"
	}

	return "未知" + util.ConvertI2S(devtype)
}

func GetDevSrvDes(devtype int) string {
	switch devtype {
	case models.DEVTYPE_IO:
		return "IO服务"
	case models.DEVTYPE_PLATE:
		return "车牌服务"
	case models.DEVTYPE_PRINTERUP:
		return "打印机UP服务"
	case models.DEVTYPE_PRINTERDN:
		return "打印机DN服务"
	case models.DEVTYPE_SCANUP:
		return "扫码UP服务"
	case models.DEVTYPE_SCANDN:
		return "扫码DN服务"
	case models.DEVTYPE_EQUIPMENT:
		return "GRPC自助缴费设备"
	case models.DEVTYPE_FEEDISP:
		return "GRPC费显设备"
	case models.DEVTYPE_ANTEP:
		return "EP电子车牌天线设备"
	case models.DEVTYPE_ETC:
		return "ETC天线集群"
	}

	return "未知" + util.ConvertI2S(devtype)
}

func GetCmdDes(stype string) string {
	switch stype {
	case models.GRPCTYPE_Init:
		return stype + "-初始化"
	case models.GRPCTYPE_FEE:
		return stype + "-计费"

	case models.GRPCTYPE_CHECKPAY:
		return stype + "-车牌付校验"
	case models.GRPCTYPE_PAY:
		return stype + "-支付"
	case models.GRPCTYPE_EXITFLOW:
		{
			switch config.ConfigData["tollmode"].(int) {
			case models.TOLLMODE_OPEN:
				return stype + "-开放道流水"
			case models.TOLLMODE_ENTRY:
				return stype + "-入口流水"
			case models.TOLLMODE_FLAG:
				return stype + "-标识站流水"
			default:
				return stype + "-出口流水"
			}
		}
	case models.GRPCTYPE_EXITIMG:
		return stype + "-出口图片流水"
	case models.GRPCTYPE_DEVSTATE:
		return stype + "-设备状态"
	case models.GRPCTYPE_CONTROL:
		return stype + "-远程控制"
	case models.GRPCTYPE_CARPATH:
		return stype + "-车辆通行信息"
	case models.GRPCTYPE_CHECKFREEFLOW:
		return stype + "-自由流校验"
	case models.GRPCTYPE_NOTIFYFREEFLOW:
		return stype + "-自由流通知过车"

	case models.GRPCTYPE_IOSTATE:
		return stype + "-IO状态"
	case models.GRPCTYPE_PLATESTATE:
		return stype + "-PLATE状态"
	case models.GRPCTYPE_PRINTERSTATE:
		return stype + "-PRINTER状态"
	case models.GRPCTYPE_SCANSTATE:
		return stype + "-SCAN状态"

	case models.GRPCTYPE_OPENSCAN:
		return stype + "-开启扫码"
	case models.GRPCTYPE_CLOSESCAN:
		return stype + "-关闭扫码"
	case models.GRPCTYPE_PRINTTICKET:
		return stype + "-出票"
	case models.GRPCTYPE_IOCONTROL:
		return stype + "-IO控制"

	case models.GRPCTYPE_EPAntInit:
		return stype + "-EP天线初始化"
	case models.GRPCTYPE_EPAntRealRead:
		return stype + "-EP天线读取信息"
	case models.GRPCTYPE_EPAntState:
		return stype + "-EP天线状态信息"

	case models.GRPCTYPE_FEEDISPSTATE:
		return stype + "-费显状态信息"
	case models.GRPCTYPE_FEEDISPSHOW:
		return stype + "-费显屏显信息"
	case models.GRPCTYPE_FEEDISPALARM:
		return stype + "-费显报警信息"

	case models.GRPCTYPE_READERSTATE:
		return stype + "-读写器状态信息"
	case models.GRPCTYPE_READERM1READ:
		return stype + "-读写器读M1卡"
	case models.GRPCTYPE_READERM1WRITE:
		return stype + "-读写器写M1卡"
	case models.GRPCTYPE_READERETCREAD:
		return stype + "-读写器读ETC卡"
	case models.GRPCTYPE_READERETCWRITE:
		return stype + "-读写器写ETC卡"
	case models.GRPCTYPE_READERETCPAY:
		return stype + "-读写器ETC支付"
	case models.GRPCTYPE_READERETCBALANCE:
		return stype + "-读写器ETC卡余额"
	case models.GRPCTYPE_READERCARDTYPE:
		return stype + "-读写器卡片类型"
	case models.GRPCTYPE_READERCARDCLOSE:
		return stype + "-读写器关闭卡片"

	case models.GRPCTYPE_PLATE:
		return stype + "-车脸识别结果通知"

	case models.GRPCTYPE_ETCAntInit:
		return stype + "-ETC天线初始化"
	case models.GRPCTYPE_ETCAntState:
		return stype + "-ETC天线状态"
	case models.GRPCTYPE_ETCMSG:
		return stype + "-ETC交互信息"
	default:
		return stype + "-未知"
	}
}

func GoAutoDevState() {
	for {
		var info models.AutoDevStateData
		info.ExitNetwork = ""
		info.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
		info.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
		info.EquipmentDev = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.PrinterDev1 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.PrinterDev2 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.PlateDev = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.Coil1 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.Coil2 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.Scan1 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.Scan2 = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.EpDev = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.FeeDisp = util.ConvertI2S(models.DEVSTATE_UNKNOWN)
		info.FeeAlarm = util.ConvertI2S(models.DEVSTATE_UNKNOWN)

		if PEquipmentObj != nil {
			info.EquipmentDev = util.ConvertI2S(PEquipmentObj.State)
		}

		if PPrinterObj != nil {
			info.PrinterDev1 = util.ConvertI2S(PPrinterObj.UpState)
			info.PrinterDev2 = util.ConvertI2S(PPrinterObj.DnState)
		}

		if PPlateObj != nil {
			info.PlateDev = util.ConvertI2S(PPlateObj.State)
		}

		if PIOObj != nil {
			info.Coil1 = util.ConvertI2S(PIOObj.StateCoil1)
			info.Coil2 = util.ConvertI2S(PIOObj.StateCoil2)
		}

		if PScanObj != nil {
			info.Scan1 = util.ConvertI2S(PScanObj.UpState)
			info.Scan2 = util.ConvertI2S(PScanObj.DnState)
		}

		if PEPAntObj != nil {
			info.EpDev = util.ConvertI2S(PEPAntObj.FuncGetEPAntState())
		}

		if PPCloudObj != nil {
			PPCloudObj.FuncReqDevState(info)
		}
		util.MySleep_s(300)
	}
}

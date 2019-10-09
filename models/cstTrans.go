package models

type GRPCTaskData struct {
	Reqtime string
	Sno     string
	SType   string

	ReqData interface{}
	//返回结果
	Result  ResultInfo
	RstData []byte
}

const (
	GRPCRESULT_OK   = "0"
	GRPCRESULT_FAIL = "1"
)

const (
	//云服务使用
	GRPCTYPE_Init = "0"
	GRPCTYPE_FEE  = "1"

	GRPCTYPE_CHECKPAY       = "2"
	GRPCTYPE_PAY            = "3"
	GRPCTYPE_EXITFLOW       = "4"
	GRPCTYPE_EXITIMG        = "5"
	GRPCTYPE_DEVSTATE       = "6"
	GRPCTYPE_CONTROL        = "7"
	GRPCTYPE_CARPATH        = "8"
	GRPCTYPE_CHECKFREEFLOW  = "9"
	GRPCTYPE_NOTIFYFREEFLOW = "10"

	//扫码用
	GRPCTYPE_SCANSTATE = "101"
	GRPCTYPE_OPENSCAN  = "102"
	GRPCTYPE_CLOSESCAN = "103"

	//打印机使用
	GRPCTYPE_PRINTERSTATE = "110"
	GRPCTYPE_PRINTTICKET  = "111"

	//IO控制用
	GRPCTYPE_IOSTATE   = "120"
	GRPCTYPE_IOCONTROL = "121"

	//车牌用
	GRPCTYPE_PLATESTATE = "130"
	GRPCTYPE_PLATE      = "131"

	//天线电子车牌
	GRPCTYPE_EPAntInit     = "140"
	GRPCTYPE_EPAntRealRead = "141"
	GRPCTYPE_EPAntState    = "142"

	//费显
	GRPCTYPE_FEEDISPSTATE = "150"
	GRPCTYPE_FEEDISPSHOW  = "151"
	GRPCTYPE_FEEDISPALARM = "152"

	//读卡器
	GRPCTYPE_READERSTATE      = "160"
	GRPCTYPE_READERM1READ     = "161"
	GRPCTYPE_READERM1WRITE    = "162"
	GRPCTYPE_READERETCREAD    = "163"
	GRPCTYPE_READERETCWRITE   = "164"
	GRPCTYPE_READERETCPAY     = "165"
	GRPCTYPE_READERETCBALANCE = "166"
	GRPCTYPE_READERCARDTYPE   = "168"
	GRPCTYPE_READERCARDCLOSE  = "169"

	//ETC天线
	GRPCTYPE_ETCAntInit  = "170"
	GRPCTYPE_ETCAntState = "171"
	GRPCTYPE_ETCMSG      = "172"
)

//公共部分
type HeaderInfo struct {
	Reqtime string `json:"reqTime"`
	Rsttime string `json:"rstTime"`
	Key     string `json:"key"`
	No      string `json:"no"`
}

type ResultInfo struct {
	ResultValue string `json:"code"`
	ResultDes   string `json:"msg"`
}

//计费接口//////////////////////////////////////////////////////////////
//计费
type ReqCalcFeeData struct {
	Vehplate   string `json:"vehPlate"`
	Vehclass   string `json:"vehClass"`
	Entrystaid string `json:"entryStationId"`
	Exitstaid  string `json:"exitStationId"`
	Flagstaid  string `json:"flagStationId"`
}

type RstCalcFeeData struct {
	Toll string `json:"toll"`
}

//支付接口//////////////////////////////////////////////////////////////
//车牌付校验
type ReqCheckPlatepayData struct {
	Vehplate string `json:"vehPlate"`
	Vehclass string `json:"vehClass"`
	TID      string `json:"tid"` //标签唯一码
}

type RstCheckPlatepayData struct {
	CheckResult string `json:"payResult"`
	Vehplate    string `json:"vehPlate"`
	Vehclass    string `json:"vehClass"`
}

//请求支付
type ReqPayData struct {
	Operatorid string `json:"operatorId"`
	Shiftid    string `json:"shiftId"`
	TransDate  string `json:"statisticsDate"`
	Paytype    string `json:"payType"`
	PayCode    string `json:"payCode"`
	Vehplate   string `json:"vehPlate"`
	Toll       string `json:"toll"`
	Transtime  string `json:"statisticsTime"`
}

//支付应答结果
type RstPayData struct {
	Paytype   string `json:"payType"`
	Paymethod string `json:"payMethod"`
	Goodsno   string `json:"goodsNo"`
	Tradeno   string `json:"tradeNo"`
	Toll      string `json:"toll"`
	Paytime   string `json:"payTime"`
}

////////////////////////////////////////////////////////////////
//MTransData 交易数据
type MTransData struct {
	FlowNo string `json:"flowNo"`

	//正确的车牌，即电子车牌/etc卡片中记录
	Vehclass string `json:"vehClass"`
	Vehplate string `json:"vehPlate"`
	VehColor string `json:"vehColor"`

	ExitNetwork   string `json:"exitNetwork"`
	ExitStationid string `json:"exitStationId"`
	ExitLandid    string `json:"exitLaneId"`
	ExitOperator  string `json:"exitOperator"`
	ExitShiftdate string `json:"exitShiftDate"`
	ExitShift     string `json:"exitShift"`

	Paytype   string `json:"payType"`   //1.车脸付；2.扫码付；3.电子车牌付；4.卡片付；
	Paymethod string `json:"payMethod"` //1.微信；2.支付宝；3.虚拟卡钱包支付
	Payid     string `json:"payId"`     //支付标识，卡支付为卡号，扫码支付为二维码（即支付码）
	Transtime string `json:"exitTime"`
	Paytime   string `json:"payTime"`
	Toll      string `json:"totalToll"`

	Goodsno   string `json:"goodsNo"`
	Tradeno   string `json:"tradeNo"`
	PrintFlag string `json:"printFlag"`

	OBUID      string `json:"obuid"`
	Psamtermid string `json:"psamtermid"`
	Cardtradno string `json:"cardtradno"`
	Psamtradno string `json:"psamtradno"`
	Tac        string `json:"tac"`

	ProCardId string `json:"proCardId"` //使用自助缴费机，卡号
	CardId    string `json:"cardId"`
	Cardtype  int    `json:"cardtype"`

	EntryNetwork   string `json:"entryNetwork"`   //入口网络编号
	EntryStationId string `json:"entryStationId"` //入口收费站编号
	EntryLaneId    string `json:"entryLaneId"`    //入口车道编号
	EntryOperator  string `json:"entryOperator"`  //入口收费员工号
	EntryShift     string `json:"entryShift"`     //入口班次
	EntryTime      string `json:"entryTime"`      //入口时间
	FlagStationid  string `json:"flagStationid"`  //标识站编号
	FlagTime       string `json:"flagTime"`       //标识站时间
	HasCard        string `json:"hasCard"`        //入口发卡标志

	RegVehclass string `json:"RegVehclass"` //车脸识别的车型
	RegVehplate string `json:"RegVehplate"` //车脸识别的车牌
	AntennaID   string `json:"AntennaID"`   //天线id
	TID         string `json:"TID"`         //标签唯一码

	TransFrom   string `json:"transFrom"`   //交易来源
	StationMode string `json:"stationMode"` //收费站模式

	//交易其他信息
	PayFinishtime  string `json:"payFinishtime"`  //支付成功记录返回时间，自用
	TransFlow      int    `json:"transFlow"`      //交易流程
	TransState     int    `json:"transState"`     //交易状态
	TransMemo      string `json:"transMemo"`      //交易备注 0:原创交易，1：历史交易恢复，2：自由流交易
	TransFailState int    `json:"transFailState"` //交易失败状态
	CreateBy       string `json:"createBy"`       //交易创建方式 //车检创，ep创，etc天线创等
	CoilFlag       string `json:"coilFlag"`       //车检匹配
	EPFlag         string `json:"epFlag"`         //电子车牌匹配
	ETCFlag        string `json:"epFlag"`         //ETC匹配
	CarfaceFlag    string `json:"carfaceFlag"`    //车脸识别匹配

	RstVehPlateInfo RstVehRecognizeInfo `json:"plateInfo"`

	File0015 F0015Info `json:"file0015"`
	File0019 F0019Info `json:"file0019"`

	//卡交易信息 原始数据
	StrFile0015   string `json:"strFile0015"`   //0015文件 16进制形式
	StrFile0019   string `json:"strFile0019"`   //0019文件 16进制形式
	BeforeBalance uint   `json:"beforeBalance"` //
	AfterBalance  uint   `json:"afterBalance"`

	//ETC
	//ETC信息来源
	ETCAntGrpKey string `json:"key"`
	ETCAntNo     string `json:"antno"`
	//交互信息
	ErrcodeB2 int        `json:"errcodeb2"`
	ErrcodeB3 int        `json:"errcodeb3"`
	ErrcodeB4 int        `json:"errcodeb4"`
	Frameb2   ETCFrameB2 `json:"frameb2"`
	Frameb3   ETCFrameB3 `json:"frameb3"`
	Frameb4   ETCFrameB4 `json:"frameb4"`
	Frameb5   ETCFrameB5 `json:"frameb5"`
	Frameb7   ETCFrameB7 `json:"frameb7"`
	Framec3   ETCFrameC3 `json:"framec3"`
	Framec5   ETCFrameC6 `json:"framec6"`
}

//出口交易数据
type ReqExitFlowData struct {
	FlowNo   string `json:"flowNo"`
	Vehclass string `json:"vehClass"`
	Vehplate string `json:"vehPlate"`
	VehColor string `json:"vehColor"`

	ExitNetwork   string `json:"exitNetwork"`
	ExitStationid string `json:"exitStationId"`
	ExitLandid    string `json:"exitLaneId"`
	ExitOperator  string `json:"exitOperator"`
	ExitShiftdate string `json:"exitShiftDate"`
	ExitShift     string `json:"exitShift"`

	Paytype   string `json:"payType"`   //1.车牌付；2.扫码付；3.卡片付
	Paymethod string `json:"payMethod"` //1.微信；2.支付宝
	Payid     string `json:"payId"`     //支付标识，卡支付为卡号，扫码支付为二维码（即支付码）
	Transtime string `json:"exitTime"`
	Paytime   string `json:"payTime"`
	Toll      string `json:"totalToll"`

	Goodsno   string `json:"goodsNo"`
	Tradeno   string `json:"tradeNo"`
	PrintFlag string `json:"printFlag"`

	Psamtermid string `json:"psamtermid"`
	Cardtradno string `json:"cardtradno"`
	Psamtradno string `json:"psamtradno"`
	Tac        string `json:"tac"`

	CardId string `json:"cardId"`

	EntryNetwork   string `json:"entryNetwork"`   //入口网络编号
	EntryStationId string `json:"entryStationId"` //入口收费站编号
	EntryLaneId    string `json:"entryLaneId"`    //入口车道编号
	EntryOperator  string `json:"entryOperator"`  //入口收费员工号
	EntryShift     string `json:"entryShift"`     //入口班次
	EntryTime      string `json:"entryTime"`      //入口时间
	FlagStationid  string `json:"flagStationid"`  //标识站编号
	FlagTime       string `json:"flagTime"`       //标识站时间

	RegVehclass string `json:"RegVehclass"` //车脸识别的车型
	RegVehplate string `json:"RegVehplate"` //车脸识别的车牌
	AntennaID   string `json:"AntennaID"`   //天线id
	TID         string `json:"TID"`         //标签唯一码

	TransFrom   string `json:"transFrom"`   //交易来源
	StationMode string `json:"stationMode"` //收费站模式
	TransMemo   string `json:"transMemo"`   //交易备注 0:原创交易，1：历史交易恢复，2：自由流交易
}

//出口图片数据
type ExitImgAddsData struct {
	FlowNo        string `json:"flowNo"`
	ExitNetwork   string `json:"exitNetwork"`
	ExitStationid string `json:"exitStationId"`
	ExitLandid    string `json:"exitLaneId"`
	ExitOperator  string `json:"exitOperator"`
	ExitShiftdate string `json:"exitShiftDate"`
	ExitShift     string `json:"exitShift"`
	Transtime     string `json:"exitTime"`
}

type ReqExitImgData struct {
	AddsInfo        ExitImgAddsData     `json:"addsInfo"`
	RstVehPlateInfo RstVehRecognizeInfo `json:"plateInfo"`
}

//设备状态数据
type AutoDevStateData struct {
	ExitNetwork   string `json:"exitNetwork"`
	ExitStationid string `json:"exitStationId"`
	ExitLandid    string `json:"exitLaneId"`

	EquipmentDev string `json:"equipmentDev"`
	PrinterDev1  string `json:"printerDev1"`
	PrinterDev2  string `json:"printerDev2"`
	PlateDev     string `json:"plateDev"`
	Coil1        string `json:"coil1"`
	Coil2        string `json:"coil2"`
	Scan1        string `json:"scan1"`
	Scan2        string `json:"scan2"`

	EpDev    string `json:"epDev"`
	FeeDisp  string `json:"feeDisp"`
	FeeAlarm string `json:"feeAlarm"`
}

//远程控制
type ReqControlData struct {
	ExitNetwork   string `json:"exitNetwork"`
	ExitStationid string `json:"exitStationId"`
	ExitLandid    string `json:"exitLaneId"`

	Cmd string `json:"cmd"`
}

type ReqCarpathData struct {
	Vehclass string `json:"vehClass"`
	Vehplate string `json:"vehPlate"`
}

type RstCarpathData struct {
	EntryNetwork   string `json:"entryNetwork"`
	EntryStationId string `json:"entryStationId"`
	EntryLaneId    string `json:"entryLaneId"`
	EntryOperator  string `json:"entryOperator"`
	EntryShift     string `json:"entryShift"`
	EntryTime      string `json:"entryTime"`
	FlagStationid  string `json:"flagStationid"`
	HasCard        string `json:"hasCard"`
}

//自由流校验//////////////////////////////////////////////////////
type ReqChkFreeflowData struct {
	Vehplate    string `json:"vehPlate"`
	Vehclass    string `json:"vehClass"`
	ExitNetwork string `json:"exitNetwork"`
	Exitstaid   string `json:"exitStationId"`
	TID         string `json:"tid"` //标签唯一码
}

type RstChkFreeflowData struct {
	Result   string `json:"result"`
	Toll     string `json:"toll"`
	Vehplate string `json:"vehPlate"`
	Vehclass string `json:"vehClass"`
}

type ReqNotifyFreeflowData struct {
	Vehplate    string `json:"vehPlate"`
	Vehclass    string `json:"vehClass"`
	ExitNetwork string `json:"exitNetwork"`
	Exitstaid   string `json:"exitStationId"`
	TID         string `json:"tid"` //标签唯一码
}

type RstNotifyFreeflowData struct {
	Result string `json:"result"`
}

package models

const MAX_SIZE1024 = 1024
const MAX_BUFFERSIZE = 1024 * 100
const MAX_LBUFFERSIZE = 1024 * 500

const SELFDEFAULTTIP = "1.请放通行卡;2.请出示付款码;3.请取发票;4.欢迎再次光临;"

/*
const FLOWDIR = "D:\\test\\flow\\"
const IMGDIR = "D:\\test\\img\\"

const FAIL_FLOWDIR = "D:\\test\\flowFail\\"
const FAIL_IMGDIR = "D:\\test\\imgFail\\"
*/

const FLOWDIR = "./flow/flow/"
const IMGDIR = "./flow/img/"
const FAIL_FLOWDIR = "./flow/flowFail/"
const FAIL_IMGDIR = "./flow/imgFail/"

const GDebugxl = true
const GDebugant = true

//模式定义
const (
	TOLLMODE_OPEN  = 1
	TOLLMODE_ENTRY = 2
	TOLLMODE_FLAG  = 3
	TOLLMODE_EXIT  = 4
)

//交易交易来源
const (
	TRANSFROM_EP        = "1" //“1”:电子车牌
	TRANSFROM_PLATE     = "2" //“2”:车脸识别
	TRANSFROM_EQUIPMENT = "3" //“3”:自助缴费机
	TRANSFROM_ETC       = "4" //“4”:ETC
)

const (
	//车检创，ep创，etc天线创,车牌创
	CREATEBY_COIL  = "1"
	CREATEBY_EP    = "2"
	CREATEBY_ETC   = "3"
	CREATEBY_PLATE = "4"
	CREATEBY_CARD  = "5"
)

const (
	TRANSMATCH_NO  = "0"
	TRANSMATCH_YES = "1"
)

const (
	TRANSSTATE_INIT = 0
	TRANSSTATE_SUC  = 1
	TRANSSTATE_FAIL = 2
)

const (
	TRANSMEMO_INIT         = "0"
	TRANSMEMO_HISTORY      = "1"
	TRANSMEMO_FREEFLOW     = "2"
	TRANSMEMO_FREEFLOWPASS = "3"
)

const (
	TRANSFLOW_INIT      = 0
	TRANSFLOW_CARPATH   = 1
	TRANSFLOW_FEE       = 2
	TRANSFLOW_PUTCARD   = 3
	TRANSFLOW_READCARD  = 4
	TRANSFLOW_SCAN      = 5
	TRANSFLOW_PAY       = 6
	TRANSFLOW_WRITECARD = 7
	TRANSFLOW_TICKET    = 8
	TRANSFLOW_OUTCARD   = 9  //退卡
	TRANSFLOW_END       = 10 //结束 成功或失败查看交易状态标记
	TRANSFLOW_FINISH    = 11 //交易成功，入库成功，交易结束
)

const (
	TRANSPROCSTATE_INIT    = 0
	TRANSPROCSTATE_PROCING = 1
	TRANSPROCSTATE_END     = 2
)

const (
	TRANSFAIL_NONE               = 0
	TRANSFAIL_NOTSUPPORTPLATEPAY = 1 //非车牌付
	TRANSFAIL_PATHLOST           = 2 //路径缺失
	TRANSFAIL_FEECALC            = 3 //计费失败
	TRANSFAIL_FEEPAY             = 4 //支付失败
	TRANSFAIL_WRITECARD          = 5 //写卡失败

	TRANSFAIL_PLATE = 6 //车牌不一致
	TRANSFAIL_ISSUE = 7 //发行商不一致
)

const (
	STATIONMODE_OTHER        = 0
	STATIONMODE_EXIT         = 1 //无自由流的出口
	STATIONMODE_FREEFLOWEXIT = 2 //自由流出口
	STATIONMODE_FREEPLUSEXIT = 3 //有自由流站的出口
)

//单位ms
const (
	GOSLEEP_DEFAULT = 100
	TIMEOUT_DEFAULT = 5000
	TIMEOUT_PUTCARD = 30000
	TIMEOUT_SCAN    = 15000
	TIMEOUT_TICKET  = 10000
)

const (
	CHECK_YES = 1
	CHECK_NO  = 0
)

const (
	FLOW_EXIT    = 0
	FLOW_EXITIMG = 1
)

//各类超时处理
const (
	SCAN_TIMEOUT_SEC = 60 //s
	NET_TIMEOUT_SEC  = 15 //
)

const (
	DEVGRPCTYPE_CLOUD   = 1
	DEVGRPCTYPE_SCAN    = 2
	DEVGRPCTYPE_IO      = 3
	DEVGRPCTYPE_PRINTER = 4
	DEVGRPCTYPE_PLATE   = 5
	DEVGRPCTYPE_FEEDISP = 6
	DEVGRPCTYPE_EP      = 7
	DEVGRPCTYPE_READER  = 8

	DEVGRPCTYPE_ETC = 9
)

//设备
const (
	DEVTYPE_IO        = 1
	DEVTYPE_PLATE     = 2
	DEVTYPE_PRINTERUP = 3
	DEVTYPE_PRINTERDN = 4
	DEVTYPE_SCANUP    = 5
	DEVTYPE_SCANDN    = 6
	DEVTYPE_EQUIPMENT = 7
	DEVTYPE_FEEDISP   = 8
	DEVTYPE_ANTEP     = 9
	DEVTYPE_READER    = 10

	DEVTYPE_ETC = 11
)

//车检信号
const (
	COIL_OPEN  = 1
	COIL_CLOSE = 0
)

//工位
const (
	WORKSTATION_DN = 0
	WORKSTATION_UP = 1
)

//设备状态
const (
	DEVSTATE_UNKNOWN = -1
	DEVSTATE_OK      = 0
	DEVSTATE_TROUBLE = 1
)

//车牌付校验结果
const (
	PLATEPAY_INIT = 0 //0未校验
	PLATEPAY_YES  = 1 //1.车牌付车辆
	PLATEPAY_NO   = 2 //2.非车牌付车辆
)

//支付方式
const (
	PAYTYPE_UNKNOWN = "0" //未知
	PAYTYPE_PLATE   = "1" //1.车脸付
	PAYTYPE_SCAN    = "2" //2.扫码付
	PAYTYPE_EP      = "3" //3.电子车牌付
	PAYTYPE_ETC     = "4" //4.ETC卡片付

)

//支付渠道
const (
	PAYMETHOD_UNKNOWN = "0" //未知
	PAYMETHOD_WX      = "1" //1.微信
	PAYMETHOD_ZFB     = "2" //2.支付宝
	PAYMETHOD_VC      = "3" //3.虚拟卡钱包
	PAYMETHOD_ETC     = "4" //4.ETC钱包
)

//有无卡
const (
	CARD_YES = "1" //1.有卡
	CARD_NO  = "2" //2.无卡
)

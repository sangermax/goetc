package models

const (
	CMD_4C = 0x4C
	CMD_C0 = 0xC0
	CMD_C1 = 0xC1
	CMD_C2 = 0xC2
	CMD_C3 = 0xC3
	CMD_C4 = 0xC4
	CMD_C5 = 0xC5
	CMD_C6 = 0xC6
	CMD_C7 = 0xC7
	CMD_B0 = 0xB0
	CMD_B1 = 0xB1
	CMD_B2 = 0xB2
	CMD_B3 = 0xB3
	CMD_B4 = 0xB4
	CMD_B5 = 0xB5
	CMD_B7 = 0xB7
)

const (
	ETCLENGTH_B0 = 24
	ETCLENGTH_B2 = 49
	ETCLENGTH_B3 = 69
	ETCLENGTH_B4 = 100
	ETCLENGTH_B5 = 36
	ETCLENGTH_B7 = 11

	ETCLENGTH_0015 = 43
	ETCLENGTH_0019 = 43
	ETCLENGTH_0008 = 128
	ETCLENGTH_0018 = 23
)

const (
	CODE_OK  = 0X00
	CODE_ERR = 0X01
)

const (
	FRAMERLT_INIT = 0X00
	FRAMERLT_OK   = 0X01
	FRAMERLT_ERR  = 0X02
)

const (
	DEFAULT_NET = "3201"
	BW_NET      = "3202"

	SH_NETWORK = "3101"
	ZJ_NETWORK = "3301"
	AH_NETWORK = "3401"
	FJ_NETWORK = "3501"
	JX_NETWORK = "3601"
	SD_NETWORK = "3701"

	BJ_NETWORK  = "1101"
	TJ_NETWORK  = "1201"
	HEB_NETWORK = "1301"
	SX_NETWORK  = "1401"
	NM_NETWORK  = "1501"

	LN_NETWORK  = "2101"
	JL_NETWORK  = "2201"
	HLJ_NETWORK = "2301"

	HEN_NETWORK  = "4101"
	HUB_NETWORK  = "4201"
	HUN_NETWORK  = "4301"
	GD_NETWORK   = "4401"
	GX_NETWORK   = "4501"
	HAIN_NETWORK = "4601"

	CQ_NETWORK = "5001"
	SC_NETWORK = "5101"
	GZ_NETWORK = "5201"
	YN_NETWORK = "5301"
	XZ_NETWORK = "5401"

	SHANXI_NETWORK = "6101"
	GS_NETWORK     = "6201"
	QH_NETWORK     = "6301"
	NX_NETWORK     = "6401"
	XJ_NETWORK     = "6501"

	ARMY_CARDNETWORK = "501"
)
const ARMY_PROVIDER = "BEFCB3B5"

type F0008Info struct {
	FlagNums int     `json:"flagnums"`
	LastFlag int     `json:"lastflag"`
	FlagRsds [62]int `json:"flagrsds"`
}

type F0015Info struct {
	CardIssue   string `json:"cardissue"`
	CardType    int    `json:"cardtype"`
	CardVersion int    `json:"cardversion"`
	CardNetWork string `json:"cardnetwork"`
	CardId      string `json:"cardid"`
	StartTm     string `json:"starttm"`
	EndTm       string `json:"endtm"`
	VehPlate    string `json:"vehplate"`
	UserType    int    `json:"usertype"`
	VehColor    int    `json:"vehcolor"`
	VehClass    int    `json:"vehclass"`
}

type F0018Info struct {
	EtcTradNo  string `json:"etctradno"`
	OverToll   int    `json:"overToll"`
	Toll       int    `json:"toll"`
	Transtype  int    `json:"transtype"`
	PsamTermID string `json:"psamtermid"`
	TransTime  string `json:"transtime"`
}

type F0019Info struct {
	InStationNetWork  string `json:"instationnetwork"`
	InStation         string `json:"instation"`
	InLane            string `json:"inlane"`
	InTime            string `json:"intime"`
	VehClass          int    `json:"vehclass"`
	FlowState         int    `json:"flowstate"`
	AuxSta            string `json:"auxstation"`
	FlagSta           string `json:"flagstation"`
	OutStationNetWork string `json:"outstationnetwork"`
	OutStation        string `json:"outstation"`
	InOperator        string `json:"inoperator"`
	InBanci           int    `json:"inbanci"`
	VehPlate          string `json:"vehplate"`
}

//4c 控制天线开关
type ETCFrame4C struct {
	Antennastatus int `json:"antennastatus"`
}

//C0H	对RSU关键参数如功率、车道模式等进行初始化/设置
type ETCFrameC0 struct {
	Seconds  int    `json:"seconds"`
	Datetime string `json:"datetime"`
	LaneMode int    `json:"lanemode"`

	WaitTime     int `json:"waittime"`
	TxPower      int `json:"txpower"`
	PLLChannelID int `json:"pllchannelid"`
}

//C1H	对PC收到RSU发来的信息的应答，表示收到信息并要求继续处理指定OBU
type ETCFrameC1 struct {
	OBUID string `json:"obuid"`
}

//C2H	对PC收到RSU发来的信息的应答，表示收到信息并要求当前不再继续处理指定OBU
type ETCFrameC2 struct {
	OBUID    string `json:"obuid"`
	StopType int    `json:"stopType"`
}

//C3H	将站信息写入指定OBU（备用）
type ETCFrameC3 struct {
	OBUID     string    `json:"obuid"`
	F0019Info F0019Info `json:"f0019"`
	Datetime  string    `json:"datetime"`
}

//C6H	对指定OBU的电子钱包扣费，并向指定的OBU写站信息
type ETCFrameC6 struct {
	OBUID        string    `json:"obuid"`
	ConsumeMoney int       `json:"consumeMoney"`
	F0019Info    F0019Info `json:"f0019"`
	Datetime     string    `json:"datetime"`
}

//C7H	读取指定OBU中CPU卡的交易记录文件
type ETCFrameC7 struct {
	OBUID     string `json:"obuid"`
	RecordNUM int    `json:"recordnum"`
}

//B0	RSU的设备状态信息，含PSAM卡号等
type ETCFrameB0 struct {
	RSCTL           int    `json:"rsctl"`
	FrameType       int    `json:"frametype"`
	RSUStatus       int    `json:"rsuStatus"`
	RSUTerminalId   string `json:"psamtermid"`
	RSUAlgId        int    `json:"rsualgid"`
	RSUManuID       int    `json:"rsumanuid"`
	RSUIndividualID string `json:"rsuindivid"`
	RSUVersion      string `json:"rsuversion"`
	Reserved        string `json:"reserved"`
}

//B1	车辆压上或离开地感的状态信息
type ETCFrameB1 struct {
	RSCTL       int `json:"rsctl"`
	FrameType   int `json:"frametype"`
	RsuIoStatus int `json:"rsuiostatus"`
	RsuIoChgSum int `json:"rsuiochgsum"`
}

//B2	主要包括OBU系统信息文件内容
type ETCFrameB2 struct {
	RSCTL                int    `json:"rsctl"`
	FrameType            int    `json:"frametype"`
	OBUID                string `json:"obuid"`
	ErrorCode            int    `json:"errorcode"`
	ContractProvider     string `json:"contractprovider"`
	ContractType         int    `json:"contracttype"`
	ContractVersion      int    `json:"contractversion"`
	ContractSerialNumber string `json:"contractserialnumber"`
	ContractSignedDate   string `json:"contractsigneddate"`
	ContractExpiredDate  string `json:"contractexpireddate"`
	CPUCardID            string `json:"cpucardid"`
	Equitmentstatus      int    `json:"equitmentstatus"`
	OBUStatus            []byte `json:"obustatus"`
}

//B3	主要包括车辆信息文件内容
type ETCFrameB3 struct {
	RSCTL       int    `json:"rsctl"`
	FrameType   int    `json:"frametype"`
	OBUID       string `json:"obuid"`
	ErrorCode   int    `json:"errorCode"`
	VehPlate    string `json:"vehplate"`
	VehColor    int    `json:"vehcolor"`
	VehClass    int    `json:"vehclass"`
	VehUserType int    `json:"vehusertype"`

	VehDimens       string `json:"vehdimens"`
	VehWheels       int    `json:"vehwheels"`
	VehAxies        int    `json:"vehaxies"`
	VehWheelBases   string `json:"vehwheelBases"`
	VehWeightLimits string `json:"vehweightLimits"`
	VehSpecificinfo string `json:"vehspecificinfo"`
	VehEngineNum    string `json:"vehengineNum"`
}

//B4	主要包括IC卡关键信息文件内容
type ETCFrameB4 struct {
	RSCTL         int       `json:"rsctl"`
	FrameType     int       `json:"frametype"`
	OBUID         string    `json:"obuid"`
	ErrorCode     int       `json:"errorcode"`
	CardRestMoney int       `json:"cardrestmoney"`
	F0015Info     F0015Info `json:"f0015"`
	F0019Info     F0019Info `json:"f0019"`
}

//B5	RSU与OBU交易完成后的结果信息
type ETCFrameB5 struct {
	RSCTL     int    `json:"rsctl"`
	FrameType int    `json:"frametype"`
	OBUID     string `json:"obuid"`
	ErrorCode int    `json:"errorCode"`

	TransTime     string `json:"transtime"`
	PsamTransNo   string `json:"psamtransno"`
	EtcTradNo     string `json:"etctradno"`
	Transtype     int    `json:"transtype"`
	CardRestMoney int    `json:"cardrestmoney"`
	Tac           string `json:"tac"`
	WrFileTime    string `json:"wrfiletime"`
}

//B7	包括IC卡交易记录信息文件内容
type ETCFrameB7 struct {
	RSCTL     int         `json:"rsctl"`
	FrameType int         `json:"frametype"`
	OBUID     string      `json:"obuid"`
	ErrorCode int         `json:"errorCode"`
	RecordNum int         `json:"recordNum"`
	F0018Rsds []F0018Info `json:"f0018rsds"`
}

type ETCCommInfo struct {
	Key        string `json:"key"`
	AntNo      string `json:"antno"`
	Psamtermid string `json:"psamtermid"`
	Msg        []byte `json:"msg"`
}

const (
	ENTRY_FLOWETC = 0x03
	EXIT_FLOWETC  = 0x04
	OPEN_FLOWETC  = 0x06
)

package models

const (
	RW_SECMD30 = 0X30
	RW_SECMD31 = 0X31

	//缴费机 -> pc
	RW_SECMD41 = 0X41
	RW_SECMD42 = 0X42
	RW_SECMD43 = 0X43
	RW_SECMD44 = 0X44
	RW_SECMD45 = 0X45
	RW_SECMD46 = 0X46
	RW_SECMD47 = 0X47
	RW_SECMD49 = 0X49
	RW_SECMD56 = 0X56
	//pc -> 缴费机
	RW_SECMD61 = 0X61
	RW_SECMD62 = 0X62
	RW_SECMD63 = 0X63
	RW_SECMD64 = 0X64
	RW_SECMD65 = 0X65
	RW_SECMD66 = 0X66
)

//自助设备工位代号
const (
	SEUP1 = 0X31 //上工位
	SEUP2 = 0X32 //上工位
	SEDN1 = 0X33 //下工位
	SEDN2 = 0X34 //下工位
)

//卡机状态
const (
	KGSTATE_OK      = 0X30 //正常
	KGSTATE_TROUBLE = 0X31 //故障
	KGSTATE_RESERVE = 0X32 //保留
	KGSTATE_OFFLINE = 0X33 //离线
)

//卡夹状态
const (
	KJSTATE_IN  = 0X30 //卡夹已装上
	KJSTATE_OFF = 0X31 //卡夹已卸下
)

//有卡状态
const (
	KSTATE_NO      = 0X30 //无卡
	KSTATE_ANT     = 0X31 //天线有卡
	KSTATE_BAYONET = 0X32 //卡口有卡
)

type KGInfo struct {
	KGState byte `json:"KGState"`
	KJState byte `json:"KJState"`
	KNums   int  `json:"KNums"`
}

type EquipmentKGInfo struct {
	UpCurWorkstation byte      `json:"UpCurWorkstation"`
	DnCurWorkstation byte      `json:"DnCurWorkstation"`
	KGRsds           [4]KGInfo `json:"KGRsds"`
}

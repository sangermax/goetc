package models

type EPInfo struct {
	DeviceSN string `json:"DeviceSN"` //设备SN (string)
	MsgID    int    `json:"MsgID"`    //消息ID(int)
	MsgName  string `json:"MsgName"`  //消息名(string)
	MsgType  int    `json:"MsgType"`  //消息类型(int)
	Version  string `json:"Version"`  //SDK版本(string)

	MessageValue interface{} `json:"MessageValue"` //消息内容
}

//读卡数据
type EPReadDataInfo struct {
	ApprovedLoad               string `json:"ApprovedLoad"`               //核定载客/总质量/总牵引质量参数
	CardNumber                 string `json:"CardNumber"`                 //卡号
	Color                      string `json:"Color"`                      //车身颜色
	CompulsoryRetirementPeriod string `json:"CompulsoryRetirementPeriod"` //强制报废期
	Emissions                  string `json:"Emissions,omitempty"`
	LicenseCode                string `json:"LicenseCode"`       //发牌机关代号
	ManufactureDate            string `json:"ManufactureDate"`   //出厂日期
	PlateNumber                string `json:"PlateNumber"`       //号牌号码序号
	PlateType                  string `json:"PlateType"`         //号牌种类
	Power                      string `json:"Power,omitempty"`   //功率
	UseCharacteristic          string `json:"UseCharacteristic"` //使用性质
	ValidityPeriod             string `json:"ValidityPeriod"`    //检验有效期
	VehicleType                string `json:"VehicleType"`       //车辆类型
	MHGLJ                      string `json:"MHGLJ,omitempty"`   //民航管理局代码//该条仅为民航车辆使用.普通社会车辆无需关注此字段
}

type EPCustomizedSelectFTCcResult struct {
	ReadDataInfo EPReadDataInfo `json:"ReadDataInfo"` //读卡数据
	Result       int            `json:"Result"`       //操作状态码(int)
}

//标签选择规则结果
type EPSelectFTCcResult struct {
	CustomizedSelectFTCcResult EPCustomizedSelectFTCcResult `json:"CustomizedSelectFTCcResult"` //个性化读选择规则结果
}

type EPTagReportUnit struct {
	AntennaID             int                `json:"AntennaID"`             //天线ID
	FirstSeenTimestampUTC string             `json:"FirstSeenTimestampUTC"` //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	LastSeenTimestampUTC  string             `json:"LastSeenTimestampUTC"`  //标签末次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	PeakRSSI              int                `json:"PeakRSSI"`              //RSSI(int)
	RfFTCcID              int                `json:"RfFTCcID"`              //射频规则ID:(int)
	SelectFTCcID          int                `json:"SelectFTCcID"`          //标签选择规则ID(int)
	SelectFTCcResult      EPSelectFTCcResult `json:"SelectFTCcResult"`      //

	FTCcIndex    int    `json:"FTCcIndex"`    //天线规则索引(int)
	TID          string `json:"TID"`          //TID:标签唯一码(string)
	TagSeenCount int    `json:"TagSeenCount"` //标签总操作次数(int)
}

type EPTagReportData struct {
	ReportRsds []EPTagReportUnit `json:"TagReportData"` //标签上报数据[0-N条]
}

//解析后提取的电子车牌信息 自定义使用
type EPResultReadInfo struct {
	AntennaID int    `json:"AntennaID"` //天线ID
	TID       string `json:"TID"`       //TID:标签唯一码(string)

	ReadDataInfo EPReadDataInfo `json:"ReadDataInfo"` //读卡数据
}

//后增加天线，usr1和usr5结构体定义///////////////////////////////////////////////////////////////////////
type EPMultiStatus struct {
	StatusCode int `json:"StatusCode"` //
}

type EPMultiHbCustomizedReadFTCcResult struct {
	OpFTCcID     int            `json:"OpFTCcID"`     //天线ID
	ReadDataInfo EPReadDataInfo `json:"ReadDataInfo"` //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	Result       int            `json:"Result"`       //操作状态码(int)
}

type EPMultiHbReadFTCcResult struct {
	OpFTCcID int    `json:"OpFTCcID"` //天线ID
	ReadData string `json:"ReadData"` //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	Result   int    `json:"Result"`   //操作状态码(int)
}

type EPMultiAccessFTCcResultUnit struct {
	HbCustomizedReadFTCcResult EPMultiHbCustomizedReadFTCcResult `json:"HbCustomizedReadFTCcResult,omitempty"` //天线ID
	HbReadFTCcResult           EPMultiHbReadFTCcResult           `json:"HbReadFTCcResult,omitempty"`           //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.
}

type EPMultiTagReportDataUnit struct {
	AccessFTCcID     int                           `json:"AccessFTCcID"`     //天线ID
	AccessFTCcResult []EPMultiAccessFTCcResultUnit `json:"AccessFTCcResult"` //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.

	AntennaID             int                `json:"AntennaID"`                  //天线ID
	FirstSeenTimestampUTC string             `json:"FirstSeenTimestampUTC"`      //标签首次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	LastSeenTimestampUTC  string             `json:"LastSeenTimestampUTC"`       //标签末次操作时间戳(string):微秒,字符串代表16进制的long long类型.
	PeakRSSI              int                `json:"PeakRSSI"`                   //RSSI(int)
	RfFTCcID              int                `json:"RfFTCcID"`                   //射频规则ID:(int)
	SelectFTCcID          int                `json:"SelectFTCcID"`               //标签选择规则ID(int)
	SelectFTCcResult      EPSelectFTCcResult `json:"SelectFTCcResult,omitempty"` //

	FTCcIndex    int    `json:"FTCcIndex"`    //天线规则索引(int)
	TID          string `json:"TID"`          //TID:标签唯一码(string)
	TagSeenCount int    `json:"TagSeenCount"` //标签总操作次数(int)
}

type EPMessageValue500 struct {
	IsLastedFrame int           `json:"IsLastedFrame"` //
	SequenceID    int           `json:"SequenceID"`    //
	Status        EPMultiStatus `json:"Status"`        //

	ReportRsds []EPMultiTagReportDataUnit `json:"TagReportData"` //标签上报数据[0-N条]
}

package models

// ColorRed 费显红色
const ColorRed byte = 0

// ColorGreen 绿色
const ColorGreen byte = 1

// ColorYellow 黄色
const ColorYellow byte = 2

// Datalen 数据长度,永远是36个字节
const Datalen byte = 36

// Displaylightday 默认亮度
const Displaylightday byte = 255

// Displaylightnight 晚上亮度
const Displaylightnight byte = 200

// Alarmoff 报警关
const Alarmoff byte = 0

// Alarmon 报警开
const Alarmon byte = 1

// Alarmtimelong 报警时间长30秒
const Alarmtimelong int = 30

// Alarmtimeshort 报警时间短 10秒
const Alarmtimeshort int = 10

// heartbeattimeout 30秒没心跳，认为超时
const Heartbeattimeout int = 30

//heartfrequence 10秒一次心跳
const Heartfrequence int = 10

const AlarmStateoff int = 0

// Alarmon 报警开
const AlarStatemon int = 1

type FeedispShowData struct {
	Line1  string `json:"Line1"`
	Line2  string `json:"Line2"`
	Line3  string `json:"Line3"`
	Color1 string `json:"color1"`
	Color2 string `json:"color2"`
	Color3 string `json:"color3"`
}

type FeedispAlarmData struct {
	AlarmValue string `json:"alarmValue"`
	AlarmTm    string `json:"alarmTm"`
}

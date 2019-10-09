package models

const MAX_PRINTER_LEN = 48

const (
	PrinterOK                = 0 //正常
	PrinterNoPaper           = 1 //缺纸
	PrinterMachanicalTrouble = 2 //机械故障
	PrinterKeyDn             = 3 //按键按下
	PrinterHighTmp           = 4 //温度过高
	PrinterErrKnife          = 5 //切刀发生错误
	PrinterErr               = 6
)

//打印机字体
const (
	FONTSIZE_1 = 1 //选择7*9字体
	FONTSIZE_2 = 2 //选择5*9字体
	FONTSIZE_3 = 3 //选择5*9字体
)

const (
	SizeNormal      = 0
	SizeTimes       = 1
	SizeDoubleTimes = 2
)

type PrinterContent struct {
	Aligyntype string `json:"aligyntype"` //0：左对齐；1：居中；2：右对齐
	Fontsize   string `json:"fontsize"`   //字体大小,0：正常；1:倍高倍宽；2:4倍
	Content    string `json:"content"`
}

type ReqPrinterTicketData struct {
	WorkStation string           `json:"workStation"`
	LineNums    string           `json:"nums"`
	PrintRsds   []PrinterContent `json:"printerdata"`
}

type PrinterStateData struct {
	UpState string `json:"upState"`
	DnState string `json:"dnState"`
}

package models

const (
	SECMD30 = 0X30
	SECMD31 = 0X31

	//缴费机 -> pc
	SECMD41 = 0X41
	SECMD42 = 0X42
	SECMD43 = 0X43
	SECMD44 = 0X44
	SECMD45 = 0X45
	SECMD46 = 0X46
	SECMD47 = 0X47
	SECMD48 = 0X48
	SECMD49 = 0X49
	SECMD4A = 0X4A
	SECMD4B = 0X4B
	SECMD5A = 0X5A
	//pc -> 缴费机
	SECMD61 = 0X61
	SECMD62 = 0X62
	SECMD63 = 0X63
	SECMD64 = 0X64
	SECMD65 = 0X65
	SECMD66 = 0X66
	SECMD67 = 0X67
	SECMD68 = 0X68
	SECMD69 = 0X69
	SECMD6A = 0X6A
	SECMD6B = 0X6B
	SECMD6C = 0X6C
	SECMD6D = 0X6D
	SECMD70 = 0X70
	SECMD72 = 0X72
	SECMD73 = 0X73
	SECMD74 = 0X74
	SECMD75 = 0X75
	SECMD7A = 0X7A
	SECMD7B = 0X7B
	SECMD7C = 0X7C
)

//自助设备工位代号
const (
	SEUP = 0X31 //上工位
	SEDN = 0X32 //下工位
)

const (
	SHOW1  = 0x01 //收卡盒伸出
	SHOW2  = 0x02 //收卡盒伸出后静态显示
	SHOW3  = 0x03 //ETC卡文字显示
	SHOW4  = 0x04 //收卡盒刷ETC卡动画
	SHOW5  = 0x05 //通行卡文字显示
	SHOW6  = 0x06 //收卡盒放入通行卡
	SHOW7  = 0x07 //收卡盒收回，扫码臂伸出
	SHOW8  = 0x08 //扫码动画
	SHOW9  = 0x09 //打印发票动画
	SHOW10 = 0x0A //等待取票静态显示
	SHOW11 = 0x0B //取票动画
	SHOW12 = 0x0C //缴费机初始状态静态显示
	SHOW13 = 0x0D //按键获取发票动画
	SHOW14 = 0x0E //降级天线刷卡
	SHOW15 = 0x0F //降级卡道收卡
)

//缴费机报文结构
type FrameSelfEquipment struct {
	Stx   byte
	Rsctl byte
	Ctl   byte
	Data  []byte
	Etx   byte
}

type FrameSE74 struct {
	Aligyntype int
	SpaceValue int
	Xpos       int
	Ypos       int
	WidthShow  int
	HeightShow int
	Fontsize   int
	RedColor   int
	GreenColor int
	BlueColor  int
	Content    []byte
}

//对齐方式
const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

//行间距
const (
	SPACE0  = 0
	SPACE1  = 1
	SPACE2  = 2
	SPACE3  = 3
	SPACE4  = 4
	SPACE5  = 5
	SPACE6  = 6
	SPACE7  = 7
	SPACE8  = 8
	SPACE9  = 9
	SPACE10 = 10
	SPACE11 = 11
	SPACE12 = 12
	SPACE13 = 13
	SPACE14 = 14
	SPACE15 = 15
)

//字体大小
const (
	FONT8  = 0
	FONT12 = 1
	FONT16 = 2
	FONT24 = 3
	FONT32 = 4
	FONT40 = 5
	FONT48 = 6
	FONT56 = 7
	FONT64 = 7
)

const (
	COMPONENT_READER   = "10"
	DNCOMPONENT_READER = "11" //下工位 读写臂
	UPCOMPONENT_READER = "12" //上工位 读写臂

	COMPONENT_SCAN   = "20"
	DNCOMPONENT_SCAN = "21" //下工位 扫码臂
	UPCOMPONENT_SCAN = "22" //上工位 扫码臂

	COMPONENT_TICKET   = "30"
	DNCOMPONENT_TICKET = "31" //下工位 票夹臂
	UPCOMPONENT_TICKET = "32" //上工位 票夹臂

)

//状态 移动还是原状
const (
	CSTATE_KEEP = 0
	CSTATE_MOVE = 1
)

type EquipSubComponent struct {
	DstState    int   //目的状态
	CurState    int   //当前状态
	Workstation byte  //工位
	UpdateTm    int64 //最近一次控制命令发送时间
}

type EquipFrameinfo struct {
	Workstation byte   //工位
	LastTm      int64  //最近一次控制命令发送时间
	Trynums     int    //尝试次数
	Rsctl       byte   //命令字
	Frame       []byte //帧内容
}

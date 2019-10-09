package device

import (
	"errors"
	"sync"

	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//智能缴费机
type DevSelfEquipment struct {
	SerialBase

	State       int
	Frameno     byte
	FramenoLock *sync.Mutex

	ComponentMap *util.BeeMap
}

func (p *DevSelfEquipment) InitDev() {
	p.State = models.DEVSTATE_UNKNOWN
	p.FramenoLock = new(sync.Mutex)
	p.Frameno = 0
	p.ComponentMap = util.NewBeeMap()
	p.InitComponent()

	p.InitSerial(models.DEVTYPE_EQUIPMENT, config.ConfigData["selfEquipmentCom"].(string), 115200, p.Recvproc)

	if p.IsConn() {
		p.State = models.DEVSTATE_OK
		p.Package61(0x30)
		//设备自检
		p.Package72()
	} else {
		p.State = models.DEVSTATE_TROUBLE
	}

	go p.goComponentCheck()
}

func (p *DevSelfEquipment) GetRsctl() byte {
	var b byte
	b = 0x30
	p.FramenoLock.Lock()

	if p.Frameno >= 9 {
		p.Frameno = 0
	} else {
		p.Frameno += 1
	}
	b += p.Frameno

	p.FramenoLock.Unlock()

	return b
}

func (p *DevSelfEquipment) GetCmdDes(cmd byte) string {
	switch cmd {
	case models.SECMD30:
		return "正应答"
	case models.SECMD31:
		return "负应答"

	case models.SECMD41:
		return "(缴费机->PC) 设备加电上报"
	case models.SECMD42:
		return "状态信息上报"
	case models.SECMD43:
		return "收卡结果上报"
	case models.SECMD44:
		return "有卡插入"
	case models.SECMD45:
		return "卡取走"
	case models.SECMD46:
		return "上报卡夹号信息"
	case models.SECMD47:
		return "退卡结果上报"
	case models.SECMD48:
		return "扫码机伸缩结果上报"
	case models.SECMD49:
		return "收卡盒伸出结果上报"
	case models.SECMD4A:
		return "夹票手伸出结果上报"
	case models.SECMD4B:
		return "按键上报"
	case models.SECMD5A:
		return "驱动单元到上位机转发"
	//pc -> 缴费机
	case models.SECMD61:
		return "初始化信息"
	case models.SECMD62:
		return "控制收卡"
	case models.SECMD63:
		return "控制退卡"
	case models.SECMD64:
		return "读卡成功"
	case models.SECMD65:
		return "查询卡机状态"
	case models.SECMD66:
		return "发送扫码指令"
	case models.SECMD67:
		return "设置卡夹卡数"
	case models.SECMD68:
		return "车道线圈信号"
	case models.SECMD69:
		return "票据打印并出票完成"
	case models.SECMD6A:
		return "缴费完成"
	case models.SECMD6B:
		return "栏杆机状态"
	case models.SECMD6C:
		return "设置卡夹卡箱编号"
	case models.SECMD6D:
		return "发送流程结束指令"
	case models.SECMD70:
		return "控制卡机报警"
	case models.SECMD72:
		return "控制卡机复位"
	case models.SECMD73:
		return "控制语音播报"
	case models.SECMD74:
		return "控制屏显示"
	case models.SECMD75:
		return "控制屏按预制条目显示"
	case models.SECMD7A:
		return "上位机到驱动单元数据转发"
	case models.SECMD7B:
		return "缴费机重启"
	case models.SECMD7C:
		return "总控制板重启"

	}

	return "未知:" + util.ConvertI2S(int(cmd))
}

func (p *DevSelfEquipment) GetShowDes(value byte) string {
	s := util.ConvertI2S(int(value))
	switch value {
	case models.SHOW1:
		return s + "-" + "收卡盒伸出"

	case models.SHOW2:
		return s + "-" + "收卡盒伸出后静态显示"

	case models.SHOW3:
		return s + "-" + "ETC卡文字显示"

	case models.SHOW4:
		return s + "-" + "收卡盒刷ETC卡动画"

	case models.SHOW5:
		return s + "-" + "通行卡文字显示"

	case models.SHOW6:
		return s + "-" + "收卡盒放入通行卡"

	case models.SHOW7:
		return s + "-" + "收卡盒收回，扫码臂伸出"

	case models.SHOW8:
		return s + "-" + "扫码动画"

	case models.SHOW9:
		return s + "-" + "打印发票动画"

	case models.SHOW10:
		return s + "-" + "等待取票静态显示"

	case models.SHOW11:
		return s + "-" + "取票动画"

	case models.SHOW12:
		return s + "-" + "缴费机初始状态静态显示"

	case models.SHOW13:
		return s + "-" + "按键获取发票动画"

	case models.SHOW14:
		return s + "-" + "降级天线刷卡"

	case models.SHOW15:
		return s + "-" + "降级卡道收卡"
	}

	return s + "-" + "未知"
}

//解析前转义
func (p *DevSelfEquipment) BeforeParseFrame(inbuf []byte) []byte {
	outbuf := make([]byte, 1024)
	len1 := len(inbuf)
	i := 0
	j := 0

	for ; i < len1; i = i + 1 {
		if i+1 < len1 {
			if inbuf[i] == 0x3D && inbuf[i+1] == 0x5C {
				outbuf[j] = 0x3C
				j += 1
				i += 1
			} else if inbuf[i] == 0x3D && inbuf[i+1] == 0x5D {
				outbuf[j] = 0x3D
				j += 1
				i += 1
			} else if inbuf[i] == 0x3D && inbuf[i+1] == 0x5E {
				outbuf[j] = 0x3E
				j += 1
				i += 1
			} else {
				outbuf[j] = inbuf[i]
				j += 1
			}
		} else {
			outbuf[j] = inbuf[i]
			j += 1
		}
	}

	return outbuf[0:j]
}

//打包前转义
func (p *DevSelfEquipment) BeforePackageFrame(inbuf []byte) []byte {
	outbuf := make([]byte, 1024)
	len1 := len(inbuf)
	i := 0
	j := 0

	for ; i < len1; i = i + 1 {
		if i+1 < len1 {
			if inbuf[i] == 0x3C {
				outbuf[j] = 0x3D
				j += 1
				outbuf[j] = 0x5C
				j += 1
				i += 1
			} else if inbuf[i] == 0x3D {
				outbuf[j] = 0x3D
				j += 1
				outbuf[j] = 0x5D
				j += 1
				i += 1
			} else if inbuf[i] == 0x3E {
				outbuf[j] = 0x3D
				j += 1
				outbuf[j] = 0x5E
				j += 1
				i += 1
			} else {
				outbuf[j] = inbuf[i]
				j += 1
			}
		} else {
			outbuf[j] = inbuf[i]
			j += 1
		}
	}

	return outbuf[0:j]
}

//解析
func (p *DevSelfEquipment) ParseFrame(inbuf []byte) (models.FrameSelfEquipment, int, error) {
	length := len(inbuf)
	baselen := 4

	//util.FileLogs.Info("ParseFrame:%d,%d,%s.\r\n",length,baselen,util.ConvertByte2Hexstring(inbuf,true))

	if length < baselen {
		return models.FrameSelfEquipment{}, 0, errors.New("还未收完")
	}

	i := 0
	for {
		startpos := 0
		endpos := 0
		for ; i < length; i = i + 1 {
			if inbuf[i] == 0x3C {
				startpos = i
				break
			}
		}

		if i >= length {
			return models.FrameSelfEquipment{}, length, errors.New("未找到帧头")
		}

		for i = startpos + 1; i < length; i = i + 1 {
			if inbuf[i] == 0x3E {
				endpos = i
				break
			}
		}

		if i >= length {
			return models.FrameSelfEquipment{}, startpos, errors.New("未找到帧尾")
		}

		var info models.FrameSelfEquipment
		pos := startpos

		info.Stx = inbuf[pos]
		pos += 1
		info.Rsctl = inbuf[pos]
		pos += 1
		info.Ctl = inbuf[pos]
		pos += 1

		content := p.BeforeParseFrame(inbuf[pos:endpos])
		contentlen := len(content)
		info.Data = make([]byte, contentlen)
		copy(info.Data, content)

		info.Etx = inbuf[endpos]
		pos += 1

		return info, endpos + 1, nil
	}
}

//组帧
func (p *DevSelfEquipment) PackageFrame(rsctl byte, cmd byte, inbuf []byte) ([]byte, error) {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = 0x3C
	pos += 1

	buf[pos] = rsctl
	pos += 1

	buf[pos] = cmd
	pos += 1

	if inbuf != nil {
		content := p.BeforePackageFrame(inbuf)
		contentlen := len(content)
		copy(buf[pos:], content)
		pos += contentlen
	}

	buf[pos] = 0x3E
	pos += 1

	return buf[0:pos], nil
}

//负应答
func (p *DevSelfEquipment) Package31(rsctl byte) bool {
	return p.FuncSend(rsctl, models.SECMD31, nil)
}

//正应答
func (p *DevSelfEquipment) Package30(rsctl byte) bool {
	return p.FuncSend(rsctl, models.SECMD30, nil)
}

//设备加电上报
func (p *DevSelfEquipment) Parse41(info models.FrameSelfEquipment) bool {
	return p.Package61(info.Rsctl)
}

//状态信息上报
func (p *DevSelfEquipment) Parse42(info models.FrameSelfEquipment) {
}

//收卡结果上报
func (p *DevSelfEquipment) Parse43(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//有卡插入
func (p *DevSelfEquipment) Parse44(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//卡取走
func (p *DevSelfEquipment) Parse45(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//上报卡夹号信息
func (p *DevSelfEquipment) Parse46(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//退卡结果上报
func (p *DevSelfEquipment) Parse47(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//扫码机伸缩结果上报
func (p *DevSelfEquipment) Parse48(info models.FrameSelfEquipment) bool {
	datalen := len(info.Data)
	if datalen < 3 {
		//回复否认
		return p.Package31(info.Rsctl)
	}

	pos := 0
	b1 := info.Data[pos]
	pos += 1
	//b2 := info.Data[pos]
	pos += 1
	b3 := int(info.Data[pos])
	pos += 1

	switch b3 {
	case 0x30: //回缩
		p.UpdateComponent(models.COMPONENT_SCAN, b1, models.CSTATE_KEEP, false)
	case 0x31: //前伸
		p.UpdateComponent(models.COMPONENT_SCAN, b1, models.CSTATE_MOVE, false)
	}

	//直接回复确认
	return p.Package30(info.Rsctl)
}

//收卡盒伸出结果上报
func (p *DevSelfEquipment) Parse49(info models.FrameSelfEquipment) bool {
	datalen := len(info.Data)
	if datalen < 3 {
		//回复否认
		return p.Package31(info.Rsctl)
	}

	pos := 0
	b1 := info.Data[pos]
	pos += 1
	//b2 := info.Data[pos]
	pos += 1
	b3 := int(info.Data[pos])
	pos += 1

	switch b3 {
	case 0x30: //回缩
		p.UpdateComponent(models.COMPONENT_READER, b1, models.CSTATE_KEEP, false)
	case 0x31: //前伸
		p.UpdateComponent(models.COMPONENT_READER, b1, models.CSTATE_MOVE, false)
	}

	//直接回复确认
	return p.Package30(info.Rsctl)
}

//夹票手伸出结果上报
func (p *DevSelfEquipment) Parse4A(info models.FrameSelfEquipment) bool {
	datalen := len(info.Data)
	if datalen < 3 {
		//回复否认
		return p.Package31(info.Rsctl)
	}

	pos := 0
	b1 := info.Data[pos]
	pos += 1
	//b2 := info.Data[pos]
	pos += 1
	b3 := int(info.Data[pos])
	pos += 1

	switch b3 {
	case 0x30: //回缩
		p.UpdateComponent(models.COMPONENT_TICKET, b1, models.CSTATE_KEEP, false)
	case 0x31: //前伸
		p.UpdateComponent(models.COMPONENT_TICKET, b1, models.CSTATE_MOVE, false)
	}

	//直接回复确认
	return p.Package30(info.Rsctl)
}

//按键上报
func (p *DevSelfEquipment) Parse4B(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//驱动单元到上位机转发
func (p *DevSelfEquipment) Parse5A(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//////////////////////////////////////////////////////////////////////////////////////
//初始化信息帧
func (p *DevSelfEquipment) Package61(rsctl byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	sNums := "500"
	copy(buf[pos:], []byte(sNums))
	pos += 3

	sTime := util.GetNow(true)
	copy(buf[pos:], []byte(sTime))
	pos += len(sTime)

	return p.FuncSend(rsctl, models.SECMD61, buf[0:pos])
}

//控制收卡
func (p *DevSelfEquipment) Package62(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD62, buf[0:pos])

}

//控制退卡
func (p *DevSelfEquipment) Package63(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD63, buf[0:pos])
}

//读卡成功
func (p *DevSelfEquipment) Package64(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD64, buf[0:pos])
}

//查询卡机状态
func (p *DevSelfEquipment) Package65() bool {
	return p.FuncSend(p.GetRsctl(), models.SECMD65, nil)
}

//发送扫码指令
func (p *DevSelfEquipment) Package66(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD66, buf[0:pos])
}

//设置卡夹卡数
func (p *DevSelfEquipment) Package67(bdata byte, nums int) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	byNums := util.Int2Bytes_B(nums)
	copy(buf[pos:], byNums[1:3])
	pos += 3

	return p.FuncSend(p.GetRsctl(), models.SECMD67, buf[0:pos])
}

//车道线圈信号 控制卡盒伸出
func (p *DevSelfEquipment) Package68(bdata, bcoil1, bcoil2 byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1
	buf[pos] = util.Convertb2sb(bcoil1)
	pos += 1
	buf[pos] = util.Convertb2sb(bcoil2)
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD68, buf[0:pos])
}

//票据打印并出票完成
func (p *DevSelfEquipment) Package69(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD69, buf[0:pos])
}

//缴费完成
func (p *DevSelfEquipment) Package6A(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD6A, buf[0:pos])
}

//栏杆机状态
func (p *DevSelfEquipment) Package6B(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD6B, buf[0:pos])
}

//设置卡夹卡箱信号编号
func (p *DevSelfEquipment) Package6C(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	bs := util.ConvertString2Byte("000000")
	copy(buf[pos:], bs)
	pos += len(bs)

	return p.FuncSend(p.GetRsctl(), models.SECMD6C, buf[0:pos])
}

//发送流程结束
func (p *DevSelfEquipment) Package6D(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD6D, buf[0:pos])
}

//控制卡机报警
func (p *DevSelfEquipment) Package70(bdata, bvalue byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1
	buf[pos] = bvalue
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD70, buf[0:pos])
}

//控制卡机复位
func (p *DevSelfEquipment) Package72() bool {
	return p.FuncSend(p.GetRsctl(), models.SECMD72, nil)
}

//控制语音播报
func (p *DevSelfEquipment) Package73(content string) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	bs := util.ConvertString2Byte(content)
	gbk_bs, _ := util.Utf8ToGbk(bs)
	copy(buf[pos:], gbk_bs)
	pos += len(gbk_bs)

	util.FileLogs.Info("语音播报内容:%s", content)
	return p.FuncSend(p.GetRsctl(), models.SECMD73, buf[0:pos])
}

//控制屏显示
func (p *DevSelfEquipment) Package74(info models.FrameSE74) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = byte((info.SpaceValue)<<4 + (info.SpaceValue))
	pos += 1

	b1 := util.Short2Bytes_B(info.Xpos)
	copy(buf[pos:], b1)
	pos += 2

	b2 := util.Short2Bytes_B(info.Ypos)
	copy(buf[pos:], b2)
	pos += 2

	b3 := util.Short2Bytes_B(info.WidthShow)
	copy(buf[pos:], b3)
	pos += 2

	b4 := util.Short2Bytes_B(info.HeightShow)
	copy(buf[pos:], b4)
	pos += 2

	buf[pos] = byte(info.Fontsize)
	pos += 1

	buf[pos] = byte(info.RedColor)
	pos += 1
	buf[pos] = byte(info.GreenColor)
	pos += 1
	buf[pos] = byte(info.BlueColor)
	pos += 1
	copy(buf[pos:], info.Content)
	pos += len(info.Content)

	buf[pos] = 0x00 //补空格清屏后续数据
	pos += 1

	return p.FuncSend(p.GetRsctl(), models.SECMD74, buf[0:pos])
}

//控制屏按预制条目显示
func (p *DevSelfEquipment) Package75(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	util.FileLogs.Info("屏显内容:%s", p.GetShowDes(bdata))
	return p.FuncSend(p.GetRsctl(), models.SECMD75, buf[0:pos])
}

//上位机到驱动单元
func (p *DevSelfEquipment) Package7A(bdata byte, inbuf []byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = bdata
	pos += 1

	copy(buf[pos:], inbuf)
	pos += len(inbuf)

	return p.FuncSend(p.GetRsctl(), models.SECMD75, buf[0:pos])
}

//缴费机重启
func (p *DevSelfEquipment) Package7B() bool {
	return p.FuncSend(p.GetRsctl(), models.SECMD7B, nil)
}

//总控制板重启
func (p *DevSelfEquipment) Package7C() bool {
	return p.FuncSend(p.GetRsctl(), models.SECMD7C, nil)
}

func (p *DevSelfEquipment) FuncSend(rsctl byte, cmd byte, inbuf []byte) bool {
	outbuf, _ := p.PackageFrame(rsctl, cmd, inbuf)

	util.FileLogs.Info("DevSelfEquipment:(%d-%s)-%d,%s.\r\n", cmd, p.GetCmdDes(cmd), len(outbuf), util.ConvertByte2Hexstring(outbuf, true))

	return p.SendProc(outbuf, false)
}

func (p *DevSelfEquipment) Recvproc(inbuf []byte) (int, error) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, errors.New("长度不足")
	}

	info, offset, err := p.ParseFrame(inbuf[0:inlen])
	if err != nil {
		//util.FileLogs.Info("DevSelfEquipment run parse:%d,%s.\r\n", offset, err.Error())
		return offset, err
	}

	//util.FileLogs.Info("DevSelfEquipment Recvproc recv:%02X-%s.\r\n", info.Ctl, p.GetCmdDes(info.Ctl))
	//util.FileLogs.Info("DevSelfEquipment recv:%d,%s.\r\n", inlen, util.ConvertByte2Hexstring(inbuf[0:inlen], true))

	switch info.Ctl {
	case models.SECMD30:
		break
	case models.SECMD41:
		p.Parse41(info)
		break
	case models.SECMD42:
		p.Parse42(info)
		break
	case models.SECMD43:
		p.Parse43(info)
		break
	case models.SECMD44:
		p.Parse44(info)
		break
	case models.SECMD45:
		p.Parse45(info)
		break
	case models.SECMD46:
		p.Parse46(info)
		break
	case models.SECMD47:
		p.Parse47(info)
		break
	case models.SECMD48:
		p.Parse48(info)
		break
	case models.SECMD49:
		p.Parse49(info)
		break
	case models.SECMD4A:
		p.Parse4A(info)
		break
	case models.SECMD4B:
		p.Parse4B(info)
		break
	case models.SECMD5A:
		p.Parse5A(info)
		break

	}

	return offset, nil
}

func (p *DevSelfEquipment) Recovery() {
	p.UpdateComponent(models.COMPONENT_READER, models.SEUP, models.CSTATE_KEEP, true)
	p.UpdateComponent(models.COMPONENT_SCAN, models.SEUP, models.CSTATE_KEEP, true)
	p.UpdateComponent(models.COMPONENT_TICKET, models.SEUP, models.CSTATE_KEEP, true)

	p.UpdateComponent(models.COMPONENT_READER, models.SEDN, models.CSTATE_KEEP, true)
	p.UpdateComponent(models.COMPONENT_SCAN, models.SEDN, models.CSTATE_KEEP, true)
	p.UpdateComponent(models.COMPONENT_TICKET, models.SEDN, models.CSTATE_KEEP, true)
}

//dstflag true:更改dst false:更改cur
func (p *DevSelfEquipment) UpdateComponent(component string, workstation byte, dstState int, dstflag bool) {
	sc := ""
	switch component {
	case models.COMPONENT_READER:
		if workstation == models.SEUP {
			sc = models.UPCOMPONENT_READER
		} else if workstation == models.SEDN {
			sc = models.DNCOMPONENT_READER
		}

	case models.COMPONENT_SCAN:
		if workstation == models.SEUP {
			sc = models.UPCOMPONENT_SCAN
		} else if workstation == models.SEDN {
			sc = models.DNCOMPONENT_SCAN
		}

	case models.COMPONENT_TICKET:
		if workstation == models.SEUP {
			sc = models.UPCOMPONENT_TICKET
		} else if workstation == models.SEDN {
			sc = models.DNCOMPONENT_TICKET
		}
	}

	e := p.ComponentMap.Get(sc)
	if e == nil {
		return
	}

	ev := e.(models.EquipSubComponent)
	if dstflag {
		ev.DstState = dstState
	} else {
		ev.CurState = dstState
	}

	p.ComponentMap.ReSet(sc, ev)
}

func (p *DevSelfEquipment) InitComponent() {
	p.InitComponentValue(models.DNCOMPONENT_READER, models.SEDN) //下工位 读写臂
	p.InitComponentValue(models.DNCOMPONENT_SCAN, models.SEDN)   //下工位 扫码臂
	p.InitComponentValue(models.DNCOMPONENT_TICKET, models.SEDN) //下工位 票夹臂
	p.InitComponentValue(models.UPCOMPONENT_READER, models.SEUP) //上工位 读写臂
	p.InitComponentValue(models.UPCOMPONENT_SCAN, models.SEUP)   //上工位 扫码臂
	p.InitComponentValue(models.UPCOMPONENT_TICKET, models.SEUP) //上工位 票夹臂
}

func (p *DevSelfEquipment) InitComponentValue(component string, bdata byte) {
	var unit models.EquipSubComponent
	unit.DstState = models.CSTATE_KEEP
	unit.CurState = models.CSTATE_KEEP
	unit.Workstation = bdata
	unit.UpdateTm = 0
	p.ComponentMap.ReSet(component, unit)
}

func (p *DevSelfEquipment) goComponentCheck() {
	for {
		p.ComponentMap.Lock.Lock()
		for k, v := range p.ComponentMap.BM {
			if v == nil {
				continue
			}

			ev := v.(models.EquipSubComponent)
			if ev.CurState == ev.DstState {
				continue
			}

			switch k {
			case models.UPCOMPONENT_READER, models.DNCOMPONENT_READER:
				if ev.DstState == models.CSTATE_MOVE {
					util.FileLogs.Info("读写器支臂未伸出，重新发送指令")
					p.FuncCReaderMove(ev.Workstation, false)
				}
				if ev.DstState == models.CSTATE_KEEP {
					util.FileLogs.Info("读写器支臂未收回，重新发送指令")
					p.FuncCReaderKeep(ev.Workstation, false)

					//取消交易
					p.Package6D(ev.Workstation)
				}

			case models.UPCOMPONENT_SCAN, models.DNCOMPONENT_SCAN:
				if ev.DstState == models.CSTATE_MOVE {
					util.FileLogs.Info("扫码器支臂未伸出，重新发送指令")
					p.FuncCScanMove(ev.Workstation, false)
				}
				if ev.DstState == models.CSTATE_KEEP {
					util.FileLogs.Info("扫码器支臂未收回，重新发送指令")
					p.FuncCScanKeep(ev.Workstation, false)
				}

			case models.DNCOMPONENT_TICKET:
			case models.UPCOMPONENT_TICKET:
				if ev.DstState == models.CSTATE_MOVE {
					util.FileLogs.Info("票夹支臂未伸出，重新发送指令")
					p.FuncCTicketMove(ev.Workstation, false)
				}

				if ev.DstState == models.CSTATE_KEEP {
					util.FileLogs.Info("票夹支臂未收回，重新发送指令")
					p.FuncCTicketKeep(ev.Workstation, false)
				}
			}
		}

		p.ComponentMap.Lock.Unlock()

		util.MySleep_s(3)
	}
}

//伸读写设备 联动
func (p *DevSelfEquipment) FuncCReaderMoveLinked(bdata byte) {
	//语音播报
	p.Package73("请放通行卡或刷ETC卡")

	//卡臂伸出
	p.FuncCReaderMove(bdata, true)

	//屏显 提示用户放卡
	//p.Package75(models.SHOW2)
	p.Package74(ShowTip16("请放通行卡;或刷ETC卡", models.ColorGreen))
}

func (p *DevSelfEquipment) FuncCReaderMove(bdata byte, bUpdate bool) {
	p.Package68(bdata, 0X01, 0X01)

	//更改状态
	if bUpdate {
		p.UpdateComponent(models.COMPONENT_READER, bdata, models.CSTATE_MOVE, true)
	}
}

//回收读写设备
func (p *DevSelfEquipment) FuncCReaderKeep(bdata byte, bUpdate bool) {
	p.Package62(bdata)

	//更改状态
	if bUpdate {
		p.UpdateComponent(models.COMPONENT_READER, bdata, models.CSTATE_KEEP, true)
	}
}

//伸扫码设备
func (p *DevSelfEquipment) FuncCScanMoveLinked(bdata byte) {
	p.Package73("请出示付款码")
	p.FuncCScanMove(bdata, true)
	//p.Package75(models.SHOW8)
}

func (p *DevSelfEquipment) FuncCScanMove(bdata byte, bUpdate bool) {
	p.Package66(bdata)

	//更改状态
	if bUpdate {
		p.UpdateComponent(models.COMPONENT_SCAN, bdata, models.CSTATE_MOVE, true)
	}
}

//回收扫码设备
func (p *DevSelfEquipment) FuncCScanKeep(bdata byte, bUpdate bool) {
	p.Package6A(bdata)

	//更改状态
	if bUpdate {
		p.UpdateComponent(models.COMPONENT_SCAN, bdata, models.CSTATE_KEEP, true)
	}
}

//伸票夹设备
func (p *DevSelfEquipment) FuncCTicketMoveLinked(bdata byte) {
	p.Package73("请取发票")
	p.FuncCTicketMove(bdata, true)

	p.Package74(ShowTip16("请取发票", models.ColorGreen))
	//p.Package75(models.SHOW11)
}

func (p *DevSelfEquipment) FuncCTicketMove(bdata byte, bUpdate bool) {
	p.Package69(bdata)
	//更改状态
	if bUpdate {
		p.UpdateComponent(models.COMPONENT_TICKET, bdata, models.CSTATE_MOVE, true)
	}
}

//回收票夹设备 //该设备为司机取票后卡机自动控制票夹手缩回
func (p *DevSelfEquipment) FuncCTicketKeep(bdata byte, bUpdate bool) {

}

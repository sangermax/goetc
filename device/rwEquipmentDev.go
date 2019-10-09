package device

import (
	"errors"
	"sync"

	"FTC/config"
	"FTC/models"
	"FTC/util"
)

//朗为智能缴费机
type DevRWEquipment struct {
	SerialBase

	State       int
	Frameno     byte
	FramenoLock *sync.Mutex

	KGStateInfo models.EquipmentKGInfo
	FrameMap    *util.BeeMap
}

func (p *DevRWEquipment) InitDev() {
	p.FramenoLock = new(sync.Mutex)
	p.Frameno = 0
	p.InitSerial(models.DEVTYPE_EQUIPMENT, config.ConfigData["selfEquipmentCom"].(string), 9600, p.Recvproc)

	if p.IsConn() {
		p.Package61(0x30)
		go p.goFrameCheck()
	}
}

func (p *DevRWEquipment) GetRsctl() byte {
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

func (p *DevRWEquipment) GetCmdDes(cmd byte) string {
	switch cmd {
	case models.RW_SECMD30:
		return "正应答"
	case models.RW_SECMD31:
		return "负应答"

	case models.RW_SECMD41:
		return "(缴费机->PC) 设备上电"
	case models.RW_SECMD42:
		return "(缴费机->PC) 状态信息上报"
	case models.RW_SECMD43:
		return "(缴费机->PC) 退卡信息"
	case models.RW_SECMD44:
		return "(缴费机->PC) 按键信息"
	case models.RW_SECMD45:
		return "(缴费机->PC) 卡取走信息"
	case models.RW_SECMD46:
		return "(缴费机->PC) 卡夹信息"
	case models.RW_SECMD47:
		return "(缴费机->PC) 回收卡完成"
	case models.RW_SECMD49:
		return "(缴费机->PC) 收卡完成"
	case models.RW_SECMD56:
		return "(缴费机->PC) 核心板软件版本"
	//pc -> 缴费机
	case models.RW_SECMD61:
		return "(PC->缴费机) 初始化信息"
	case models.RW_SECMD62:
		return "(PC->缴费机) 回收卡信息"
	case models.RW_SECMD63:
		return "(PC->缴费机) 退卡信息"
	case models.RW_SECMD64:
		return "(PC->缴费机) 收卡信息"
	case models.RW_SECMD65:
		return "(PC->缴费机) 查询卡机状态"
	case models.RW_SECMD66:
		return "(PC->缴费机) 查询卡夹"
	}

	return "未知:" + util.ConvertI2S(int(cmd))
}

//解析前转义
func (p *DevRWEquipment) BeforeParseFrame(inbuf []byte) []byte {
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
func (p *DevRWEquipment) BeforePackageFrame(inbuf []byte) []byte {
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
func (p *DevRWEquipment) ParseFrame(inbuf []byte) (models.FrameSelfEquipment, int, error) {
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
func (p *DevRWEquipment) PackageFrame(rsctl byte, cmd byte, inbuf []byte) ([]byte, error) {
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

func (p *DevRWEquipment) FuncSend(rsctl byte, cmd byte, inbuf []byte) bool {
	outbuf, _ := p.PackageFrame(rsctl, cmd, inbuf)

	util.FileLogs.Info("DevRWEquipment:(%d-%s)-%d,%s.\r\n",
		cmd, p.GetCmdDes(cmd), len(outbuf), util.ConvertByte2Hexstring(outbuf, true))

	return p.SendProc(outbuf, true)
}

func (p *DevRWEquipment) FuncSend2(rsctl byte, cmd byte, inbuf []byte) bool {
	outbuf, _ := p.PackageFrame(rsctl, cmd, inbuf)

	//添加到map中
	pInfo := new(models.EquipFrameinfo)
	pInfo.LastTm = util.GetTimeStampSec()
	pInfo.Trynums = 1
	pInfo.Rsctl = rsctl
	pInfo.Frame = make([]byte, len(outbuf))
	copy(pInfo.Frame, outbuf)
	p.FrameMap.ReSet(util.Convertb2s(cmd), pInfo)

	util.FileLogs.Info("DevRWEquipment2:(%d-%s)-%d,%s.\r\n",
		cmd, p.GetCmdDes(cmd), len(outbuf), util.ConvertByte2Hexstring(outbuf, true))

	return p.SendProc(outbuf, true)
}

//负应答
func (p *DevRWEquipment) Package31(rsctl byte) bool {
	return p.FuncSend(rsctl, models.RW_SECMD31, nil)
}

//正应答
func (p *DevRWEquipment) Package30(rsctl byte) bool {
	return p.FuncSend(rsctl, models.RW_SECMD30, nil)
}

func (p *DevRWEquipment) Parse30(info models.FrameSelfEquipment) bool {
	//查看是否有对应帧，有则删除，
	p.FrameMap.Lock.Lock()
	defer p.FrameMap.Lock.Unlock()

	for k, v := range p.FrameMap.BM {
		if v == nil {
			continue
		}

		ev := v.(*models.EquipFrameinfo)
		if ev.Rsctl == info.Rsctl {
			delete(p.FrameMap.BM, k)
			break
		}
	}

	return false
}

func (p *DevRWEquipment) Parse31(info models.FrameSelfEquipment) bool {
	p.FrameMap.Lock.Lock()
	defer p.FrameMap.Lock.Unlock()

	for _, v := range p.FrameMap.BM {
		if v == nil {
			continue
		}

		ev := v.(models.EquipFrameinfo)
		if ev.Rsctl == info.Rsctl {
			ev.Trynums += 1
			p.SendProc(ev.Frame, true)
			break
		}
	}
	return false
}

//设备加电上报
func (p *DevRWEquipment) Parse41(info models.FrameSelfEquipment) bool {
	return p.Package61(info.Rsctl)
}

func (p *DevRWEquipment) GetCardnums(s []byte) int {
	slen := len(s)
	i := 0
	for i = 0; i < slen; i += 1 {
		if s[i] != 0x30 {
			break
		}
	}

	return util.ConvertS2I(string(s[i:slen]))
}

func (p *DevRWEquipment) GetWorkstation(bdata byte) byte {
	if bdata == models.SEUP {
		return p.KGStateInfo.UpCurWorkstation
	}

	return p.KGStateInfo.DnCurWorkstation
}

//状态信息上报
func (p *DevRWEquipment) Parse42(info models.FrameSelfEquipment) bool {
	var stainfo models.EquipmentKGInfo
	pos := 0
	inbuf := info.Data

	stainfo.UpCurWorkstation = inbuf[pos]
	pos += 1
	stainfo.DnCurWorkstation = inbuf[pos]
	pos += 1

	for i := 0; i < 4; i += 1 {
		stainfo.KGRsds[i].KGState = inbuf[pos]
		pos += 1
		stainfo.KGRsds[i].KJState = inbuf[pos]
		pos += 1
		stainfo.KGRsds[i].KNums = p.GetCardnums(inbuf[pos : pos+3])
		pos += 3
	}

	p.KGStateInfo = stainfo
	return true
}

//退卡信息
func (p *DevRWEquipment) Parse43(info models.FrameSelfEquipment) bool {
	//工位信息：31H 上工位；32H 为下工位；33H 为退卡失败
	//当前卡机编号： 31H 为 1#通道；32H 为 2#通道；33H为 3#通道；34 为 4#通道；
	util.FileLogs.Info("退卡:工位:%02x,卡机编号:%02x.", info.Data[0], info.Data[1])
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//按键信息
func (p *DevRWEquipment) Parse44(info models.FrameSelfEquipment) bool {
	util.FileLogs.Info("有卡:工位:%02x,卡机编号:%02x.", info.Data[0], info.Data[1])

	//直接回复确认
	return p.Package30(info.Rsctl)
}

//卡取走
func (p *DevRWEquipment) Parse45(info models.FrameSelfEquipment) bool {
	util.FileLogs.Info("卡取走:工位:%02x,卡机编号:%02x.", info.Data[0], info.Data[1])

	//直接回复确认
	return p.Package30(info.Rsctl)
}

//卡夹信息
func (p *DevRWEquipment) Parse46(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//回收卡完成
func (p *DevRWEquipment) Parse47(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//收卡完成
func (p *DevRWEquipment) Parse49(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//核心板软件版本
func (p *DevRWEquipment) Parse56(info models.FrameSelfEquipment) bool {
	//直接回复确认
	return p.Package30(info.Rsctl)
}

//初始化信息帧
func (p *DevRWEquipment) Package61(rsctl byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	sNums := "500"
	copy(buf[pos:], []byte(sNums))
	pos += 3

	sTime := util.GetNow(true)
	copy(buf[pos:], []byte(sTime))
	pos += len(sTime)

	return p.FuncSend2(rsctl, models.RW_SECMD61, buf[0:pos])
}

//回收卡信息
func (p *DevRWEquipment) Package62(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = 0x30 //p.GetWorkstation(bdata)
	pos += 1

	return p.FuncSend2(p.GetRsctl(), models.RW_SECMD62, buf[0:pos])
}

//退卡信息
func (p *DevRWEquipment) Package63(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = 0x30 //p.GetWorkstation(bdata)
	pos += 1

	return p.FuncSend2(p.GetRsctl(), models.RW_SECMD63, buf[0:pos])
}

//收卡信息
func (p *DevRWEquipment) Package64(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = 0x30 //p.GetWorkstation(bdata)
	pos += 1

	return p.FuncSend2(p.GetRsctl(), models.RW_SECMD64, buf[0:pos])
}

//查询卡机状态
func (p *DevRWEquipment) Package65() bool {
	return p.FuncSend2(p.GetRsctl(), models.RW_SECMD65, nil)
}

//查询卡夹
func (p *DevRWEquipment) Package66(bdata byte) bool {
	buf := make([]byte, models.MAX_SIZE1024)
	pos := 0

	buf[pos] = 0x30 //p.GetWorkstation(bdata)
	pos += 1

	return p.FuncSend2(p.GetRsctl(), models.RW_SECMD66, buf[0:pos])
}

func (p *DevRWEquipment) Recvproc(inbuf []byte) (int, error) {
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
	case models.RW_SECMD30:
		p.Parse30(info)
		break
	case models.RW_SECMD31:
		p.Parse31(info)
		break
	case models.RW_SECMD41:
		p.Parse41(info)
		break
	case models.RW_SECMD42:
		p.Parse42(info)
		break
	case models.RW_SECMD43:
		p.Parse43(info)
		break
	case models.RW_SECMD44:
		p.Parse44(info)
		break
	case models.RW_SECMD45:
		p.Parse45(info)
		break
	case models.RW_SECMD46:
		p.Parse46(info)
		break
	case models.RW_SECMD47:
		p.Parse47(info)
		break
	case models.RW_SECMD49:
		p.Parse49(info)
		break
	case models.RW_SECMD56:
		p.Parse56(info)
		break

	}

	return offset, nil
}

func (p *DevRWEquipment) goFrameCheck() {
	for {
		p.FrameMap.Lock.Lock()
		for _, v := range p.FrameMap.BM {
			if v == nil {
				continue
			}

			ev := v.(*models.EquipFrameinfo)
			ev.Trynums += 1
			p.SendProc(ev.Frame, false)
		}

		p.FrameMap.Lock.Unlock()

		util.MySleep_s(3)
	}
}

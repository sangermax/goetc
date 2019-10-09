package util

import (
	"errors"
	"fmt"
	"strings"

	"FTC/config"
	"FTC/models"
)

func ETCParseB0(inbuf []byte) (models.ETCFrameB0, error) {
	var frame models.ETCFrameB0
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B0 {
		return models.ETCFrameB0{}, errors.New("B0,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.RSUStatus = int(inbuf[pos])
	pos += 1
	frame.RSUTerminalId = ConvertByte2Hexstring(inbuf[pos:pos+6], false)
	pos += 6
	frame.RSUAlgId = int(inbuf[pos])
	pos += 1
	frame.RSUManuID = int(inbuf[pos])
	pos += 1
	frame.RSUIndividualID = ConvertByte2Hexstring(inbuf[pos:pos+3], false)
	pos += 3
	frame.RSUVersion = ConvertByte2Hexstring(inbuf[pos:pos+2], false)
	pos += 2
	frame.Reserved = ConvertByte2Hexstring(inbuf[pos:pos+5], false)
	pos += 5

	return frame, nil
}

func ETCParseB2(inbuf []byte) (models.ETCFrameB2, error) {
	var frame models.ETCFrameB2
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B2 {
		return models.ETCFrameB2{}, errors.New("B2,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.OBUID = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ErrorCode = int(inbuf[pos])
	pos += 1
	frame.ContractProvider = ConvertByte2Hexstring(inbuf[pos:pos+8], false)
	pos += 8
	frame.ContractType = int(inbuf[pos])
	pos += 1
	frame.ContractVersion = int(inbuf[pos])
	pos += 1
	frame.ContractSerialNumber = ConvertByte2Hexstring(inbuf[pos:pos+8], false)
	pos += 8
	frame.ContractSignedDate = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ContractExpiredDate = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.CPUCardID = ConvertByte2Hexstring(inbuf[pos:pos+10], false)
	pos += 10
	frame.Equitmentstatus = int(inbuf[pos])
	pos += 1
	frame.OBUStatus = make([]byte, 2)
	copy(frame.OBUStatus, inbuf[pos:pos+2])
	pos += 2

	return frame, nil
}

func ETCParseB3(inbuf []byte) (models.ETCFrameB3, error) {
	var frame models.ETCFrameB3
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B3 {
		return models.ETCFrameB3{}, errors.New("B3,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.OBUID = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ErrorCode = int(inbuf[pos])
	pos += 1

	utf8str, err := GbkToUtf8(TrimPlate(inbuf[pos : pos+12]))
	pos += 12
	if err == nil {
		frame.VehPlate = string(Remove0(utf8str))
	}

	frame.VehColor = Bytes2Int_L(inbuf[pos : pos+2])
	pos += 2
	frame.VehClass = int(inbuf[pos])
	pos += 1
	frame.VehUserType = int(inbuf[pos])
	pos += 1
	frame.VehDimens = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.VehWheels = int(inbuf[pos])
	pos += 1
	frame.VehAxies = int(inbuf[pos])
	pos += 1
	frame.VehWheelBases = ConvertByte2Hexstring(inbuf[pos:pos+2], false)
	pos += 2
	frame.VehWeightLimits = ConvertByte2Hexstring(inbuf[pos:pos+3], false)
	pos += 3
	frame.VehSpecificinfo = ConvertByte2Hexstring(inbuf[pos:pos+16], false)
	pos += 16
	frame.VehEngineNum = ConvertByte2Hexstring(inbuf[pos:pos+16], false)
	pos += 16

	return frame, nil
}

func ETCParse0008(f0008 []byte) (models.F0008Info, error) {
	if len(f0008) < models.ETCLENGTH_0008 {
		return models.F0008Info{}, errors.New("0008,文件长度错误")
	}
	var tmpinfo models.F0008Info
	pos := 0

	tmpinfo.FlagNums = int(f0008[pos])
	pos += 1
	tmpinfo.LastFlag = Bytes2Short_L(f0008[pos : pos+2])
	pos += 2

	for i := 0; i < tmpinfo.FlagNums; i += 1 {
		tmpinfo.FlagRsds[i] = Bytes2Short_L(f0008[pos : pos+2])
		pos += 2
	}
	return tmpinfo, nil
}

func ETCParse0015(f0015 []byte) (models.F0015Info, error) {
	if len(f0015) < models.ETCLENGTH_0015 {
		return models.F0015Info{}, errors.New("0015,文件长度错误")
	}
	var tmp0015info models.F0015Info
	pos := 0

	utf8Cardissue, err := GbkToUtf8(f0015[pos : pos+8])
	pos += 8
	if err == nil {
		tmp0015info.CardIssue = string(utf8Cardissue)
	}

	tmp0015info.CardType = int(f0015[pos])
	pos += 1
	tmp0015info.CardVersion = int(f0015[pos])
	pos += 1

	strnet := ETCConverCardNetworkFromIssue(tmp0015info.CardIssue)
	if strnet == "" {
		tmp0015info.CardNetWork = ConvertByte2Hexstring(f0015[pos:pos+2], false)
	} else {
		if strnet == models.ARMY_CARDNETWORK {
			tmp0015info.CardNetWork = strnet
		} else {
			tmp0015info.CardNetWork = strnet[0:2] + ConvertByte2Hexstring(f0015[pos+1:pos+2], false)
		}
	}
	pos += 2

	tmp0015info.CardId = ConvertByte2Hexstring(f0015[pos:pos+8], false)
	pos += 8
	tmp0015info.StartTm = ConvertByte2UnixTmstring(f0015[pos:pos+4], false)
	pos += 4
	tmp0015info.EndTm = ConvertByte2UnixTmstring(f0015[pos:pos+4], false)
	pos += 4

	utf8str, err := GbkToUtf8(TrimPlate(f0015[pos : pos+12]))
	if err == nil {
		tmp0015info.VehPlate = string(Remove0(utf8str))
	}

	pos += 12
	tmp0015info.UserType = int(f0015[pos])
	pos += 1

	if tmp0015info.CardVersion >= 0x40 {
		tmp0015info.VehColor = int(f0015[pos])
		pos += 1
		tmp0015info.VehClass = int(f0015[pos])
		pos += 1
	} else {
		//卡网络编号调整

		tmp0015info.VehColor = Bytes2Int_L(f0015[pos : pos+2])
		pos += 2
	}

	return tmp0015info, nil
}

func ETCParse0019(f0019 []byte) (models.F0019Info, error) {
	if len(f0019) < models.ETCLENGTH_0019 {
		return models.F0019Info{}, errors.New("0019,文件长度错误")
	}
	var tmp0019info models.F0019Info
	pos := 0

	pos = 3
	tmp0019info.InStationNetWork = ConvertByte2Hexstring(f0019[pos:pos+2], false)
	pos += 2
	tmp0019info.InStation = StaByte2StaStr(f0019[pos : pos+2])
	pos += 2

	tmp0019info.InLane = ConvertI2S(int(f0019[pos]))
	pos += 1

	intimesec := Bytes2Int_B(f0019[pos : pos+4])
	tmp0019info.InTime = ConvertTimestmp2Time(int64(intimesec))
	pos += 4

	tmp0019info.VehClass = int(f0019[pos])
	pos += 1
	tmp0019info.FlowState = int(f0019[pos])
	pos += 1
	tmp0019info.FlagSta = ConvertI2S(Bytes2Int_B(f0019[pos : pos+4]))
	pos += 4
	tmp0019info.AuxSta = StaByte2StaStr(f0019[pos : pos+2])
	pos += 2

	tmp0019info.OutStationNetWork = ConvertByte2Hexstring(f0019[pos:pos+2], false)
	pos += 2
	tmp0019info.OutStation = StaByte2StaStr(f0019[pos : pos+2])
	pos += 2

	tmp0019info.InOperator = ConvertI2S(Bytes2Short_B(f0019[pos : pos+2]))
	pos += 2
	tmp0019info.InBanci = int(f0019[pos])
	pos += 1

	utf8str, err := GbkToUtf8(TrimPlate(f0019[pos : pos+12]))
	pos += 12
	if err == nil {
		tmp0019info.VehPlate = string(Remove0(utf8str))
	}

	pos += 4
	return tmp0019info, nil
}

func ETCParse0018(inbuf []byte) (models.F0018Info, error) {
	if len(inbuf) < models.ETCLENGTH_0018 {
		return models.F0018Info{}, errors.New("0018,文件长度错误")
	}
	var frame models.F0018Info
	pos := 0

	frame.EtcTradNo = ConvertByte2Hexstring(inbuf[pos:pos+2], false)
	pos += 2

	tmpbuf := make([]byte, 4)
	tmpbuf[0] = 0x00
	copy(tmpbuf[1:], inbuf[pos:pos+3])
	frame.OverToll = Bytes2Int_B(tmpbuf)
	pos += 3

	frame.Toll = Bytes2Int_B(inbuf[pos : pos+4])
	pos += 4
	frame.Transtype = int(inbuf[pos])
	pos += 1
	frame.PsamTermID = ConvertByte2Hexstring(inbuf[pos:pos+6], false)
	pos += 6
	frame.TransTime = ConvertByte2Hexstring(inbuf[pos:pos+7], false)
	pos += 7

	return frame, nil
}

func ETCParseB4(inbuf []byte) (models.ETCFrameB4, error) {
	var frame models.ETCFrameB4
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B4 {
		return models.ETCFrameB4{}, errors.New("B4,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.OBUID = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ErrorCode = int(inbuf[pos])
	pos += 1
	frame.CardRestMoney = Bytes2Int_B(inbuf[pos : pos+4])
	pos += 4

	tmp0015, err := ETCParse0015(inbuf[pos : pos+43])
	pos += 43
	if err != nil {
		fmt.Printf("ParseB4 0015 failed:%s.\n", err.Error())
		return models.ETCFrameB4{}, errors.New("B4," + err.Error())
	}

	var tmp0019 models.F0019Info
	tmp0019, err = ETCParse0019(inbuf[pos : pos+43])
	pos += 43
	if err != nil {
		fmt.Printf("ParseB4 read 0019 failed:%s.\n", err.Error())
		return models.ETCFrameB4{}, errors.New("B4," + err.Error())
	}

	frame.F0015Info = tmp0015
	frame.F0019Info = tmp0019

	return frame, nil
}

func ETCParseB5(inbuf []byte) (models.ETCFrameB5, error) {
	var frame models.ETCFrameB5
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B5 {
		return models.ETCFrameB5{}, errors.New("B5,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.OBUID = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ErrorCode = int(inbuf[pos])
	pos += 1
	frame.TransTime = ConvertByte2Hexstring(inbuf[pos:pos+7], false)
	pos += 7
	frame.PsamTransNo = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.EtcTradNo = ConvertByte2Hexstring(inbuf[pos:pos+2], false)
	pos += 2
	frame.Transtype = int(inbuf[pos])
	pos += 1
	frame.CardRestMoney = Bytes2Int_B(inbuf[pos : pos+4])
	pos += 4
	frame.Tac = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	wrtm := Bytes2Int_B(inbuf[pos : pos+4])
	frame.WrFileTime = ConvertTimestmp2Time(int64(wrtm))
	pos += 4

	return frame, nil
}

func ETCParseB7(inbuf []byte) (models.ETCFrameB7, error) {
	var frame models.ETCFrameB7
	framelen := len(inbuf)
	if framelen < models.ETCLENGTH_B7 {
		return models.ETCFrameB7{}, errors.New("B7,长度不足")
	}

	pos := 0
	pos += 1
	frame.RSCTL = int(inbuf[pos])
	pos += 1
	frame.FrameType = int(inbuf[pos])
	pos += 1
	frame.OBUID = ConvertByte2Hexstring(inbuf[pos:pos+4], false)
	pos += 4
	frame.ErrorCode = int(inbuf[pos])
	pos += 1
	frame.RecordNum = int(inbuf[pos])
	pos += 1

	if framelen < 11+23*frame.RecordNum {
		return models.ETCFrameB7{}, errors.New("B7,长度不足2")
	}

	for i := 0; i < frame.RecordNum; i += 1 {
		f0018, err := ETCParse0018(inbuf[pos : pos+23])
		if err == nil {
			frame.F0018Rsds = append(frame.F0018Rsds, f0018)
		}
		pos += 23
	}

	return frame, nil
}

func ETCConverCardNetworkFromIssue(strIssue string) string {
	if strIssue == "" {
		return ""
	}

	if strings.Index(strIssue, "军车") >= 0 {
		return models.ARMY_CARDNETWORK
	}

	if strings.Index(strIssue, "北京") >= 0 {
		return models.BJ_NETWORK
	}

	if strings.Index(strIssue, "天津") >= 0 {
		return models.TJ_NETWORK
	}

	if strings.Index(strIssue, "河北") >= 0 {
		return models.HEB_NETWORK
	}

	if strings.Index(strIssue, "山西") >= 0 {
		return models.SX_NETWORK
	}

	if strings.Index(strIssue, "内蒙") >= 0 {
		return models.NM_NETWORK
	}

	if strings.Index(strIssue, "辽宁") >= 0 {
		return models.LN_NETWORK
	}

	if strings.Index(strIssue, "吉林") >= 0 {
		return models.JL_NETWORK
	}

	if strings.Index(strIssue, "龙江") >= 0 {
		return models.HLJ_NETWORK
	}

	if strings.Index(strIssue, "上海") >= 0 {
		return models.SH_NETWORK
	}

	if strings.Index(strIssue, "江苏") >= 0 {
		return models.DEFAULT_NET
	}

	if strings.Index(strIssue, "浙江") >= 0 {
		return models.ZJ_NETWORK
	}

	if strings.Index(strIssue, "安徽") >= 0 {
		return models.AH_NETWORK
	}

	if strings.Index(strIssue, "福建") >= 0 {
		return models.FJ_NETWORK
	}

	if strings.Index(strIssue, "江西") >= 0 {
		return models.JX_NETWORK
	}

	if strings.Index(strIssue, "山东") >= 0 {
		return models.SD_NETWORK
	}

	if strings.Index(strIssue, "河南") >= 0 {
		return models.HEN_NETWORK
	}

	if strings.Index(strIssue, "湖北") >= 0 {
		return models.HUB_NETWORK
	}

	if strings.Index(strIssue, "湖南") >= 0 {
		return models.HUN_NETWORK
	}

	if strings.Index(strIssue, "广东") >= 0 {
		return models.GD_NETWORK
	}

	if strings.Index(strIssue, "广西") >= 0 {
		return models.GX_NETWORK
	}

	if strings.Index(strIssue, "海南") >= 0 {
		return models.HAIN_NETWORK
	}

	if strings.Index(strIssue, "重庆") >= 0 {
		return models.CQ_NETWORK
	}

	if strings.Index(strIssue, "四川") >= 0 {
		return models.SC_NETWORK
	}

	if strings.Index(strIssue, "贵州") >= 0 {
		return models.GZ_NETWORK
	}

	if strings.Index(strIssue, "云南") >= 0 {
		return models.YN_NETWORK
	}

	if strings.Index(strIssue, "西藏") >= 0 {
		return models.XZ_NETWORK
	}

	if strings.Index(strIssue, "陕西") >= 0 {
		return models.SHANXI_NETWORK
	}

	if strings.Index(strIssue, "甘肃") >= 0 {
		return models.GS_NETWORK
	}

	if strings.Index(strIssue, "青海") >= 0 {
		return models.QH_NETWORK
	}

	if strings.Index(strIssue, "宁夏") >= 0 {
		return models.NX_NETWORK
	}

	if strings.Index(strIssue, "新疆") >= 0 {
		return models.XJ_NETWORK
	}

	return ""
}

func ETCPackage0019(info models.F0019Info) []byte {
	buf := make([]byte, 43)
	pos := 0

	buf[pos] = 0xAA
	pos += 1
	buf[pos] = 0x29
	pos += 1
	buf[pos] = 0x00
	pos += 1

	if info.InStationNetWork != "" {
		innet := ConvertS2I(info.InStationNetWork)
		buf[pos] = (byte)(innet / 100)
		pos += 1
		buf[pos] = (byte)(innet % 100)
		pos += 1
	} else {
		pos += 2
	}

	if info.InStation != "" {
		insta := ConvertS2I(info.InStation)
		buf[pos] = (byte)(insta / 10000)
		pos += 1
		buf[pos] = (byte)((insta%10000)%100) | ((byte)((insta%10000)/100))<<5
		pos += 1
	} else {
		pos += 2
	}

	buf[pos] = Converts2b(info.InLane)
	pos += 1

	bsIntime := ConvertTime2Unix(info.InTime, true)
	copy(buf[pos:], bsIntime)
	pos += 4

	buf[pos] = (byte)(info.VehClass)
	pos += 1

	buf[pos] = (byte)(info.FlowState)
	pos += 1

	pos += 4
	pos += 2

	if info.OutStationNetWork != "" {
		outnet := ConvertS2I(info.OutStationNetWork)
		buf[pos] = (byte)(outnet / 100)
		pos += 1
		buf[pos] = (byte)(outnet % 100)
		pos += 1
	} else {
		pos += 2
	}

	if info.OutStation != "" {
		outsta := ConvertS2I(info.OutStation)
		buf[pos] = (byte)(outsta / 10000)
		pos += 1
		buf[pos] = (byte)((outsta%10000)%100) | ((byte)((outsta%10000)/100))<<5
		pos += 1
	} else {
		pos += 2
	}

	if info.InOperator != "" {
		op := ConvertS2I(info.InOperator) % 1000
		bsop := Short2Bytes_B(op)
		copy(buf[pos:], bsop)
	}
	pos += 2

	buf[pos] = (byte)(info.InBanci)
	pos += 1

	bsplate, err := Utf8ToGbk(([]byte)(info.VehPlate))
	if err == nil {
		copy(buf[pos:], bsplate)
	}
	pos += 12

	pos += 4

	//fmt.Println(ConvertByte2Hexstring(buf, true))
	return buf
}

func ETCBcc(buf []byte) byte {
	var val byte
	length := len(buf)
	for i := 0; i < length; i++ {
		val ^= buf[i]
	}

	return val
}

func ETCReversalRsctl(rsctl byte) byte {
	l := (rsctl & 0x0F)
	h := (rsctl & 0xF0) >> 4

	return (l<<4 | h)
}

//协议
func ETCEnpack(inbuf []byte) ([]byte, error) {
	outbuf := make([]byte, 1024)
	pos := 0

	outbuf[pos] = 0xFF
	pos += 1
	copy(outbuf[pos:], inbuf)
	pos += len(inbuf)
	outbuf[pos] = ETCBcc(outbuf[1 : pos+1])
	pos += 1
	outbuf[pos] = 0xFF
	pos += 1

	databuf := ETCBeforePackFrame(outbuf[0:pos])
	return databuf, nil
}

func ETCDepack(inbuf []byte) ([]byte, int, error) {
	length := len(inbuf)
	baselen := 4

	if length < baselen {
		return nil, 0, errors.New("还未收完")
	}

	i := 0
	for {
		startpos := 0
		endpos := 0
		for ; i < length; i = i + 1 {
			if inbuf[i] == 0xFF {
				startpos = i
				break
			}
		}

		if i >= length {
			return nil, startpos, errors.New("未找到帧头")
		}

		for i = startpos + 1; i < length; i = i + 1 {
			if inbuf[i] == 0xFF {
				endpos = i
				break
			}
		}

		if i >= length {
			return nil, startpos, errors.New("未找到帧尾")
		}

		framelen := endpos - startpos + 1
		frame := make([]byte, framelen)
		copy(frame, inbuf[startpos:endpos+1])

		//转义
		//记录日志，然后转义处理
		FileLogs.Info("ETCDepack：%02X_%s", frame[2], ConvertByte2Hexstring(frame, false))
		outbuf := ETCBeforeParseFrame(frame)

		//bcc校验
		buflen := len(outbuf)
		b1 := frame[buflen-2]
		b2 := ETCBcc(frame[1 : buflen-2])
		if b1 != b2 {
			FileLogs.Info("ETC天线解析：BCC校验码失败:%02X,%02X", b1, b2)
			return nil, endpos + 1, errors.New("bcc校验失败")
		}

		return outbuf, endpos + 1, nil
	}
}

//解析前转义
func ETCBeforeParseFrame(inbuf []byte) []byte {
	outbuf := make([]byte, 1024)
	len1 := len(inbuf)
	i := 0
	j := 0

	for ; i < len1; i = i + 1 {
		if i+1 < len1 {
			if inbuf[i] == 0xFE && inbuf[i+1] == 0x01 {
				outbuf[j] = 0xFF
				j += 1
				i += 1
			} else if inbuf[i] == 0xFE && inbuf[i+1] == 0x00 {
				outbuf[j] = 0xFE
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
func ETCBeforePackFrame(inbuf []byte) []byte {
	outbuf := make([]byte, 1024)
	len1 := len(inbuf)
	i := 0
	j := 0

	//帧头
	outbuf[j] = inbuf[i]
	i += 1
	j += 1

	for ; i < len1-1; i = i + 1 {
		if inbuf[i] == 0xFF {
			outbuf[j] = 0xFE
			j += 1
			outbuf[j] = 0x01
			j += 1

		} else if inbuf[i] == 0xFE {
			outbuf[j] = 0xFE
			j += 1
			outbuf[j] = 0x00
			j += 1

		} else {
			outbuf[j] = inbuf[i]
			j += 1
		}
	}

	//帧尾
	outbuf[j] = inbuf[len1-1]
	j += 1
	return outbuf[0:j]
}

func ETCPackNull(rsctl byte) []byte {
	outbuf := make([]byte, 4)

	pos := 0
	outbuf[pos] = 0xFF
	pos += 1
	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = ETCBcc(outbuf[1:pos])
	pos += 1
	outbuf[pos] = 0xFF
	pos += 1

	return outbuf[0:pos]
}

func ETCPackC0(rsctl byte) []byte {
	var req models.ETCFrameC0

	stm, secs := GetTimeStampSecAndTm(true)
	req.Seconds = int(secs)
	req.Datetime = stm
	req.LaneMode = config.ConfigData["tollmode"].(int)
	req.WaitTime = config.ConfigData["etcWaittime"].(int)
	req.TxPower = config.ConfigData["etcPower"].(int)
	req.PLLChannelID = config.ConfigData["etcChannelID"].(int)

	outbuf := make([]byte, 17)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C0
	pos += 1
	copy(outbuf[pos:], Int2Bytes_B(req.Seconds))
	pos += 4
	copy(outbuf[pos:], ConvertHexstring2Byte(req.Datetime))
	pos += 7
	outbuf[pos] = byte(req.LaneMode)
	pos += 1
	outbuf[pos] = byte(req.WaitTime)
	pos += 1
	outbuf[pos] = byte(req.TxPower)
	pos += 1
	outbuf[pos] = byte(req.PLLChannelID)
	pos += 1

	framebuf, _ := ETCEnpack(outbuf[0:pos])
	return framebuf
}

func ETCPackC1(rsctl byte, obuid []byte) []byte {
	outbuf := make([]byte, 6)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C1
	pos += 1
	copy(outbuf[pos:], obuid)
	pos += 4

	framebuf, _ := ETCEnpack(outbuf[0:pos])
	return framebuf
}

func ETCPackC2(rsctl byte, obuid []byte, stoptype byte) []byte {
	outbuf := make([]byte, 7)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C2
	pos += 1
	copy(outbuf[pos:], obuid)
	pos += 4
	outbuf[pos] = stoptype
	pos += 1

	framebuf, _ := ETCEnpack(outbuf[0:pos])
	return framebuf
}

func ETCPackC3(rsctl byte, req models.ETCFrameC3) []byte {
	outbuf := make([]byte, 53)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C3
	pos += 1
	copy(outbuf[pos:], ConvertHexstring2Byte(req.OBUID))
	pos += 4
	b0019 := ETCPackage0019(req.F0019Info)
	copy(outbuf[pos:], b0019[3:])
	pos += 40
	copy(outbuf[pos:], ConvertHexstring2Byte(req.Datetime))
	pos += 7

	framebuf, _ := ETCEnpack(outbuf[0:pos])
	return framebuf
}

func ETCPackC6(rsctl byte, req models.ETCFrameC6) []byte {
	outbuf := make([]byte, 57)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C6
	pos += 1
	copy(outbuf[pos:], ConvertHexstring2Byte(req.OBUID))
	pos += 4
	copy(outbuf[pos:], Int2Bytes_B(req.ConsumeMoney))
	pos += 4
	b0019 := ETCPackage0019(req.F0019Info)
	copy(outbuf[pos:], b0019[3:])
	pos += 40

	copy(outbuf[pos:], ConvertHexstring2Byte(req.Datetime))
	pos += 7

	framebuf, _ := ETCEnpack(outbuf[0:pos])
	return framebuf
}

func ETCPackC7(rsctl byte, req models.ETCFrameC7) []byte {
	outbuf := make([]byte, 7)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_C7
	pos += 1
	copy(outbuf[pos:], ConvertHexstring2Byte(req.OBUID))
	pos += 4
	outbuf[pos] = byte(req.RecordNUM)
	pos += 1

	framebuf, _ := ETCEnpack(outbuf[0:pos])

	return framebuf
}

func ETCPack4C(rsctl byte, req models.ETCFrame4C) []byte {
	outbuf := make([]byte, 3)
	pos := 0

	outbuf[pos] = rsctl
	pos += 1
	outbuf[pos] = models.CMD_4C
	pos += 1
	outbuf[pos] = byte(req.Antennastatus)
	pos += 1

	framebuf, _ := ETCEnpack(outbuf[0:pos])

	return framebuf
}

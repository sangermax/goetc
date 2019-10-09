package device

import (
	"FTC/models"
	"FTC/util"
	"strings"
)

//缴费机屏显示
//补空格
func AppendSpaceTo(buf []byte, nums int) []byte {
	length := len(buf)
	if len(buf) > nums {
		return buf[0:nums]
	}

	var databuf []byte
	for i := 0; i < nums; i++ {
		if i < length {
			databuf = append(databuf, buf[i])
		} else {
			databuf = append(databuf, ' ')
		}
	}

	return databuf
}

func ShowTip16(str string, color byte) models.FrameSE74 {
	util.FileLogs.Info("自助机控制屏信息:%s", str)
	var info models.FrameSE74
	info.Aligyntype = models.AlignLeft
	info.SpaceValue = models.SPACE0
	info.Xpos = 0
	info.Ypos = 0
	info.WidthShow = 128
	info.HeightShow = 64

	info.Fontsize = models.FONT16
	switch color {
	case models.ColorRed:
		info.RedColor = 255
		info.GreenColor = 0
		info.BlueColor = 0
	case models.ColorYellow:
		info.RedColor = 255
		info.GreenColor = 255
		info.BlueColor = 0
	default: //models.ColorGreen
		info.RedColor = 0
		info.GreenColor = 255
		info.BlueColor = 0
	}

	buf := make([]byte, 65)
	pos := 0
	grps := strings.Split(str, ";")
	//fmt.Println("len(grps):%d...\r\n", len(grps))
	for i := 0; i < len(grps) && i < 4; i++ {
		bs, _ := util.Utf8ToGbk([]byte(grps[i]))
		bline := AppendSpaceTo(bs, 16)
		copy(buf[pos:], bline)
		pos += len(bline)
	}

	info.Content = make([]byte, pos)
	copy(info.Content, buf[0:pos])

	return info
}

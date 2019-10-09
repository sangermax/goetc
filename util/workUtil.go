package util

import (
	"FTC/models"
	"fmt"
	"strings"
	"time"
)

//工位，返回值1表示系统定义的工位，返回值2用于自助机定义的工位
func GetSE(vehclass int) (byte, byte) {
	switch vehclass {
	case 1, 2, 3, 4:
		return models.WORKSTATION_DN, models.SEDN
	case 11, 12, 13, 14, 15:
		return models.WORKSTATION_UP, models.SEUP
	default:
		return models.WORKSTATION_DN, models.SEDN
	}

}

func GetPlatePayChkDes(plateChk int) string {
	switch plateChk {
	case models.PLATEPAY_YES:
		return "车牌付"
	default:
		return "非车牌付"
	}

}

func GetShiftId(strTm string) string {
	tm, _ := time.Parse("2006-01-02 15:04:05", strTm)
	if tm.Hour() >= 0 && tm.Hour() < 8 {
		return "1"
	} else if tm.Hour() >= 8 && tm.Hour() < 16 {
		return "2"
	} else if tm.Hour() >= 16 && tm.Hour() <= 23 {
		return "3"
	}

	return "1"
}

func GetShiftDate(strTm string) string {
	tm, _ := time.Parse("2006-01-02 15:04:05", strTm)
	return tm.Format("20060102")
}

func GetVehclassDes(vehclass int) string {
	switch vehclass {
	case 1, 2, 3, 4:
		return "客" + ConvertI2S(vehclass)
	case 11, 12, 13, 14, 15:
		return "货" + ConvertI2S(vehclass-10)
	default:
		return "未知" + ConvertI2S(vehclass)
	}
}

func GetPayMethodDes(paymethod string) string {
	switch paymethod {
	case models.PAYMETHOD_WX:
		return paymethod + "-支付宝"
	case models.PAYMETHOD_ZFB:
		return paymethod + "-微信"
	default:
		return paymethod
	}
}

func GetEPVehColor(strColor string) string {
	str := Unicode2Utf8(strColor)
	iRet := models.PCOLOR_BLUE
	switch str {
	case "蓝":
		iRet = models.PCOLOR_BLUE
	case "黄":
		iRet = models.PCOLOR_YELLOW
	case "绿":
		iRet = models.PCOLOR_GREEN
	case "白":
		iRet = models.PCOLOR_WHITE
	case "黑":
		iRet = models.PCOLOR_BLACK

	}

	return ConvertI2S(iRet)
}

func GetEPVehClass(strApprovedLoad, strVehtype string) string {
	FileLogs.Info("GetEPVehClass:%s,%s.", strVehtype, strApprovedLoad)

	if strings.HasPrefix(strVehtype, "K") {
		iLoad := ConvertS2I(strApprovedLoad)

		if iLoad >= 1 && iLoad <= 7 {
			return "1"
		}

		if iLoad >= 8 && iLoad <= 19 {
			return "2"
		}

		if iLoad >= 20 && iLoad <= 39 {
			return "3"
		}

		if iLoad >= 40 {
			return "4"
		}
	}
	return "1"
}

func GetCreateByDes(flag string) string {
	switch flag {
	case models.CREATEBY_COIL:
		return "线圈捕获"
	case models.CREATEBY_EP:
		return "电子车牌捕获"
	case models.CREATEBY_ETC:
		return "ETC捕获"
	case models.CREATEBY_PLATE:
		return "车辆识别捕获"
	case models.CREATEBY_CARD:
		return "自助机卡片捕获"
	}

	return flag
}

func GetFeeShow(toll string) string {
	return fmt.Sprintf("费额 %.2f元", (float32(ConvertS2I(toll))+float32(0.005))/100.0)
}

//comparePlateLen 为0则全比较，
func ComparePlate(str1, str2 string, samelen int) bool {
	l := 0
	l1 := len(str1)
	l2 := len(str2)
	if l1 <= l2 {
		l = l1
	} else {
		l = l2
	}

	if (l <= 3) || (samelen == 0 && l1 != l2) {
		return false
	}

	j := 0
	if str1[0] == str2[0] && str1[1] == str2[1] && str1[2] == str2[2] {
		j++
	}

	for i := 3; i < l; i++ {
		if str1[i] == str2[i] {
			j++
		}
	}

	if j >= samelen {
		return true
	}

	return false
}

//true :不带短横；false:带短横
func GetTimeStampSecAndTm(bflag bool) (string, int64) {
	now := time.Now()
	s := ""
	if bflag {
		s = now.Format("20060102150405")
	} else {
		s = now.Format("2006-01-02 15:04:05")
	}

	return s, now.Unix()
}

func TrimPlate(s []byte) []byte {
	slen := len(s)
	i := 0
	for i = 0; i < slen; i++ {
		if s[i] == 0x00 {
			break
		}
	}

	return s[0 : i+1]
}

package main

import (
	"FTC/models"
	"FTC/util"
	"encoding/json"
	"fmt"
)

// 【基类】
//定义一个最基础的struct类MsgModel，里面包含一个成员变量msgId
type MsgModel struct {
	msgId   int
	msgType int
}

// MsgModel的一个成员方法，用来设置msgId
func (msg *MsgModel) SetId(msgId int) {
	msg.msgId = msgId
}

func (msg *MsgModel) SetType(msgType int) {
	msg.msgType = msgType
}

//【子类】
// 再定义一个struct为GroupMsgModel，包含了MsgModel，即组合，但是并没有给定MsgModel任何名字，因此是匿名组合
type GroupMsgModel struct {
	MsgModel

	// 如果子类也包含一个基类的一样的成员变量，那么通过子类设置和获取得到的变量都是基类的
	msgId int
}

func (group *GroupMsgModel) GetId() int {
	return group.msgId
}

/*
func (group *GroupMsgModel) SetId(msgId int) {
    group.msgId = msgId
}
*/

type tb struct {
	Bytes []byte
}

func main() {

	/*
		group := &GroupMsgModel{}

		group.SetId(123)
		group.SetType(1)

		fmt.Println("group.msgId =", group.msgId, "\tgroup.MsgModel.msgId =", group.MsgModel.msgId)
		fmt.Println("group.msgType =", group.msgType, "\tgroup.MsgModel.msgType =", group.MsgModel.msgType)
	*/
	//testep()

	/*
		temp := tb{Bytes: []byte("abcde")}
		js, _ := json.Marshal(temp)
		fmt.Println(string(js))
		var t2 tb
		json.Unmarshal(js, &t2)
		fmt.Println(string(t2.Bytes))
	*/

	//二进制流读文件
	b1, err := util.ReadFile("D:\\1.jpg")
	//编码
	s1 := util.EncodeBase64(b1)

	//解码
	b2, err := util.DncodeBase64(s1)
	//写文件
	if err == nil {
		util.WriteFile("D:\\12.jpg", b2)
		fmt.Printf("%d,%d,finish", len(s1), len(b1))
	}

}

func chkEpMultiMsg(epinfo models.EPInfo) bool {

	if epinfo.MsgType == 300 {
		return true
	}

	if epinfo.MsgType == 500 {
		rlt := false

		var epdata models.EPMessageValue500
		tmpbuf, err := json.Marshal(epinfo.MessageValue)
		if err != nil {
			util.MyPrintf("chkEpMultiMsg MessageValue解析失败")
			return false
		}

		err = json.Unmarshal(tmpbuf, &epdata)
		if err != nil {
			util.MyPrintf("chkEpMultiMsg EPTagReportData解析失败")
			return false
		}

		if epdata.ReportRsds == nil {
			util.MyPrintf("chkEpMultiMsg ReportRsds is nil")
			return false
		}

		isize := len(epdata.ReportRsds)
		var tmp1 models.EPMultiHbCustomizedReadFTCcResult
		if isize > 0 {
			for _, v := range epdata.ReportRsds {
				fmt.Println(v.AccessFTCcResult)
				fmt.Println("-------------------------------1")

				fmt.Println("AccessFTCcResult :%d.", len(v.AccessFTCcResult))

				for _, v1 := range v.AccessFTCcResult {
					if v1.HbCustomizedReadFTCcResult == tmp1 {
						util.FileLogs.Info("chkEpMultiMsg 标签选择规则结果为空，抛弃")
						continue
					}

					fmt.Println("////////////////////////////////////")
					fmt.Println(v1.HbCustomizedReadFTCcResult.ReadDataInfo)

					var unit models.EPResultReadInfo
					unit.AntennaID = v.AntennaID
					unit.TID = v.TID
					unit.ReadDataInfo = v1.HbCustomizedReadFTCcResult.ReadDataInfo

				}
			}

			return rlt
		}
	}

	return false
}

func testep() bool {
	//jsonstr := "{\"DeviceSN\" : \"424c1208000186be\",\"MessageValue\" :{\"TagReportData\" : [{\"AntennaID\" : 1,\"FirstSeenTimestampUTC\" : \"16972dd3a67\",		\"LastSeenTimestampUTC\" : \"16972dd3a71\",\"PeakRSSI\" : 24,\"RfFTCcID\" : 0,\"SelectFTCcID\" : 12345,	\"FTCcIndex\" : 0,\"TID\" : \"E881000A2D8462EF\",	\"TagSeenCount\" : 1}]},\"MsgID\" : 62390,\"MsgName\" : \"TagSelectAccessReport\",\"MsgType\" : 500,\"Version\" : \"V2.0\"}"

	jsonstr := "{\"DeviceSN\" : \"424c120c000186a1\",\"MessageValue\" : {\"IsLastedFrame\" : 0,\"SequenceID\" : 2533,\"Status\" : {\"StatusCode\" : 0},\"TagReportData\" : [{\"AccessFTCcID\" : 12345,\"AccessFTCcResult\" : [{\"HbCustomizedReadFTCcResult\" : {\"OpFTCcID\" : 0,\"ReadDataInfo\" : {\"ApprovedLoad\" : \"0\",\"CardNumber\" : \"No.320000000100\",\"Color\" : \"\\u767d\",\"CompulsoryRetirementPeriod\" : \"0\",\"Emissions\" : \"0\",\"LicenseCode\" : \"\\u82cfD\",\"ManufactureDate\" : \"0\",\"PlateNumber\" : \"953MQ\",\"PlateType\" : \"\\u5927\\u578b\\u6c7d\\u8f66\",\"UseCharacteristic\" : \"\\u672a\\u627e\\u5230\",\"ValidityPeriod\" : \"0\",\"VehicleType\" : \"X99\\u5176\\u4ed6\"},\"Result\" : 0}},{\"HbReadFTCcResult\" : {\"OpFTCcID\" : 5,\"ReadData\" : \"AA552C29895C0000000000000000000000000000000000000000000000000000000000000000000000000000\",\"Result\" : 0}}],\"AntennaID\" : 4,\"FirstSeenTimestampUTC\" : \"16978058610\",\"LastSeenTimestampUTC\" : \"1697805861a\",\"PeakRSSI\" : 24,\"RfFTCcID\" : 0,\"SelectFTCcID\" : 12345,\"FTCcIndex\" : 0,\"TID\" : \"E881000A2D8462EF\",\"TagSeenCount\" : 1}]},\"MsgID\" : 49581,\"MsgName\" : \"CachedSelectAccessReport\",\"MsgType\" : 500,\"Version\" : \"V2.0\"}"

	var epinfo models.EPInfo
	err := json.Unmarshal([]byte(jsonstr), &epinfo)
	if err != nil {
		fmt.Println("ReceivedMsgCallback EPInfo 解析失败")
		return false
	}

	chkEpMultiMsg(epinfo)

	/*
		if epinfo.MsgType == 581 {
			var epdata models.EPMessageValue500
			tmpbuf, err := json.Marshal(epinfo.MessageValue)
			if err != nil {
				fmt.Println("chkEpMsg MessageValue解析失败")
				return false
			}

			err = json.Unmarshal(tmpbuf, &epdata)
			if err != nil {
				fmt.Println("chkEpMsg EPTagReportData解析失败")
				return false
			}

			if epdata.ReportRsds == nil {
				fmt.Println("chkEpMsg ReportRsds is nil")
				return false
			}

			isize := len(epdata.ReportRsds)
			fmt.Println("chkEpMsg ReportRsds :%d.", isize)
			if isize > 0 {
				for _, v := range epdata.ReportRsds {
					fmt.Println(v.AccessFTCcResult)
					fmt.Println("-------------------------------1")

					fmt.Println("AccessFTCcResult :%d.", len(v.AccessFTCcResult))

					for _, v1 := range v.AccessFTCcResult {
						var tmp1 models.EPMultiHbCustomizedReadFTCcResult
						var tmp2 models.EPMultiHbReadFTCcResult
						if v1.HbCustomizedReadFTCcResult != tmp1 {
							fmt.Println(v1.HbCustomizedReadFTCcResult.ReadDataInfo)
							fmt.Println("-------------------------------21")
						}

						if v1.HbReadFTCcResult != tmp2 {
							fmt.Println(v1.HbReadFTCcResult)
							fmt.Println("-------------------------------22")
						}

					}

					return true
				}
			}
		}
	*/
	return false
}

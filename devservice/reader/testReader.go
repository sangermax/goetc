package main

import (
	"FTC/config"
	"FTC/models"
	"FTC/util"
	"fmt"
)

var gTestReader ReaderService

func testPackage0019(info models.F0019Info) []byte {
	buf := make([]byte, 43)
	pos := 0

	buf[pos] = 0xAA
	pos += 1
	buf[pos] = 0x29
	pos += 1
	buf[pos] = 0x00
	pos += 1

	if info.InStationNetWork != "" {
		innet := util.ConvertS2I(info.InStationNetWork)
		buf[pos] = (byte)(innet / 100)
		pos += 1
		buf[pos] = (byte)(innet % 100)
		pos += 1
	} else {
		pos += 2
	}

	if info.InStation != "" {
		insta := util.ConvertS2I(info.InStation)
		buf[pos] = (byte)(insta / 10000)
		pos += 1
		buf[pos] = (byte)((insta%10000)%100) | ((byte)((insta%10000)/100))<<5
		pos += 1
	} else {
		pos += 2
	}

	buf[pos] = util.Converts2b(info.InLane)
	pos += 1

	//

	bsIntime := util.ConvertTime2Unix(info.InTime, false)
	copy(buf[pos:], bsIntime)
	pos += 4

	buf[pos] = (byte)(info.VehClass)
	pos += 1

	buf[pos] = (byte)(info.FlowState)
	pos += 1

	pos += 4
	pos += 2

	if info.OutStationNetWork != "" {
		outnet := util.ConvertS2I(info.OutStationNetWork)
		buf[pos] = (byte)(outnet / 100)
		pos += 1
		buf[pos] = (byte)(outnet % 100)
		pos += 1
	} else {
		pos += 2
	}

	if info.OutStation != "" {
		outsta := util.ConvertS2I(info.OutStation)
		buf[pos] = (byte)(outsta / 10000)
		pos += 1
		buf[pos] = (byte)((outsta%10000)%100) | ((byte)((outsta%10000)/100))<<5
		pos += 1
	} else {
		pos += 2
	}

	if info.InOperator != "" {
		op := util.ConvertS2I(info.InOperator) % 1000
		bsop := util.Short2Bytes_B(op)
		copy(buf[pos:], bsop)
	}
	pos += 2

	buf[pos] = (byte)(info.InBanci)
	pos += 1

	bsplate, err := util.Utf8ToGbk(([]byte)(info.VehPlate))
	if err == nil {
		copy(buf[pos:], bsplate)
	}
	pos += 12

	pos += 4

	fmt.Println(util.ConvertByte2Hexstring(buf, true))
	return buf
}

func TestReader() {
	var tmp0019 models.F0019Info

	tmp0019.InStationNetWork = "3201"
	tmp0019.InStation = "1500206"
	tmp0019.InLane = "6"
	tmp0019.InTime = "20010405112015"
	tmp0019.VehClass = 1
	tmp0019.FlowState = 4
	tmp0019.OutStationNetWork = "3201"
	tmp0019.OutStation = "1500206"
	tmp0019.InOperator = "1250108028"
	tmp0019.InBanci = 1
	tmp0019.VehPlate = "苏A12345"
	testPackage0019(tmp0019)

	brlt := gTestReader.InitReader(models.DnReader1, config.ConfigData["readerDnCom1"].(string), 1)
	if !brlt {
		return
	}
	util.FileLogs.Info("读卡器InitReader suc.")

	for {

		if gTestReader.pReaderInfObj.FuncJT_OpenCard() != 0 {
			util.MySleep_ms(10)
			continue
		}

		iCardtype := gTestReader.pReaderInfObj.FuncJT_GetCardType()

		//主动上报监测到卡片
		sCardid := ""
		switch iCardtype {
		case models.MifareS50, models.MifareS70:
			{
				//读卡号
				out := gTestReader.pReaderInfObj.FuncJT_GetCardSer()
				if out == nil {
					util.MySleep_ms(10)
					continue
				}

				sCardid = util.ConvertByte2Hexstring(out, false)
			}
		case models.MifarePro, models.MifareProX:
			{
				//读卡号
				rlt, out := gTestReader.pReaderInfObj.FuncJT_ProGetCardID()
				if rlt != 0 {
					util.MySleep_ms(10)
					continue
				}

				sCardid = util.ConvertByte2Hexstring(out, false)
			}
		}

		util.FileLogs.Info("读卡器open card suc:%d,%s.", iCardtype, sCardid)

		var f0015 models.ReaderETCReadReqData
		f0015.FileID = 0x0015
		f0015.Length = 50
		r1, r2 := gTestReader.pReaderInfObj.FuncJT_ProReadFile(f0015.FileID, f0015.Length)
		if r1 != 0 {
			util.FileLogs.Info("读卡器FuncJT_ProReadFile f0015 failed. rlt:%d.", r1)
			return
		}
		util.FileLogs.Info("读卡器FuncJT_ProReadFile f0015 suc:%s.", util.ConvertByte2Hexstring(r2, true))

		var f0019 models.ReaderETCReadReqData
		f0019.FileID = 0x0015
		f0019.Length = 50
		r3, r4 := gTestReader.pReaderInfObj.FuncJT_ProReadFile(f0019.FileID, f0019.Length)
		if r3 != 0 {
			util.FileLogs.Info("读卡器FuncJT_ProReadFile f0019 failed. rlt:%d.", r3)
			return
		}
		util.FileLogs.Info("读卡器FuncJT_ProReadFile f0019 suc:%s.", util.ConvertByte2Hexstring(r4, true))

		r5, r6 := gTestReader.pReaderInfObj.FuncJT_ProQueryBalance()
		if r5 != 0 {
			util.FileLogs.Info("读卡器FuncJT_ProQueryBalance failed. rlt:%d.", r5)
			return
		}
		util.FileLogs.Info("读卡器FuncJT_ProQueryBalance suc:%d.", r6)

		var fpay models.ReaderETCPayReqData
		fpay.Money = 1
		fpay.Data = make([]byte, 43)
		fpay.Data[13] = 0x04
		fpay.Paytime = util.GetNow(true)

		//copy(fpay.Data, f0019)
		r11, r12, r13, r14, r15 := gTestReader.pReaderInfObj.FuncJT_ProDecrement(fpay.Money, fpay.Data, fpay.Paytime)
		if r11 != 0 {
			util.FileLogs.Info("读卡器FuncJT_ProDecrement failed. rlt:%d.", r11)
			return
		}

		util.FileLogs.Info("读卡器FuncJT_ProDecrement suc.")
		fmt.Println(r12)
		fmt.Println(r13)
		fmt.Println(r14)
		fmt.Println(r15)
		break
	}

}

package tollmode

import (
	"FTC/config"
	"FTC/device"
	"FTC/models"
	"FTC/util"
)

func ETCSignalFunc() {
	GETCFrameListLock.Lock()
	device.PETCAntObj.ETCReadListLock.Lock()
	for e := device.PETCAntObj.ETCReadList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(models.ETCCommInfo)
		GETCFrameList.PushBack(ev)
	}

	device.PETCAntObj.ETCReadList.Init()
	device.PETCAntObj.ETCReadListLock.Unlock()
	GETCFrameListLock.Unlock()
}

//接收到etc报文，单独开协程处理
func goEtcFrameProc() {
	for {
		GETCFrameListLock.Lock()

		for e := GETCFrameList.Front(); e != nil; e = e.Next() {
			ev := e.Value.(models.ETCCommInfo)
			doEtcFrameProc(ev)
		}

		GETCFrameList.Init()
		GETCFrameListLock.Unlock()

		util.MySleep_ms(10)
	}
}

func EtcResponseNull(rsctl byte, key string, antno string) {
	//发送空应答
	rltbuf := util.ETCPackNull(rsctl)
	if device.PETCAntObj != nil {
		device.PETCAntObj.FuncGrpcEtcAntInfo(key, antno, rltbuf)
	}
}

func EtcResponseC1(rsctl byte, obuid []byte, key string, antno string) {
	//发送继续
	rltbuf := util.ETCPackC1(rsctl, obuid)
	if device.PETCAntObj != nil {
		device.PETCAntObj.FuncGrpcEtcAntInfo(key, antno, rltbuf)
	}
}

func EtcResponseC2(rsctl byte, obuid []byte, key string, antno string) {
	//发送停止
	rltbuf := util.ETCPackC2(rsctl, obuid, 1)
	if device.PETCAntObj != nil {
		device.PETCAntObj.FuncGrpcEtcAntInfo(key, antno, rltbuf)
	}
}

func doEtcFrameProc(info models.ETCCommInfo) {
	rsctl := util.ETCReversalRsctl(info.Msg[1])
	cmd := info.Msg[2]
	obuid := info.Msg[3:7]
	sobuid := util.ConvertByte2Hexstring(obuid, false)
	errcode := info.Msg[7]

	util.FileLogs.Info("开始处理ETC:%s,%02x", sobuid, cmd)
	switch cmd {
	case models.CMD_B2:
		//检查历史队列/添加到当前队列
		b, _ := chkETCInHistoryList(sobuid)
		if b {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		if errcode != models.CODE_OK {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		//添加到当前队列
		if EtcAddFrame(cmd, info) {
			EtcResponseC1(rsctl, obuid, info.Key, info.AntNo)
		} else {
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
		}

	case models.CMD_B3:
		if errcode != models.CODE_OK {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		//更新当前队列
		if EtcUpdateFrame(cmd, info) {
			EtcResponseC1(rsctl, obuid, info.Key, info.AntNo)
		} else {
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
		}

	case models.CMD_B4:
		if errcode != models.CODE_OK {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		//更新当前队列
		if EtcUpdateFrame(cmd, info) {
		} else {
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
		}

	case models.CMD_B5:
		if errcode != models.CODE_OK {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		//更新当前队列
		if EtcUpdateFrame(cmd, info) {
			//EtcResponseNull(rsctl, info.Key, info.AntNo)
		} else {
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
		}

	case models.CMD_B7:
		if errcode != models.CODE_OK {
			//发送停止
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
			return
		}

		//更新当前队列
		if EtcUpdateFrame(cmd, info) {
		} else {
			EtcResponseC2(rsctl, obuid, info.Key, info.AntNo)
		}
	}
}

//etc数据帧添加到当前队列
func EtcAddFrame(cmd byte, info models.ETCCommInfo) bool {
	frameb2, err := util.ETCParseB2(info.Msg)

	transdata := new(models.MTransData)
	transdata.StationMode = util.ConvertI2S(models.TOLLMODE_EXIT)

	transdata.ExitStationid = util.ConvertI2S(config.ConfigData["stationid"].(int))
	transdata.ExitLandid = util.ConvertI2S(config.ConfigData["laneid"].(int))
	transdata.ExitNetwork = util.ConvertI2S(config.ConfigData["network"].(int))
	transdata.ExitOperator = util.ConvertI2S(config.ConfigData["operator"].(int))

	transdata.Transtime, transdata.FlowNo = util.GetNowAndMs()
	transdata.ExitShiftdate = util.GetShiftDate(transdata.Transtime)
	transdata.ExitShift = util.GetShiftId(transdata.Transtime)

	transdata.CreateBy = models.CREATEBY_ETC
	transdata.ETCFlag = models.TRANSMATCH_YES
	transdata.TransFrom = models.TRANSFROM_ETC
	transdata.TransState = models.TRANSSTATE_INIT
	transdata.ETCAntGrpKey = info.Key
	transdata.ETCAntNo = info.AntNo
	transdata.ErrcodeB2 = models.FRAMERLT_INIT
	transdata.ErrcodeB2 = models.FRAMERLT_INIT
	transdata.ErrcodeB2 = models.FRAMERLT_INIT

	if err != nil {
		transdata.ErrcodeB2 = models.FRAMERLT_ERR
	} else {
		transdata.ErrcodeB2 = models.FRAMERLT_OK
		transdata.OBUID = frameb2.OBUID
		transdata.Frameb2 = frameb2
	}

	//遍历如果有，则覆盖；没有则追加
	GTransListLock.Lock()
	defer GTransListLock.Unlock()

	bfind := false
	for e := GTransList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(*models.MTransData)
		if ev == nil {
			continue
		}

		if ev.OBUID == frameb2.OBUID {
			if ev.ErrcodeB2 == models.FRAMERLT_OK {
				//抛弃该次，拒绝
				util.FileLogs.Info("EtcAddFrame B2帧已在当前交易中，抛弃")
				return false
			}

			ev = transdata
			bfind = true

			break
		}
	}

	if !bfind {
		//如果obu是空，则抛弃
		if ChkStrObuidZero(frameb2.OBUID) {
			util.FileLogs.Info("EtcAddFrame obuid is null ,abort.")
			return false
		}

		GTransList.PushBack(transdata)
	}

	isize := GTransList.Len()
	util.FileLogs.Info("%s 创建第%d笔交易:%s,%s", util.GetCreateByDes(transdata.CreateBy), isize, transdata.OBUID, transdata.Vehplate)

	return true
}

func ChkObuidZero(obuid []byte) bool {
	for i := 0; i < 4; i++ {
		if obuid[i] != 0x00 {
			return false
		}
	}

	return true
}

func ChkStrObuidZero(sobuid string) bool {
	if sobuid != "00000000" {
		return false
	}

	return true
}

func EtcUpdateFrame(cmd byte, info models.ETCCommInfo) bool {
	rsctl := util.ETCReversalRsctl(info.Msg[1])
	sobuid := util.ConvertByte2Hexstring(info.Msg[3:7], false)
	bret := false

	GTransListLock.Lock()
	defer GTransListLock.Unlock()

	for e := GTransList.Front(); e != nil; e = e.Next() {
		ev := e.Value.(*models.MTransData)
		if ev == nil {
			continue
		}

		if ev.OBUID == sobuid && ev.ETCAntGrpKey == info.Key && ev.ETCAntNo == info.AntNo {
			switch cmd {
			case models.CMD_B3:
				frameb3, err := util.ETCParseB3(info.Msg)
				if err != nil {
					ev.ErrcodeB3 = models.FRAMERLT_ERR
				} else {
					ev.ErrcodeB3 = models.FRAMERLT_OK
					bret = true
				}
				ev.Frameb3 = frameb3
				ev.Vehclass = util.ConvertI2S(frameb3.VehClass)
				ev.Vehplate = frameb3.VehPlate

			case models.CMD_B4:
				frameb4, err := util.ETCParseB4(info.Msg)
				ev.Frameb4 = frameb4

				if err != nil {
					ev.ErrcodeB4 = models.FRAMERLT_ERR
				} else {
					ev.ErrcodeB4 = models.FRAMERLT_OK
					ev.CardId = frameb4.F0015Info.CardId

					if ev.ErrcodeB2 == models.FRAMERLT_OK &&
						ev.ErrcodeB3 == models.FRAMERLT_OK &&
						ev.ErrcodeB4 == models.FRAMERLT_OK {
						bret = true
						EtcBusinessB4Proc(rsctl, ev)
					} else {
						bret = false
					}
				}

			case models.CMD_B5:
				frameb5, err := util.ETCParseB5(info.Msg)
				ev.Frameb5 = frameb5
				ev.Psamtermid = info.Psamtermid
				if err != nil {
					ev.TransState = models.TRANSFAIL_WRITECARD
					ev.TransFlow = models.TRANSFLOW_END
					return false
				}

				bret = true
				EtcBusinessB5Proc(rsctl, ev)

			case models.CMD_B7:
				frameb7, err := util.ETCParseB7(info.Msg)
				if err == nil {
					bret = true
				}
				ev.Frameb7 = frameb7
				EtcBusinessB7Proc(rsctl, ev)
			}

			return bret
		}
	}

	return false
}

func EtcBusinessB4Proc(rsctl byte, ev *models.MTransData) {
	/*
		//验发行日期
		//验发行商
		//验车型车牌
		if ev.Frameb3.VehClass != ev.Frameb4.F0015Info.VehClass ||
			ev.Frameb3.VehPlate != ev.Frameb4.F0015Info.VehPlate {
			ev.TransFlow = models.TRANSFLOW_END
			ev.TransState = models.TRANSFAIL_PLATE

			util.FileLogs.Info("%s,EtcBusinessB4Proc 车型(%s,%s)或车牌(%s,%s)不一致",
				ev.OBUID,
				ev.Frameb3.VehClass, ev.Frameb4.F0015Info.VehClass,
				ev.Frameb3.VehPlate, ev.Frameb4.F0015Info.VehPlate)
			return
		}
	*/

	switch config.ConfigData["tollmode"].(int) {
	case models.TOLLMODE_OPEN:
		EtcC6Proc(rsctl, ev)
	case models.TOLLMODE_ENTRY:
		EtcC3Proc(rsctl, ev)
	case models.TOLLMODE_FLAG:
		//
	case models.TOLLMODE_EXIT:
		EtcC6Proc(rsctl, ev)
	}
}

func EtcC3Proc(rsctl byte, ev *models.MTransData) bool {
	util.FileLogs.Info("%s,C3处理", ev.OBUID)

	//校验合法性，流通状态等，暂不考虑
	f0019 := ev.Frameb4.F0019Info
	f0019.InStationNetWork = ev.ExitNetwork
	f0019.InStation = ev.ExitStationid
	f0019.InLane = ev.ExitLandid
	f0019.InTime = ev.Transtime
	f0019.VehClass = ev.Frameb3.VehClass
	f0019.FlowState = models.ENTRY_FLOWETC
	f0019.FlagSta = "0"
	f0019.AuxSta = "0"
	f0019.OutStation = "0"
	f0019.OutStationNetWork = "0"
	f0019.InOperator = ev.ExitOperator
	f0019.InBanci = util.ConvertS2I(ev.ExitShift)
	f0019.VehPlate = ev.Frameb3.VehPlate

	var info models.ETCFrameC3
	info.OBUID = ev.OBUID
	info.F0019Info = f0019
	info.Datetime = util.ConvertTime_2(ev.Transtime)
	sendbuf := util.ETCPackC3(rsctl, info)
	if device.PETCAntObj != nil {
		device.PETCAntObj.FuncGrpcEtcAntInfo(ev.ETCAntGrpKey, ev.ETCAntNo, sendbuf)
	}
	return true
}

func EtcC6Proc(rsctl byte, ev *models.MTransData) bool {
	util.FileLogs.Info("%s,C6处理", ev.OBUID)

	//先计费，再发扣款指令
	
		if !FuncFEE(ev) {
			ev.TransFlow = models.TRANSFLOW_END
			ev.TransState = models.TRANSFAIL_FEECALC
			return false
		}
	

	f0019 := ev.Frameb4.F0019Info
	f0019.FlowState = models.EXIT_FLOWETC
	f0019.OutStationNetWork = ev.ExitNetwork
	f0019.OutStation = ev.ExitStationid

	var info models.ETCFrameC6
	info.OBUID = ev.OBUID
	info.ConsumeMoney = util.ConvertS2I(ev.Toll)
	info.F0019Info = f0019
	info.Datetime = util.ConvertTime_2(ev.Transtime)
	sendbuf := util.ETCPackC6(rsctl, info)
	if device.PETCAntObj != nil {
		device.PETCAntObj.FuncGrpcEtcAntInfo(ev.ETCAntGrpKey, ev.ETCAntNo, sendbuf)
	}
	return true
}

func EtcBusinessB5Proc(rsctl byte, ev *models.MTransData) {
	_, strms := util.GetNowAndMs()
	util.FileLogs.Info("%s,收到成功B5帧，耗时:%d (ms)", ev.OBUID, util.Diffms(ev.FlowNo, strms))

	ev.Tac = ev.Frameb5.Tac
	ev.Cardtradno = ev.Frameb5.EtcTradNo
	ev.Psamtradno = ev.Frameb5.PsamTransNo

	ev.TransFlow = models.TRANSFLOW_END
	ev.TransState = models.TRANSSTATE_SUC
	ev.Paytype = models.PAYTYPE_ETC
	ev.Paytime = util.GetNow(false)
	ev.PayFinishtime = util.GetNow(false)
}

func EtcBusinessB7Proc(rsctl byte, ev *models.MTransData) {
	util.FileLogs.Info("%s,B7处理", ev.OBUID)
}

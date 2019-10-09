package main

/*
#cgo CFLAGS: -I./
#cgo LDFLAGS: -L./ -lMyEP
#include <stdio.h>
#include <stdlib.h>
#include "myep.h"
void ReceivedMsgCallback_cgo(const char * jsonMsg);
*/
import "C"
import (
	"FTC/config"
	"FTC/device"
	"FTC/pb"
	"FTC/util"
	"container/list"
	"encoding/json"
	"sync"
	"unsafe"

	"FTC/models"
)

type EPCacheInfo struct {
	TID    string
	Lasttm int64
}

var gCacheEPReadList *list.List

func chkEpMsg(epinfo models.EPInfo) bool {
	nowtm := util.GetTimeStampSec()
	for e1 := gCacheEPReadList.Front(); e1 != nil; {
		ev := e1.Value.(*EPCacheInfo)
		if nowtm > ev.Lasttm+10 {
			en := e1
			e1 = e1.Next()
			gCacheEPReadList.Remove(en)
		} else {
			e1 = e1.Next()
		}
	}

	if epinfo.MsgType == 300 {
		return true
	}

	if epinfo.MsgType == 500 {
		var epdata models.EPTagReportData
		tmpbuf, err := json.Marshal(epinfo.MessageValue)
		if err != nil {
			util.MyPrintf("chkEpMsg MessageValue解析失败")
			return false
		}

		err = json.Unmarshal(tmpbuf, &epdata)
		if err != nil {
			util.MyPrintf("chkEpMsg EPTagReportData解析失败")
			return false
		}

		if epdata.ReportRsds == nil {
			util.MyPrintf("chkEpMsg ReportRsds is nil")
			return false
		}

		isize := len(epdata.ReportRsds)
		//var tmpSelectFTCcResult models.EPSelectFTCcResult
		if isize > 0 {
			for _, v := range epdata.ReportRsds {
				/*
					if v.SelectFTCcResult == tmpSelectFTCcResult {
						util.FileLogs.Info("标签选择规则结果,标签选择规则ID:%d，忽略", v.SelectFTCcID)
						continue
					}
				*/

				var unit models.EPResultReadInfo
				unit.AntennaID = v.AntennaID
				unit.TID = v.TID
				//unit.ReadDataInfo = v.SelectFTCcResult.CustomizedSelectFTCcResult.ReadDataInfo

				bfind := false
				tmpinfo := new(EPCacheInfo)
				tmpinfo.TID = unit.TID
				tmpinfo.Lasttm = util.GetTimeStampSec()
				for e := gCacheEPReadList.Front(); e != nil; e = e.Next() {
					ev := e.Value.(*EPCacheInfo)
					if ev.TID == tmpinfo.TID {
						//ev.Lasttm = tmpinfo.Lasttm
						bfind = true

						util.FileLogs.Info("%s 已检测到，忽略", ev.TID)
						break
					}
				}

				if !bfind {
					gCacheEPReadList.PushBack(tmpinfo)
				}

				return !bfind
			}
		}
	}

	return false
}

func chkEpMultiMsg(epinfo models.EPInfo) bool {
	nowtm := util.GetTimeStampSec()
	for e1 := gCacheEPReadList.Front(); e1 != nil; {
		ev := e1.Value.(*EPCacheInfo)
		if nowtm > ev.Lasttm+10 {
			en := e1
			e1 = e1.Next()
			gCacheEPReadList.Remove(en)
		} else {
			e1 = e1.Next()
		}
	}

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
		//var tmp1 models.EPMultiHbCustomizedReadFTCcResult
		if isize > 0 {
			for _, v := range epdata.ReportRsds {
				/*for _, v1 := range v.AccessFTCcResult {

				if v1.HbCustomizedReadFTCcResult == tmp1 {
					util.FileLogs.Info("chkEpMultiMsg 标签选择规则结果为空，抛弃")
					continue
				}
				*/
				var unit models.EPResultReadInfo
				unit.AntennaID = v.AntennaID
				unit.TID = v.TID
				//unit.ReadDataInfo = v1.HbCustomizedReadFTCcResult.ReadDataInfo

				bfind := false
				tmpinfo := new(EPCacheInfo)
				tmpinfo.TID = unit.TID
				tmpinfo.Lasttm = util.GetTimeStampSec()
				for e := gCacheEPReadList.Front(); e != nil; e = e.Next() {
					ev := e.Value.(*EPCacheInfo)
					if ev.TID == tmpinfo.TID {
						ev.Lasttm = tmpinfo.Lasttm
						bfind = true

						util.FileLogs.Info("chkEpMultiMsg %s 已检测到，忽略", ev.TID)
						break
					}
				}

				if !bfind {
					util.FileLogs.Info("ReceivedMsgCallback ADD")
					gCacheEPReadList.PushBack(tmpinfo)
					rlt = true
				}

				//}
			}

			return rlt
		}
	}

	return false
}

//export ReceivedMsgCallback
func ReceivedMsgCallback(jsonMsg *C.char) {
	jsonstr := C.GoString(jsonMsg)

	var epinfo models.EPInfo
	err := json.Unmarshal([]byte(jsonstr), &epinfo)
	if err != nil {
		util.MyPrintf("ReceivedMsgCallback EPInfo 解析失败")
		return
	}

	if epinfo.MsgType == 300 {
		gEPSrv.SendHeartBeat()
	}

	if epinfo.MsgType == 500 {
		//util.FileLogs.Info("ReceivedMsgCallback:", jsonstr)
	}

	if gEPSrv.IsGrpcConn() {
		m := &pb.EPAntReadRequest{Msg: jsonstr}

		//由于读卡太频繁，此处过滤下，连续10s内读到同一张卡，则抛弃，不处理
		if models.GDebugant {
			if chkEpMsg(epinfo) {
				gEPSrv.GrpcSendproc(models.GRPCTYPE_EPAntRealRead, "0", m)
			}
		} else {
			if chkEpMultiMsg(epinfo) {
				gEPSrv.GrpcSendproc(models.GRPCTYPE_EPAntRealRead, "0", m)
			}
		}

	} else {
		util.FileLogs.Info("ReceivedMsgCallback:grpc 连接断开，抛弃")
	}

	return
}

//电子车牌服务
type EPService struct {
	device.ConnBaseCli //连接电子天线
	device.GrpcClient  //连接工控机

	AntState  int
	LastQueTm int64
	AntLock   *sync.RWMutex
}

func (p *EPService) InitEP() {
	p.AntState = models.DEVSTATE_UNKNOWN
	p.LastQueTm = 0
	p.AntLock = new(sync.RWMutex)

	//连接EP天线
	C.Loadso()
	C.RegisterCallbackFunction((C.pReceivedMsgCallback_t)(unsafe.Pointer(C.ReceivedMsgCallback_cgo)))
	p.InitConn(models.DEVTYPE_ANTEP, config.ConfigData["epAntSrvUrl"].(string), p.Recvproc)
	go p.goAutoState()

	/////////////////////////////////////grpc client
	p.GrpcInit2(models.DEVGRPCTYPE_EP, config.ConfigData["epAntGrpcUrlCli"].(string), p.FuncEPGRPCInit, p.FuncEPGrpcProc)

}

func (p *EPService) goAutoState() {
	key := config.ConfigData["epAntKey"].(string)

	for {
		now := util.GetTimeStampMs()
		if !p.IsConn() {
			p.LastQueTm = now
			util.MySleep_s(5)
			continue
		}

		if now > p.LastQueTm && now > p.LastQueTm+20*1000 {
			p.LastQueTm = now
			util.FileLogs.Info("epant timeout disconn")
			p.DisConnect()
		}

		m := &pb.EPAntStateRequest{Antkey: key,
			State: util.ConvertI2S(p.AntState)}
		p.GrpcSendproc(models.GRPCTYPE_EPAntState, "0", m)

		util.MySleep_s(5)
	}
}

func (p *EPService) Recvproc(inbuf []byte) (int, bool) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, true
	}

	//fmt.Println("ep recv................................")
	//fmt.Println(util.ConvertByte2Hexstring(inbuf, true))

	p.LastQueTm = util.GetTimeStampMs()

	p.AntLock.Lock()
	C.ReceiveMessages((*C.uchar)(&inbuf[0]), C.int(inlen))
	p.AntLock.Unlock()

	return inlen, true
}

func (p *EPService) SendHeartBeat() {
	var clen C.int
	var pbuf *C.uchar
	C.GenerateKeepaliveAckData(&pbuf, &clen)

	buflen := int(clen)
	buf := C.GoBytes(unsafe.Pointer(pbuf), clen)
	p.SendProc(buf[0:buflen], false)
}

func (p *EPService) FuncEPGRPCInit() {
	if !p.IsGrpcConn() {
		return
	}

	key := config.ConfigData["epAntKey"].(string)
	m := &pb.EPAntInitRequest{Antkey: key}
	p.GrpcSendproc(models.GRPCTYPE_EPAntInit, "0", m)
}

func (p *EPService) FuncEPGrpcProc(msg *pb.Message) {
	sType := msg.Type
	//inbuf := msg.Data
	//sno := msg.No

	//util.FileLogs.Info("FuncEPGrpcProc %s-收到GRPC服务应答 开始处理.\r\n", device.GetCmdDes(sType))
	switch sType {
	case models.GRPCTYPE_EPAntInit:
		break

	case models.GRPCTYPE_EPAntRealRead:
		break

	case models.GRPCTYPE_EPAntState:
		break
	}

	return
}

var gEPSrv EPService

func main() {
	util.FileLogs.Info("电子车牌服务启动中...")
	config.InitConfig("../../conf/config.conf")
	gCacheEPReadList = list.New()
	gEPSrv.InitEP()

	for {
		util.MySleep_s(5)
	}
}

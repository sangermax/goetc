package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//天线设备
type DevEPAnt struct {
	PGrpcMgr *GrpcManage

	//队列
	EPReadListLock *sync.Mutex
	EPReadList     *list.List
}

func (p *DevEPAnt) InitDevAnt() {
	p.EPReadListLock = new(sync.Mutex)
	p.EPReadList = list.New()

	p.PGrpcMgr = new(GrpcManage)
	p.PGrpcMgr.GrpcManageInit(models.DEVGRPCTYPE_EP)
	go p.goListen()
}

//目前暂时按1个天线计，先取第一个
func (p *DevEPAnt) FuncGetEPAntState() int {
	if p.PGrpcMgr == nil {
		return models.DEVSTATE_UNKNOWN
	}

	e := p.PGrpcMgr.GrpcMap.GetFirst()
	if e == nil {
		return models.DEVSTATE_TROUBLE
	}

	return e.(*GrpcConnInfo).State
}

func (p *DevEPAnt) goListen() {

	anturl := config.ConfigData["epAntGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", anturl)
	if err != nil {
		util.ConsoleLogs.Info("EP天线服务监听grpc端口失败:%s", anturl)
		os.Exit(-1)
		return
	}
	util.ConsoleLogs.Info("EP天线服务监听grpc端口成功:%s", anturl)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &EPServer{})
	s.Serve(lis)

}

type EPServer struct {
}

func (p *EPServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	key := ""
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			util.FileLogs.Info("Communite read done")

			if PEPAntObj != nil && PEPAntObj.PGrpcMgr != nil {
				PEPAntObj.PGrpcMgr.RemoveGrpcConn(key)
			}
			return nil
		}

		if err != nil {
			util.FileLogs.Info("Communite ERR", err)

			if PEPAntObj != nil && PEPAntObj.PGrpcMgr != nil {
				PEPAntObj.PGrpcMgr.RemoveGrpcConn(key)
			}
			//断开连接处理
			return err
		}

		if in.Type != models.GRPCTYPE_EPAntState {
			//util.FileLogs.Info("收到来自EP天线的信息")
			//fmt.Println(in)
		}

		switch in.Type {
		case models.GRPCTYPE_EPAntInit:
			key, _ = FuncGrpcEpAntInitProc(stream, in)
			break

		case models.GRPCTYPE_EPAntRealRead:
			FuncGrpcEpAntReadProc(stream, in)
			break

		case models.GRPCTYPE_EPAntState:
			FuncGrpcEpAntStateProc(stream, in)
			break
		}

	}

}

func GrpcSrvSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("EPANT GrpcSendproc")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

func FuncGrpcEpAntReadProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) {
	sType := in.Type
	no := in.No
	inbuf := in.Data

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_FAIL
	result.ResultDes = "json格式错误"

	m := &pb.EPAntReadResponse{}

	req := &pb.EPAntReadRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEpAntReadProc err:%s.\r\n", err.Error())
		GrpcSrvSendproc(stream, sType, no, m, result)
		return
	}

	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""
	GrpcSrvSendproc(stream, sType, no, m, result)

	//util.FileLogs.Info("EPANT 读到电子车牌信息:%s.\r\n", req.Msg)

	var epinfo models.EPInfo
	err := json.Unmarshal([]byte(req.Msg), &epinfo)
	if err != nil {
		util.MyPrintf("FuncGrpcEpAntReadProc EPInfo 解析失败")
		return
	}

	if epinfo.MsgType != 500 {
		//util.MyPrintf("FuncGrpcEpAntReadProc 非车辆类型信息，忽略.:%d", epinfo.MsgType)
		return
	}

	//var epdata models.EPTagReportData
	var epdata models.EPMessageValue500
	tmpbuf, err := json.Marshal(epinfo.MessageValue)
	if err != nil {
		util.MyPrintf("FuncGrpcEpAntReadProc MessageValue解析失败")
		return
	}

	err = json.Unmarshal(tmpbuf, &epdata)
	if err != nil {
		util.MyPrintf("FuncGrpcEpAntReadProc EPTagReportData解析失败")
		return
	}

	if epdata.ReportRsds == nil {
		util.MyPrintf("FuncGrpcEpAntReadProc ReportRsds is nil")
		return
	}

	//var tmp1 models.EPMultiHbCustomizedReadFTCcResult
	isize := len(epdata.ReportRsds)
	if isize > 0 {

		PEPAntObj.EPReadListLock.Lock()
		if models.GDebugant {
			for _, v := range epdata.ReportRsds {
				var unit models.EPResultReadInfo
				unit.AntennaID = v.AntennaID
				unit.TID = v.TID
				//unit.ReadDataInfo = v.SelectFTCcResult.CustomizedSelectFTCcResult.ReadDataInfo

				if PEPAntObj != nil {
					PEPAntObj.EPReadList.PushBack(unit)
				}

				fmt.Println(unit)

			}

		} else {
			for _, v := range epdata.ReportRsds {
				/*
					for _, v1 := range v.AccessFTCcResult {
						if v1.HbCustomizedReadFTCcResult == tmp1 {
							continue
						}
				*/

				var unit models.EPResultReadInfo
				unit.AntennaID = v.AntennaID
				unit.TID = v.TID
				//unit.ReadDataInfo = v1.HbCustomizedReadFTCcResult.ReadDataInfo
				if PEPAntObj != nil {
					PEPAntObj.EPReadList.PushBack(unit)
				}

				fmt.Println(unit)
				//}
			}
		}

		PEPAntObj.EPReadListLock.Unlock()

		ChanEP <- true
	}

}

func FuncGrpcEpAntInitProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) (string, error) {
	sType := in.Type
	no := in.No
	inbuf := in.Data

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_FAIL
	result.ResultDes = "json格式错误"

	m := &pb.EPAntInitResponse{}

	req := &pb.EPAntInitRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEpAntInitProc err:%s.\r\n", err.Error())
		GrpcSrvSendproc(stream, sType, no, m, result)
		return "", err
	}

	util.FileLogs.Info("EPANT 电子车牌初始化.\r\n")

	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""
	GrpcSrvSendproc(stream, sType, no, m, result)

	if PEPAntObj != nil && PEPAntObj.PGrpcMgr != nil {
		PEPAntObj.PGrpcMgr.AddGrpcConn(req.Antkey, stream)
		return req.Antkey, nil
	}

	return req.Antkey, nil
}

func FuncGrpcEpAntStateProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) {
	inbuf := in.Data

	req := &pb.EPAntStateRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEpAntStateProc err:%s.\r\n", err.Error())
		return
	}

	if PEPAntObj != nil && PEPAntObj.PGrpcMgr != nil {
		PEPAntObj.PGrpcMgr.UpdateDevState(req.Antkey, req.State)
	}

	//util.FileLogs.Info("EPANT 电子车牌状态:%s,%s.\r\n", req.Antkey, req.State)
}

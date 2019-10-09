package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"container/list"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//ETC天线设备
type DevETCAnt struct {
	PGrpcMgr *GrpcManage

	//队列
	ETCReadListLock *sync.Mutex
	ETCReadList     *list.List
}

func (p *DevETCAnt) InitDevAnt() {
	p.PGrpcMgr = new(GrpcManage)
	p.ETCReadListLock = new(sync.Mutex)
	p.ETCReadList = list.New()

	p.PGrpcMgr.GrpcManageInit(models.DEVGRPCTYPE_ETC)
	go p.goListen()
}

//目前暂时按1个天线计，先取第一个
func (p *DevETCAnt) FuncGetEtcAntState() int {
	if p.PGrpcMgr == nil {
		return models.DEVSTATE_UNKNOWN
	}

	e := p.PGrpcMgr.GrpcMap.GetFirst()
	if e == nil {
		return models.DEVSTATE_TROUBLE
	}

	return models.DEVSTATE_TROUBLE
	return e.(*GrpcConnInfo).State
}

func (p *DevETCAnt) goListen() {
	anturl := config.ConfigData["etcGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", anturl)
	if err != nil {
		util.ConsoleLogs.Info("ETC天线服务监听grpc端口失败:%s", anturl)
		os.Exit(-1)
		return
	}
	util.ConsoleLogs.Info("ETC天线服务监听grpc端口成功:%s", anturl)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &ETCServer{})
	s.Serve(lis)
}

func (p *DevETCAnt) FuncGrpcEtcAntInfo(key, antno string, msg []byte) {
	sType := models.GRPCTYPE_ETCMSG
	util.FileLogs.Info("FuncGrpcEtcAntInfo %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	var info models.ETCCommInfo
	info.Key = key
	info.AntNo = antno
	info.Msg = make([]byte, len(msg))
	copy(info.Msg, msg)

	if p.PGrpcMgr != nil {
		stream := p.PGrpcMgr.GetGrpcConnStream(info.Key)

		var result models.ResultInfo
		result.ResultValue = models.GRPCRESULT_OK
		result.ResultDes = ""

		m := &pb.ETCAntReadResponse{Key: info.Key,
			AntNo: info.AntNo,
			Msg:   util.ConvertByte2Hexstring(info.Msg, false)}

		ETCGrpcSrvSendproc(stream, sType, "0", m, result)

	}
}

type ETCServer struct {
}

func (p *ETCServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	key := ""
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			util.FileLogs.Info("ETCServer Communite read done")

			if PETCAntObj != nil && PETCAntObj.PGrpcMgr != nil {
				PETCAntObj.PGrpcMgr.RemoveGrpcConn(key)
			}
			return nil
		}

		if err != nil {
			util.FileLogs.Info("ETCServer Communite ERR", err)

			if PETCAntObj != nil && PETCAntObj.PGrpcMgr != nil {
				PETCAntObj.PGrpcMgr.RemoveGrpcConn(key)
			}
			//断开连接处理
			return err
		}

		if in.Type != models.GRPCTYPE_EPAntState {
			util.FileLogs.Info("收到来自ETC天线的信息")
			fmt.Println(in)
		}

		switch in.Type {
		case models.GRPCTYPE_ETCAntInit:
			key, _ = FuncGrpcEtcAntInitProc(stream, in)
			break

		case models.GRPCTYPE_ETCMSG:
			FuncGrpcEtcAntReadProc(stream, in)
			break

		case models.GRPCTYPE_ETCAntState:
			FuncGrpcEtcAntStateProc(stream, in)
			break
		}

	}

}

func ETCGrpcSrvSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("ETCGrpcSrvSendproc error:%s.\r\n", err.Error())
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

func FuncGrpcEtcAntReadProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) {
	sType := in.Type
	no := in.No
	inbuf := in.Data

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_FAIL
	result.ResultDes = "json格式错误"

	m := &pb.ETCAntReadResponse{}
	req := &pb.ETCAntReadRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEtcAntReadProc err:%s.\r\n", err.Error())
		ETCGrpcSrvSendproc(stream, sType, no, m, result)
		return
	}

	//主业务处理
	var info models.ETCCommInfo
	info.Key = req.Key
	info.AntNo = req.AntNo
	info.Psamtermid = req.PsamTermid
	b := util.ConvertHexstring2Byte(req.Msg)
	info.Msg = make([]byte, len(b))
	copy(info.Msg, b)

	PETCAntObj.ETCReadListLock.Lock()
	PETCAntObj.ETCReadList.PushBack(info)
	PETCAntObj.ETCReadListLock.Unlock()
	ChanETC <- true
}

func FuncGrpcEtcAntInitProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) (string, error) {
	sType := in.Type
	no := in.No
	inbuf := in.Data

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_FAIL
	result.ResultDes = "json格式错误"

	m := &pb.ETCAntInitResponse{}

	req := &pb.ETCAntInitRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEtcAntInitProc err:%s.\r\n", err.Error())
		ETCGrpcSrvSendproc(stream, sType, no, m, result)
		return "", err
	}

	util.FileLogs.Info("ETC天线初始化.\r\n")

	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""
	ETCGrpcSrvSendproc(stream, sType, no, m, result)

	if PETCAntObj != nil && PETCAntObj.PGrpcMgr != nil {
		PETCAntObj.PGrpcMgr.AddGrpcConn(req.Key, stream)
		return req.Key, nil
	}

	return req.Key, nil
}

func FuncGrpcEtcAntStateProc(stream pb.GrpcMsg_CommuniteServer, in *pb.Message) {
	inbuf := in.Data

	req := &pb.ETCAntStateReport{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncGrpcEtcAntStateProc err:%s.\r\n", err.Error())
		return
	}

	if PETCAntObj != nil && PETCAntObj.PGrpcMgr != nil {
		//PETCAntObj.PGrpcMgr.UpdateDevState(req.Key, req.State)
	}
}

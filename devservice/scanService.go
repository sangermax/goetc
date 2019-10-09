package main

import (
	"FTC/config"
	"FTC/device"
	"FTC/pb"
	"FTC/util"
	"errors"
	"fmt"
	"io"
	"net"

	"FTC/models"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//扫码服务
type ScanService struct {
	device.SerialBase

	State int
}

//协议
func (p *ScanService) enpack(cmd int) ([]byte, error) {
	return nil, nil
}

func (p *ScanService) depack(inbuf []byte) ([]byte, int, error) {
	//找到ODOA ,即认为
	length := len(inbuf)
	baselen := 2

	if length < baselen {
		return nil, 0, errors.New("还未收完")
	}

	i := 0
	for {
		endpos := 0
		for i = 0; i < length-1; i = i + 1 {
			if inbuf[i] == 0x0D && inbuf[i+1] == 0x0A {
				endpos = i
				break
			}
		}

		if i >= length-1 {
			return nil, 0, errors.New("还未收完2")
		}

		return inbuf[0:endpos], length, nil
	}

}

func (p *ScanService) InitScan() {
	p.State = models.DEVSTATE_UNKNOWN
}

func (p *ScanService) FuncSend() bool {
	buf := make([]byte, models.MAX_BUFFERSIZE)
	pos := 0

	return p.SendProc(buf[0:pos], false)
}

func (p *ScanService) Recvproc(inbuf []byte) (int, error) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, errors.New("长度不足")
	}

	outbuf, offset, err := p.depack(inbuf[0:inlen])
	if err != nil {
		return offset, err
	}

	//扫码返回
	p.FuncRstBars(outbuf)

	return offset, nil
}

func (p *ScanService) FuncRstBars(inbuf []byte) {
	util.FileLogs.Info("FuncRstBars:扫码返回处理,%s", string(inbuf))

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""

	m := &pb.OpenScanResponse{Scanbar: string(inbuf)}
	if g_scanstream != nil {
		ScanSrvGrpcSendproc(g_scanstream, models.GRPCTYPE_OPENSCAN, "0", m, result)
	}
}

var gScanUpSrv ScanService
var gScanDnSrv ScanService
var g_scanstream pb.GrpcMsg_CommuniteServer

func main() {
	util.FileLogs.Info("扫码服务启动中...")
	config.InitConfig("../conf/config.conf")

	r1 := gScanUpSrv.InitSerial(models.DEVTYPE_SCANUP, config.ConfigData["scanUpCom"].(string), 9600, gScanUpSrv.Recvproc)
	gScanUpSrv.InitScan()

	r2 := gScanDnSrv.InitSerial(models.DEVTYPE_SCANDN, config.ConfigData["scanDnCom"].(string), 9600, gScanDnSrv.Recvproc)
	gScanDnSrv.InitScan()

	//由于没有心跳之类，只要串口打开成功，则认为设备ok
	if r1 {
		gScanUpSrv.State = models.DEVSTATE_OK
	} else {
		gScanUpSrv.State = models.DEVSTATE_TROUBLE
	}

	if r2 {
		gScanDnSrv.State = models.DEVSTATE_OK
	} else {
		gScanDnSrv.State = models.DEVSTATE_TROUBLE
	}

	/////////////////////////////////////grpc
	sAddr := config.ConfigData["scanGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("扫码服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("扫码服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &ScanServer{})
	s.Serve(lis)

}

type ScanServer struct {
}

func (p *ScanServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道扫码GRPC已连接")
	bConn := true
	g_scanstream = stream

	go func() {
		for {
			if !bConn {
				break
			}

			//发送设备状态
			sUpState := util.ConvertI2S(gScanUpSrv.State)
			sDnState := util.ConvertI2S(gScanDnSrv.State)
			m := &pb.PrinterStateReport{UpState: sUpState, DnState: sDnState}
			ScanSrvGrpcSendproc(stream, models.GRPCTYPE_SCANSTATE, "0", m, models.ResultInfo{})

			util.MySleep_s(5)
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			util.FileLogs.Info("Communite read done")
			bConn = false
			break
		}

		if err != nil {
			util.FileLogs.Info("Communite ERR:%s", err.Error())
			bConn = false
			break
		}

		util.FileLogs.Info("收到来自车道扫码信息")
		fmt.Println(in)

		switch in.Type {
		case models.GRPCTYPE_OPENSCAN:
			FuncOpenScanProc(in.Data)

		case models.GRPCTYPE_CLOSESCAN:
			FuncCloseScanProc(in.Data)
		}

	}

	return nil
}

func ScanSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("scan GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("scan GrpcSendproc ")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("scan GrpcSendproc")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

func FuncOpenScanProc(inbuf []byte) {

}

func FuncCloseScanProc(inbuf []byte) {

}

package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"errors"

	"github.com/golang/protobuf/proto"
)

//IO信号
type DevIO struct {
	GrpcClient

	StateCoil1 int
	StateCoil2 int
	StateCoil3 int
	StateCoil4 int
	StateCoil5 int
	StateCoil6 int
	StateCoil7 int
	StateCoil8 int
}

func (p *DevIO) InitDevIO() {
	p.StateCoil1 = models.COIL_CLOSE
	p.StateCoil2 = models.COIL_CLOSE
	p.StateCoil3 = models.COIL_CLOSE
	p.StateCoil4 = models.COIL_CLOSE
	p.StateCoil5 = models.COIL_CLOSE
	p.StateCoil6 = models.COIL_CLOSE
	p.StateCoil7 = models.COIL_CLOSE
	p.StateCoil8 = models.COIL_CLOSE

	p.GrpcInit(models.DEVGRPCTYPE_IO, config.ConfigData["ioGrpcUrlCli"].(string), p.FuncGrpcIOProc)
}

func (p *DevIO) AllDevIOWithout() bool {
	if p.StateCoil1 == models.COIL_CLOSE &&
		p.StateCoil2 == models.COIL_CLOSE &&
		p.StateCoil3 == models.COIL_CLOSE &&
		p.StateCoil4 == models.COIL_CLOSE &&
		p.StateCoil5 == models.COIL_CLOSE &&
		p.StateCoil6 == models.COIL_CLOSE &&
		p.StateCoil7 == models.COIL_CLOSE &&
		p.StateCoil8 == models.COIL_CLOSE {
		return true
	}

	return false
}

func (p *DevIO) FuncGrpcIOProc(msg *pb.Message) {
	sType := msg.Type

	if sType != models.GRPCTYPE_IOSTATE {
		util.FileLogs.Info("%s-收到GRPC应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_IOSTATE:
		p.FuncIOState(msg)
	case models.GRPCTYPE_IOCONTROL:
		p.FuncRstIOControl(msg)
	}

	return
}

func (p *DevIO) FuncIOState(msg *pb.Message) {
	//sType := models.GRPCTYPE_IOSTATE
	//util.FileLogs.Info("%s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.IOStateReport{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	if p.StateCoil1 != util.ConvertS2I(recvmsg.State1) {
		util.FileLogs.Info("IOState1 (b-a):(%d->%s).\r\n", p.StateCoil1, recvmsg.State1)
		p.StateCoil1 = util.ConvertS2I(recvmsg.State1)

		if p.StateCoil1 == 1 {
			ChanVehIn <- true
		}
	}

	if p.StateCoil2 != util.ConvertS2I(recvmsg.State2) {
		util.FileLogs.Info("IOState2 (b-a):(%d->%s).\r\n", p.StateCoil2, recvmsg.State2)
		p.StateCoil2 = util.ConvertS2I(recvmsg.State2)

		//落杆，车检信号从有到无
		if p.StateCoil2 == 0 {
			ChanVehOut <- true
		}
	}

	if p.StateCoil3 != util.ConvertS2I(recvmsg.State3) {
		util.FileLogs.Info("IOState3 (b-a):(%d->%s).\r\n", p.StateCoil3, recvmsg.State3)
		p.StateCoil3 = util.ConvertS2I(recvmsg.State3)
	}

	if p.StateCoil4 != util.ConvertS2I(recvmsg.State4) {
		util.FileLogs.Info("IOState4 (b-a):(%d->%s).\r\n", p.StateCoil4, recvmsg.State4)
		p.StateCoil4 = util.ConvertS2I(recvmsg.State4)
	}

	if p.StateCoil5 != util.ConvertS2I(recvmsg.State5) {
		util.FileLogs.Info("IOState5 (b-a):(%d->%s).\r\n", p.StateCoil5, recvmsg.State5)
		p.StateCoil5 = util.ConvertS2I(recvmsg.State5)
	}

	if p.StateCoil6 != util.ConvertS2I(recvmsg.State6) {
		util.FileLogs.Info("IOState6 (b-a):(%d->%s).\r\n", p.StateCoil6, recvmsg.State6)
		p.StateCoil6 = util.ConvertS2I(recvmsg.State6)
	}

	if p.StateCoil7 != util.ConvertS2I(recvmsg.State7) {
		util.FileLogs.Info("IOState7 (b-a):(%d->%s).\r\n", p.StateCoil7, recvmsg.State7)
		p.StateCoil7 = util.ConvertS2I(recvmsg.State7)
	}

	if p.StateCoil8 != util.ConvertS2I(recvmsg.State8) {
		util.FileLogs.Info("IOState8 (b-a):(%d->%s).\r\n", p.StateCoil8, recvmsg.State8)
		p.StateCoil8 = util.ConvertS2I(recvmsg.State8)
	}
}

func (p *DevIO) FuncLanganProc(val int) {
	var req models.ReqIOControlData

	req.IONo = "0"
	req.IOControl = util.ConvertI2S(val)
	p.FuncReqIOControl(req)
}

func (p *DevIO) FuncReqIOControl(req models.ReqIOControlData) error {
	sType := models.GRPCTYPE_IOCONTROL
	util.FileLogs.Info("FuncReqIOControl %s-GRPC车道请求:(%s-%s).\r\n", GetCmdDes(sType), req.IONo, req.IOControl)

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	m := &pb.IOControlRequest{IONo: req.IONo,
		IOControl: req.IOControl}

	return p.GrpcSendproc(sType, "0", m)
}

func (p *DevIO) FuncRstIOControl(msg *pb.Message) {
	sType := models.GRPCTYPE_IOCONTROL
	util.FileLogs.Info("FuncRstIOControl %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.IOControlResponse{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

}

func (p *DevIO) GetCoilState(i int) bool {
	switch i {
	case 1:
		if p.StateCoil1 == 1 {
			return true
		} else {
			return false
		}
	case 2:
		if p.StateCoil2 == 1 {
			return true
		} else {
			return false
		}

	case 3:
		if p.StateCoil3 == 1 {
			return true
		} else {
			return false
		}

	case 4:
		if p.StateCoil4 == 1 {
			return true
		} else {
			return false
		}

	case 5:
		if p.StateCoil5 == 1 {
			return true
		} else {
			return false
		}

	case 6:
		if p.StateCoil6 == 1 {
			return true
		} else {
			return false
		}

	case 7:
		if p.StateCoil7 == 1 {
			return true
		} else {
			return false
		}

	case 8:
		if p.StateCoil8 == 1 {
			return true
		} else {
			return false
		}
	}

	return false
}

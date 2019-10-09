package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"errors"

	"github.com/golang/protobuf/proto"
)

//打印机设备
type DevPrinter struct {
	GrpcClient

	UpState int
	DnState int
}

func (p *DevPrinter) InitDevPrinter() {
	p.UpState = models.DEVSTATE_UNKNOWN
	p.DnState = models.DEVSTATE_UNKNOWN

	p.GrpcInit(models.DEVGRPCTYPE_PRINTER, config.ConfigData["printerGrpcUrlCli"].(string), p.FuncGrpcPrinterProc)
}

func (p *DevPrinter) FuncGrpcPrinterProc(msg *pb.Message) {
	sType := msg.Type

	if sType != models.GRPCTYPE_PRINTERSTATE {
		util.FileLogs.Info("FuncGrpcPrinterProc %s-收到GRPC应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_PRINTERSTATE:
		p.FuncPrinterState(msg)
	}

	return
}

func (p *DevPrinter) FuncPrinterState(msg *pb.Message) {
	//sType := models.GRPCTYPE_PRINTERSTATE
	//util.FileLogs.Info("FuncPrinterState %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.PrinterStateReport{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	p.UpState = util.ConvertS2I(recvmsg.UpState)
	p.DnState = util.ConvertS2I(recvmsg.DnState)
}

func (p *DevPrinter) FuncReqPrintTicket(req models.ReqPrinterTicketData) error {
	sType := models.GRPCTYPE_PRINTTICKET
	util.FileLogs.Info("FuncReqPrintTicket %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	rsds := new(pb.PrinterTicketRequest)
	rsds.WorkStation = req.WorkStation
	rsds.LineNums = req.LineNums
	icnt := util.ConvertS2I(req.LineNums)
	for i := 0; i < icnt; i += 1 {
		rsds.PrintRsds = append(rsds.PrintRsds, &pb.PrinterContent{
			Aligyntype: req.PrintRsds[i].Aligyntype,
			Fontsize:   req.PrintRsds[i].Fontsize,
			Content:    req.PrintRsds[i].Content})
	}

	return p.GrpcSendproc(sType, "0", rsds)
}

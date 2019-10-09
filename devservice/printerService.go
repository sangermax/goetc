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

//打印机服务
type PrinterService struct {
	device.SerialBase

	State     int
	LastQueTm int64
}

//协议
func (p *PrinterService) enpack(cmd int) ([]byte, error) {
	return nil, nil
}

func (p *PrinterService) depack(inbuf []byte) ([]byte, int, error) {
	return nil, 0, nil
}

func (p *PrinterService) InitPrinter() {
	p.State = models.DEVSTATE_UNKNOWN
	p.LastQueTm = 0

	p.initPrinter()
	p.setPrinterFont(models.FONTSIZE_1)
	p.setLeftMargin(6)
}

func (p *PrinterService) FuncSend(info models.ReqPrinterTicketData) bool {
	buf := make([]byte, models.MAX_BUFFERSIZE)
	pos := 0

	return p.SendProc(buf[0:pos], true)
}

func (p *PrinterService) Recvproc(inbuf []byte) (int, error) {
	p.LastQueTm = util.GetTimeStampMs()

	inlen := len(inbuf)
	if inlen < 4 {
		return 0, errors.New("长度不足")
	}

	//22 18 18 18

	if inbuf[0] == 0x16 && inbuf[1] == 0x12 && inbuf[2] == 0x12 && inbuf[3] == 0x12 {
		p.State = models.DEVSTATE_OK
		return inlen, nil
	}

	i1 := 0
	//正常状态下打印机返回 16H，当返回的字节的 BIT3=1 时，打印机处于脱机状态。
	if inbuf[0]&0x04 == 0x04 {
		i1 = 1 //脱机
	} else if inbuf[1]&0x10 == 0x10 {
		i1 = 2 //无纸
	} else if inbuf[1]&0x20 == 0x20 {
		i1 = 3 //机械故障
	} else if inbuf[3]&0x20 == 0x20 {
		i1 = 2 //无纸
	}

	if 0 == i1 {
		p.State = models.DEVSTATE_OK
	} else {
		p.State = models.DEVSTATE_TROUBLE
	}

	return inlen, nil
}

func (p *PrinterService) initPrinter() {
	buf := make([]byte, 2)
	pos := 0

	buf[pos] = 0x1B
	pos += 1
	buf[pos] = 0x40
	pos += 1
	p.SendProc(buf[0:pos], true)
}

func (p *PrinterService) ClearCache() {
	buf := make([]byte, 1)
	pos := 0

	buf[pos] = 0x18
	pos += 1
	p.SendProc(buf[0:pos], true)
}

func (p *PrinterService) goReqPrinterState() {
	for {
		if !p.IsConn() {
			util.MySleep_s(5)
			continue
		}

		nowms := util.GetTimeStampMs()
		if nowms > p.LastQueTm+10000 {
			p.State = models.DEVSTATE_TROUBLE
		}

		buf := make([]byte, 20)
		pos := 0

		buf[pos] = 0x10
		pos += 1
		buf[pos] = 0x04
		pos += 1
		buf[pos] = 0x01
		pos += 1
		buf[pos] = 0x10
		pos += 1
		buf[pos] = 0x04
		pos += 1
		buf[pos] = 0x02
		pos += 1
		buf[pos] = 0x10
		pos += 1
		buf[pos] = 0x04
		pos += 1
		buf[pos] = 0x03
		pos += 1
		buf[pos] = 0x10
		pos += 1
		buf[pos] = 0x04
		pos += 1
		buf[pos] = 0x04
		pos += 1

		p.SendProc(buf[0:pos], true)

		util.MySleep_s(3)
	}
}

func (p *PrinterService) setPrinterFont(nFontType int) {
	inBuf := make([]byte, 10)
	pos := 0

	switch nFontType {
	case models.FONTSIZE_1:
		{
			inBuf[pos] = 0x1B
			pos += 1
			inBuf[pos] = 0x4D
			pos += 1
		}
		break

	case models.FONTSIZE_2:
		{
			inBuf[pos] = 0x1B
			pos += 1
			inBuf[pos] = 0x50
			pos += 1
		}
		break

	case models.FONTSIZE_3:
		{
			inBuf[pos] = 0x1B
			pos += 1
			inBuf[pos] = 0x3A
			pos += 1
		}
		break

	default:
		{
			inBuf[pos] = 0x1B
			pos += 1
			inBuf[pos] = 0x4D
			pos += 1
		}

		break
	}

	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) setPrinterLine(nLine int) {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x61
	pos += 1
	inBuf[pos] = byte(nLine)
	pos += 1

	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) setLeftMargin(nLeft int) {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x6C
	pos += 1
	inBuf[pos] = byte(nLeft)
	pos += 1

	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) setRightMargin(nRight int) {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x51
	pos += 1
	inBuf[pos] = byte(nRight)
	pos += 1

	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) setExpand(bFlag bool) {
	inBuf := make([]byte, 10)
	pos := 0

	if bFlag {
		inBuf[pos] = 0x0E
		pos += 1
	} else {
		inBuf[pos] = 0x14
		pos += 1
	}

	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) setCenter(szOld string) string {
	nLen := len(szOld)

	if models.MAX_PRINTER_LEN <= nLen {
		return fmt.Sprintf("%s", szOld)
	}

	nLeft := (models.MAX_PRINTER_LEN - nLen) / 2
	return fmt.Sprintf("%*s", nLeft, szOld)
}

func (p *PrinterService) QueryState() {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x04
	pos += 1
	p.SendProc(inBuf[0:pos], false)
}

//走纸
func (p *PrinterService) golines(lines int) {
	inBuf := make([]byte, 1024)
	pos := 0

	for i := 0; i < lines; i += 1 {
		inBuf[pos] = 0x0A
		pos += 1
	}

	p.SendProc(inBuf[0:pos], true)
}

//半切
func (p *PrinterService) cutHalf() {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x6D
	pos += 1

	p.SendProc(inBuf[0:pos], true)
}

//全切
func (p *PrinterService) cutAll() {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x69
	pos += 1

	p.SendProc(inBuf[0:pos], true)
}

//进纸n/144英寸
func (p *PrinterService) goDistance(value byte) {
	inBuf := make([]byte, 10)
	pos := 0

	inBuf[pos] = 0x1B
	pos += 1
	inBuf[pos] = 0x4A
	pos += 1
	inBuf[pos] = value
	pos += 1
	p.SendProc(inBuf[0:pos], true)
}

func (p *PrinterService) printContent(pc models.ReqPrinterTicketData) {
	p.golines(15)
	cnts := len(pc.PrintRsds)
	for i := 0; i < cnts; i += 1 {
		lineRsd := pc.PrintRsds[i]
		var newline string

		ia := util.ConvertS2I(lineRsd.Aligyntype)
		switch ia {
		case models.AlignCenter:
			newline = p.setCenter(lineRsd.Content)
			break

		default:
			newline = lineRsd.Content
			break
		}

		ifont := util.ConvertS2I(lineRsd.Fontsize)
		switch ifont {
		case models.SizeTimes, models.SizeDoubleTimes:
			p.setExpand(true)
			break

		default:
			p.setExpand(false)
			break
		}

		bsgbk, _ := util.Utf8ToGbk([]byte(newline))
		p.SendProc(bsgbk, true)
		p.golines(1)
		p.setExpand(false)

		//发太多，打印机异常？
		util.MySleep_ms(200)
		util.FileLogs.Info("打印:%s", newline)
	}

	p.golines(17 - cnts)

	//p.goDistance(0x4A)
	//切
	//p.cutHalf()
	p.cutAll()
}

func (p *PrinterService) testp() {
	var ticketinfo models.ReqPrinterTicketData

	//出票
	var pc1 models.PrinterContent
	pc1.Aligyntype = util.ConvertI2S(models.AlignCenter)
	pc1.Fontsize = util.ConvertI2S(models.SizeTimes)
	pc1.Content = fmt.Sprintf("%s", util.GetVehclassDes(1))
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	pc1.Aligyntype = util.ConvertI2S(models.AlignCenter)
	pc1.Fontsize = util.ConvertI2S(models.SizeTimes)
	pc1.Content = fmt.Sprintf("%.2f元", (float32(1)+float32(0.005))/100.00)
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
	pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
	pc1.Content = fmt.Sprintf("时间:%s  工号:%s %s", util.GetNow(false), "981", "1101")
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)
	/*
		pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
		pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
		pc1.Content = fmt.Sprintf("支付方式:%s  交易序号:%s", "支付宝", "1234567890")
		ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)
	*/
	pc1.Aligyntype = util.ConvertI2S(models.AlignLeft)
	pc1.Fontsize = util.ConvertI2S(models.SizeNormal)
	pc1.Content = fmt.Sprintf("二维码:%s", "123456789012345")
	ticketinfo.PrintRsds = append(ticketinfo.PrintRsds, pc1)

	ticketinfo.WorkStation = util.ConvertI2S(models.WORKSTATION_DN)
	ticketinfo.LineNums = util.ConvertI2S(len(ticketinfo.PrintRsds))

	gPrinterDnSrv.printContent(ticketinfo)
}

var gPrinterUpSrv PrinterService
var gPrinterDnSrv PrinterService

func main() {
	util.FileLogs.Info("打印机服务启动中...")
	config.InitConfig("../conf/config.conf")

	gPrinterUpSrv.InitSerial(models.DEVTYPE_PRINTERUP, config.ConfigData["PrinterUpCom"].(string), 9600, gPrinterUpSrv.Recvproc)
	gPrinterUpSrv.InitPrinter()
	go gPrinterUpSrv.goReqPrinterState()

	gPrinterDnSrv.InitSerial(models.DEVTYPE_PRINTERDN, config.ConfigData["PrinterDnCom"].(string), 9600, gPrinterDnSrv.Recvproc)
	gPrinterDnSrv.InitPrinter()
	go gPrinterDnSrv.goReqPrinterState()
	//gPrinterDnSrv.testp()

	/////////////////////////////////////grpc
	sAddr := config.ConfigData["printerGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("打印机服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("打印机服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &PrinterServer{})
	s.Serve(lis)

}

type PrinterServer struct {
}

func (p *PrinterServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道打印机GRPC已连接")
	bConn := true

	go func() {
		for {
			if !bConn {
				break
			}

			//发送设备状态
			sUpState := util.ConvertI2S(gPrinterUpSrv.State)
			sDnState := util.ConvertI2S(gPrinterDnSrv.State)
			m := &pb.PrinterStateReport{UpState: sUpState, DnState: sDnState}
			PrinterSrvGrpcSendproc(stream, models.GRPCTYPE_PRINTERSTATE, "0", m, models.ResultInfo{})

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

		util.FileLogs.Info("收到来自车道打印信息")
		fmt.Println(in)

		switch in.Type {
		case models.GRPCTYPE_PRINTTICKET:
			FuncPrintTicketProc(in.Data)
			break
		}

	}

	return nil
}

func PrinterSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("printer GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("printer GrpcSendproc ")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("printer GrpcSendproc")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

func FuncPrintTicketProc(inbuf []byte) {
	recvmsg := &pb.PrinterTicketRequest{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		fmt.Printf("FuncPrintTicketProc err:%s.\r\n", err.Error())
		return
	}

	var rst models.ReqPrinterTicketData

	i := 0
	icnt := util.ConvertS2I(recvmsg.LineNums)
	rst.LineNums = recvmsg.LineNums
	for i = 0; i < icnt; i += 1 {
		var unit models.PrinterContent
		unit.Aligyntype = recvmsg.PrintRsds[i].Aligyntype
		unit.Fontsize = recvmsg.PrintRsds[i].Fontsize
		unit.Content = recvmsg.PrintRsds[i].Content

		rst.PrintRsds = append(rst.PrintRsds, unit)
	}

	ipos := util.ConvertS2I(recvmsg.WorkStation)
	if ipos == models.WORKSTATION_DN {
		gPrinterDnSrv.printContent(rst)
	} else {
		gPrinterUpSrv.printContent(rst)
	}
}

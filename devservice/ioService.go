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
	"sync"

	"FTC/models"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

const IOHEAD = 0xFE
const (
	CMDIO_01 = 0x01 //继电器状态
	CMDIO_05 = 0x05 //继电器输出

	CMDIO_02 = 0x02 //光耦输入
	CMDIO_10 = 0x10 //闪开闪闭
	CMDIO_0F = 0x0F //全开全关

	CMDIO_81 = 0x81 //继电器状态 查询错误
	CMDIO_82 = 0x82 //错误
)

const IORATE = 100  //单位ms IO问询频率
const IOONKEEP = 2  //200*2 单位ms  IO有信号 持续时间
const IOOFFKEEP = 3 //200*3 单位ms IO无信号 持续时间

//IO服务
type IOService struct {
	device.SerialBase

	State [8]int

	//过滤信号
	onCoil   [8]int
	delayCnt [8]int

	ioLock    *sync.Mutex
	LastQueTm int64
}

//协议
func (p *IOService) enpack(inbuf []byte) ([]byte, error) {
	inlen := len(inbuf)
	outbuf := make([]byte, inlen+2)
	copy(outbuf, inbuf)
	icrc := util.CRC16(inbuf, inlen)
	copy(outbuf[inlen:], util.Short2Bytes_L(icrc))
	return outbuf, nil
}

func (p *IOService) depack(inbuf []byte) ([]byte, int, error) {
	inlen := len(inbuf)
	if inlen < 4 {
		return nil, 0, errors.New("帧未收完")
	}

	i := 0
	framelen := 0
	for {
		for ; i < inlen; i = i + 1 {
			if inbuf[i] == IOHEAD {
				break
			}
		}

		if i < inlen {
			if i+4 > inlen {
				return nil, i, errors.New("帧未收完1")
			}

			bFinish := false
			pos := i
			pos += 1
			cmd := inbuf[pos]
			pos += 1

			switch cmd {
			case CMDIO_01, CMDIO_02:
				if pos+1 < inlen {
					b := int(inbuf[pos])
					pos += 1

					framelen = 5 + b
					if i+framelen <= inlen {
						bFinish = true
					}
				}

				if !bFinish {
					return nil, i, errors.New("帧未收完2")
				}

			case CMDIO_05:
				framelen = 8
			case CMDIO_10:
				framelen = 8
			case CMDIO_0F:
				framelen = 8
			case CMDIO_81, CMDIO_82:
				framelen = 5
			default:
				return nil, i, errors.New("命令字错误")
			}

			if i+framelen > inlen {
				return nil, i, errors.New("帧未收完3")
			}

			//判断校验码
			crc := util.CRC16(inbuf[i:i+framelen-2], framelen-2)
			bscrc := util.Short2Bytes_L(crc)

			if bscrc != nil && bscrc[0] == inbuf[i+framelen-2] && bscrc[1] == inbuf[i+framelen-1] {
				return inbuf[i : i+framelen], i + framelen, nil
			} else {
				return nil, i + 1, errors.New("校验码错误")
			}
		} else {
			return nil, inlen, errors.New("未找到帧头")
		}
	}
	return nil, 0, nil
}

func (p *IOService) InitIO() {
	p.ioLock = new(sync.Mutex)
	p.LastQueTm = 0
	for i := 0; i < 8; i++ {
		p.State[i] = models.COIL_CLOSE
		p.onCoil[i] = 0
		p.delayCnt[i] = 0
	}
}

func (p *IOService) Recvproc(inbuf []byte) (int, error) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, errors.New("长度不足")
	}

	//util.FileLogs.Info("Recvproc:%d,%s.\r\n", inlen, util.ConvertByte2Hexstring(inbuf, true))

	outbuf, offset, err := p.depack(inbuf[0:inlen])
	if err != nil {
		//util.FileLogs.Info("Recvproc:depack err.%d,%s.\r\n", offset, err.Error())
		return offset, err
	}

	switch outbuf[1] {
	case CMDIO_05: //控制返回
		p.FuncRstControl(outbuf)
	case CMDIO_02:
		p.FuncRstState(outbuf)

	}
	return offset, nil
}

//根据实际情况实现//////////////////////////////////////
func (p *IOService) GetNoByAddr(inbuf []byte) int {
	return 0
}

func (p *IOService) GetAddrByNo(no int) []byte {
	buf := make([]byte, 2)
	switch no {
	case 0:
		buf[0] = 0x00
		buf[1] = 0x00

	case 1:
		buf[0] = 0x00
		buf[1] = 0x01
	case 2:
		buf[0] = 0x00
		buf[1] = 0x02
	case 3:
		buf[0] = 0x00
		buf[1] = 0x03
	case 4:
		buf[0] = 0x00
		buf[1] = 0x04
	case 5:
		buf[0] = 0x00
		buf[1] = 0x05
	case 6:
		buf[0] = 0x00
		buf[1] = 0x06
	case 7:
		buf[0] = 0x00
		buf[1] = 0x07
	}
	return buf
}

func (p *IOService) FuncRstControl(inbuf []byte) {
	util.FileLogs.Info("FuncRstControl:IO控制应答处理")

	var rst models.ReqIOControlData
	//地址 //地址与序号对应关系
	rst.IONo = util.ConvertI2S(p.GetNoByAddr(inbuf[2:4]))
	if inbuf[4] == 0xFF && inbuf[5] == 0x00 { //落杆
		rst.IOControl = util.ConvertI2S(models.COIL_CLOSE)
	}
	if inbuf[4] == 0x00 && inbuf[5] == 0x00 { //抬杆
		rst.IOControl = util.ConvertI2S(models.COIL_OPEN)
	}

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""

	m := &pb.IOControlResponse{IONo: rst.IONo, IOControl: rst.IOControl}
	if g_iostream != nil {
		IOSrvGrpcSendproc(g_iostream, models.GRPCTYPE_IOCONTROL, "0", m, result)
	}
}

func (p *IOService) FuncRstState(inbuf []byte) {
	state := inbuf[3]
	var i byte = 0x00

	for i = 0x00; i < 8; i++ {
		c := int((state >> i) & 0x01)
		/*
			//信号过滤
			if c == 1 {
				p.delayCnt[i] += 1
				if p.delayCnt[i] >= IOONKEEP {
					if p.onCoil[i] == 0 {
						p.State[i] = 1
						p.onCoil[i] = IOOFFKEEP
					}
				}
			} else if p.onCoil[i] > 0 {
				p.delayCnt[i] = 0
				p.onCoil[i] -= 1
				if p.onCoil[i] == 0 {
					p.State[i] = 0
				}
			}
		*/

		p.State[i] = c
	}

	/*
		p.State[0] = int(state & 0x01)
		p.State[1] = int((state & 0x02) >> 1)
		p.State[2] = int((state & 0x04) >> 2)
		p.State[3] = int((state & 0x08) >> 3)
		p.State[4] = int((state & 0x10) >> 4)
		p.State[5] = int((state & 0x20) >> 5)
		p.State[6] = int((state & 0x40) >> 6)
		p.State[7] = int((state & 0x80) >> 7)
	*/
	util.FileLogs.Info("IO状态(%d,%d,%d,%d,%d,%d,%d,%d):", p.State[0], p.State[1], p.State[2], p.State[3], p.State[4], p.State[5], p.State[6], p.State[7])
}

//光耦输入信号，1次4个寄存器状态
func (p *IOService) goReqIOState() {
	for {
		if !p.IsConn() {
			util.MySleep_s(3)
			continue
		}

		outbuf := make([]byte, 6)
		outbuf[0] = 0xFE
		outbuf[1] = CMDIO_02
		outbuf[2] = 0x00
		outbuf[3] = 0x00
		outbuf[4] = 0x00
		outbuf[5] = 0x04

		sendbuf, _ := p.enpack(outbuf)

		p.ioLock.Lock()
		p.LastQueTm = util.GetTimeStampMs()
		p.SendProc(sendbuf, false)
		p.ioLock.Unlock()

		util.MySleep_ms(IORATE)
	}
}

var gIOSrv IOService
var g_iostream pb.GrpcMsg_CommuniteServer

func main() {
	util.FileLogs.Info("IO服务启动中...")
	config.InitConfig("../conf/config.conf")
	gIOSrv.InitIO()
	gIOSrv.InitSerial(models.DEVTYPE_IO, config.ConfigData["ioCom"].(string), 38400, gIOSrv.Recvproc)
	//io状态
	go gIOSrv.goReqIOState()

	///////////	//////////////////////////grpc
	sAddr := config.ConfigData["ioGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("IO服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("IO服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &IOServer{})
	s.Serve(lis)

}

type IOServer struct {
}

func (p *IOServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道IO GRPC已连接")
	g_iostream = stream
	bConn := true

	go func() {
		for {
			if !bConn {
				break
			}

			//发送设备状态
			sState1 := util.ConvertI2S(gIOSrv.State[0])
			sState2 := util.ConvertI2S(gIOSrv.State[1])
			sState3 := util.ConvertI2S(gIOSrv.State[2])
			sState4 := util.ConvertI2S(gIOSrv.State[3])
			sState5 := util.ConvertI2S(gIOSrv.State[4])
			sState6 := util.ConvertI2S(gIOSrv.State[5])
			sState7 := util.ConvertI2S(gIOSrv.State[6])
			sState8 := util.ConvertI2S(gIOSrv.State[7])
			m := &pb.IOStateReport{State1: sState1,
				State2: sState2,
				State3: sState3,
				State4: sState4,
				State5: sState5,
				State6: sState6,
				State7: sState7,
				State8: sState8}
			IOSrvGrpcSendproc(stream, models.GRPCTYPE_IOSTATE, "0", m, models.ResultInfo{})

			util.MySleep_ms(500)
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

		util.FileLogs.Info("收到来自车道IO控制信息")
		fmt.Println(in)

		switch in.Type {
		case models.GRPCTYPE_IOCONTROL:
			FuncReqIOControl(in.Data)
		}

	}

	return nil
}

func IOSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("io GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			if sType != models.GRPCTYPE_IOSTATE {
				util.FileLogs.Info("io GrpcSendproc ")
				fmt.Println(notes[0])
			}

			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			if sType != models.GRPCTYPE_IOSTATE {
				util.FileLogs.Info("io GrpcSendproc")
				fmt.Println(notes[0])
			}

			stream.Send(notes[0])
		}
	}

}

func FuncReqIOControl(inbuf []byte) {
	req := &pb.IOControlRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		fmt.Printf("FuncReqIOControl err:%s.\r\n", err.Error())
		return
	}

	outbuf := make([]byte, 6)
	addr := gIOSrv.GetAddrByNo(util.ConvertS2I(req.IONo))
	ctl := util.ConvertS2I(req.IOControl)

	outbuf[0] = 0xFE
	outbuf[1] = CMDIO_05
	copy(outbuf[2:], addr)

	switch ctl {
	case models.COIL_OPEN:
		outbuf[4] = 0x00
		outbuf[5] = 0x00
	case models.COIL_CLOSE:
		outbuf[4] = 0xFF
		outbuf[5] = 0x00
	}

	sendbuf, _ := gIOSrv.enpack(outbuf)

	util.FileLogs.Info("FuncReqIOControl:%d,%s.\r\n", len(sendbuf), util.ConvertByte2Hexstring(sendbuf, true))
	gIOSrv.ioLock.Lock()
	nowms := util.GetTimeStampMs()
	diffms := nowms - gIOSrv.LastQueTm
	if diffms < 30 {
		util.MySleep_ms(30 - diffms)
	}
	gIOSrv.SendProc(sendbuf, true)
	util.MySleep_ms(30)
	gIOSrv.ioLock.Unlock()
}

func testio() {
	outbuf := make([]byte, 6)
	addr := gIOSrv.GetAddrByNo(util.ConvertS2I("0"))
	ctl := models.COIL_OPEN

	outbuf[0] = 0xFE
	outbuf[1] = CMDIO_05
	copy(outbuf[2:], addr)

	switch ctl {
	case models.COIL_OPEN:
		outbuf[4] = 0x00
		outbuf[5] = 0x00
	case models.COIL_CLOSE:
		outbuf[4] = 0xFF
		outbuf[5] = 0x00
	}

	sendbuf, _ := gIOSrv.enpack(outbuf)

	util.FileLogs.Info("FuncReqIOControl:%d,%s.\r\n", len(sendbuf), util.ConvertByte2Hexstring(sendbuf, true))
	gIOSrv.ioLock.Lock()
	nowms := util.GetTimeStampMs()
	diffms := nowms - gIOSrv.LastQueTm
	if diffms < 30 {
		util.MySleep_ms(30 - diffms)
	}
	gIOSrv.SendProc(sendbuf, true)
	util.MySleep_ms(30)
	gIOSrv.ioLock.Unlock()
}

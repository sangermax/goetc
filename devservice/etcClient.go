package main

import (
	"FTC/config"
	"FTC/device"
	"FTC/pb"
	"FTC/util"
	"errors"
	"fmt"

	"FTC/models"

	"github.com/golang/protobuf/proto"
)

//ETC天线
type ETCSerial struct {
	device.SerialBase

	AntGroupKey string
	AntNo       int
	State       int
	PsamTermid  string
	LastQueTm   int64

	InitAntFlag bool //天线初始化状态
}

func (p *ETCSerial) InitEtcDev(no int, comdes string, band int) {
	p.AntGroupKey = config.ConfigData["etcAntKey"].(string)
	p.AntNo = no
	p.State = models.DEVSTATE_UNKNOWN
	p.InitAntFlag = false
	p.LastQueTm = 0
	p.PsamTermid = ""

	p.InitSerial2(models.DEVTYPE_ETC, comdes, band, p.ConnectedInit, p.Recvproc)
}

func (p *ETCSerial) FuncSend(inbuf []byte) bool {
	if !p.IsConn() {
		return false
	}

	util.FileLogs.Info("etc send：%02X_%s", inbuf[2], util.ConvertByte2Hexstring(inbuf, false))
	return p.SendProc(inbuf, false)
}

func (p *ETCSerial) ConnectedInit() {
	p.SendProc(util.ETCPackC0(0x80), true)
}

func (p *ETCSerial) Recvproc(inbuf []byte) (int, error) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, errors.New("长度不足")
	}

	p.LastQueTm = util.GetTimeStampSec()

	outbuf, offset, err := util.ETCDepack(inbuf[0:inlen])
	if err != nil {
		//fmt.Println("ETCDepack err:" + err.Error())
		return offset, err
	}

	//处理
	cmd := outbuf[2]
	util.FileLogs.Info("etc recvproc :%02x", cmd)
	switch cmd {
	case models.CMD_B0:
		p.ProcParseB0(outbuf)
		return offset, nil

	case models.CMD_B2:
		//心跳，不处理
		if outbuf[7] == 0x80 {
			return offset, nil
		}
	}

	//发送到服务端
	m := &pb.ETCAntReadRequest{Key: p.AntGroupKey,
		AntNo:      util.ConvertI2S(p.AntNo),
		PsamTermid: p.PsamTermid,
		Msg:        util.ConvertByte2Hexstring(outbuf, false)}

	gEtcGrpcCli.GrpcSendproc(models.GRPCTYPE_ETCMSG, "0", m)

	return offset, nil
}

func (p *ETCSerial) ProcParseB0(inbuf []byte) {
	rsctl := inbuf[1]
	errcode := inbuf[3]
	//rsu上电初始化
	if rsctl == 0x98 {
		p.SendProc(util.ETCPackC0(0x89), true)
		return
	}

	reversalRsctl := util.ETCReversalRsctl(rsctl)
	p.SendProc(util.ETCPackC1(reversalRsctl, []byte{0, 0, 0, 0}), true)

	frameb0, err := util.ETCParseB0(inbuf)
	if errcode != models.CODE_OK || err != nil {
		p.InitAntFlag = false
	} else {
		p.InitAntFlag = true
		p.PsamTermid = frameb0.RSUTerminalId
	}
}

func (p *ETCSerial) chkFrameError(frametype byte, inbuf []byte) bool {
	bsuc := true
	framelen := 0
	errcode := inbuf[7]

	if errcode != models.CODE_OK {
		bsuc = false
	}

	if bsuc {
		switch frametype {
		case models.CMD_B2:
			framelen = models.ETCLENGTH_B2
		case models.CMD_B3:
			framelen = models.ETCLENGTH_B3
		case models.CMD_B4:
			framelen = models.ETCLENGTH_B4
		case models.CMD_B5:
			framelen = models.ETCLENGTH_B5
		case models.CMD_B7:
			framelen = models.ETCLENGTH_B7 + int(inbuf[8])*models.ETCLENGTH_0018
		default:
			bsuc = false
		}

		if len(inbuf) < framelen {
			bsuc = false
		}

	}

	return bsuc
}

//ETC天线GRPC客户端////////////////////////////////////////////////////////
type ETCGrpcClient struct {
	device.GrpcClient
}

func (p *ETCGrpcClient) InitEtcGrpc() {
	p.GrpcInit2(models.DEVGRPCTYPE_ETC, config.ConfigData["etcGrpcUrlCli"].(string), p.FuncEtcGRPCInit, p.FuncETCGrpcProc)
}

func (p *ETCGrpcClient) sendETCAntState() {
	gEtcAntMap.Lock.Lock()
	defer gEtcAntMap.Lock.Unlock()

	m := new(pb.ETCAntStateReport)
	for _, v := range gEtcAntMap.BM {
		if v == nil {
			continue
		}

		ev := v.(*ETCSerial)
		m.State = append(m.State, &pb.ETCAntState{
			AntNo: util.ConvertI2S(ev.AntNo),
			State: util.ConvertI2S(ev.State)})
	}

	p.GrpcSendproc(models.GRPCTYPE_ETCAntState, "0", m)
}

func (p *ETCGrpcClient) goAutoState() {
	for {
		if !p.IsGrpcConn() {
			util.MySleep_s(5)
			continue
		}

		p.sendETCAntState()
		util.MySleep_s(5)
	}
}

func (p *ETCGrpcClient) FuncETCGrpcProc(msg *pb.Message) {
	sType := msg.Type
	sno := msg.No

	util.FileLogs.Info("FuncETCGrpcProc %s,%s-收到GRPC服务应答 开始处理.\r\n", sno, device.GetCmdDes(sType))
	switch sType {
	case models.GRPCTYPE_ETCAntInit:
		break

	case models.GRPCTYPE_ETCAntState:
		break

	case models.GRPCTYPE_ETCMSG:
		fmt.Println(msg)
		p.FuncEtcMsgProc(msg)
		break
	}

	return
}

func (p *ETCGrpcClient) FuncEtcMsgProc(msg *pb.Message) {
	sType := msg.Type
	inbuf := msg.Data
	util.FileLogs.Info("FuncEtcMsgProc %s-GRPC服务应答.\r\n", sType)

	if msg.Resultvalue != models.GRPCRESULT_OK {
		fmt.Println("FuncEtcMsgProc err resultvalue:%d.", msg.Resultvalue)
		return
	}

	recvmsg := &pb.ETCAntReadResponse{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		fmt.Println("FuncEtcMsgProc Unmarshal err .")
		return
	}

	v := gEtcAntMap.Get(recvmsg.AntNo)
	if v != nil {
		ev := v.(*ETCSerial)
		ev.SendProc(util.ConvertHexstring2Byte(recvmsg.Msg), true)
	} else {
		fmt.Println("gEtcAntMap.Get nil.")
	}

}

func (p *ETCGrpcClient) FuncEtcGRPCInit() {
	if !p.IsGrpcConn() {
		return
	}

	key := config.ConfigData["etcAntKey"].(string)
	m := &pb.ETCAntInitRequest{Key: key}
	p.GrpcSendproc(models.GRPCTYPE_ETCAntInit, "0", m)
}

var gEtcGrpcCli ETCGrpcClient

//etc天线集群
var gEtcAntMap *util.BeeMap

func main() {
	util.FileLogs.Info("ETC天线客户端启动中...")
	config.InitConfig("../conf/config.conf")
	gEtcAntMap = util.NewBeeMap()

	antnums := config.ConfigData["etcNums"].(int)
	for i := 1; i <= antnums; i++ {
		s0 := util.ConvertI2S(i)
		s1 := "etcCom" + s0
		s2 := config.ConfigData[s1].(string)

		pETCSerial := new(ETCSerial)
		pETCSerial.InitEtcDev(i, s2, 115200)
		gEtcAntMap.Set(s0, pETCSerial)
	}

	gEtcGrpcCli.InitEtcGrpc()

	for {
		util.MySleep_s(5)
	}
}

package main

import (
	"FTC/config"
	"FTC/device"
	"FTC/pb"
	"FTC/util"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"FTC/models"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//费显服务
type FeedispService struct {
	device.SerialBase

	FeedispState int
	AlarmState   int
	Hearttimeout int
	Alarmtime    int
}

func (p *FeedispService) InitFeedisp() {
	p.FeedispState = models.DEVSTATE_UNKNOWN
	p.AlarmState = models.DEVSTATE_UNKNOWN
	p.Hearttimeout = models.Heartbeattimeout
	p.Alarmtime = 10

	p.Showdefault()
}

// Showdefault 默认显示
func (c *FeedispService) Showdefault() {
	line1 := "自助缴费车道"
	line2 := "  系统维护  "
	line3 := "请转其它车道"

	c.Showcontent(line1, line2, line3, models.ColorRed, models.ColorRed, models.ColorRed)
	c.Setalarm(models.Alarmoff, 0)
}

// Showerr 显示错误信息
func (c *FeedispService) Showerr(plate string, reson string, line3 string) {
	c.Showcontent(plate, reson, line3, models.ColorRed, models.ColorRed, models.ColorRed)
}

// Showcontent 显示内容
func (c *FeedispService) Showcontent(line1, line2, line3 string, color1, color2, color3 byte) {
	util.FileLogs.Info("费显信息:%s,%s,%s.\r\n", line1, line2, line3)

	bs1, _ := util.Utf8ToGbk([]byte(line1))
	bs2, _ := util.Utf8ToGbk([]byte(line2))
	bs3, _ := util.Utf8ToGbk([]byte(line3))
	bline1 := c.cutorfill12(bs1)
	bline2 := c.cutorfill12(bs2)
	bline3 := c.cutorfill12(bs3)

	var disp Feeproto
	disp.Head[0] = 0xA0
	disp.Head[1] = 0xB0
	disp.Head[2] = 0xC0
	disp.Head[3] = 0xD0
	disp.Cmd = 'D'
	disp.Light = models.Displaylightday
	disp.Length = 36
	disp.Color[0] = color1
	disp.Color[1] = color2
	disp.Color[2] = color3

	c.assigndata(bline1, bline2, bline3, &disp.Data)
	disp.Tail = 0xE0

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, disp)

	bufffinal := buf.Bytes()
	bufffinal[len(bufffinal)-1] = Bcc(bufffinal)
	c.SendProc(bufffinal, false)
}

// Setalarm 报警 1开, 0关
func (c *FeedispService) Setalarm(control byte, alarmtime int) {
	util.FileLogs.Info("费显报警器信息:%d,%d.\r\n", control, alarmtime)

	var alarmcmd [13]byte
	alarmcmd[0] = 0xA0
	alarmcmd[1] = 0xB0
	alarmcmd[2] = 0xC0
	alarmcmd[3] = 0xD0
	alarmcmd[4] = 0x06
	alarmcmd[5] = models.Displaylightday
	alarmcmd[6] = 1
	alarmcmd[7] = models.ColorRed
	alarmcmd[8] = models.ColorRed
	alarmcmd[9] = models.ColorRed
	alarmcmd[10] = control
	alarmcmd[11] = 0xE0

	time.Sleep(time.Millisecond * 50)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, alarmcmd)

	buffwrite := buf.Bytes()
	buffwrite[len(buffwrite)-1] = Bcc(buffwrite)

	c.SendProc(buffwrite, false)

	if control == models.Alarmoff {
		c.AlarmState = models.AlarmStateoff
	} else {
		c.AlarmState = models.AlarStatemon
	}

	c.Alarmtime = alarmtime
}

// putheart 心跳包
func (c *FeedispService) putheart() {
	var heart [12]byte
	heart[0] = 0xA0
	heart[1] = 0xB0
	heart[2] = 0xC0
	heart[3] = 0xD0

	heart[4] = 0x42
	heart[5] = 0xff
	heart[6] = 0x00
	heart[7] = models.ColorRed
	heart[8] = models.ColorRed
	heart[9] = models.ColorRed
	heart[10] = 0xE0

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, heart)

	buffwrite := buf.Bytes()
	buffwrite[len(buffwrite)-1] = Bcc(buffwrite)
	c.SendProc(buffwrite, false)
}

// Feeproto 费显的原始协议结构
type Feeproto struct {
	Head   [4]byte
	Cmd    byte
	Light  byte
	Length byte
	Color  [3]byte
	Data   [36]byte
	Tail   byte
	Bcc    byte
}

// Bcc 计算bcc
func Bcc(buf []byte) byte {
	var val byte
	length := len(buf)
	for i := 0; i < length-1; i++ {
		val ^= buf[i]
	}

	return val
}

// cutorfill12 裁剪或填补
func (c *FeedispService) cutorfill12(buf []byte) []byte {
	length := len(buf)
	if len(buf) > 12 {
		return buf[0:12]
	}

	var databuf []byte
	for i := 0; i < 12; i++ {
		if i < length {
			databuf = append(databuf, buf[i])
		} else {
			databuf = append(databuf, ' ')
		}
	}

	return databuf

}

// assigndata 赋值
func (c *FeedispService) assigndata(line1, line2, line3 []byte, data *[36]byte) {
	copy(data[0:], line1)
	copy(data[12:], line2)
	copy(data[24:], line3)
}

// heartbeat 关报警做定时器,并且做心跳
func (c *FeedispService) goHeartbeat() {
	if !c.IsConn() {
		return
	}

	hearttimer := models.Heartfrequence
	for {
		util.MySleep_s(1)

		if c.Alarmtime > 0 {
			c.Alarmtime--
			if c.Alarmtime == 0 {
				c.Setalarm(models.Alarmoff, 0)
			}
		}

		// 更新心跳状态
		if c.Hearttimeout > 0 {
			c.Hearttimeout--
			if c.Hearttimeout == 0 {
				c.FeedispState = models.DEVSTATE_TROUBLE
			}
		}

		if hearttimer > 0 {
			hearttimer--
			if hearttimer == 0 {
				c.putheart()
				hearttimer = models.Heartfrequence
			}
		}

	}

}

//费显接收进程,不管内容,收到报文则认为是ok的
func (p *FeedispService) Recvproc(inbuf []byte) (int, error) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, errors.New("长度不足")
	}

	p.Hearttimeout = models.Heartbeattimeout
	p.FeedispState = models.DEVSTATE_OK

	return inlen, nil
}

var gFeedispSrv FeedispService
var g_feedispstream pb.GrpcMsg_CommuniteServer

func main() {
	util.FileLogs.Info("费显服务启动中...")
	config.InitConfig("../conf/config.conf")

	gFeedispSrv.InitSerial(models.DEVTYPE_FEEDISP, config.ConfigData["feeDispCom"].(string), 9600, gFeedispSrv.Recvproc)
	gFeedispSrv.InitFeedisp()
	go gFeedispSrv.goHeartbeat()

	/////////////////////////////////////grpc
	sAddr := config.ConfigData["feedispGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("费显服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("费显服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &FeedispServer{})
	s.Serve(lis)

}

type FeedispServer struct {
}

func (p *FeedispServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道费显GRPC已连接")
	bConn := true
	g_feedispstream = stream

	go func() {
		for {
			if !bConn {
				break
			}

			//发送设备状态
			State1 := util.ConvertI2S(gFeedispSrv.FeedispState)
			State2 := util.ConvertI2S(gFeedispSrv.AlarmState)
			m := &pb.FeedispStateReport{FeedispState: State1, AlarmState: State2}
			FeedispSrvGrpcSendproc(stream, models.GRPCTYPE_FEEDISPSTATE, "0", m, models.ResultInfo{})

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

		util.FileLogs.Info("收到来自车道费显信息")
		fmt.Println(in)

		switch in.Type {
		case models.GRPCTYPE_FEEDISPSHOW:
			FuncFeedispShowProc(in.Data)

		case models.GRPCTYPE_FEEDISPALARM:
			FuncFeeAlarmProc(in.Data)
		}

	}

	return nil
}

func FeedispSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("feedisp GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("feedisp GrpcSendproc ")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			//util.FileLogs.Info("feedisp GrpcSendproc")
			//fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

func FuncFeedispShowProc(inbuf []byte) {
	recvmsg := &pb.FeedispShowRequest{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		fmt.Printf("FuncFeedispShowProc err:%s.\r\n", err.Error())
		return
	}

	gFeedispSrv.Showcontent(recvmsg.Line1, recvmsg.Line2, recvmsg.Line3, util.Converts2b(recvmsg.Color1), util.Converts2b(recvmsg.Color2), util.Converts2b(recvmsg.Color3))

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""

	m := &pb.FeedispShowResponse{}
	if g_feedispstream != nil {
		FeedispSrvGrpcSendproc(g_feedispstream, models.GRPCTYPE_FEEDISPSHOW, "0", m, result)
	}
}

func FuncFeeAlarmProc(inbuf []byte) {
	recvmsg := &pb.FeedispAlarmRequest{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		fmt.Printf("FuncFeeAlarmProc err:%s.\r\n", err.Error())
		return
	}

	gFeedispSrv.Setalarm(util.Converts2b(recvmsg.AlarmValue), util.ConvertS2I(recvmsg.AlarmTm))

	var result models.ResultInfo
	result.ResultValue = models.GRPCRESULT_OK
	result.ResultDes = ""

	m := &pb.FeedispAlarmResponse{}
	if g_feedispstream != nil {
		FeedispSrvGrpcSendproc(g_feedispstream, models.GRPCTYPE_FEEDISPALARM, "0", m, result)
	}
}

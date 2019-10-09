package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"errors"

	"github.com/golang/protobuf/proto"
)

//费显信号
type DevFeedisp struct {
	GrpcClient

	StateFeedisp int
	StateAlarm   int
}

func (p *DevFeedisp) InitDevFeedisp() {
	p.StateFeedisp = models.DEVSTATE_UNKNOWN
	p.StateAlarm = models.DEVSTATE_UNKNOWN

	p.GrpcInit2(models.DEVGRPCTYPE_FEEDISP, config.ConfigData["feedispGrpcUrlCli"].(string),p.FuncSendGRPCInit, p.FuncGrpcFeedispProc)
}

func (p *DevFeedisp) FuncSendGRPCInit() {
	if !p.IsGrpcConn() {
		return
	}

	var req models.FeedispShowData
	req.Line1= config.ConfigData["feeDispLine1"].(string)
	req.Line2= config.ConfigData["feeDispLine2"].(string)
	req.Line3= config.ConfigData["feeDispLine3"].(string)
	req.Color1 = util.Convertb2s(models.ColorGreen)
	req.Color2 = util.Convertb2s(models.ColorGreen)
	req.Color3 = util.Convertb2s(models.ColorGreen)
	p.FuncFeedispShow(req)

	var reqAlarm models.FeedispAlarmData
	reqAlarm.AlarmValue = util.Convertb2s(models.Alarmoff)
	reqAlarm.AlarmTm = util.ConvertI2S(0)
	p.FuncFeedispAlarm(reqAlarm)
}

func (p *DevFeedisp) FuncGrpcFeedispProc(msg *pb.Message) {
	sType := msg.Type

	if sType != models.GRPCTYPE_FEEDISPSTATE {
		util.FileLogs.Info("%s-收到GRPC应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_FEEDISPSTATE:
		p.FuncFeedispState(msg)
	}

	return
}

func (p *DevFeedisp) FuncFeedispState(msg *pb.Message) {
	//sType := models.GRPCTYPE_IOSTATE
	//util.FileLogs.Info("%s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.FeedispStateReport{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	p.StateFeedisp = util.ConvertS2I(recvmsg.FeedispState)
	p.StateAlarm = util.ConvertS2I(recvmsg.AlarmState)
}

func (p *DevFeedisp) FuncFeedispShow(req models.FeedispShowData) error {
	sType := models.GRPCTYPE_FEEDISPSHOW
	//util.FileLogs.Info("FuncFeedispShow %s-GRPC车道请求.\r\n", GetCmdDes(sType))
	util.FileLogs.Info("费显信息:%s,%s,%s.\r\n", req.Line1, req.Line2, req.Line3)

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	m := &pb.FeedispShowRequest{Line1: req.Line1,
		Line2:  req.Line2,
		Line3:  req.Line3,
		Color1: req.Color1,
		Color2: req.Color2,
		Color3: req.Color3}
	p.GrpcSendproc(sType, "0", m)

	return nil
}

func (p *DevFeedisp) FuncFeedispAlarm(req models.FeedispAlarmData) error {
	sType := models.GRPCTYPE_FEEDISPALARM
	util.FileLogs.Info("费显报警器信息:%s,%s.\r\n", req.AlarmValue, req.AlarmTm)

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	m := &pb.FeedispAlarmRequest{AlarmValue: req.AlarmValue,
		AlarmTm: req.AlarmTm}
	p.GrpcSendproc(sType, "0", m)

	return nil
}

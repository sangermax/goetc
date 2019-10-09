package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
)

//扫码设备
type DevScan struct {
	GrpcClient

	//请求处理grpc命令
	GrpcProcMap *util.BeeMap //key stype类型,GRPCTaskData为内容体

	UpState int
	DnState int
}

func (p *DevScan) InitDevScan() {
	p.UpState = models.DEVSTATE_UNKNOWN
	p.DnState = models.DEVSTATE_UNKNOWN

	p.GrpcProcMap = util.NewBeeMap()
	p.GrpcInit(models.DEVGRPCTYPE_SCAN, config.ConfigData["scanGrpcUrlCli"].(string), p.FuncGrpcScanProc)
}

func (p *DevScan) FuncCheckChan(sType string) bool {
	switch sType {
	case models.GRPCTYPE_OPENSCAN:
		return true
	}

	return false
}

func (p *DevScan) FuncGrpcScanProc(msg *pb.Message) {
	sType := msg.Type
	inbuf := msg.Data

	if p.FuncCheckChan(sType) {
		unit := p.GrpcProcMap.Get(sType)
		if unit == nil {
			util.FileLogs.Info("收到云平台服务应答，但是查找不到该请求，抛弃：(%s).\r\n", GetCmdDes(sType))
			return
		}
		unitData := unit.(models.GRPCTaskData)
		unitData.Result.ResultValue = msg.Resultvalue
		unitData.Result.ResultDes = msg.Resultdes

		inlen := len(inbuf)
		if inlen > 0 {
			unitData.RstData = make([]byte, inlen)
			copy(unitData.RstData, inbuf)
		}
		p.GrpcProcMap.ReSet(sType, unitData)
	}

	if sType != models.GRPCTYPE_SCANSTATE {
		util.FileLogs.Info("FuncGrpcScanProc %s-收到GRPC应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_SCANSTATE:
		p.FuncScanState(msg)
	case models.GRPCTYPE_OPENSCAN:
		ChanScan <- true

	case models.GRPCTYPE_CLOSESCAN:
	}

	return
}

func (p *DevScan) FuncScanState(msg *pb.Message) {
	//sType := models.GRPCTYPE_SCANSTATE
	//util.FileLogs.Info("FuncScanState %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.ScanStateReport{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	p.UpState = util.ConvertS2I(recvmsg.UpState)
	p.DnState = util.ConvertS2I(recvmsg.DnState)
}

//请求扫码
func (p *DevScan) FuncReqScanOpen(req models.ReqScanOpenData) error {
	sType := models.GRPCTYPE_OPENSCAN
	util.FileLogs.Info("FuncReqScanOpen %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = "0"
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	/*
		m := &pb.OpenScanRequest{WorkStation: req.WorkStation,
			ScanSpanTm: req.ScanSpanTm}

		return p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)
	*/
	return nil
}

//扫码结果返回处理
func (p *DevScan) FuncProcScanOpen() (string, string, models.RstScanOpenData) {
	sType := models.GRPCTYPE_OPENSCAN
	util.FileLogs.Info("FuncProcScanOpen %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	rst := models.RstScanOpenData{}
	unit := p.GrpcProcMap.Get(sType)
	if unit == nil {
		return models.GRPCRESULT_FAIL, "任务丢失", rst
	}
	unitData := unit.(models.GRPCTaskData)
	p.GrpcProcMap.Delete(sType)

	if unitData.Result.ResultValue != models.GRPCRESULT_OK {
		return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
	}

	recvmsg := &pb.OpenScanResponse{}
	if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
		return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
	}

	rst.ScanBar = recvmsg.Scanbar
	return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
}

//请求与返回扫码 封装处理
func (p *DevScan) FuncScanOpen(req models.ReqScanOpenData) (string, string, models.RstScanOpenData) {
	sType := models.GRPCTYPE_OPENSCAN
	util.FileLogs.Info("FuncScanOpen %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstScanOpenData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = "0"
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	m := &pb.OpenScanRequest{WorkStation: req.WorkStation,
		ScanSpanTm: req.ScanSpanTm}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanScan:
		{
			unit := p.GrpcProcMap.Get(reqUnit.SType)
			if unit == nil {
				return models.GRPCRESULT_FAIL, "任务丢失", rst
			}
			unitData := unit.(models.GRPCTaskData)
			p.GrpcProcMap.Delete(reqUnit.SType)

			if unitData.Result.ResultValue != models.GRPCRESULT_OK {
				return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
			}

			recvmsg := &pb.OpenScanResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			rst.ScanBar = recvmsg.Scanbar
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.SCAN_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//关闭扫码
func (p *DevScan) FuncReqScanClose(req models.ReqScanCloseData) error {
	sType := models.GRPCTYPE_CLOSESCAN
	util.FileLogs.Info("FuncReqScanOpen %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	if !p.IsGrpcConn() {
		return errors.New("云服务断连")
	}

	m := &pb.CloseScanRequest{WorkStation: req.WorkStation}

	return p.GrpcSendproc(sType, "0", m)
}

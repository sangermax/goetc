package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"encoding/json"
	"time"

	"github.com/golang/protobuf/proto"
)

//读写器设备
type DevReader struct {
	GrpcClient

	iNo          uint64
	StateGrp     models.ReaderState
	Cardtypeinfo models.ReaderCardtype

	//请求处理grpc命令
	GrpcProcMap *util.BeeMap //key stype类型,GRPCTaskData为内容体
	ChanReader  chan bool
}

func (p *DevReader) getNo() string {
	p.iNo += 1
	return util.Convert64I2S(p.iNo)
}

func (p *DevReader) InitDevReader() {
	p.iNo = 0
	p.StateGrp.DnReaderState1 = models.DEVSTATE_UNKNOWN
	p.StateGrp.DnReaderState2 = models.DEVSTATE_UNKNOWN
	p.StateGrp.UpReaderState1 = models.DEVSTATE_UNKNOWN
	p.StateGrp.UpReaderState2 = models.DEVSTATE_UNKNOWN
	p.GrpcProcMap = util.NewBeeMap()
	p.ChanReader = make(chan bool, 1)

	p.GrpcInit(models.DEVGRPCTYPE_READER, config.ConfigData["readerGrpcUrlCli"].(string), p.FuncGrpcReaderProc)
}

func (p *DevReader) FuncCheckChan(sType string) bool {
	switch sType {
	case models.GRPCTYPE_READERM1READ:
		return true
	case models.GRPCTYPE_READERM1WRITE:
		return true
	case models.GRPCTYPE_READERETCREAD:
		return true
	case models.GRPCTYPE_READERETCWRITE:
		return true
	case models.GRPCTYPE_READERETCPAY:
		return true
	case models.GRPCTYPE_READERETCBALANCE:
		return true
	}

	return false
}

func (p *DevReader) FuncGrpcReaderProc(msg *pb.Message) {
	sType := msg.Type
	inbuf := msg.Data
	sno := msg.No

	if p.FuncCheckChan(sType) {
		unit := p.GrpcProcMap.Get(sType)
		if unit == nil {
			util.FileLogs.Info("收到读写器服务应答，但是查找不到该请求，抛弃：(%s).\r\n", GetCmdDes(sType))
			return
		}
		unitData := unit.(models.GRPCTaskData)
		if unitData.Sno != sno {
			util.FileLogs.Info("收到读写器服务应答，但是sno不一致，抛弃：%s,(%s,%s).\r\n", GetCmdDes(sType), unitData.Sno, sno)
			return
		}

		unitData.Result.ResultValue = msg.Resultvalue
		unitData.Result.ResultDes = msg.Resultdes

		inlen := len(inbuf)
		if inlen > 0 {
			unitData.RstData = make([]byte, inlen)
			copy(unitData.RstData, inbuf)
		}
		p.GrpcProcMap.ReSet(sType, unitData)
	}

	if sType != models.GRPCTYPE_READERSTATE {
		util.FileLogs.Info("FuncGrpcReaderProc %s-收到GRPC服务应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_READERM1READ:
		p.ChanReader <- true
	case models.GRPCTYPE_READERM1WRITE:
		p.ChanReader <- true
	case models.GRPCTYPE_READERETCREAD:
		p.ChanReader <- true
	case models.GRPCTYPE_READERETCWRITE:
		p.ChanReader <- true
	case models.GRPCTYPE_READERETCPAY:
		p.ChanReader <- true
	case models.GRPCTYPE_READERETCBALANCE:
		p.ChanReader <- true

	case models.GRPCTYPE_READERSTATE:
		p.FuncReaderStateProc(inbuf)
	case models.GRPCTYPE_READERCARDTYPE:
		p.FuncReaderCardtypeProc(inbuf)
	case models.GRPCTYPE_READERCARDCLOSE:
		break
	}

	return
}

func (p *DevReader) FuncReqM1Read(req models.ReaderM1ReadReqData) (string, string, models.ReaderM1ReadRstData) {
	sType := models.GRPCTYPE_READERM1READ
	util.FileLogs.Info("FuncReqM1Read %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderM1ReadRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) ReaderM1WriteReqData(req models.ReaderM1WriteReqData) (string, string, models.ReaderM1WriteRstData) {
	sType := models.GRPCTYPE_READERM1WRITE
	util.FileLogs.Info("ReaderM1WriteReqData %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderM1WriteRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) ReaderETCReadReqData(req models.ReaderETCReadReqData) (string, string, models.ReaderETCReadRstData) {
	sType := models.GRPCTYPE_READERETCREAD
	util.FileLogs.Info("ReaderETCReadReqData %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderETCReadRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			util.FileLogs.Info("ReaderETCReadReqData:%s", util.ConvertByte2Hexstring(rst.Data, true))
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) ReaderETCPayReqData(req models.ReaderETCPayReqData) (string, string, models.ReaderETCPayRstData) {
	sType := models.GRPCTYPE_READERETCPAY
	util.FileLogs.Info("ReaderETCPayReqData %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderETCPayRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			util.FileLogs.Info("ReaderETCPayReqData:%s,%s,%s,%s",
				util.ConvertByte2Hexstring(rst.TradNo, true),
				util.ConvertByte2Hexstring(rst.TermTradNo, true),
				util.ConvertByte2Hexstring(rst.Tac, true),
				rst.Paytime)
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) ReaderETCBalanceReqData(req models.ReaderETCBalanceReqData) (string, string, models.ReaderETCBalanceRstData) {
	sType := models.GRPCTYPE_READERETCBALANCE
	util.FileLogs.Info("ReaderETCBalanceReqData %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderETCBalanceRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			util.FileLogs.Info("ReaderETCBalanceReqData:%d", rst.Balance)
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) ReaderClosecardReqData(req models.ReaderClosecardReqData) (string, string, models.ReaderClosecardRstData) {
	sType := models.GRPCTYPE_READERCARDCLOSE
	util.FileLogs.Info("ReaderClosecardReqData %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.ReaderClosecardRstData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "Reader服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	jsonBytes, _ := json.Marshal(req)
	m := &pb.ReaderDataRequest{Msg: string(jsonBytes)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-p.ChanReader:
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

			recvmsg := &pb.ReaderDataResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败2", rst
			}

			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(3) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

func (p *DevReader) FuncReaderStateProc(inbuf []byte) {
	//sType := models.GRPCTYPE_READERSTATE
	rst := models.ReaderState{}

	recvmsg := &pb.ReaderDataResponse{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		return
	}

	err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
	if err != nil {
		return
	}

	p.StateGrp = rst
	return
}

func (p *DevReader) FuncReaderCardtypeProc(inbuf []byte) {
	//sType := models.GRPCTYPE_READERCARDTYPE
	rst := models.ReaderCardtype{}

	recvmsg := &pb.ReaderDataResponse{}
	if err := proto.Unmarshal(inbuf, recvmsg); err != nil {
		return
	}

	err := json.Unmarshal([]byte(recvmsg.Msg), &rst)
	if err != nil {
		return
	}

	//通知有卡
	p.Cardtypeinfo = rst
	ChanCard <- true

	return
}

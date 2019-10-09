package main

import (
	"FTC/config"
	"FTC/pb"
	"FTC/util"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"FTC/models"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//读写器服务
type ReaderService struct {
	ID    int
	State int

	Psamtermid    string
	Cardid        string
	pReaderInfObj *ReaderInf
}

func (p *ReaderService) GetReaderDes() string {
	switch p.ID {
	case models.DnReader1:
		return "down 支架读写器"
	case models.DnReader2:
		return "down ETC读写器"
	case models.UpReader1:
		return "up 支架读写器"
	case models.UpReader2:
		return "up ETC读写器"
	default:
		return "未知:" + util.ConvertI2S(p.ID)
	}
}

func (p *ReaderService) InitReader(ID int, comdes string, islot int) bool {
	p.ID = ID
	p.State = models.DEVSTATE_UNKNOWN
	p.Cardid = ""
	p.Psamtermid = ""
	p.pReaderInfObj = new(ReaderInf)

	//打开读写器
	icom := util.GetDigitsByStr(comdes) + 1
	if p.pReaderInfObj.FuncJT_OpenReader(comdes, icom, 115200) != 0 {
		util.FileLogs.Info("(%d-%s)读卡器OpenReader失败", p.ID, p.GetReaderDes())
		p.State = models.DEVSTATE_TROUBLE
		return false
	}
	util.FileLogs.Info("(%d-%s)读卡器OpenReader suc", p.ID, p.GetReaderDes())

	//psam卡复位
	rlt, _, _ := p.pReaderInfObj.FuncJT_SamReset()
	if rlt != 0 {
		util.FileLogs.Info("(%d-%s)读卡器SamReset失败", p.ID, p.GetReaderDes())
		p.State = models.DEVSTATE_TROUBLE
		return false
	}
	util.FileLogs.Info("(%d-%s)读卡器SamReset suc", p.ID, p.GetReaderDes())

	//选择卡槽
	rlt = p.pReaderInfObj.FuncJT_SelectPSAMSlot(islot)
	if rlt != 0 {
		util.FileLogs.Info("(%d-%s)读卡器SelectPSAMSlot失败", p.ID, p.GetReaderDes())
		p.State = models.DEVSTATE_TROUBLE
		return false
	}
	util.FileLogs.Info("(%d-%s)读卡器FuncJT_SelectPSAMSlot suc", p.ID, p.GetReaderDes())

	//获取psam卡号
	rlt, out1 := p.pReaderInfObj.FuncJT_SamGetTermID()
	if rlt != 0 {
		util.FileLogs.Info("(%d-%s)读卡器GetTermID失败", p.ID, p.GetReaderDes())
		p.State = models.DEVSTATE_TROUBLE
		return false
	}
	p.Psamtermid = util.ConvertByte2Hexstring(out1, false)
	util.FileLogs.Info("(%d-%s)读卡器FuncJT_SamGetTermID suc:%s", p.ID, p.GetReaderDes(), p.Psamtermid)

	p.State = models.DEVSTATE_OK

	go p.goDetecCard()
	return true
}

//实时监测卡片
func (p *ReaderService) goDetecCard() {
	for {
		//监测到该读写器已经在处理卡，则中断监测卡片
		if p.ID == g_ReaderIdBusy {
			util.MySleep_ms(500)
			continue
		}

		if p.pReaderInfObj.FuncJT_OpenCard() != 0 {
			util.MySleep_ms(500)
			continue
		}

		iCardtype := p.pReaderInfObj.FuncJT_GetCardType()

		//主动上报监测到卡片
		sCardid := ""
		switch iCardtype {
		case models.MifareS50, models.MifareS70:
			{
				//读卡号
				out := p.pReaderInfObj.FuncJT_GetCardSer()
				if out == nil {
					continue
				}

				sCardid = util.ConvertByte2Hexstring(out, false)
			}
		case models.MifarePro, models.MifareProX:
			{
				//读卡号
				rlt, out := p.pReaderInfObj.FuncJT_ProGetCardID()
				if rlt != 0 {
					continue
				}

				sCardid = util.ConvertByte2Hexstring(out, false)
			}
		}

		g_ReaderIdBusy = p.ID
		if true { // sCardid != "" && sCardid != p.Cardid {
			p.Cardid = sCardid

			var info models.ReaderCardtype
			info.Cardtype = iCardtype
			info.PsamTermId = p.Psamtermid
			info.Cardid = p.Cardid

			jsonBytes, _ := json.Marshal(info)
			m := &pb.ReaderDataResponse{Msg: string(jsonBytes)}
			ReaderSrvGrpcSendproc(g_readerstream, models.GRPCTYPE_READERCARDTYPE, "0", m, getResult(models.GRPCRESULT_OK, ""))
		}

		util.MySleep_ms(1000)
	}
}

//读写器管理map
var gReaderMap map[int]ReaderService = map[int]ReaderService{}
var g_ReaderIdBusy int
var g_bConn bool
var g_readerstream pb.GrpcMsg_CommuniteServer

func initReaders() {
	g_ReaderIdBusy = models.ReaderUnknown

	var r1 ReaderService
	r1.InitReader(models.DnReader1, config.ConfigData["readerDnCom1"].(string), 1)
	gReaderMap[models.DnReader1] = r1
	/*
		var r2 ReaderService
		r2.InitReader(models.DnReader2, config.ConfigData["readerDnCom2"].(string), 1)
		gReaderMap[models.DnReader2] = r2
	*/
}

func getResult(str1 string, str2 string) models.ResultInfo {
	var result models.ResultInfo
	result.ResultValue = str1
	result.ResultDes = str2

	return result
}

func main() {
	util.FileLogs.Info("读卡器服务启动中...")
	config.InitConfig("../../conf/config.conf")

	//TestReader()
	initReaders()

	/////////////////////////////////////grpc
	g_bConn = false
	sAddr := config.ConfigData["readerGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("读卡器服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("读卡器服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &ReaderServer{})
	s.Serve(lis)
}

type ReaderServer struct {
}

func (p *ReaderServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道读卡器GRPC已连接")

	g_readerstream = stream
	g_bConn = true
	//重新初始化
	g_ReaderIdBusy = models.ReaderUnknown

	go func() {
		for {
			if !g_bConn {
				break
			}

			//发送设备状态
			var info models.ReaderState
			info.DnReaderState1 = gReaderMap[models.DnReader1].State
			info.DnReaderState2 = gReaderMap[models.DnReader2].State
			info.UpReaderState1 = models.DEVSTATE_UNKNOWN
			info.UpReaderState2 = models.DEVSTATE_UNKNOWN

			jsonBytes, _ := json.Marshal(info)
			m := &pb.ReaderDataResponse{Msg: string(jsonBytes)}
			ReaderSrvGrpcSendproc(stream, models.GRPCTYPE_READERSTATE, "0", m, getResult(models.GRPCRESULT_OK, ""))

			util.MySleep_s(5)
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			util.FileLogs.Info("Communite read done")
			g_bConn = false
			break
		}

		if err != nil {
			util.FileLogs.Info("Communite ERR:%s", err.Error())
			g_bConn = false
			break
		}

		util.FileLogs.Info("收到来自车道读卡器信息")
		fmt.Println(in)

		switch in.Type {
		case models.GRPCTYPE_READERM1READ:
			FuncM1Read(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERM1WRITE:
			FuncM1Write(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERETCREAD:
			FuncETCRead(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERETCPAY:
			FuncETCPay(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERETCBALANCE:
			FuncETCBalance(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERCARDTYPE:
			FuncCardtype(in.Type, in.No, in.Data)
			break
		case models.GRPCTYPE_READERCARDCLOSE:
			FuncCardClose(in.Type, in.No, in.Data)
			break

		}
	}

	return nil
}

func ReaderSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("reader GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			util.FileLogs.Info("reader GrpcSendproc ")
			fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			util.FileLogs.Info("reader GrpcSendproc")
			fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

func FuncM1Read(sType string, no string, inbuf []byte) {
	rst := models.ReaderM1ReadRstData{}
	m := new(pb.ReaderDataResponse)

	req := &pb.ReaderDataRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		util.FileLogs.Info("FuncM1Read err:%s.\r\n", err.Error())
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误1"))
		return
	}

	var info models.ReaderM1ReadReqData
	err := json.Unmarshal([]byte(req.Msg), &info)
	if err != nil {
		util.FileLogs.Info("FuncM1Read 解析失败")
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误2"))
		return
	}

	r1, r2 := gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_ReadFile(info.FileID, info.KeyID, info.FileType, info.Addr, info.Length)

	rst.Result = r1
	if r2 != nil {
		rst.Data = make([]byte, len(r2))
		copy(rst.Data, r2)
	}

	jsonBytes, _ := json.Marshal(rst)
	m.Msg = string(jsonBytes)
	ReaderSrvGrpcSendproc(g_readerstream, models.GRPCTYPE_READERM1READ, "0", m, getResult(models.GRPCRESULT_OK, ""))
}

func FuncM1Write(sType string, no string, inbuf []byte) {
	rst := models.ReaderM1WriteRstData{}
	m := new(pb.ReaderDataResponse)

	req := &pb.ReaderDataRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		util.FileLogs.Info("FuncM1Write err:%s.\r\n", err.Error())
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误1"))
		return
	}

	var info models.ReaderM1WriteReqData
	err := json.Unmarshal([]byte(req.Msg), &info)
	if err != nil {
		util.FileLogs.Info("FuncM1Write 解析失败")
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误2"))
		return
	}

	r1 := gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_WriteFile(info.FileID, info.KeyID, info.FileType, info.Addr, info.Length, info.Data)

	rst.Result = r1
	jsonBytes, _ := json.Marshal(rst)
	m.Msg = string(jsonBytes)
	ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_OK, ""))
}

func FuncETCRead(sType string, no string, inbuf []byte) {
	rst := models.ReaderETCReadRstData{}
	m := new(pb.ReaderDataResponse)

	req := &pb.ReaderDataRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		util.FileLogs.Info("FuncETCRead err:%s.\r\n", err.Error())
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误1"))
		return
	}

	var info models.ReaderETCReadReqData
	err := json.Unmarshal([]byte(req.Msg), &info)
	if err != nil {
		util.FileLogs.Info("FuncETCRead 解析失败")
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误2"))
		return
	}

	r1, r2 := gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_ProReadFile(info.FileID, info.Length)

	rst.Result = r1
	if r2 != nil {
		rst.Data = make([]byte, len(r2))
		copy(rst.Data, r2)
	}
	jsonBytes, _ := json.Marshal(rst)
	m.Msg = string(jsonBytes)
	ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_OK, ""))
}

func FuncETCPay(sType string, no string, inbuf []byte) {
	rst := models.ReaderETCPayRstData{}
	m := new(pb.ReaderDataResponse)

	req := &pb.ReaderDataRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		util.FileLogs.Info("FuncETCPay err:%s.\r\n", err.Error())
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误1"))
		return
	}

	var info models.ReaderETCPayReqData
	err := json.Unmarshal([]byte(req.Msg), &info)
	if err != nil {
		util.FileLogs.Info("FuncETCPay 解析失败")
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误2"))
		return
	}

	r1, r2, r3, r4, r5 := gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_ProDecrement(info.Money, info.Data, info.Paytime)
	rst.Result = r1
	if r2 != nil {
		rst.TradNo = make([]byte, len(r2))
		copy(rst.TradNo, r2)
	}

	if r3 != nil {
		rst.TermTradNo = make([]byte, len(r3))
		copy(rst.TermTradNo, r3)
	}

	rst.Paytime = r4

	if r5 != nil {
		rst.Tac = make([]byte, len(r5))
		copy(rst.Tac, r5)
	}

	jsonBytes, _ := json.Marshal(rst)
	m.Msg = string(jsonBytes)
	ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_OK, ""))
}

func FuncETCBalance(sType string, no string, inbuf []byte) {
	rst := models.ReaderETCBalanceRstData{}
	m := new(pb.ReaderDataResponse)

	req := &pb.ReaderDataRequest{}
	if err := proto.Unmarshal(inbuf, req); err != nil {
		util.FileLogs.Info("FuncETCBalance err:%s.\r\n", err.Error())
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误1"))
		return
	}

	var info models.ReaderETCBalanceReqData
	err := json.Unmarshal([]byte(req.Msg), &info)
	if err != nil {
		util.FileLogs.Info("FuncETCBalance 解析失败")
		ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_FAIL, "请求参数错误2"))
		return
	}

	r1, r2 := gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_ProQueryBalance()

	rst.Result = r1
	rst.Balance = r2
	jsonBytes, _ := json.Marshal(rst)
	m.Msg = string(jsonBytes)
	ReaderSrvGrpcSendproc(g_readerstream, sType, no, m, getResult(models.GRPCRESULT_OK, ""))
}

func FuncCardtype(sType string, no string, inbuf []byte) {

}

func FuncCardClose(sType string, no string, inbuf []byte) {
	gReaderMap[g_ReaderIdBusy].pReaderInfObj.FuncJT_CloseCard()
	g_ReaderIdBusy = models.ReaderUnknown
}

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

type DevCloud struct {
	GrpcClient

	iNo uint64

	//请求处理grpc命令
	GrpcProcMap *util.BeeMap //key stype类型,GRPCTaskData为内容体

}

func (p *DevCloud) getNo() string {
	p.iNo += 1
	return util.Convert64I2S(p.iNo)
}

func (p *DevCloud) InitDevCloud() {
	p.GrpcProcMap = util.NewBeeMap()
	p.GrpcInit2(models.DEVGRPCTYPE_CLOUD, config.ConfigData["cloudGrpcUrlCli"].(string), p.FuncSendGRPCInit, p.FuncGrpcCloudProc)
}

func (p *DevCloud) FuncSendGRPCInit() {
	if !p.IsGrpcConn() {
		return
	}

	staid := util.ConvertI2S(config.ConfigData["stationid"].(int))
	laneid := util.ConvertI2S(config.ConfigData["laneid"].(int))
	key := staid + "_" + laneid
	m := &pb.InitRequest{LaneKey: key}
	p.GrpcSendproc(models.GRPCTYPE_Init, p.getNo(), m)
}

func (p *DevCloud) FuncCheckChan(sType string) bool {
	switch sType {
	case models.GRPCTYPE_FEE:
		return true
	case models.GRPCTYPE_CHECKPAY:
		return true
	case models.GRPCTYPE_PAY:
		return true
	case models.GRPCTYPE_CARPATH:
		return true
	case models.GRPCTYPE_CHECKFREEFLOW:
		return true
	case models.GRPCTYPE_NOTIFYFREEFLOW:
		return true
	}

	return false
}

func (p *DevCloud) FuncGrpcCloudProc(msg *pb.Message) {
	sType := msg.Type
	inbuf := msg.Data
	sno := msg.No

	if p.FuncCheckChan(sType) {
		unit := p.GrpcProcMap.Get(sType)
		if unit == nil {
			util.FileLogs.Info("收到云平台服务应答，但是查找不到该请求，抛弃：(%s).\r\n", GetCmdDes(sType))
			return
		}
		unitData := unit.(models.GRPCTaskData)
		if unitData.Sno != sno {
			util.FileLogs.Info("收到云平台服务应答，但是sno不一致，抛弃：%s,(%s,%s).\r\n", GetCmdDes(sType), unitData.Sno, sno)
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

	util.FileLogs.Info("FuncGrpcCloudProc %s-收到GRPC服务应答 开始处理.\r\n", GetCmdDes(sType))
	switch sType {
	case models.GRPCTYPE_Init:

	case models.GRPCTYPE_FEE:
		ChanFee <- true

	case models.GRPCTYPE_CHECKPAY:
		ChanPayCheck <- true

	case models.GRPCTYPE_PAY:
		ChanPay <- true

	case models.GRPCTYPE_CARPATH:
		ChanCarpath <- true

	case models.GRPCTYPE_CHECKFREEFLOW:
		ChanChkFreeflow <- true

	case models.GRPCTYPE_NOTIFYFREEFLOW:
		ChanNotifyFreeflow <- true

	case models.GRPCTYPE_EXITFLOW:
		p.FuncProcExitflow(msg)
	case models.GRPCTYPE_EXITIMG:
		p.FuncProcExitimg(msg)
	case models.GRPCTYPE_DEVSTATE:

	case models.GRPCTYPE_CONTROL:
		p.FuncProcControl(msg)
	}

	return
}

//请求计费
func (p *DevCloud) FuncReqFee(req models.ReqCalcFeeData) (string, string, models.RstCalcFeeData) {
	sType := models.GRPCTYPE_FEE
	util.FileLogs.Info("FuncReqFee %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstCalcFeeData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	m := &pb.FeeRequest{Vehplate: req.Vehplate,
		Vehclass:   req.Vehclass,
		Entrystaid: req.Entrystaid,
		Exitstaid:  req.Exitstaid,
		Flagstaid:  req.Flagstaid}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanFee:
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

			recvmsg := &pb.FeeResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			rst.Toll = recvmsg.Toll
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//请求车牌付校验
func (p *DevCloud) FuncReqPlatePayCheck(req models.ReqCheckPlatepayData) (string, string, models.RstCheckPlatepayData) {
	sType := models.GRPCTYPE_CHECKPAY
	util.FileLogs.Info("FuncReqPlatePayCheck %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstCheckPlatepayData{}

	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	m := &pb.PlatePayCheckRequest{Vehplate: req.Vehplate, Vehclass: req.Vehclass, TID: req.TID}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanPayCheck:
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

			recvmsg := &pb.PlatePayCheckResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			rst.CheckResult = recvmsg.CheckResult
			rst.Vehclass = recvmsg.Vehclass
			rst.Vehplate = recvmsg.Vehplate
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//请求支付
func (p *DevCloud) FuncReqPay_test(req models.ReqPayData) (string, string, models.RstPayData) {
	sType := models.GRPCTYPE_PAY
	util.FileLogs.Info("FuncReqPay %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstPayData{}
	rst.Paytype = models.PAYTYPE_SCAN
	rst.Paymethod = "1"
	rst.Goodsno = "111"
	rst.Tradeno = "111"
	rst.Toll = "100"
	rst.Paytime = util.GetNow(false)

	return models.GRPCRESULT_OK, "", rst

}

func (p *DevCloud) FuncReqPay(req models.ReqPayData) (string, string, models.RstPayData) {
	sType := models.GRPCTYPE_PAY
	util.FileLogs.Info("FuncReqPay %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstPayData{}

	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)
	//,req.Vehplate
	m := &pb.PayRequest{Operatorid: req.Operatorid,
		Shiftid:   req.Shiftid,
		TransDate: req.TransDate,
		Paytype:   req.Paytype,
		PayCode:   req.PayCode,
		Vehplate:  "苏D953MQ",
		Toll:      req.Toll,
		Transtime: req.Transtime}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanPay:
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

			recvmsg := &pb.PayResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			rst.Paytype = recvmsg.Paytype
			rst.Paymethod = recvmsg.Paymethod
			rst.Goodsno = recvmsg.Goodsno
			rst.Tradeno = recvmsg.Tradeno
			rst.Toll = recvmsg.Toll
			rst.Paytime = recvmsg.Paytime
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//发送出口流水
func (p *DevCloud) FuncReqExitflow(req models.ReqExitFlowData) bool {
	sType := models.GRPCTYPE_EXITFLOW
	util.FileLogs.Info("FuncReqExitflow %s-GRPC车道发送流水:%s.\r\n", GetCmdDes(sType), req.FlowNo)

	m := &pb.ExitFlowRequest{
		Vehclass:       req.Vehclass,
		Vehplate:       req.Vehplate,
		ExitNetwork:    req.ExitNetwork,
		ExitStationid:  req.ExitStationid,
		ExitLandid:     req.ExitLandid,
		ExitOperator:   req.ExitOperator,
		ExitShiftdate:  req.ExitShiftdate,
		ExitShift:      req.ExitShift,
		Paytype:        req.Paytype,
		Paymethod:      req.Paymethod,
		Payid:          req.Payid,
		Transtime:      req.Transtime,
		Paytime:        req.Paytime,
		Toll:           req.Toll,
		Goodsno:        req.Goodsno,
		Tradeno:        req.Tradeno,
		PrintFlag:      req.PrintFlag,
		Termid:         req.Psamtermid,
		Termno:         req.Psamtradno,
		Tac:            req.Tac,
		CardTradeNo:    req.Cardtradno,
		CardId:         req.CardId,
		FlowNo:         req.FlowNo,
		TransFrom:      req.TransFrom,
		StationMode:    req.StationMode,
		EntryNetwork:   req.EntryNetwork,
		EntryStationId: req.EntryStationId,
		EntryLaneId:    req.EntryLaneId,
		EntryOperator:  req.EntryOperator,
		EntryShift:     req.EntryShift,
		EntryTime:      req.EntryTime,
		FlagStationid:  req.FlagStationid,
		FlagTime:       req.FlagTime,
		RegVehclass:    req.RegVehclass,
		RegVehplate:    req.RegVehplate,
		AntennaID:      req.AntennaID,
		TID:            req.TID,
		VehColor:       req.VehColor,
		TransMemo:      req.TransMemo}
	p.GrpcSendproc(sType, p.getNo(), m)

	//等待应答，如果有应答，则删除，如果没有应答，则保留至
	select {
	case <-ChanExitflow:
		return true
	case <-time.After(time.Duration(3) * time.Second):
		{
			util.FileLogs.Info("FuncReqExitflow %s-GRPC服务超时未应答.\r\n", GetCmdDes(sType))

			filenm := req.FlowNo
			nowpath := models.FLOWDIR + filenm
			newpath := models.FAIL_FLOWDIR + filenm
			util.MoveFile(nowpath, newpath)
		}
	}

	return false
}

func (p *DevCloud) FuncProcExitflow(msg *pb.Message) {
	sType := models.GRPCTYPE_EXITFLOW
	util.FileLogs.Info("FuncProcExitflow %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	if msg.Resultvalue != models.GRPCRESULT_OK {
		return
	}

	recvmsg := &pb.ExitFlowResponse{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	//删除文件
	filenm := recvmsg.FlowNo
	util.FileLogs.Info("FuncProcExitflow,filenm:%s.\r\n", filenm)

	if !util.RemoveFile(models.FLOWDIR + filenm) {
		util.RemoveFile(models.FAIL_FLOWDIR + filenm)
	}

	ChanExitflow <- true
}

//发送出口图片
func (p *DevCloud) FuncReqExitimg(req models.ReqExitImgData) bool {
	sType := models.GRPCTYPE_EXITIMG
	util.FileLogs.Info("FuncReqExitimg %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	adds := &pb.ExitImgAdds{
		ExitNetwork:   req.AddsInfo.ExitNetwork,
		ExitStationid: req.AddsInfo.ExitStationid,
		ExitLandid:    req.AddsInfo.ExitLandid,
		ExitOperator:  req.AddsInfo.ExitOperator,
		ExitShiftdate: req.AddsInfo.ExitShiftdate,
		ExitShift:     req.AddsInfo.ExitShift,
		Transtime:     req.AddsInfo.Transtime,
		FlowNo:        req.AddsInfo.FlowNo}

	carbaseinfo := &pb.CarInfo{
		Lpn:          req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Lpn,          //车牌号码
		LpnScore:     req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnScore,     //车牌分数
		LpnColor:     req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnColor,     //车牌颜色
		LpnposLeft:   req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnposLeft,   //车牌位置坐标left
		LpnposTop:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnposTop,    //车牌位置坐标top
		LpnposRight:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnposRight,  //车牌位置坐标right
		LpnposBottom: req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.LpnposBottom, //车牌位置坐标bottom
		Color:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Color,        //车身颜色
		ColorScore:   req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.ColorScore,   //车身颜色分数
		Brand:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand,        //车辆品牌
		Brand0:       req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand0,       //车辆品牌0
		Brand1:       req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand1,       //车辆品牌1
		Brand2:       req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand2,       //车辆品牌2
		Brand3:       req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Brand3,       //车辆品牌3
		Vehtype:      req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Vehtype,      //车型
		VehClass:     req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.VehClass,     //车辆类型
		Type0:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Type0,        //车辆类型0
		Type1:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Type1,        //车辆类型1
		Type2:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Type2,        //车辆类型2
		Type3:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Type3,        //车辆类型3
		Subbrand:     req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Subbrand,     //车辆子品牌
		Subbrand0:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Subbrand0,    //车辆子品牌0
		Subbrand1:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Subbrand1,    //车辆子品牌1
		Subbrand2:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Subbrand2,    //车辆子品牌2
		Subbrand3:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Subbrand3,    //车辆子品牌3
		Year:         req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Year,         //车辆年份
		Year0:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Year0,        //车辆年份0
		Year1:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Year1,        //车辆年份1
		Year2:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Year2,        //车辆年份2
		Year3:        req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Year3,        //车辆年份3
		BrandScore:   req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.BrandScore,   //品牌分数
		BrandScore0:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.BrandScore0,  //品牌分数0
		BrandScore1:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.BrandScore1,  //品牌分数1
		BrandScore2:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.BrandScore2,  //品牌分数2
		BrandScore3:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.BrandScore3,  //品牌分数3
		Pose:         req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.Pose,         //车辆整体位置
		CarposLeft:   req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.CarposLeft,   //车辆坐标left
		CarposTop:    req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.CarposTop,    //车辆坐标top
		CarposRight:  req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.CarposRight,  //车辆坐标right
		CarposBottom: req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.CarposBottom, //车辆坐标bottom
		CarRectScore: req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.CarRectScore, //车辆位置分数
		ImgQuality:   req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.ImgQuality,   //车辆或车牌图像质量
		IDNumber:     req.RstVehPlateInfo.VehicleInfo.CarBaseInfo.IDNumber,     //Json文件唯一值标志
	}
	carfeature := &pb.CarFeature{
		GJ:  req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.GJ,  //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
		NJB: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.NJB, //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
		TC:  req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.TC,  //天窗	同GJ字段
		AQD: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.AQD, //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
		DDH: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.DDH, //开车打电话	同AQD字段
		ZYB: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.ZYB, //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
		CZH: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.CZH, //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
		CRZ: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.CRZ, //出入证	同GJ字段
		XSB: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.XSB, //新手标	同GJ字段
		JSY: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.JSY, //驾驶员	同AQD字段
		LT:  req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.LT,  //轮胎	同CZH字段
		CD:  req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.CD,  //车灯	同CZH字段
		JSC: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.JSC, //驾驶窗	同CZH字段
		HSJ: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.HSJ, //后视镜	同CZH字段
		XLJ: req.RstVehPlateInfo.VehicleInfo.CarFeatureInfo.XLJ, //行李架	同CZH字段
	}
	car := &pb.Car{
		CarBaseInfo:    carbaseinfo,
		CarFeatureInfo: carfeature}

	i := 0
	icnt := 0
	otherBike := new(pb.TotalBike)
	otherBike.BikeCount = req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeCount
	icnt = util.ConvertS2I(otherBike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherBike.BikeRsds = append(otherBike.BikeRsds, &pb.BikeInfo{
			Posleft:   req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posleft,
			Postop:    req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Postop,
			Posright:  req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posright,
			Posbottom: req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posbottom,
			Score:     req.RstVehPlateInfo.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Score})
	}

	otherMotoBike := new(pb.TotalBike)
	otherMotoBike.BikeCount = req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeCount
	icnt = util.ConvertS2I(otherMotoBike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherMotoBike.BikeRsds = append(otherMotoBike.BikeRsds, &pb.BikeInfo{
			Posleft:   req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posleft,
			Postop:    req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Postop,
			Posright:  req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posright,
			Posbottom: req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posbottom,
			Score:     req.RstVehPlateInfo.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Score})
	}

	otherTribike := new(pb.TotalBike)
	otherTribike.BikeCount = req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeCount
	icnt = util.ConvertS2I(otherTribike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherTribike.BikeRsds = append(otherTribike.BikeRsds, &pb.BikeInfo{
			Posleft:   req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posleft,
			Postop:    req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Postop,
			Posright:  req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posright,
			Posbottom: req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posbottom,
			Score:     req.RstVehPlateInfo.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Score})
	}

	others := &pb.OtherVehicleInfo{
		TotalMotobikeInfo: otherMotoBike,
		TotalTribikeInfo:  otherTribike,
		TotalBikeInfo:     otherBike}

	faces := new(pb.TotalFaceInfo)
	faces.FaceCount = req.RstVehPlateInfo.VehicleFaceInfo.FaceCount
	icnt = util.ConvertS2I(faces.FaceCount)
	for i = 0; i < icnt; i += 1 {
		faces.FaceRsds = append(faces.FaceRsds, &pb.FaceInfo{
			FaceRectScore:  req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceRectScore,
			FaceposLeft:    req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceposLeft,
			FaceposTop:     req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceposTop,
			FaceposRight:   req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceposRight,
			FaceposBottom:  req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceposBottom,
			FacekeyPoint1X: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1X,
			FacekeyPoint1Y: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1Y,
			FacekeyPoint2X: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2X,
			FacekeyPoint2Y: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2Y,
			FacekeyPoint3X: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3X,
			FacekeyPoint3Y: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3Y,
			FacekeyPoint4X: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4X,
			FacekeyPoint4Y: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4Y,
			FacekeyPoint5X: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5X,
			FacekeyPoint5Y: req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5Y,
			FaceFeature:    req.RstVehPlateInfo.VehicleFaceInfo.FaceRsds[i].FaceFeature})
	}

	img := &pb.Img{
		PlateImgBuffer: req.RstVehPlateInfo.VehicleImg.PlateImgBuffer,
		CarImgBuffer:   req.RstVehPlateInfo.VehicleImg.CarImgBuffer}

	plateinfo := &pb.RstVehRecognize{
		VehicleInfo:      car,
		VehicleOtherInfo: others,
		VehicleFaceInfo:  faces,
		VehicleImg:       img}

	m := &pb.ExitImgRequest{
		AddsInfo:        adds,
		RstVehPlateInfo: plateinfo}

	p.GrpcSendproc(sType, p.getNo(), m)

	//等待应答，如果有应答，则删除，如果没有应答，则保留至
	select {
	case <-ChanExitimg:
		return true
	case <-time.After(time.Duration(3) * time.Second):
		{
			util.FileLogs.Info("FuncReqExitimg %s-GRPC服务超时未应答.\r\n", GetCmdDes(sType))

			filenm := adds.FlowNo
			nowpath := models.IMGDIR + filenm
			newpath := models.FAIL_IMGDIR + filenm
			util.MoveFile(nowpath, newpath)
		}

	}

	return false
}

func (p *DevCloud) FuncProcExitimg(msg *pb.Message) {
	sType := models.GRPCTYPE_EXITIMG
	util.FileLogs.Info("FuncProcExitimg %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	if msg.Resultvalue != models.GRPCRESULT_OK {
		return
	}

	recvmsg := &pb.ExitImgResponse{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	//删除文件
	filenm := recvmsg.FlowNo
	if !util.RemoveFile(models.IMGDIR + filenm) {
		util.RemoveFile(models.FAIL_IMGDIR + filenm)
	}

	ChanExitimg <- true
}

//发送设备状态
func (p *DevCloud) FuncReqDevState(req models.AutoDevStateData) {
	m := &pb.DevStateRequest{
		ExitNetwork:   req.ExitNetwork,
		ExitStationid: req.ExitStationid,
		ExitLandid:    req.ExitLandid,
		EquipmentDev:  req.EquipmentDev,
		PrinterDev1:   req.PrinterDev1,
		PrinterDev2:   req.PrinterDev2,
		PlateDev:      req.PlateDev,
		Coil1:         req.Coil1,
		Coil2:         req.Coil2,
		Scan1:         req.Scan1,
		Scan2:         req.Scan2}
	p.GrpcSendproc(models.GRPCTYPE_DEVSTATE, p.getNo(), m)
}

func (p *DevCloud) FuncProcControl(msg *pb.Message) {
	sType := models.GRPCTYPE_CONTROL
	util.FileLogs.Info("FuncProcControl %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.ControlRequest{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	var rlt models.ResultInfo
	rlt.ResultValue = models.GRPCRESULT_OK
	rlt.ResultDes = ""

	m := &pb.ControlResponse{}
	p.GrpcSendprocWithResult(sType, msg.GetNo(), m, rlt)
}

//请求通行信息
func (p *DevCloud) FuncReqCarpath(req models.ReqCarpathData) (string, string, models.RstCarpathData) {
	sType := models.GRPCTYPE_CARPATH
	util.FileLogs.Info("FuncReqCarpath %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstCarpathData{}
	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	m := &pb.CarpathRequest{Vehplate: req.Vehplate, Vehclass: req.Vehclass}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanCarpath:
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

			recvmsg := &pb.CarpathResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			rst.EntryNetwork = recvmsg.EntryNetwork
			rst.EntryStationId = recvmsg.EntryStationId
			rst.EntryLaneId = recvmsg.EntryLaneId
			rst.EntryOperator = recvmsg.EntryOperator
			rst.EntryShift = recvmsg.EntryShift
			rst.EntryTime = recvmsg.EntryTime
			rst.FlagStationid = recvmsg.FlagStationid
			rst.HasCard = recvmsg.HasCard
			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//自由流校验
func (p *DevCloud) FuncChkFreeflow(req models.ReqChkFreeflowData) (string, string, models.RstChkFreeflowData) {
	sType := models.GRPCTYPE_CHECKFREEFLOW
	util.FileLogs.Info("FuncChkFreeflow %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstChkFreeflowData{}

	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	reqjson, _ := json.Marshal(req)
	m := &pb.ChkFreeflowRequest{Msgbody: string(reqjson)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanChkFreeflow:
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

			recvmsg := &pb.ChkFreeflowResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msgbody), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析json失败", rst
			}

			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

//自由流校验
func (p *DevCloud) FuncNotifyFreeflow(req models.ReqNotifyFreeflowData) (string, string, models.RstNotifyFreeflowData) {
	sType := models.GRPCTYPE_NOTIFYFREEFLOW
	util.FileLogs.Info("FuncNotifyFreeflow %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	rst := models.RstNotifyFreeflowData{}

	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "云服务断连", rst
	}

	var reqUnit models.GRPCTaskData
	reqUnit.Reqtime = util.GetNow(false)
	reqUnit.Sno = p.getNo()
	reqUnit.SType = sType
	reqUnit.ReqData = req
	p.GrpcProcMap.ReSet(reqUnit.SType, reqUnit)

	reqjson, _ := json.Marshal(req)
	m := &pb.NotifyFreeflowRequest{Msgbody: string(reqjson)}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	select {
	case <-ChanNotifyFreeflow:
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

			recvmsg := &pb.NotifyFreeflowResponse{}
			if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
				return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
			}

			err := json.Unmarshal([]byte(recvmsg.Msgbody), &rst)
			if err != nil {
				return models.GRPCRESULT_FAIL, "解析json失败", rst
			}

			return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
		}

	case <-time.After(time.Duration(models.NET_TIMEOUT_SEC) * time.Second):
		p.GrpcProcMap.Delete(reqUnit.SType)
		return models.GRPCRESULT_FAIL, "超时无应答", rst
	}
}

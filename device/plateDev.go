package device

import (
	"FTC/config"
	"FTC/models"
	"FTC/pb"
	"FTC/util"

	"github.com/golang/protobuf/proto"
)

//车脸识别设备
type DevPlate struct {
	GrpcClient
	State int

	//请求处理grpc命令
	GrpcProcMap *util.BeeMap //key stype类型,GRPCTaskData为内容体
}

func (p *DevPlate) InitDevPlate() {
	p.State = models.DEVSTATE_UNKNOWN
	p.GrpcProcMap = util.NewBeeMap()

	p.GrpcInit(models.DEVGRPCTYPE_PLATE, config.ConfigData["plateGrpcUrlCli"].(string), p.FuncGrpcPlateProc)
}

func (p *DevPlate) FuncCheckChan(sType string) bool {
	switch sType {
	case models.GRPCTYPE_PLATE:
		return true
	}

	return false
}

func (p *DevPlate) FuncGrpcPlateProc(msg *pb.Message) {
	sType := msg.Type
	inbuf := msg.Data

	if p.FuncCheckChan(sType) {
		var unitData models.GRPCTaskData
		unitData.Reqtime = util.GetNow(false)
		unitData.SType = sType
		unitData.Result.ResultValue = msg.Resultvalue
		unitData.Result.ResultDes = msg.Resultdes

		inlen := len(inbuf)
		if inlen > 0 {
			unitData.RstData = make([]byte, inlen)
			copy(unitData.RstData, inbuf)
		}
		p.GrpcProcMap.ReSet(sType, unitData)
	}

	if sType != models.GRPCTYPE_PLATESTATE {
		util.FileLogs.Info("FuncGrpcPlateProc %s-收到GRPC应答 开始处理.\r\n", GetCmdDes(sType))
	}

	switch sType {
	case models.GRPCTYPE_PLATESTATE:
		p.FuncPlateState(msg)
	case models.GRPCTYPE_PLATE:
		ChanPlate <- true
	}

	return
}

func (p *DevPlate) FuncPlateState(msg *pb.Message) {
	//sType := models.GRPCTYPE_PLATESTATE
	//util.FileLogs.Info("FuncPlateState %s-GRPC服务应答.\r\n", GetCmdDes(sType))

	recvmsg := &pb.PlateStateReport{}
	if err := proto.Unmarshal(msg.Data, recvmsg); err != nil {
		return
	}

	p.State = util.ConvertS2I(recvmsg.State)
}

/*
//车脸识别请求
func (p *DevPlate) FuncCarFaceReqProc() (string, string) {
	sType := models.GRPCTYPE_PLATE
	util.FileLogs.Info("FuncCarFaceReqProc %s-GRPC车道请求.\r\n", GetCmdDes(sType))

	if !p.IsGrpcConn() {
		return models.GRPCRESULT_FAIL, "车辆识别服务断连"
	}



	m := &pb.ReqVehRecognize{}
	p.GrpcSendproc(reqUnit.SType, reqUnit.Sno, m)

	return models.GRPCRESULT_OK, ""
}
*/

//车脸识别返回
func (p *DevPlate) FuncCarFaceRstProc() (string, string, models.RstVehRecognizeInfo) {
	sType := models.GRPCTYPE_PLATE
	util.FileLogs.Info("FuncCarFaceRstProc %s-GRPC车道处理.\r\n", GetCmdDes(sType))

	rst := models.RstVehRecognizeInfo{}

	unit := p.GrpcProcMap.Get(sType)
	if unit == nil {
		return models.GRPCRESULT_FAIL, "任务丢失", rst
	}
	unitData := unit.(models.GRPCTaskData)
	p.GrpcProcMap.Delete(sType)

	if unitData.Result.ResultValue != models.GRPCRESULT_OK {
		return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
	}

	recvmsg := &pb.RstVehRecognize{}
	if err := proto.Unmarshal(unitData.RstData, recvmsg); err != nil {
		return models.GRPCRESULT_FAIL, "解析返回结果失败", rst
	}

	rst.VehicleInfo.CarBaseInfo.Lpn = recvmsg.VehicleInfo.CarBaseInfo.Lpn                   //车牌号码
	rst.VehicleInfo.CarBaseInfo.LpnScore = recvmsg.VehicleInfo.CarBaseInfo.LpnScore         //车牌分数
	rst.VehicleInfo.CarBaseInfo.LpnColor = recvmsg.VehicleInfo.CarBaseInfo.LpnColor         //车牌颜色
	rst.VehicleInfo.CarBaseInfo.LpnposLeft = recvmsg.VehicleInfo.CarBaseInfo.LpnposLeft     //车牌位置坐标left
	rst.VehicleInfo.CarBaseInfo.LpnposTop = recvmsg.VehicleInfo.CarBaseInfo.LpnposTop       //车牌位置坐标top
	rst.VehicleInfo.CarBaseInfo.LpnposRight = recvmsg.VehicleInfo.CarBaseInfo.LpnposRight   //车牌位置坐标right
	rst.VehicleInfo.CarBaseInfo.LpnposBottom = recvmsg.VehicleInfo.CarBaseInfo.LpnposBottom //车牌位置坐标bottom
	rst.VehicleInfo.CarBaseInfo.Color = recvmsg.VehicleInfo.CarBaseInfo.Color               //车身颜色
	rst.VehicleInfo.CarBaseInfo.ColorScore = recvmsg.VehicleInfo.CarBaseInfo.ColorScore     //车身颜色分数
	rst.VehicleInfo.CarBaseInfo.Brand = recvmsg.VehicleInfo.CarBaseInfo.Brand               //车辆品牌
	rst.VehicleInfo.CarBaseInfo.Brand0 = recvmsg.VehicleInfo.CarBaseInfo.Brand0             //车辆品牌0
	rst.VehicleInfo.CarBaseInfo.Brand1 = recvmsg.VehicleInfo.CarBaseInfo.Brand1             //车辆品牌1
	rst.VehicleInfo.CarBaseInfo.Brand2 = recvmsg.VehicleInfo.CarBaseInfo.Brand2             //车辆品牌2
	rst.VehicleInfo.CarBaseInfo.Brand3 = recvmsg.VehicleInfo.CarBaseInfo.Brand3             //车辆品牌3
	rst.VehicleInfo.CarBaseInfo.Vehtype = recvmsg.VehicleInfo.CarBaseInfo.Vehtype           //车型
	rst.VehicleInfo.CarBaseInfo.VehClass = recvmsg.VehicleInfo.CarBaseInfo.VehClass         //车辆类型
	rst.VehicleInfo.CarBaseInfo.Type0 = recvmsg.VehicleInfo.CarBaseInfo.Type0               //车辆类型0
	rst.VehicleInfo.CarBaseInfo.Type1 = recvmsg.VehicleInfo.CarBaseInfo.Type1               //车辆类型1
	rst.VehicleInfo.CarBaseInfo.Type2 = recvmsg.VehicleInfo.CarBaseInfo.Type2               //车辆类型2
	rst.VehicleInfo.CarBaseInfo.Type3 = recvmsg.VehicleInfo.CarBaseInfo.Type3               //车辆类型3
	rst.VehicleInfo.CarBaseInfo.Subbrand = recvmsg.VehicleInfo.CarBaseInfo.Subbrand         //车辆子品牌
	rst.VehicleInfo.CarBaseInfo.Subbrand0 = recvmsg.VehicleInfo.CarBaseInfo.Subbrand0       //车辆子品牌0
	rst.VehicleInfo.CarBaseInfo.Subbrand1 = recvmsg.VehicleInfo.CarBaseInfo.Subbrand1       //车辆子品牌1
	rst.VehicleInfo.CarBaseInfo.Subbrand2 = recvmsg.VehicleInfo.CarBaseInfo.Subbrand2       //车辆子品牌2
	rst.VehicleInfo.CarBaseInfo.Subbrand3 = recvmsg.VehicleInfo.CarBaseInfo.Subbrand3       //车辆子品牌3
	rst.VehicleInfo.CarBaseInfo.Year = recvmsg.VehicleInfo.CarBaseInfo.Year                 //车辆年份
	rst.VehicleInfo.CarBaseInfo.Year0 = recvmsg.VehicleInfo.CarBaseInfo.Year0               //车辆年份0
	rst.VehicleInfo.CarBaseInfo.Year1 = recvmsg.VehicleInfo.CarBaseInfo.Year1               //车辆年份1
	rst.VehicleInfo.CarBaseInfo.Year2 = recvmsg.VehicleInfo.CarBaseInfo.Year2               //车辆年份2
	rst.VehicleInfo.CarBaseInfo.Year3 = recvmsg.VehicleInfo.CarBaseInfo.Year3               //车辆年份3
	rst.VehicleInfo.CarBaseInfo.BrandScore = recvmsg.VehicleInfo.CarBaseInfo.BrandScore     //品牌分数
	rst.VehicleInfo.CarBaseInfo.BrandScore0 = recvmsg.VehicleInfo.CarBaseInfo.BrandScore0   //品牌分数0
	rst.VehicleInfo.CarBaseInfo.BrandScore1 = recvmsg.VehicleInfo.CarBaseInfo.BrandScore1   //品牌分数1
	rst.VehicleInfo.CarBaseInfo.BrandScore2 = recvmsg.VehicleInfo.CarBaseInfo.BrandScore2   //品牌分数2
	rst.VehicleInfo.CarBaseInfo.BrandScore3 = recvmsg.VehicleInfo.CarBaseInfo.BrandScore3   //品牌分数3
	rst.VehicleInfo.CarBaseInfo.Pose = recvmsg.VehicleInfo.CarBaseInfo.Pose                 //车辆整体位置
	rst.VehicleInfo.CarBaseInfo.CarposLeft = recvmsg.VehicleInfo.CarBaseInfo.CarposLeft     //车辆坐标left
	rst.VehicleInfo.CarBaseInfo.CarposTop = recvmsg.VehicleInfo.CarBaseInfo.CarposTop       //车辆坐标top
	rst.VehicleInfo.CarBaseInfo.CarposRight = recvmsg.VehicleInfo.CarBaseInfo.CarposRight   //车辆坐标right
	rst.VehicleInfo.CarBaseInfo.CarposBottom = recvmsg.VehicleInfo.CarBaseInfo.CarposBottom //车辆坐标bottom
	rst.VehicleInfo.CarBaseInfo.CarRectScore = recvmsg.VehicleInfo.CarBaseInfo.CarRectScore //车辆位置分数
	rst.VehicleInfo.CarBaseInfo.ImgQuality = recvmsg.VehicleInfo.CarBaseInfo.ImgQuality     //车辆或车牌图像质量
	rst.VehicleInfo.CarBaseInfo.IDNumber = recvmsg.VehicleInfo.CarBaseInfo.IDNumber

	rst.VehicleInfo.CarFeatureInfo.GJ = recvmsg.VehicleInfo.CarFeatureInfo.GJ   //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.NJB = recvmsg.VehicleInfo.CarFeatureInfo.NJB //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.TC = recvmsg.VehicleInfo.CarFeatureInfo.TC   //天窗	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.AQD = recvmsg.VehicleInfo.CarFeatureInfo.AQD //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.DDH = recvmsg.VehicleInfo.CarFeatureInfo.DDH //开车打电话	同AQD字段
	rst.VehicleInfo.CarFeatureInfo.ZYB = recvmsg.VehicleInfo.CarFeatureInfo.ZYB //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
	rst.VehicleInfo.CarFeatureInfo.CZH = recvmsg.VehicleInfo.CarFeatureInfo.CZH //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
	rst.VehicleInfo.CarFeatureInfo.CRZ = recvmsg.VehicleInfo.CarFeatureInfo.CRZ //出入证	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.XSB = recvmsg.VehicleInfo.CarFeatureInfo.XSB //新手标	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.JSY = recvmsg.VehicleInfo.CarFeatureInfo.JSY //驾驶员	同AQD字段
	rst.VehicleInfo.CarFeatureInfo.LT = recvmsg.VehicleInfo.CarFeatureInfo.LT   //轮胎	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.CD = recvmsg.VehicleInfo.CarFeatureInfo.CD   //车灯	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.JSC = recvmsg.VehicleInfo.CarFeatureInfo.JSC //驾驶窗	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.HSJ = recvmsg.VehicleInfo.CarFeatureInfo.HSJ //后视镜	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.XLJ = recvmsg.VehicleInfo.CarFeatureInfo.XLJ //行李架	同CZH字段

	i := 0
	icnt := util.ConvertS2I(recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeCount)
	rst.VehicleOtherInfo.TotalMotobikeInfo.BikeCount = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeCount
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posbottom
		unit.Posleft = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posleft
		unit.Postop = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Postop
		unit.Posright = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posright
		unit.Score = recvmsg.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Score

		rst.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds, unit)
	}

	icnt = util.ConvertS2I(recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeCount)
	rst.VehicleOtherInfo.TotalTribikeInfo.BikeCount = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeCount
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posbottom
		unit.Posleft = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posleft
		unit.Postop = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Postop
		unit.Posright = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posright
		unit.Score = recvmsg.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Score

		rst.VehicleOtherInfo.TotalTribikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalTribikeInfo.BikeRsds, unit)
	}

	icnt = util.ConvertS2I(recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeCount)
	rst.VehicleOtherInfo.TotalBikeInfo.BikeCount = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeCount
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posbottom
		unit.Posleft = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posleft
		unit.Postop = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Postop
		unit.Posright = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posright
		unit.Score = recvmsg.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Score

		rst.VehicleOtherInfo.TotalBikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalBikeInfo.BikeRsds, unit)
	}

	icnt = util.ConvertS2I(recvmsg.VehicleFaceInfo.FaceCount)
	rst.VehicleFaceInfo.FaceCount = recvmsg.VehicleFaceInfo.FaceCount
	for i = 0; i < icnt; i += 1 {
		var unit models.FaceInfo
		unit.FaceRectScore = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceRectScore
		unit.FaceposLeft = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceposLeft
		unit.FaceposTop = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceposTop
		unit.FaceposRight = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceposRight
		unit.FaceposBottom = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceposBottom
		unit.FacekeyPoint1X = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1X
		unit.FacekeyPoint1Y = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1Y
		unit.FacekeyPoint2X = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2X
		unit.FacekeyPoint2Y = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2Y
		unit.FacekeyPoint3X = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3X
		unit.FacekeyPoint3Y = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3Y
		unit.FacekeyPoint4X = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4X
		unit.FacekeyPoint4Y = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4Y
		unit.FacekeyPoint5X = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5X
		unit.FacekeyPoint5Y = recvmsg.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5Y
		unit.FaceFeature = recvmsg.VehicleFaceInfo.FaceRsds[i].FaceFeature

		rst.VehicleFaceInfo.FaceRsds = append(rst.VehicleFaceInfo.FaceRsds, unit)
	}

	rst.VehicleImg.CarImgBuffer = recvmsg.VehicleImg.CarImgBuffer
	rst.VehicleImg.PlateImgBuffer = recvmsg.VehicleImg.PlateImgBuffer

	return unitData.Result.ResultValue, unitData.Result.ResultDes, rst
}

func TestPlate(addsinfo models.ExitImgAddsData) (string, string, models.RstVehRecognizeInfo) {
	var rst models.RstVehRecognizeInfo
	rst.VehicleInfo.CarBaseInfo.Lpn = "苏A12345"     //车牌号码
	rst.VehicleInfo.CarBaseInfo.LpnScore = "12"     //车牌分数
	rst.VehicleInfo.CarBaseInfo.LpnColor = "13"     //车牌颜色
	rst.VehicleInfo.CarBaseInfo.LpnposLeft = "14"   //车牌位置坐标left
	rst.VehicleInfo.CarBaseInfo.LpnposTop = "15"    //车牌位置坐标top
	rst.VehicleInfo.CarBaseInfo.LpnposRight = "16"  //车牌位置坐标right
	rst.VehicleInfo.CarBaseInfo.LpnposBottom = "17" //车牌位置坐标bottom
	rst.VehicleInfo.CarBaseInfo.Color = "18"        //车身颜色
	rst.VehicleInfo.CarBaseInfo.ColorScore = "19"   //车身颜色分数
	rst.VehicleInfo.CarBaseInfo.Brand = "20"        //车辆品牌
	rst.VehicleInfo.CarBaseInfo.Brand0 = "21"       //车辆品牌0
	rst.VehicleInfo.CarBaseInfo.Brand1 = "22"       //车辆品牌1
	rst.VehicleInfo.CarBaseInfo.Brand2 = "23"       //车辆品牌2
	rst.VehicleInfo.CarBaseInfo.Brand3 = "24"       //车辆品牌3
	rst.VehicleInfo.CarBaseInfo.Vehtype = "1"       //车型
	rst.VehicleInfo.CarBaseInfo.VehClass = "1"      //车辆类型
	rst.VehicleInfo.CarBaseInfo.Type0 = "27"        //车辆类型0
	rst.VehicleInfo.CarBaseInfo.Type1 = "28"        //车辆类型1
	rst.VehicleInfo.CarBaseInfo.Type2 = "29"        //车辆类型2
	rst.VehicleInfo.CarBaseInfo.Type3 = "30"        //车辆类型3
	rst.VehicleInfo.CarBaseInfo.Subbrand = "31"     //车辆子品牌
	rst.VehicleInfo.CarBaseInfo.Subbrand0 = "32"    //车辆子品牌0
	rst.VehicleInfo.CarBaseInfo.Subbrand1 = "33"    //车辆子品牌1
	rst.VehicleInfo.CarBaseInfo.Subbrand2 = "34"    //车辆子品牌2
	rst.VehicleInfo.CarBaseInfo.Subbrand3 = "35"    //车辆子品牌3
	rst.VehicleInfo.CarBaseInfo.Year = "36"         //车辆年份
	rst.VehicleInfo.CarBaseInfo.Year0 = "37"        //车辆年份0
	rst.VehicleInfo.CarBaseInfo.Year1 = "38"        //车辆年份1
	rst.VehicleInfo.CarBaseInfo.Year2 = "39"        //车辆年份2
	rst.VehicleInfo.CarBaseInfo.Year3 = "40"        //车辆年份3
	rst.VehicleInfo.CarBaseInfo.BrandScore = "41"   //品牌分数
	rst.VehicleInfo.CarBaseInfo.BrandScore0 = "42"  //品牌分数0
	rst.VehicleInfo.CarBaseInfo.BrandScore1 = "43"  //品牌分数1
	rst.VehicleInfo.CarBaseInfo.BrandScore2 = "44"  //品牌分数2
	rst.VehicleInfo.CarBaseInfo.BrandScore3 = "45"  //品牌分数3
	rst.VehicleInfo.CarBaseInfo.Pose = "46"         //车辆整体位置
	rst.VehicleInfo.CarBaseInfo.CarposLeft = "47"   //车辆坐标left
	rst.VehicleInfo.CarBaseInfo.CarposTop = "48"    //车辆坐标top
	rst.VehicleInfo.CarBaseInfo.CarposRight = "49"  //车辆坐标right
	rst.VehicleInfo.CarBaseInfo.CarposBottom = "50" //车辆坐标bottom
	rst.VehicleInfo.CarBaseInfo.CarRectScore = "51" //车辆位置分数
	rst.VehicleInfo.CarBaseInfo.ImgQuality = "52"   //车辆或车牌图像质量
	rst.VehicleInfo.CarBaseInfo.IDNumber = "53"

	rst.VehicleInfo.CarFeatureInfo.GJ = "54"  //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.NJB = "55" //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.TC = "56"  //天窗	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.AQD = "57" //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
	rst.VehicleInfo.CarFeatureInfo.DDH = "58" //开车打电话	同AQD字段
	rst.VehicleInfo.CarFeatureInfo.ZYB = "59" //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
	rst.VehicleInfo.CarFeatureInfo.CZH = "60" //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
	rst.VehicleInfo.CarFeatureInfo.CRZ = "61" //出入证	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.XSB = "62" //新手标	同GJ字段
	rst.VehicleInfo.CarFeatureInfo.JSY = "63" //驾驶员	同AQD字段
	rst.VehicleInfo.CarFeatureInfo.LT = "64"  //轮胎	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.CD = "65"  //车灯	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.JSC = "66" //驾驶窗	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.HSJ = "67" //后视镜	同CZH字段
	rst.VehicleInfo.CarFeatureInfo.XLJ = "68" //行李架	同CZH字段

	i := 0
	rst.VehicleOtherInfo.TotalMotobikeInfo.BikeCount = "2"
	icnt := util.ConvertS2I(rst.VehicleOtherInfo.TotalMotobikeInfo.BikeCount)
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = util.ConvertI2S(i * 1)
		unit.Posleft = util.ConvertI2S(i * 10)
		unit.Postop = util.ConvertI2S(i * 100)
		unit.Posright = util.ConvertI2S(i * 10)
		unit.Score = util.ConvertI2S(i * 100)

		rst.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds, unit)
	}

	rst.VehicleOtherInfo.TotalTribikeInfo.BikeCount = "2"
	icnt = util.ConvertS2I(rst.VehicleOtherInfo.TotalTribikeInfo.BikeCount)
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = util.ConvertI2S(i)
		unit.Posleft = util.ConvertI2S(10 + i)
		unit.Postop = util.ConvertI2S(100 + i)
		unit.Posright = util.ConvertI2S(1000 + i)
		unit.Score = util.ConvertI2S(10000 + i)

		rst.VehicleOtherInfo.TotalTribikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalTribikeInfo.BikeRsds, unit)
	}

	rst.VehicleOtherInfo.TotalBikeInfo.BikeCount = "2"
	icnt = util.ConvertS2I(rst.VehicleOtherInfo.TotalBikeInfo.BikeCount)
	for i = 0; i < icnt; i += 1 {
		var unit models.BikeInfo
		unit.Posbottom = util.ConvertI2S(20 + i)
		unit.Posleft = util.ConvertI2S(200 + i)
		unit.Postop = util.ConvertI2S(2000 + i)
		unit.Posright = util.ConvertI2S(20000 + i)
		unit.Score = util.ConvertI2S(200000 + i)

		rst.VehicleOtherInfo.TotalBikeInfo.BikeRsds = append(rst.VehicleOtherInfo.TotalBikeInfo.BikeRsds, unit)
	}

	rst.VehicleFaceInfo.FaceCount = "2"
	icnt = util.ConvertS2I(rst.VehicleFaceInfo.FaceCount)
	for i = 0; i < icnt; i += 1 {
		var unit models.FaceInfo
		unit.FaceRectScore = util.ConvertI2S(501 + i)
		unit.FaceposLeft = util.ConvertI2S(502 + i)
		unit.FaceposTop = util.ConvertI2S(503 + i)
		unit.FaceposRight = util.ConvertI2S(504 + i)
		unit.FaceposBottom = util.ConvertI2S(505 + i)
		unit.FacekeyPoint1X = util.ConvertI2S(506 + i)
		unit.FacekeyPoint1Y = util.ConvertI2S(507 + i)
		unit.FacekeyPoint2X = util.ConvertI2S(508 + i)
		unit.FacekeyPoint2Y = util.ConvertI2S(509 + i)
		unit.FacekeyPoint3X = util.ConvertI2S(510 + i)
		unit.FacekeyPoint3Y = util.ConvertI2S(511 + i)
		unit.FacekeyPoint4X = util.ConvertI2S(512 + i)
		unit.FacekeyPoint4Y = util.ConvertI2S(513 + i)
		unit.FacekeyPoint5X = util.ConvertI2S(514 + i)
		unit.FacekeyPoint5Y = util.ConvertI2S(515 + i)
		unit.FaceFeature = util.ConvertI2S(516 + i)

		rst.VehicleFaceInfo.FaceRsds = append(rst.VehicleFaceInfo.FaceRsds, unit)
	}

	rst.VehicleImg.CarImgBuffer = "recvmsg.VehicleImg.CarImgBuffer"
	rst.VehicleImg.PlateImgBuffer = "recvmsg.VehicleImg.PlateImgBuffer"

	var tmpImg models.ReqExitImgData
	tmpImg.AddsInfo = addsinfo
	tmpImg.RstVehPlateInfo = rst

	return models.GRPCRESULT_OK, "", rst
}

package main

import (
	"FTC/config"
	"FTC/device"
	"FTC/pb"
	"FTC/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"

	"FTC/models"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var PlateSTX = "zszt"

const (
	CMDPLATE_PlateREQ = 0x01 //应答车牌
	CMDPLATE_PlateRST = 0x11 //主动上报车牌

	CMDPLATE_StateREQ = 0x02 //问询状态
	CMDPLATE_StateRST = 0x12 //应答状态
)

//车牌服务
type PlateService struct {
	device.ConnBaseCli

	State     int
	LastQueTm int64
}

//协议
func (p *PlateService) enpack(cmd byte, no byte, inbuf []byte) ([]byte, error) {
	buf := make([]byte, models.MAX_LBUFFERSIZE)
	pos := 0

	inlen := 0
	if inbuf != nil {
		inlen = len(inbuf)
	}

	bsStx := util.ConvertString2Byte(PlateSTX)
	copy(buf[pos:], bsStx)
	pos += len(bsStx)

	buf[pos] = cmd
	pos += 1

	buf[pos] = no
	pos += 1

	bsLen := util.Int2Bytes_B(inlen)
	copy(buf[pos:], bsLen)
	pos += 4

	if inlen > 0 {
		copy(buf[pos:], inbuf)
		pos += inlen
	}

	return buf[0:pos], nil
}

func (p *PlateService) depack(inbuf []byte) ([]byte, int, error) {
	length := len(inbuf)
	baselen := 10

	//util.MyOutputFrame("DepackMtc",inbuf)

	if length < baselen {
		return nil, 0, errors.New("还未收完")
	}

	i := 0
	for {
		startpos := 0
		endpos := 0
		pos := 0
		for ; i < length-4; i = i + 1 {
			stx := string(inbuf[i : i+4])
			if PlateSTX == stx {
				startpos = i
				break
			}
		}

		if i >= length-4 {
			return nil, i, errors.New("未找到帧头")
		}

		if startpos+baselen > length {
			return nil, startpos, errors.New("还未收完1")
		}

		pos = startpos
		datalen := util.Bytes2Int_B(inbuf[pos+6 : pos+10])
		if startpos+datalen+baselen > length {
			return nil, startpos, errors.New("还未收完2")
		}

		framelen := datalen + baselen
		endpos = startpos + framelen
		frame := make([]byte, framelen)
		copy(frame, inbuf[startpos:endpos])

		return frame, endpos, nil
	}

	return nil, 0, nil
}

func (p *PlateService) InitPlate() {
	p.State = models.DEVSTATE_UNKNOWN
	p.LastQueTm = 0
}

func (p *PlateService) Recvproc(inbuf []byte) (int, bool) {
	inlen := len(inbuf)
	if inlen <= 0 {
		return 0, false
	}

	fmt.Println("recv :%d.%s..", inlen, util.ConvertByte2Hexstring(inbuf[0:10], true))

	outbuf, offset, err := p.depack(inbuf[0:inlen])
	if err != nil {
		fmt.Println("%d,%s", offset, err.Error())
		return offset, false
	}

	cmd := outbuf[4]
	switch cmd {
	case CMDPLATE_PlateRST:
		p.funcPlateProc(outbuf)
	case CMDPLATE_StateRST:
		p.State = models.DEVSTATE_OK
	}

	return offset, true
}

//应答车牌上报
func (p *PlateService) FuncReqPlateProc(strno string) {
	outbuf, _ := p.enpack(CMDPLATE_PlateREQ, util.Converts2b(strno), nil)
	p.SendProc(outbuf, true)
}

func (p *PlateService) testplatestr() {
	str := "7A737A74110200000EE37B0A202020226F7468657256656869636C65496E666F22203A207B0A20202020202022746F74616C42696B6522203A207B0A2020202020202020202262696B65436F756E7422203A2022222C0A2020202020202020202262696B65496E666F22203A205B0A2020202020202020202020207B0A20202020202020202020202020202022706F73626F74746F6D22203A2022222C0A20202020202020202020202020202022706F736C65667422203A2022222C0A20202020202020202020202020202022706F73726967687422203A2022222C0A20202020202020202020202020202022706F73746F7022203A2022222C0A2020202020202020202020202020202273636F726522203A2022220A2020202020202020202020207D0A2020202020202020205D0A2020202020207D2C0A20202020202022746F74616C4D6F746F62696B6522203A207B0A202020202020202020226D6F746F62696B65436F756E7422203A2022222C0A202020202020202020226D6F746F62696B65496E666F22203A205B0A2020202020202020202020207B0A20202020202020202020202020202022706F73626F74746F6D22203A2022222C0A20202020202020202020202020202022706F736C65667422203A2022222C0A20202020202020202020202020202022706F73726967687422203A2022222C0A20202020202020202020202020202022706F73746F7022203A2022222C0A2020202020202020202020202020202273636F726522203A2022220A2020202020202020202020207D0A2020202020202020205D0A2020202020207D2C0A20202020202022746F74616C54726962696B6522203A207B0A2020202020202020202274726962696B65436F756E7422203A2022222C0A2020202020202020202274726962696B65496E666F22203A205B0A2020202020202020202020207B0A20202020202020202020202020202022706F73626F74746F6D22203A2022222C0A20202020202020202020202020202022706F736C65667422203A2022222C0A20202020202020202020202020202022706F73726967687422203A2022222C0A20202020202020202020202020202022706F73746F7022203A2022222C0A2020202020202020202020202020202273636F726522203A2022220A2020202020202020202020207D0A2020202020202020205D0A2020202020207D0A2020207D2C0A20202022746F74616C436172496E666F22203A207B0A2020202020202263617222203A205B0A2020202020202020207B0A202020202020202020202020226361724665617475726522203A207B0A2020202020202020202020202020202241514422203A2022222C0A20202020202020202020202020202022434422203A2022222C0A2020202020202020202020202020202243525A22203A2022222C0A20202020202020202020202020202022435A4822203A2022222C0A2020202020202020202020202020202244444822203A2022222C0A20202020202020202020202020202022474A22203A2022222C0A2020202020202020202020202020202248534A22203A2022222C0A202020202020202020202020202020224A534322203A2022222C0A202020202020202020202020202020224A535922203A2022222C0A202020202020202020202020202020224C5422203A2022222C0A202020202020202020202020202020224E4A4222203A2022222C0A20202020202020202020202020202022544322203A2022222C0A20202020202020202020202020202022584C4A22203A2022222C0A2020202020202020202020202020202258534222203A2022222C0A202020202020202020202020202020225A594222203A2022220A2020202020202020202020207D2C0A20202020202020202020202022636172496E666F22203A207B0A2020202020202020202020202020202249444E756D62657222203A20223230313930333139313033393536353638222C0A202020202020202020202020202020226272616E6422203A2022B4F3D6DA222C0A202020202020202020202020202020226272616E643022203A2022222C0A202020202020202020202020202020226272616E643122203A2022222C0A202020202020202020202020202020226272616E643222203A2022222C0A202020202020202020202020202020226272616E643322203A2022222C0A202020202020202020202020202020226272616E6453636F726522203A2022222C0A202020202020202020202020202020226272616E6453636F72653022203A2022222C0A202020202020202020202020202020226272616E6453636F72653122203A2022222C0A202020202020202020202020202020226272616E6453636F72653222203A2022222C0A202020202020202020202020202020226272616E6453636F72653322203A2022222C0A2020202020202020202020202020202263617243617054696D6522203A2022323031392D30332D31395F31303A33393A3536222C0A202020202020202020202020202020226361725265637453636F726522203A2022222C0A20202020202020202020202020202022636172536176655061746822203A2022443A5C5CB3B5C1BECDBCC6AC5C5CB3B5B5C0305C5C323031395C5C30335C5C31395C5CBFCD315C5C305F305F31305F33395F35362E6A7067222C0A202020202020202020202020202020226361726E756D62657222203A20302C0A20202020202020202020202020202022636172706F73426F74746F6D22203A2022222C0A20202020202020202020202020202022636172706F734C65667422203A2022222C0A20202020202020202020202020202022636172706F73526967687422203A2022222C0A20202020202020202020202020202022636172706F73546F7022203A2022222C0A20202020202020202020202020202022636F6C6F7222203A2022B0D7C9AB222C0A20202020202020202020202020202022636F6C6F7253636F726522203A2022222C0A20202020202020202020202020202022696D675175616C69747922203A2022222C0A202020202020202020202020202020226C706E22203A2022CBD54138305A5535222C0A202020202020202020202020202020226C706E436F6C6F7222203A2022C0B6C9AB222C0A202020202020202020202020202020226C706E53636F726522203A2022222C0A202020202020202020202020202020226C706E706F73426F74746F6D22203A2022383538222C0A202020202020202020202020202020226C706E706F734C65667422203A2022383139222C0A202020202020202020202020202020226C706E706F73526967687422203A2022393932222C0A202020202020202020202020202020226C706E706F73546F7022203A2022383138222C0A20202020202020202020202020202022706F736522203A2022222C0A202020202020202020202020202020227375626272616E6422203A2022CDBEB0B2222C0A202020202020202020202020202020227375626272616E643022203A2022222C0A202020202020202020202020202020227375626272616E643122203A2022222C0A202020202020202020202020202020227375626272616E643222203A2022222C0A202020202020202020202020202020227375626272616E643322203A2022222C0A202020202020202020202020202020227479706522203A20223031222C0A20202020202020202020202020202022747970653022203A20224D5056222C0A20202020202020202020202020202022747970653122203A2022222C0A20202020202020202020202020202022747970653222203A2022222C0A20202020202020202020202020202022747970653322203A2022222C0A2020202020202020202020202020202276656869636C654361724665617475726522203A2022222C0A202020202020202020202020202020227965617222203A202232303136BFEE222C0A20202020202020202020202020202022796561723022203A2022222C0A20202020202020202020202020202022796561723122203A2022222C0A20202020202020202020202020202022796561723222203A2022222C0A20202020202020202020202020202022796561723322203A2022220A2020202020202020202020207D0A2020202020202020207D0A2020202020205D2C0A20202020202022636172436F756E7422203A202231220A2020207D2C0A20202022746F74616C46616365496E666F22203A207B0A202020202020226661636522203A205B0A2020202020202020207B0A2020202020202020202020202266616365496E666F22203A207B0A20202020202020202020202020202022666163654665617475726522203A2022222C0A20202020202020202020202020202022666163655265637453636F726522203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74315822203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74315922203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74325822203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74325922203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74335822203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74335922203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74345822203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74345922203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74355822203A2022222C0A20202020202020202020202020202022666163656B6579506F696E74355922203A2022222C0A2020202020202020202020202020202266616365706F73426F74746F6D22203A2022222C0A2020202020202020202020202020202266616365706F734C65667422203A2022222C0A2020202020202020202020202020202266616365706F73526967687422203A2022222C0A2020202020202020202020202020202266616365706F73546F7022203A2022220A2020202020202020202020207D0A2020202020202020207D0A2020202020205D2C0A2020202020202266616365436F756E7422203A2022220A2020207D0A7D0A"
	out := util.ConvertHexstring2Byte(str)
	p.funcPlateProc(out)
}

func (p *PlateService) funcPlateProc(inbuf []byte) {
	strno := util.Convertb2s(inbuf[5])
	databuf := inbuf[10:]

	util.FileLogs.Info("funcPlateProc:%s.", util.ConvertByte2Hexstring(inbuf, true))
	p.FuncReqPlateProc(strno)

	var recvObj models.RecvVehRecognizeInfo
	mbuf, err := util.GbkToUtf8(databuf)
	fmt.Println(string(mbuf))
	fmt.Println("--------------=====11")
	err = json.Unmarshal(mbuf, &recvObj)
	if err != nil {
		util.FileLogs.Info("funcPlateProc Unmarshal failed:%s.", err.Error())
		return
	}

	iveh := 0
	var infoObj models.RstVehRecognizeInfo
	infoObj.VehicleOtherInfo = recvObj.VehicleOtherInfo
	infoObj.VehicleFaceInfo = recvObj.VehicleFaceInfo
	if len(recvObj.TotalCarInfo.CarRsds) > 0 {
		infoObj.VehicleInfo.CarBaseInfo = recvObj.TotalCarInfo.CarRsds[0].CarBaseInfo
		infoObj.VehicleInfo.CarFeatureInfo = recvObj.TotalCarInfo.CarRsds[0].CarFeatureInfo
		iveh = util.ConvertS2I(recvObj.TotalCarInfo.CarRsds[0].CarBaseInfo.Type)
		infoObj.VehicleInfo.CarBaseInfo.VehClass = util.ConvertI2S(iveh)
		infoObj.VehicleInfo.CarBaseInfo.Vehtype = util.ConvertI2S(iveh)
	}

	fmt.Println(infoObj)

	//封装为grpc，发送给车道
	carbaseinfo := &pb.CarInfo{
		Lpn:          infoObj.VehicleInfo.CarBaseInfo.Lpn,          //车牌号码
		LpnScore:     infoObj.VehicleInfo.CarBaseInfo.LpnScore,     //车牌分数
		LpnColor:     infoObj.VehicleInfo.CarBaseInfo.LpnColor,     //车牌颜色
		LpnposLeft:   infoObj.VehicleInfo.CarBaseInfo.LpnposLeft,   //车牌位置坐标left
		LpnposTop:    infoObj.VehicleInfo.CarBaseInfo.LpnposTop,    //车牌位置坐标top
		LpnposRight:  infoObj.VehicleInfo.CarBaseInfo.LpnposRight,  //车牌位置坐标right
		LpnposBottom: infoObj.VehicleInfo.CarBaseInfo.LpnposBottom, //车牌位置坐标bottom
		Color:        infoObj.VehicleInfo.CarBaseInfo.Color,        //车身颜色
		ColorScore:   infoObj.VehicleInfo.CarBaseInfo.ColorScore,   //车身颜色分数
		Brand:        infoObj.VehicleInfo.CarBaseInfo.Brand,        //车辆品牌
		Brand0:       infoObj.VehicleInfo.CarBaseInfo.Brand0,       //车辆品牌0
		Brand1:       infoObj.VehicleInfo.CarBaseInfo.Brand1,       //车辆品牌1
		Brand2:       infoObj.VehicleInfo.CarBaseInfo.Brand2,       //车辆品牌2
		Brand3:       infoObj.VehicleInfo.CarBaseInfo.Brand3,       //车辆品牌3
		Vehtype:      infoObj.VehicleInfo.CarBaseInfo.Vehtype,      //车型
		VehClass:     infoObj.VehicleInfo.CarBaseInfo.VehClass,     //车辆类型
		Type0:        infoObj.VehicleInfo.CarBaseInfo.Type0,        //车辆类型0
		Type1:        infoObj.VehicleInfo.CarBaseInfo.Type1,        //车辆类型1
		Type2:        infoObj.VehicleInfo.CarBaseInfo.Type2,        //车辆类型2
		Type3:        infoObj.VehicleInfo.CarBaseInfo.Type3,        //车辆类型3
		Subbrand:     infoObj.VehicleInfo.CarBaseInfo.Subbrand,     //车辆子品牌
		Subbrand0:    infoObj.VehicleInfo.CarBaseInfo.Subbrand0,    //车辆子品牌0
		Subbrand1:    infoObj.VehicleInfo.CarBaseInfo.Subbrand1,    //车辆子品牌1
		Subbrand2:    infoObj.VehicleInfo.CarBaseInfo.Subbrand2,    //车辆子品牌2
		Subbrand3:    infoObj.VehicleInfo.CarBaseInfo.Subbrand3,    //车辆子品牌3
		Year:         infoObj.VehicleInfo.CarBaseInfo.Year,         //车辆年份
		Year0:        infoObj.VehicleInfo.CarBaseInfo.Year0,        //车辆年份0
		Year1:        infoObj.VehicleInfo.CarBaseInfo.Year1,        //车辆年份1
		Year2:        infoObj.VehicleInfo.CarBaseInfo.Year2,        //车辆年份2
		Year3:        infoObj.VehicleInfo.CarBaseInfo.Year3,        //车辆年份3
		BrandScore:   infoObj.VehicleInfo.CarBaseInfo.BrandScore,   //品牌分数
		BrandScore0:  infoObj.VehicleInfo.CarBaseInfo.BrandScore0,  //品牌分数0
		BrandScore1:  infoObj.VehicleInfo.CarBaseInfo.BrandScore1,  //品牌分数1
		BrandScore2:  infoObj.VehicleInfo.CarBaseInfo.BrandScore2,  //品牌分数2
		BrandScore3:  infoObj.VehicleInfo.CarBaseInfo.BrandScore3,  //品牌分数3
		Pose:         infoObj.VehicleInfo.CarBaseInfo.Pose,         //车辆整体位置
		CarposLeft:   infoObj.VehicleInfo.CarBaseInfo.CarposLeft,   //车辆坐标left
		CarposTop:    infoObj.VehicleInfo.CarBaseInfo.CarposTop,    //车辆坐标top
		CarposRight:  infoObj.VehicleInfo.CarBaseInfo.CarposRight,  //车辆坐标right
		CarposBottom: infoObj.VehicleInfo.CarBaseInfo.CarposBottom, //车辆坐标bottom
		CarRectScore: infoObj.VehicleInfo.CarBaseInfo.CarRectScore, //车辆位置分数
		ImgQuality:   infoObj.VehicleInfo.CarBaseInfo.ImgQuality,   //车辆或车牌图像质量
		IDNumber:     infoObj.VehicleInfo.CarBaseInfo.IDNumber,     //Json文件唯一值标志
	}
	carfeature := &pb.CarFeature{
		GJ:  infoObj.VehicleInfo.CarFeatureInfo.GJ,  //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
		NJB: infoObj.VehicleInfo.CarFeatureInfo.NJB, //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
		TC:  infoObj.VehicleInfo.CarFeatureInfo.TC,  //天窗	同GJ字段
		AQD: infoObj.VehicleInfo.CarFeatureInfo.AQD, //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
		DDH: infoObj.VehicleInfo.CarFeatureInfo.DDH, //开车打电话	同AQD字段
		ZYB: infoObj.VehicleInfo.CarFeatureInfo.ZYB, //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
		CZH: infoObj.VehicleInfo.CarFeatureInfo.CZH, //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
		CRZ: infoObj.VehicleInfo.CarFeatureInfo.CRZ, //出入证	同GJ字段
		XSB: infoObj.VehicleInfo.CarFeatureInfo.XSB, //新手标	同GJ字段
		JSY: infoObj.VehicleInfo.CarFeatureInfo.JSY, //驾驶员	同AQD字段
		LT:  infoObj.VehicleInfo.CarFeatureInfo.LT,  //轮胎	同CZH字段
		CD:  infoObj.VehicleInfo.CarFeatureInfo.CD,  //车灯	同CZH字段
		JSC: infoObj.VehicleInfo.CarFeatureInfo.JSC, //驾驶窗	同CZH字段
		HSJ: infoObj.VehicleInfo.CarFeatureInfo.HSJ, //后视镜	同CZH字段
		XLJ: infoObj.VehicleInfo.CarFeatureInfo.XLJ, //行李架	同CZH字段
	}
	car := &pb.Car{
		CarBaseInfo:    carbaseinfo,
		CarFeatureInfo: carfeature}

	i := 0
	icnt := 0
	otherBike := new(pb.TotalBike)
	otherBike.BikeCount = infoObj.VehicleOtherInfo.TotalBikeInfo.BikeCount
	icnt = util.ConvertS2I(otherBike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherBike.BikeRsds = append(otherBike.BikeRsds, &pb.BikeInfo{
			Posleft:   infoObj.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posleft,
			Postop:    infoObj.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Postop,
			Posright:  infoObj.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posright,
			Posbottom: infoObj.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Posbottom,
			Score:     infoObj.VehicleOtherInfo.TotalBikeInfo.BikeRsds[i].Score})
	}

	otherMotoBike := new(pb.TotalBike)
	otherMotoBike.BikeCount = infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeCount
	icnt = util.ConvertS2I(otherMotoBike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherMotoBike.BikeRsds = append(otherMotoBike.BikeRsds, &pb.BikeInfo{
			Posleft:   infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posleft,
			Postop:    infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Postop,
			Posright:  infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posright,
			Posbottom: infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Posbottom,
			Score:     infoObj.VehicleOtherInfo.TotalMotobikeInfo.BikeRsds[i].Score})
	}

	otherTribike := new(pb.TotalBike)
	otherTribike.BikeCount = infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeCount
	icnt = util.ConvertS2I(otherTribike.BikeCount)
	for i = 0; i < icnt; i += 1 {
		otherTribike.BikeRsds = append(otherTribike.BikeRsds, &pb.BikeInfo{
			Posleft:   infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posleft,
			Postop:    infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Postop,
			Posright:  infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posright,
			Posbottom: infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Posbottom,
			Score:     infoObj.VehicleOtherInfo.TotalTribikeInfo.BikeRsds[i].Score})
	}

	others := &pb.OtherVehicleInfo{
		TotalMotobikeInfo: otherMotoBike,
		TotalTribikeInfo:  otherTribike,
		TotalBikeInfo:     otherBike}

	faces := new(pb.TotalFaceInfo)
	faces.FaceCount = infoObj.VehicleFaceInfo.FaceCount
	icnt = util.ConvertS2I(faces.FaceCount)
	for i = 0; i < icnt; i += 1 {
		faces.FaceRsds = append(faces.FaceRsds, &pb.FaceInfo{
			FaceRectScore:  infoObj.VehicleFaceInfo.FaceRsds[i].FaceRectScore,
			FaceposLeft:    infoObj.VehicleFaceInfo.FaceRsds[i].FaceposLeft,
			FaceposTop:     infoObj.VehicleFaceInfo.FaceRsds[i].FaceposTop,
			FaceposRight:   infoObj.VehicleFaceInfo.FaceRsds[i].FaceposRight,
			FaceposBottom:  infoObj.VehicleFaceInfo.FaceRsds[i].FaceposBottom,
			FacekeyPoint1X: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1X,
			FacekeyPoint1Y: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint1Y,
			FacekeyPoint2X: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2X,
			FacekeyPoint2Y: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint2Y,
			FacekeyPoint3X: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3X,
			FacekeyPoint3Y: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint3Y,
			FacekeyPoint4X: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4X,
			FacekeyPoint4Y: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint4Y,
			FacekeyPoint5X: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5X,
			FacekeyPoint5Y: infoObj.VehicleFaceInfo.FaceRsds[i].FacekeyPoint5Y,
			FaceFeature:    infoObj.VehicleFaceInfo.FaceRsds[i].FaceFeature})
	}

	img := &pb.Img{
		PlateImgBuffer: infoObj.VehicleImg.PlateImgBuffer,
		CarImgBuffer:   infoObj.VehicleImg.CarImgBuffer}

	m := &pb.RstVehRecognize{
		VehicleInfo:      car,
		VehicleOtherInfo: others,
		VehicleFaceInfo:  faces,
		VehicleImg:       img}

	var pret models.ResultInfo
	pret.ResultValue = models.GRPCRESULT_OK
	pret.ResultDes = ""
	PlateSrvGrpcSendproc(g_platestream, models.GRPCTYPE_PLATE, strno, m, pret)
}

func (p *PlateService) goAutoState() {
	for {
		if !p.IsConn() {
			p.State = models.DEVSTATE_TROUBLE
			util.MySleep_s(5)
			continue
		}

		nowtm := util.GetTimeStampSec()
		if nowtm > p.LastQueTm+5 {
			p.LastQueTm = nowtm

			outbuf, _ := p.enpack(CMDPLATE_StateREQ, 0, nil)
			p.SendProc(outbuf, false)
		}

		util.MySleep_s(5)
	}
}

var gPlateSrv PlateService
var g_bConn bool
var g_platestream pb.GrpcMsg_CommuniteServer
var g_strNo string

func main() {
	util.FileLogs.Info("车牌服务启动中...")
	config.InitConfig("../conf/config.conf")
	gPlateSrv.InitPlate()
	//连接车牌服务
	gPlateSrv.InitConn(models.DEVTYPE_PLATE, config.ConfigData["plateRegSrvIP"].(string), gPlateSrv.Recvproc)
	go gPlateSrv.goAutoState()

	gPlateSrv.testplatestr()

	/////////////////////////////////////grpc
	g_bConn = false
	sAddr := config.ConfigData["plateGrpcUrlSrv"].(string)
	lis, err := net.Listen("tcp", sAddr)
	if err != nil {
		util.FileLogs.Info("车牌服务监听grpc端口失败:%s", sAddr)
		return
	}
	util.FileLogs.Info("车牌服务监听grpc端口成功:%s", sAddr)

	s := grpc.NewServer()
	pb.RegisterGrpcMsgServer(s, &PlateServer{})
	s.Serve(lis)

}

type PlateServer struct {
}

func (p *PlateServer) Communite(stream pb.GrpcMsg_CommuniteServer) error {
	util.FileLogs.Info("车道车牌GRPC已连接")

	g_platestream = stream
	g_bConn = true

	go func() {
		for {
			if !g_bConn {
				break
			}

			//发送设备状态
			sState := util.ConvertI2S(gPlateSrv.State)
			m := &pb.PlateStateReport{State: sState}
			PlateSrvGrpcSendproc(stream, models.GRPCTYPE_PLATESTATE, "0", m, models.ResultInfo{})

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

		util.FileLogs.Info("收到来自车道车牌信息")
		fmt.Println(in)
	}

	return nil
}

func PlateSrvGrpcSendproc(stream pb.GrpcMsg_CommuniteServer, sType, no string, msg proto.Message, result models.ResultInfo) {
	if msg != nil {
		out, err := proto.Marshal(msg)
		if err != nil {
			util.FileLogs.Info("plate GrpcSendproc error:%s.\r\n", err.Error())
			return
		}

		notes := []*pb.Message{
			{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			util.FileLogs.Info("plate GrpcSendproc ")
			fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	} else {
		notes := []*pb.Message{
			{Type: sType, No: no, Data: nil, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
		}

		if stream != nil {
			util.FileLogs.Info("plate GrpcSendproc")
			fmt.Println(notes[0])
			stream.Send(notes[0])
		}
	}

}

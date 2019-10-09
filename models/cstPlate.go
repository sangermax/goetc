package models

type CarInfo struct {
	Lpn          string `json:"lpn,omitempty"`          //车牌号码
	LpnScore     string `json:"lpnScore,omitempty"`     //车牌分数
	LpnColor     string `json:"lpnColor,omitempty"`     //车牌颜色
	LpnposLeft   string `json:"lpnposLeft,omitempty"`   //车牌位置坐标left
	LpnposTop    string `json:"lpnposTop,omitempty"`    //车牌位置坐标top
	LpnposRight  string `json:"lpnposRight,omitempty"`  //车牌位置坐标right
	LpnposBottom string `json:"lpnposBottom,omitempty"` //车牌位置坐标bottom
	Color        string `json:"color,omitempty"`        //车身颜色
	ColorScore   string `json:"colorScore,omitempty"`   //车身颜色分数
	Brand        string `json:"brand,omitempty"`        //车辆品牌
	Brand0       string `json:"brand0,omitempty"`       //车辆品牌0
	Brand1       string `json:"brand1,omitempty"`       //车辆品牌1
	Brand2       string `json:"brand2,omitempty"`       //车辆品牌2
	Brand3       string `json:"brand3,omitempty"`       //车辆品牌3
	Vehtype      string `json:"vehtype,omitempty"`      //车型
	VehClass     string `json:"vehclass,omitempty"`     //车辆类型
	Type         string `json:"type,omitempty"`         //车辆类型
	Type0        string `json:"type0,omitempty"`        //车辆类型0
	Type1        string `json:"type1,omitempty"`        //车辆类型1
	Type2        string `json:"type2,omitempty"`        //车辆类型2
	Type3        string `json:"type3,omitempty"`        //车辆类型3
	Subbrand     string `json:"subbrand,omitempty"`     //车辆子品牌
	Subbrand0    string `json:"subbrand0,omitempty"`    //车辆子品牌0
	Subbrand1    string `json:"subbrand1,omitempty"`    //车辆子品牌1
	Subbrand2    string `json:"subbrand2,omitempty"`    //车辆子品牌2
	Subbrand3    string `json:"subbrand3,omitempty"`    //车辆子品牌3
	Year         string `json:"year,omitempty"`         //车辆年份
	Year0        string `json:"year0,omitempty"`        //车辆年份0
	Year1        string `json:"year1,omitempty"`        //车辆年份1
	Year2        string `json:"year2,omitempty"`        //车辆年份2
	Year3        string `json:"year3,omitempty"`        //车辆年份3
	BrandScore   string `json:"brandScore,omitempty"`   //品牌分数
	BrandScore0  string `json:"brandScore0,omitempty"`  //品牌分数0
	BrandScore1  string `json:"brandScore1,omitempty"`  //品牌分数1
	BrandScore2  string `json:"brandScore2,omitempty"`  //品牌分数2
	BrandScore3  string `json:"brandScore3,omitempty"`  //品牌分数3
	Pose         string `json:"pose,omitempty"`         //车辆整体位置
	CarposLeft   string `json:"carposLeft,omitempty"`   //车辆坐标left
	CarposTop    string `json:"carposTop,omitempty"`    //车辆坐标top
	CarposRight  string `json:"carposRight,omitempty"`  //车辆坐标right
	CarposBottom string `json:"carposBottom,omitempty"` //车辆坐标bottom
	CarRectScore string `json:"carRectScore,omitempty"` //车辆位置分数
	ImgQuality   string `json:"imgQuality,omitempty"`   //车辆或车牌图像质量
	IDNumber     string `json:"IDNumber,omitempty"`     //Json文件唯一值标志
}

type CarFeature struct {
	GJ  string `json:"GJ,omitempty"`  //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
	NJB string `json:"NJB,omitempty"` //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
	TC  string `json:"TC,omitempty"`  //天窗	同GJ字段
	AQD string `json:"AQD,omitempty"` //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
	DDH string `json:"DDH,omitempty"` //开车打电话	同AQD字段
	ZYB string `json:"ZYB,omitempty"` //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
	CZH string `json:"CZH,omitempty"` //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
	CRZ string `json:"CRZ,omitempty"` //出入证	同GJ字段
	XSB string `json:"XSB,omitempty"` //新手标	同GJ字段
	JSY string `json:"JSY,omitempty"` //驾驶员	同AQD字段
	LT  string `json:"LT,omitempty"`  //轮胎	同CZH字段
	CD  string `json:"CD,omitempty"`  //车灯	同CZH字段
	JSC string `json:"JSC,omitempty"` //驾驶窗	同CZH字段
	HSJ string `json:"HSJ,omitempty"` //后视镜	同CZH字段
	XLJ string `json:"XLJ,omitempty"` //行李架	同CZH字段

}

type TotalBike struct {
	BikeCount string     `json:"bikeCount,omitempty"` //数量
	BikeRsds  []BikeInfo `json:"bikeInfo,omitempty"`  //自行车信息

}

type TotalMotobike struct {
	BikeCount string     `json:"motobikeCount,omitempty"` //数量
	BikeRsds  []BikeInfo `json:"motobikeInfo,omitempty"`  //自行车信息

}

type TotalTribike struct {
	BikeCount string     `json:"tribikeCount,omitempty"` //数量
	BikeRsds  []BikeInfo `json:"tribikeInfo,omitempty"`  //自行车信息

}

type BikeInfo struct {
	Posleft   string `json:"posleft,omitempty"`   //坐标left
	Postop    string `json:"postop,omitempty"`    //坐标top
	Posright  string `json:"posright,omitempty"`  //坐标right
	Posbottom string `json:"posbottom,omitempty"` //坐标bottom
	Score     string `json:"score,omitempty"`     //得分

}

type TotalFaceInfo struct {
	FaceCount string     `json:"faceCount,omitempty"`
	FaceRsds  []FaceInfo `json:"face,omitempty"`
}

type FaceInfo struct {
	FaceRectScore  string `json:"faceRectScore,omitempty"`  //人脸框检测得分
	FaceposLeft    string `json:"faceposLeft,omitempty"`    //人脸位置坐标left
	FaceposTop     string `json:"faceposTop,omitempty"`     //人脸位置坐标top
	FaceposRight   string `json:"faceposRight,omitempty"`   //人脸位置坐标right
	FaceposBottom  string `json:"faceposBottom,omitempty"`  //人脸位置坐标bottom
	FacekeyPoint1X string `json:"facekeyPoint1X,omitempty"` //人脸关键点1X坐标
	FacekeyPoint1Y string `json:"facekeyPoint1Y,omitempty"` //人脸关键点1Y坐标
	FacekeyPoint2X string `json:"facekeyPoint2X,omitempty"` //人脸关键点2X坐标
	FacekeyPoint2Y string `json:"facekeyPoint2Y,omitempty"` //人脸关键点2Y坐标
	FacekeyPoint3X string `json:"facekeyPoint3X,omitempty"` //人脸关键点3X坐标
	FacekeyPoint3Y string `json:"facekeyPoint3Y,omitempty"` //人脸关键点3Y坐标
	FacekeyPoint4X string `json:"facekeyPoint4X,omitempty"` //人脸关键点4X坐标
	FacekeyPoint4Y string `json:"facekeyPoint4Y,omitempty"` //人脸关键点4Y坐标
	FacekeyPoint5X string `json:"facekeyPoint5X,omitempty"` //人脸关键点5X坐标
	FacekeyPoint5Y string `json:"facekeyPoint5Y,omitempty"` //人脸关键点5X坐标
	FaceFeature    string `json:"faceFeature,omitempty"`    //人脸特征码

}

type Img struct {
	PlateImgBuffer string `json:"plateImgBuffer,omitempty"` //车牌图片
	CarImgBuffer   string `json:"carImgBuffer,omitempty"`   //车辆图片

}

type Car struct {
	CarBaseInfo    CarInfo    `json:"carInfo,omitempty"`    //
	CarFeatureInfo CarFeature `json:"carFeature,omitempty"` //
}

type TotalCarInfo struct {
	CarCount string `json:"carCount,omitempty"` //
	CarRsds  []Car  `json:"car,omitempty"`      //
}

type OtherVehicleInfo struct {
	TotalMotobikeInfo TotalMotobike `json:"totalMotobike,omitempty"` //
	TotalTribikeInfo  TotalTribike  `json:"totalTribike,omitempty"`  //
	TotalBikeInfo     TotalBike     `json:"totalBike,omitempty"`     //
}

type RstVehRecognizeInfo struct {
	VehicleInfo      Car              `json:"car,omitempty"`              //
	VehicleOtherInfo OtherVehicleInfo `json:"otherVehicleInfo,omitempty"` //
	VehicleFaceInfo  TotalFaceInfo    `json:"totalFaceInfo,omitempty"`    //
	VehicleImg       Img              `json:"img,omitempty"`              //
}

type RecvVehRecognizeInfo struct {
	TotalCarInfo     TotalCarInfo     `json:"totalCarInfo,omitempty"`     //
	VehicleOtherInfo OtherVehicleInfo `json:"otherVehicleInfo,omitempty"` //
	VehicleFaceInfo  TotalFaceInfo    `json:"totalFaceInfo,omitempty"`    //
	VehicleImg       Img              `json:"img,omitempty"`              //
}

//车辆品牌列表

//车牌颜色
const (
	PCOLOR_BLUE   = 1 //蓝色
	PCOLOR_YELLOW = 2 //黄色
	PCOLOR_GREEN  = 3 //绿色
	PCOLOR_WHITE  = 4 //白色
	PCOLOR_BLACK  = 5 //黑色
	PCOLOR_OTHERS = 6 //其它颜色
)

//车身颜色
const (
	VCOLOR_WHITE  = 1  //	白
	VCOLOR_GRAY   = 2  //	灰
	VCOLOR_YELLOW = 3  //	黄
	VCOLOR_PINK   = 4  //	粉
	VCOLOR_RED    = 5  //	红
	VCOLOR_PURPLE = 6  //	紫
	VCOLOR_GREEN  = 7  //	绿
	VCOLOR_BLUE   = 8  //	蓝
	VCOLOR_BROWN  = 9  //	棕
	VCOLOR_BLACK  = 10 //	黑
	VCOLOR_OTHERS = 11 //	其它
)

//车辆类型
const (
	VCLASS_LTRUCK = 1  //大货车
	VCLASS_TRUCK  = 2  //货车
	VCLASS_STRUCK = 3  //小货车
	VCLASS_LBUS   = 4  //大客车
	VCLASS_BUS    = 5  //客车
	VCLASS_CAR    = 6  //轿车
	VCLASS_SUV    = 7  //SUV
	VCLASS_MPV    = 8  //MPV
	VCLASS_VAN    = 9  //面包车
	VCLASS_PICKUP = 10 //皮卡
)

//车型
const (
	VTYPE1  = 1  //客一
	VTYPE2  = 2  //客二
	VTYPE3  = 3  //客三
	VTYPE4  = 4  //客四
	VTYPE11 = 11 //货一
	VTYPE12 = 12 //货二
	VTYPE13 = 13 //货三
	VTYPE14 = 14 //货四
	VTYPE15 = 15 //货五
)

/////////////////////////////////////////////////
const (
	PR_SUC  = "success"
	PR_FAIL = "fail"
)

type PlateStateInfo struct {
	State string `json:"state"` //
}

type PlateResult struct {
	Code   string `json:"code"`
	Errdes string `json:"errdes"`
}

type PlateRstData struct {
	//返回结果
	Result PlateResult `json:"result"` //
	Infos  interface{} `json:"info"`
}

//临时车牌//////////////////////////////////////////////////
type VehPlateflag struct {
	BufferSize string `json:"bufferSize,omitempty"` //
	FileName   string `json:"fileName,omitempty"`   //
}

type VehPlateOtherVehicleInfo struct {
	TotalMotobikeInfo VehPlateTotalMotobike `json:"totalMotobike,omitempty"` //
	TotalTribikeInfo  VehPlateTotalTribike  `json:"totalTribike,omitempty"`  //
	TotalBikeInfo     VehPlateTotalBike     `json:"totalBike,omitempty"`     //
}

type VehPlateTotalBike struct {
	BikeCount int                `json:"bikeCount,omitempty"` //数量
	BikeRsds  []VehPlateBikeInfo `json:"bikeInfo,omitempty"`  //自行车信息
}

type VehPlateTotalMotobike struct {
	BikeCount int                `json:"motobikeCount,omitempty"` //数量
	BikeRsds  []VehPlateBikeInfo `json:"motobikeInfo,omitempty"`  //自行车信息
}

type VehPlateTotalTribike struct {
	BikeCount int                `json:"tribikeCount,omitempty"` //数量
	BikeRsds  []VehPlateBikeInfo `json:"tribikeInfo,omitempty"`  //自行车信息
}

type VehPlateBikeInfo struct {
	Posleft   int `json:"posleft,omitempty"`   //坐标left
	Postop    int `json:"postop,omitempty"`    //坐标top
	Posright  int `json:"posright,omitempty"`  //坐标right
	Posbottom int `json:"posbottom,omitempty"` //坐标bottom
	Score     int `json:"score,omitempty"`     //得分
}

type VehPlateCarInfo struct {
	Lpn               string `json:"lpn,omitempty"`          //车牌号码
	LpnScore          string `json:"lpnScore,omitempty"`     //车牌分数
	LpnColor          string `json:"lpnColor,omitempty"`     //车牌颜色
	LpnposLeft        int    `json:"lpnposLeft,omitempty"`   //车牌位置坐标left
	LpnposTop         int    `json:"lpnposTop,omitempty"`    //车牌位置坐标top
	LpnposRight       int    `json:"lpnposRight,omitempty"`  //车牌位置坐标right
	LpnposBottom      int    `json:"lpnposBottom,omitempty"` //车牌位置坐标bottom
	Color             string `json:"color,omitempty"`        //车身颜色
	ColorScore        string `json:"colorScore,omitempty"`   //车身颜色分数
	Brand             string `json:"brand,omitempty"`        //车辆品牌
	Brand0            string `json:"brand0,omitempty"`       //车辆品牌0
	Brand1            string `json:"brand1,omitempty"`       //车辆品牌1
	Brand2            string `json:"brand2,omitempty"`       //车辆品牌2
	Brand3            string `json:"brand3,omitempty"`       //车辆品牌3
	Vehtype           string `json:"vehtype,omitempty"`      //车型
	VehClass          string `json:"vehclass,omitempty"`     //车辆类型
	Type              string `json:"type,omitempty"`         //车辆类型0
	Type0             string `json:"type0,omitempty"`        //车辆类型0
	Type1             string `json:"type1,omitempty"`        //车辆类型1
	Type2             string `json:"type2,omitempty"`        //车辆类型2
	Type3             string `json:"type3,omitempty"`        //车辆类型3
	Subbrand          string `json:"subbrand,omitempty"`     //车辆子品牌
	Subbrand0         string `json:"subbrand0,omitempty"`    //车辆子品牌0
	Subbrand1         string `json:"subbrand1,omitempty"`    //车辆子品牌1
	Subbrand2         string `json:"subbrand2,omitempty"`    //车辆子品牌2
	Subbrand3         string `json:"subbrand3,omitempty"`    //车辆子品牌3
	Year              string `json:"year,omitempty"`         //车辆年份
	Year0             string `json:"year0,omitempty"`        //车辆年份0
	Year1             string `json:"year1,omitempty"`        //车辆年份1
	Year2             string `json:"year2,omitempty"`        //车辆年份2
	Year3             string `json:"year3,omitempty"`        //车辆年份3
	BrandScore        string `json:"brandScore,omitempty"`   //品牌分数
	BrandScore0       string `json:"brandScore0,omitempty"`  //品牌分数0
	BrandScore1       string `json:"brandScore1,omitempty"`  //品牌分数1
	BrandScore2       string `json:"brandScore2,omitempty"`  //品牌分数2
	BrandScore3       string `json:"brandScore3,omitempty"`  //品牌分数3
	Pose              int    `json:"pose,omitempty"`         //车辆整体位置
	CarposLeft        string `json:"carposLeft,omitempty"`   //车辆坐标left
	CarposTop         string `json:"carposTop,omitempty"`    //车辆坐标top
	CarposRight       string `json:"carposRight,omitempty"`  //车辆坐标right
	CarposBottom      string `json:"carposBottom,omitempty"` //车辆坐标bottom
	CarRectScore      string `json:"carRectScore,omitempty"` //车辆位置分数
	CarCapTime        string `json:"carCapTime,omitempty"`
	CarSavePath       string `json:"carSavePath,omitempty"`
	Carnumber         int    `json:"carnumber,omitempty"`
	VehicleCarFeature string `json:"vehicleCarFeature,omitempty"`
	ImgQuality        int    `json:"imgQuality,omitempty"` //车辆或车牌图像质量
	IDNumber          string `json:"IDNumber,omitempty"`   //Json文件唯一值标志
}

type VehPlateCarFeature struct {
	GJ  int `json:"GJ,omitempty"`  //挂件	开始字符为0表示无挂件，开始字符为1表示有挂件，’|’后跟挂件坐标（左上角坐标+右下角坐标），”|”后跟分数
	NJB int `json:"NJB,omitempty"` //年检标	开始字符为0表示无年检标，开始字符为m-n(n>0)表示有m到n个年检标，’|’后是整体坐标，”|”后跟分数
	TC  int `json:"TC,omitempty"`  //天窗	同GJ字段
	AQD int `json:"AQD,omitempty"` //安全带	开始字符0表示主驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数，’ ;’后1表示副驾驶安全带，’|’后0表示无，1表示有安全带，”|”后跟分数
	DDH int `json:"DDH,omitempty"` //开车打电话	同AQD字段
	ZYB int `json:"ZYB,omitempty"` //遮阳板	开始字符0表示主驾驶位置，’|’后0表示无遮阳板，1表示有遮阳板，’|’后是遮阳板坐标，”|”后跟分数，副驾驶位置同理，以’;’分隔，如”0|1|’100,200,300,400’|95.0;1|0|’0,0,0,0’|95.0”表示主驾驶有遮阳板，并给出坐标，副驾驶无遮阳板
	CZH int `json:"CZH,omitempty"` //抽纸盒	开始字符0表示无纸巾盒，开始字符为n表示有n个纸巾盒，‘|’后是n个坐标，以’,’分隔，”|”后是n个分数
	CRZ int `json:"CRZ,omitempty"` //出入证	同GJ字段
	XSB int `json:"XSB,omitempty"` //新手标	同GJ字段
	JSY int `json:"JSY,omitempty"` //驾驶员	同AQD字段
	LT  int `json:"LT,omitempty"`  //轮胎	同CZH字段
	CD  int `json:"CD,omitempty"`  //车灯	同CZH字段
	JSC int `json:"JSC,omitempty"` //驾驶窗	同CZH字段
	HSJ int `json:"HSJ,omitempty"` //后视镜	同CZH字段
	XLJ int `json:"XLJ,omitempty"` //行李架	同CZH字段

}

type VehPlateCars struct {
	CarInfo    VehPlateCarInfo    `json:"carInfo,omitempty"`    //
	CarFeature VehPlateCarFeature `json:"carFeature,omitempty"` //
}

type VehPlateCar struct {
	CarCount int            `json:"carCount,omitempty"` //数量
	Car      []VehPlateCars `json:"car,omitempty"`      //自行车信息
}

type VehPlateTotalFaceInfo struct {
	FaceCount int                `json:"faceCount,omitempty"`
	FaceRsds  []VehPlateFaceInfo `json:"face,omitempty"`
}

type VehPlateFaceInfo struct {
	FaceRectScore  int `json:"faceRectScore,omitempty"`  //人脸框检测得分
	FaceposLeft    int `json:"faceposLeft,omitempty"`    //人脸位置坐标left
	FaceposTop     int `json:"faceposTop,omitempty"`     //人脸位置坐标top
	FaceposRight   int `json:"faceposRight,omitempty"`   //人脸位置坐标right
	FaceposBottom  int `json:"faceposBottom,omitempty"`  //人脸位置坐标bottom
	FacekeyPoint1X int `json:"facekeyPoint1X,omitempty"` //人脸关键点1X坐标
	FacekeyPoint1Y int `json:"facekeyPoint1Y,omitempty"` //人脸关键点1Y坐标
	FacekeyPoint2X int `json:"facekeyPoint2X,omitempty"` //人脸关键点2X坐标
	FacekeyPoint2Y int `json:"facekeyPoint2Y,omitempty"` //人脸关键点2Y坐标
	FacekeyPoint3X int `json:"facekeyPoint3X,omitempty"` //人脸关键点3X坐标
	FacekeyPoint3Y int `json:"facekeyPoint3Y,omitempty"` //人脸关键点3Y坐标
	FacekeyPoint4X int `json:"facekeyPoint4X,omitempty"` //人脸关键点4X坐标
	FacekeyPoint4Y int `json:"facekeyPoint4Y,omitempty"` //人脸关键点4Y坐标
	FacekeyPoint5X int `json:"facekeyPoint5X,omitempty"` //人脸关键点5X坐标
	FacekeyPoint5Y int `json:"facekeyPoint5Y,omitempty"` //人脸关键点5X坐标
	FaceFeature    int `json:"faceFeature,omitempty"`    //人脸特征码
}

type VehPlateImgbuf struct {
	PlateImgBuffer []byte `json:"plateImgBuffer,omitempty"` //车牌图片
	CarImgBuffer   []byte `json:"carImgBuffer,omitempty"`   //车辆图片
}

type VehPlateRecognizeInfo struct {
	Flag             VehPlateflag             `json:"flag,omitempty"`             //
	TotalCarInfo     VehPlateCar              `json:"totalCarInfo,omitempty"`     //
	OtherVehicleInfo VehPlateOtherVehicleInfo `json:"otherVehicleInfo,omitempty"` //
	TotalFaceInfo    VehPlateTotalFaceInfo    `json:"totalFaceInfo,omitempty"`    //
	Imgbuf           VehPlateImgbuf           `json:"img,omitempty"`
}

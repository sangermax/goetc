package main

import (
	"FTCRoute/config"
	"FTCRoute/device"
	"FTCRoute/models"
	"FTCRoute/util"
	"encoding/json"
	"time"
)

var g_pCRmq *device.RabbitMQ

func main() {
	config.InitConfig("../conf/config.conf")

	g_pCRmq = new(device.RabbitMQ)
	err := g_pCRmq.SetupRMQ(config.ConfigData["rmqurl"].(string))
	if err != nil {
		util.FileLogs.Info("SetupRMQ failed: ", err.Error())
		return
	}
	go g_pCRmq.Receive(models.CloudExchange, models.FeeReq, MsgFeeRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.PlatepaycheckReq, MsgRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.PayReq, MsgRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.FlowReq, MsgRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.ImgReq, MsgRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.DeviceReq, MsgRecvProc)
	go g_pCRmq.Receive(models.CloudExchange, models.CtlRst, MsgRecvProc)

	for {
		time.Sleep(60 * time.Second)

		testControl()
	}
}

func MsgFeeRecvProc(inbuf []byte) error {
	util.FileLogs.Info("MsgFeeRecvProc:%s.\r\n", string(inbuf))

	var rstObj models.RstCalcFeeData
	rstObj.Toll = "100"

	var reqinfo models.ReqRMQJsons
	json.Unmarshal(inbuf, &reqinfo)

	var buf models.RstRMQJsons
	buf.Headers = reqinfo.Headers
	buf.Headers.Rsttime = util.GetNow(false)
	buf.Results.ResultDes = ""
	buf.Results.ResultValue = "0"
	buf.Data = rstObj

	jsons, _ := json.Marshal(buf)
	g_pCRmq.Publish(models.CloudExchange, models.FeeRst, jsons)

	return nil
}

func MsgRecvProc(inbuf []byte) error {
	util.FileLogs.Info("MsgCtlRecvProc:%s.\r\n", string(inbuf))
	return nil
}

func testControl() {
	util.FileLogs.Info("testControl.\r\n")

	var reqObj models.ReqControlData
	reqObj.ExitNetwork = "3201"
	reqObj.ExitStationid = "50503"
	reqObj.ExitLandid = "1101"
	reqObj.Cmd = "1"

	var buf models.RstRMQJsons
	buf.Headers.No = "1"
	buf.Headers.Rsttime = util.GetNow(false)
	buf.Results.ResultDes = ""
	buf.Results.ResultValue = "0"
	buf.Data = reqObj

	jsons, _ := json.Marshal(buf)
	g_pCRmq.Publish(models.CloudExchange, models.CtlReq, jsons)
}

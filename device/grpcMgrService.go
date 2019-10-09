package device

import (
	"FTC/models"
	pb "FTC/pb"
	"FTC/util"
)

type GrpcConnInfo struct {
	key    string
	State  int
	stream pb.GrpcMsg_CommuniteServer
}

type GrpcManage struct {
	GrpcManageType int
	GrpcMap        *util.BeeMap
}

func (p *GrpcManage) GrpcManageInit(itype int) {
	p.GrpcManageType = itype
	p.GrpcMap = util.NewBeeMap()
}

func (p *GrpcManage) AddGrpcConn(key string, stream pb.GrpcMsg_CommuniteServer) {
	info := new(GrpcConnInfo)
	info.key = key
	info.State = models.DEVSTATE_UNKNOWN
	info.stream = stream
	p.GrpcMap.Set(key, info)

	util.ConsoleLogs.Info("%s GrpcManage 添加新连接:size:%d,%s.\r\n", GetGrpcDes(p.GrpcManageType), p.GrpcMap.Size(), key)
}

func (p *GrpcManage) RemoveGrpcConn(k string) {
	p.GrpcMap.Delete(k)

	util.ConsoleLogs.Info("%s GrpcManage删除连接:size:%d.\r\n", GetGrpcDes(p.GrpcManageType), p.GrpcMap.Size())
}

func (p *GrpcManage) GetGrpcConnStream(k string) pb.GrpcMsg_CommuniteServer {
	v := p.GrpcMap.Get(k)
	if v != nil {
		return v.(*GrpcConnInfo).stream
	}

	return nil
}

func (p *GrpcManage) UpdateDevState(k string, state string) {
	v := p.GrpcMap.Get(k)
	if v != nil {
		v.(*GrpcConnInfo).State = util.ConvertS2I(state)
	}

	return
}

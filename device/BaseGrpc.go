package device

import (
	"FTC/models"
	"FTC/pb"
	"FTC/util"
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

//收到应答处理
type GrpcRecvProc func(*pb.Message)

//连接初始化
type GrpcConnectedInit func()

type GrpcClient struct {
	devtype   int
	conn      *grpc.ClientConn
	Connstate bool
	stream    pb.GrpcMsg_CommuniteClient
	strAddr   string
}

func (p *GrpcClient) GrpcInit(grpcdev int, strAddr string, funcproc GrpcRecvProc) {
	p.devtype = grpcdev
	p.strAddr = strAddr
	p.Connstate = false

	go p.goGrpcRecvproc(nil, funcproc)
}

func (p *GrpcClient) GrpcInit2(grpcdev int, strAddr string, funcConnectedInit GrpcConnectedInit, funcproc GrpcRecvProc) {
	p.devtype = grpcdev
	p.strAddr = strAddr
	p.Connstate = false

	go p.goGrpcRecvproc(funcConnectedInit, funcproc)
}

func (p *GrpcClient) IsGrpcConn() bool {
	return p.Connstate
}

func (p *GrpcClient) GrpcClose() {
	p.conn.Close()
	p.Connstate = false
}

func (p *GrpcClient) GrpcConn(funcConnectedInit GrpcConnectedInit) (err error) {
	if p.Connstate {
		p.GrpcClose()
	}

	p.conn, err = grpc.Dial(p.strAddr, grpc.WithInsecure())
	if err != nil {
		//fmt.Printf("%s GrpcConn fail:%s.\r\n", GetGrpcDes(p.devtype), err.Error())
		return err
	}

	client := pb.NewGrpcMsgClient(p.conn)
	p.stream, err = client.Communite(context.Background())
	if err != nil {
		//fmt.Printf("%s GrpcConn fail:%s.\r\n", GetGrpcDes(p.devtype), err.Error())
		return err
	}

	p.Connstate = true

	fmt.Printf("%s GrpcConn suc.\r\n", GetGrpcDes(p.devtype))
	if funcConnectedInit != nil {
		funcConnectedInit()
	}

	return nil
}

func (p *GrpcClient) GrpcSendproc(sType string, no string, msg proto.Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		//fmt.Printf("GrpcSendproc error:%s.\r\n", err.Error())
		return err
	}

	notes := []*pb.Message{
		{Type: sType, No: no, Data: out},
	}

	if p.stream != nil {
		if sType != models.GRPCTYPE_DEVSTATE {
			fmt.Println("GrpcSendproc:", GetGrpcDes(p.devtype), GetCmdDes(sType))
			//fmt.Println(notes[0])
		}

		p.stream.Send(notes[0])
	}

	return nil
}

func (p *GrpcClient) GrpcSendprocWithResult(sType, no string, msg proto.Message, result models.ResultInfo) {
	out, err := proto.Marshal(msg)
	if err != nil {
		util.FileLogs.Info("%s GrpcSendprocWithResult error:%s.\r\n", GetGrpcDes(p.devtype), err.Error())
		return
	}

	notes := []*pb.Message{
		{Type: sType, No: no, Data: out, Resultvalue: result.ResultValue, Resultdes: result.ResultDes},
	}

	if p.stream != nil {
		if sType != models.GRPCTYPE_DEVSTATE {
			fmt.Println("GrpcSendprocWithResult:", GetGrpcDes(p.devtype), GetCmdDes(sType))
			//fmt.Println(notes[0])
		}

		p.stream.Send(notes[0])
	}
}

func (p *GrpcClient) goGrpcRecvproc(funcConnectedInit GrpcConnectedInit, funcproc GrpcRecvProc) {
	for {
		//判断连接
		if !p.Connstate {
			p.GrpcConn(funcConnectedInit)
			util.MySleep_s(5)
			continue
		}

		in, err := p.stream.Recv()
		if err == io.EOF {
			//fmt.Println("goGrpcRecvproc read done ")
			p.GrpcClose()
			util.MySleep_s(5)
			continue
		}

		if err != nil {
			//fmt.Printf("goGrpcRecvproc Failed to receive a note : %s.\r\n", err.Error())
			p.GrpcClose()
			util.MySleep_s(5)
			continue
		}

		if in.Type == models.GRPCTYPE_DEVSTATE || in.Type == models.GRPCTYPE_PLATESTATE ||
			in.Type == models.GRPCTYPE_IOSTATE || in.Type == models.GRPCTYPE_SCANSTATE ||
			in.Type == models.GRPCTYPE_PRINTERSTATE {

		} else {
			//fmt.Println("GrpcClient RECV ", GetGrpcDes(p.devtype))
			//fmt.Println(in)
		}

		funcproc(in)
	}

}

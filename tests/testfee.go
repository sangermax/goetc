package main

import (
	"fmt"
	"time"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "FTC/pb"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":6002"
)


type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) AskFee(ctx context.Context, in *pb.AskRequest) (*pb.AskReply, error) {
	fmt.Printf("%s.\r\n",in.Vehplate)
	time.Sleep(30 * time.Second)

	return &pb.AskReply{Toll: in.Vehplate+"100"}, nil
}

func (s *server) CheckValid(ctx context.Context, in *pb.CheckValidRequest) (*pb.CheckValidReply, error) {
	return &pb.CheckValidReply{Result:"1"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCloudFeeSrvServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

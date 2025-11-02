package main

import (
	"fmt"
	"log"
	"net"

	pb "belajar-chatting-app-1/proto"
	"belajar-chatting-app-1/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServiceServer(s, server.NewChatServer())

	fmt.Println("🚀 gRPC Chat Service running on port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

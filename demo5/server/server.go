package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "envoy-demo5/protos"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedHelloServiceServer
}

// const target string = "localhost:5050" -> this is fine for native deployment
const target string = "0.0.0.0:5050" //   -> containerized deployment

func (*server) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	name := req.Name
	resp := &pb.HelloResponse{Greeting: "Hello " + name}
	return resp, nil
}

func main() {
	lis, err := net.Listen("tcp", target)
	if err != nil {
		log.Fatalf("Backend Error %v", err)
	}
	fmt.Printf("Server is listening on %v...\n", target)

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &server{})

	s.Serve(lis)
}

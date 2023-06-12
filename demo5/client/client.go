package main

import (
	"context"
	"fmt"
	"log"

	pb "envoy-demo5/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// const target string = "localhost:5050" -> native deployment
const target string = "localhost:1337" // -> containerized deployment

func main() {
	// Create a client connection to localhost:5050 (local development)
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cli := pb.NewHelloServiceClient(conn)
	req := &pb.HelloRequest{Name: "Xingjian"}

	// Authorization header for the original request
	ctx := metadata.AppendToOutgoingContext(
		context.Background(),
		"Authorization", "Bearer foo",
		"Bar", "baz",
	)

	resp, err := cli.Hello(ctx, req)
	if err != nil {
		errStatus, isGrpcErr := status.FromError(err)
		if !isGrpcErr {
			fmt.Printf("Unknown error! %v", errStatus.Message())
			return
		}
		code := errStatus.Code()
		msg := errStatus.Message()
		fmt.Println(code)
		fmt.Println(msg)
	} else {
		fmt.Printf("Receive response => [%v]\n", resp.Greeting)
	}
}

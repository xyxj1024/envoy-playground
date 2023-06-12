package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	echo "envoy-demo4/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	_ "google.golang.org/grpc/resolver" // for "dns:///be.cluster.local:50051"
	_ "google.golang.org/grpc/xds"      // for xds-experimental:///be-srv
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 10
)

func main() {
	address := flag.String("host", "dns:///be.cluster.local:50051", "dns:///be.cluster.local:50051 or xds-experimental:///be-srv")
	flag.Parse()

	go func() {
		lis, err := net.Listen("tcp", ":19000")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()

		var grpcOptions []grpc.ServerOption
		grpcOptions = append(grpcOptions,
			grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				Time:    grpcKeepaliveTime,
				Timeout: grpcKeepaliveTimeout,
			}),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             grpcKeepaliveMinTime,
				PermitWithoutStream: true,
			}),
		)
		grpcServer := grpc.NewServer(grpcOptions...)

		cleanup, err := admin.Register(grpcServer)
		if err != nil {
			log.Fatalf("failed to register admin services: %v", err)
		}
		defer cleanup()

		log.Printf("Admin port listen on :%s", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	conn, err := grpc.Dial(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	cli := echo.NewEchoServiceClient(conn)
	ctx := context.Background()
	for i := 0; i < 60; i++ {
		res, err := cli.SayHello(ctx, &echo.EchoRequest{Name: "unary RPC msg"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("RPC Response: %v %v", i, res)
		time.Sleep(2 * time.Second)
	}
}

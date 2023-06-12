package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	echo "envoy-demo4/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

var (
	grpcPort   string
	serverName string
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 10
)

type server struct {
	echo.UnimplementedEchoServiceServer
}

func isGrpcRequest(req *http.Request) bool {
	return req.ProtoMajor == 2 && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc")
}

func (srv *server) SayHello(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	log.Println("Got RPC: -->", req.Name)
	return &echo.EchoResponse{Message: "Hello " + req.Name + " from " + serverName}, nil
}

func (srv *server) SayHelloStream(req *echo.EchoRequest, stream echo.EchoService_SayHelloStreamServer) error {
	log.Println("Got stream: -->")
	stream.Send(&echo.EchoResponse{Message: "Hello " + req.Name})
	stream.Send(&echo.EchoResponse{Message: "Hello " + req.Name})
	return nil
}

type healthServer struct{}

func (hsrv *healthServer) Check(ctx context.Context, hc *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling gRPC health check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (hsrv *healthServer) Watch(hc *healthpb.HealthCheckRequest, wsrv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func init() {
	// no default values specified
	flag.StringVar(&grpcPort, "grpcport", "", "gRPC port")
	flag.StringVar(&serverName, "servername", "", "gRPC server name")
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

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

	echo.RegisterEchoServiceServer(grpcServer, &server{})
	healthpb.RegisterHealthServer(grpcServer, &healthServer{})

	log.Println("Starting gRPC Server...")
	grpcServer.Serve(lis)
}

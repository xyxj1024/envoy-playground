package main

import (
	// Standard library
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"

	// Third-party
	"github.com/gogo/googleapis/google/rpc"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	// Envoy go-control-plane
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
)

const (
	grpcPort                 = ":4040"
	grpcMaxConcurrentStreams = 10
)

type healthServer struct{}

func (hsrv *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Printf("Handling gRPC Check request")
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (hsrv *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

type AuthorizationServer struct{}

func (asrv *AuthorizationServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	log.Println(">>> Authorization called check()")

	ibh, err := json.MarshalIndent(req.Attributes.Request.Http.Headers, "", "  ")
	if err == nil {
		log.Println("Inbound Headers: ")
		log.Println((string(ibh)))
	}

	cex, err := json.MarshalIndent(req.Attributes.ContextExtensions, "", "  ")
	if err == nil {
		log.Println("Context Extensions: ")
		log.Println((string(cex)))
	}

	authHeader, ok := req.Attributes.Request.Http.Headers["authorization"]
	var splitToken []string

	if ok {
		splitToken = strings.Split(authHeader, "Bearer ")
	}
	if len(splitToken) == 2 {
		token := splitToken[1]

		if token == "foo" {
			return &auth.CheckResponse{
				Status: &rpcstatus.Status{
					Code: int32(rpc.OK),
				},
				HttpResponse: &auth.CheckResponse_OkResponse{
					OkResponse: &auth.OkHttpResponse{
						Headers: []*core.HeaderValueOption{
							{
								Header: &core.HeaderValue{
									Key:   "x-custom-header-from-authz",
									Value: "permitted",
								},
							},
						},
					},
				},
			}, nil
		} else {
			return &auth.CheckResponse{
				Status: &rpcstatus.Status{
					Code: int32(rpc.PERMISSION_DENIED),
				},
				HttpResponse: &auth.CheckResponse_DeniedResponse{
					DeniedResponse: &auth.DeniedHttpResponse{
						Status: &envoy_type.HttpStatus{
							Code: envoy_type.StatusCode_Unauthorized,
						},
						Body: "PERMISSION_DENIED",
					},
				},
			}, nil

		}

	}
	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.UNAUTHENTICATED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Unauthorized,
				},
				Body: "Authorization Header malformed or not provided",
			},
		},
	}, nil
}

func main() {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	auth.RegisterAuthorizationServer(grpcServer, &AuthorizationServer{})
	healthpb.RegisterHealthServer(grpcServer, &healthServer{})
	log.Printf("Starting gRPC Server at %s", grpcPort)
	grpcServer.Serve(lis)
}

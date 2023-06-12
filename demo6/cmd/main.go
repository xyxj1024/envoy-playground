package main

import (
	// Standard library
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Self-defined
	internalLogger "envoy-demo4/internal/logger"
	"envoy-demo4/pkg/callback"
	"envoy-demo4/pkg/logger"
	"envoy-demo4/pkg/snapshot"

	// Third-party libraries
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	xdsClusterName string
	xdsPort        uint
	debug          bool
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func init() {
	// Name of the cluster which provides Envoy instances ADS/SDS subscription
	flag.StringVar(&xdsClusterName, "xds-cluster", "control_plane", "xDS cluster name")
	// Port number to which Envoy instances are bound for configuration updates
	flag.UintVar(&xdsPort, "xds-port", 18000, "xDS management server port")
	flag.BoolVar(&debug, "debug", true, "Enable xDS server debug logging")
}

func main() {
	flag.Parse()
	internalLogger.BootLogger(debug)
	ctx := context.Background()
	internalLogger.Infof("Starting control plane")

	signal := make(chan struct{})
	cb := &callback.Callbacks{
		Signal:         signal,
		Fetches:        0,
		Requests:       0,
		DeltaRequests:  0,
		DeltaResponses: 0,
	}

	config := cache.NewSnapshotCache(
		false,                 // disable the ADS flag
		snapshot.StaticHash{}, // use the constant node hash
		internalLogger.Instance().WithFields(logger.Fields{"area": "SnapshotCache"}),
	)
	srv := server.NewServer(ctx, config, cb)
	manager := snapshot.NewManager(
		adsProvider,
		sdsProvider,
		config,
		internalLogger.Instance().WithFields(logger.Fields{"area": "SnapshotManager"}),
	)
	events := generateWatchers(ctx)
	go manager.Listen(events)
	go runManagementServer(ctx, srv, xdsPort)

	waitForSignal(ctx)
}

func generateWatchers(ctx context.Context) chan string {
	updateEvents := make(chan string)
	log := internalLogger.Instance().WithFields(logger.Fields{"area": "UpdateEventsWatcher"})

	// go watcher.ForSwarmEvent(log).Start(ctx, updateEvents)
	// go watcher.CreateInitialStartupEvent(updateEvents)

	return updateEvents
}

func runManagementServer(ctx context.Context, srv server.Server, port uint) {
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		internalLogger.Fatalf(err.Error())
	}

	registerServices(grpcServer, srv)

	internalLogger.Infof("xDS Management server listening on %d\n", port)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			internalLogger.Errorf(err.Error())
		}
	}()
	<-ctx.Done()
	grpcServer.GracefulStop()
}

func registerServices(grpcServer *grpc.Server, srv server.Server) {
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, srv)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, srv)
}

func waitForSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	internalLogger.Infof("Shutting down control plane upon receiving SIGINT...")
	ctx.Done()
}

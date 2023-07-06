package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"envoy-swarm-control/pkg/callback"
	"envoy-swarm-control/pkg/snapshot"
	"envoy-swarm-control/pkg/watcher"

	docker "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	debug          bool
	xdsPort        uint
	ingressNetwork string
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Enable xDS server debug logging")
	flag.UintVar(&xdsPort, "xds-port", 18000, "xDS management server port")                              // Port number to which Envoy instances are bound for configuration updates
	flag.StringVar(&ingressNetwork, "ingress-network", "mesh-traffic", "Docker overlay network name/ID") // Deploy using: docker network create --driver=overlay --attachable mesh-traffic
}

func main() {
	flag.Parse()
	mainctx := context.Background()
	logrus.Infof("Starting control plane")

	// Create xDS management server
	signal := make(chan struct{})
	cb := &callback.Callbacks{
		Signal:         signal,
		Fetches:        0,
		Requests:       0,
		DeltaRequests:  0,
		DeltaResponses: 0,
	}
	config := cache.NewSnapshotCache(
		true, // enable the ADS flag
		cache.IDHash{},
		nil,
	)
	srv := server.NewServer(mainctx, config, cb)

	manager := snapshot.NewManager(config)
	update := generateWatcher(mainctx)
	go manager.Discover(update, mainctx)

	// Run xDS management server
	go runManagementServer(mainctx, srv, xdsPort)

	waitForSignal(mainctx)
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
		logrus.Fatalf(err.Error())
	}

	registerServices(grpcServer, srv)

	logrus.Infof("xDS Management server listening on %d", port)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Errorf(err.Error())
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
	// secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, srv)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, srv)
}

/* Function generateWatcher:
 * creates a new watcher for Docker events and an initial update channel.
 */
func generateWatcher(ctx context.Context) chan snapshot.ServiceLabels {
	updateChannel := make(chan snapshot.ServiceLabels)

	go watcher.StartWatcher(
		ctx,
		newDockerClient(),
		ingressNetwork,
		updateChannel,
	)

	go watcher.InitUpdateChannel(updateChannel)

	return updateChannel
}

func newDockerClient() *docker.Client {
	httpHeaders := map[string]string{
		"User-Agent": "envoy-swarm-control",
	}

	c, err := docker.NewClientWithOpts(
		docker.FromEnv,
		docker.WithHTTPHeaders(httpHeaders),
		docker.WithAPIVersionNegotiation(), // For "Maximum supported API version is 1.41"
	)
	if err != nil {
		panic(err)
	}

	return c
}

func waitForSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	logrus.Infof("Shutting down control plane upon receiving SIGINT...")
	ctx.Done()
}

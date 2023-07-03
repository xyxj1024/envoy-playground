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
	"envoy-swarm-control/pkg/logger"
	"envoy-swarm-control/pkg/snapshot"
	"envoy-swarm-control/pkg/storage"
	"envoy-swarm-control/pkg/watcher"
	"envoy-swarm-control/pkg/xds"
	"envoy-swarm-control/pkg/xds/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	debug                bool
	xdsClusterName       string
	xdsPort              uint
	ingressNetwork       string
	certificateDirectory string
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Enable xDS server debug logging")
	flag.StringVar(&xdsClusterName, "xds-cluster", "control-plane", "xDS cluster name")                  // Name of the cluster which provides Envoy instances ADS/SDS subscription
	flag.UintVar(&xdsPort, "xds-port", 18000, "xDS management server port")                              // Port number to which Envoy instances are bound for configuration updates
	flag.StringVar(&ingressNetwork, "ingress-network", "mesh-traffic", "Docker overlay network name/ID") // Deploy using: docker network create --driver=overlay --attachable mesh-traffic
	flag.StringVar(&certificateDirectory, "cert-dir", "cert", "OpenSSL X.509 certificate file path")
}

func main() {
	flag.Parse()
	logger.BootLogger(debug)
	mainctx := context.Background()
	logger.Infof("Starting control plane")

	// Deploy()

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
		logger.Instance().WithFields(logger.Fields{"area": "snapshot-cache"}),
	)
	srv := server.NewServer(mainctx, config, cb)

	// Create and run watcher for Docker events
	sdsProvider := setupTLS()
	adsProvider := setupDiscovery(sdsProvider)
	manager := snapshot.NewManager(
		adsProvider,
		sdsProvider,
		config,
		logger.Instance().WithFields(logger.Fields{"area": "snapshot-manager"}),
	)
	events := generateWatcher(mainctx)
	go manager.Listen(events)

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
		logger.Fatalf(err.Error())
	}

	registerServices(grpcServer, srv)

	logger.Infof("xDS Management server listening on %d", port)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logger.Errorf(err.Error())
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

/* Function setupDiscovery:
 * returns an ADS provider whose "dockerClient" field is handled by xds/provider.go.
 */
func setupDiscovery(sdsProvider xds.SDS) xds.ADS {
	listenerBuilder := xds.NewListenerProvider(sdsProvider)

	return xds.NewADSProvider(
		ingressNetwork,
		listenerBuilder,
		logger.Instance().WithFields(logger.Fields{"area": "ads-provider"}),
	)
}

/* Function setupTLS:
 * fetches certificate files and returns a SDS provider.
 */
func setupTLS() (sdsProvider xds.SDS) {
	fileStorage := getStorage()
	certificateStorage := &tls.Certificate{Storage: fileStorage}
	sdsProvider = tls.NewCertificateSecretsProvider(
		xdsClusterName,
		certificateStorage,
		logger.Instance().WithFields(logger.Fields{"area": "sds-provider"}),
	)

	return sdsProvider
}

/* Function generateWatcher:
 * creates a new watcher for Docker events and an initial update channel.
 */
func generateWatcher(ctx context.Context) chan snapshot.UpdateReason {
	updateChannel := make(chan snapshot.UpdateReason)
	log := logger.Instance().WithFields(logger.Fields{"area": "docker-events-watcher"})

	go watcher.NewSwarmEvent(log).StartWatcher(ctx, updateChannel)

	go watcher.InitUpdateChannel(updateChannel)

	return updateChannel
}

func getStorage() storage.Storage {
	disk := storage.NewDiskStorage(certificateDirectory, logger.Instance().WithFields(logger.Fields{"area": "disk"}))
	return disk
}

func waitForSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	logger.Infof("Shutting down control plane upon receiving SIGINT...")
	ctx.Done()
}

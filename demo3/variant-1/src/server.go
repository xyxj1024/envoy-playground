package main

import (
	// standard library
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	// third-party
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// go-control-plane
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	debug       bool
	port        uint
	gatewayPort uint
	mode        string
	version     int32
	config      cache.SnapshotCache
)

const (
	Ads                      = "ads"
	Xds                      = "xds"
	Rest                     = "rest"
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.UintVar(&gatewayPort, "gateway", 18001, "Management server port for HTTP gateway")
	flag.StringVar(&mode, "ads", Ads, "Management server type (ads only now)")
}

func RunManagementServer(ctx context.Context, server server.Server, port uint) {
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
		logrus.WithError(err).Fatal("Failed to listen")
	}
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	logrus.WithFields(logrus.Fields{"port": port}).Info("Management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Error(err)
		}
	}()
	<-ctx.Done()
	grpcServer.GracefulStop()
}

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	logrus.Printf("Starting control plane")

	signal := make(chan struct{})
	cb := &Callbacks{
		Signal:         signal,
		Fetches:        0,
		Requests:       0,
		DeltaRequests:  0,
		DeltaResponses: 0,
	}

	// create a configuration cache
	config = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
	// create an xDS server
	srv := server.NewServer(ctx, config, cb)
	// start the xDS server
	go RunManagementServer(ctx, srv, port)
	<-signal
	for {
		// read upstream IP address
		logrus.Printf("Enter remote host: ")
		reader := bufio.NewReader(os.Stdin)
		upstreamHost, err := reader.ReadString('\n')
		if err != nil {
			logrus.Fatal(err)
		}
		upstreamHost = strings.ReplaceAll(upstreamHost, "\n", "")
		// read upstream cluster name
		logrus.Printf("Enter cluster name: ")
		reader = bufio.NewReader(os.Stdin)
		clusterName, err := reader.ReadString('\n')
		if err != nil {
			logrus.Fatal(err)
		}
		clusterName = strings.ReplaceAll(clusterName, "\n", "")
		// increment version number
		atomic.AddInt32(&version, 1)
		// make and set a snapshot
		GenerateSnapshot(ctx, config, clusterName, upstreamHost, version)
		time.Sleep(60 * time.Second)
	}
}

/*
func ReadFile() [][]string {
	dat, err := os.ReadFile("domains.csv")
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(strings.NewReader(string(dat)))
	var domains [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.WithError(err).Fatal("failed to read")
		}
		if record[0] == "domain_name" {
			continue
		}
		domains = append(domains, record)
	}
	return domains
}
*/

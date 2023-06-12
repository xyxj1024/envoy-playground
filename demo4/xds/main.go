package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	debug         bool
	port          uint
	gatewayPort   uint
	mode          string
	version       int32
	upstreamPorts UpstreamPorts
	config        cache.SnapshotCache
)

const (
	localhost                = "127.0.0.1"
	backendHostName          = "be.cluster.local"
	clusterName              = "be-srv-cluster"
	virtualHostName          = "be-srv-vs"
	listenerName             = "be-srv"
	routeConfigName          = "be-srv-route"
	Ads                      = "ads"
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

type UpstreamPorts []uint

func (u *UpstreamPorts) String() string {
	return strings.Join(strings.Fields(fmt.Sprint(*u)), ",")
}

func (u *UpstreamPorts) Set(port string) error {
	logrus.Printf("[upstream port] %s", port)
	u64, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		logrus.Fatal(err)
	}
	*u = append(*u, uint(u64))
	return nil
}

type Callbacks struct {
	Signal         chan struct{}
	Debug          bool
	Fetches        int
	Requests       int
	DeltaRequests  int
	DeltaResponses int
	mu             sync.Mutex
}

var _ server.Callbacks = &Callbacks{}

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	logrus.WithFields(logrus.Fields{"fetches": cb.Fetches, "requests": cb.Requests}).Info("Report() callbacks")
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	logrus.Infof("OnStreamOpen %d of type %v", id, typ)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64, node *core.Node) {
	logrus.Infof("OnStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	logrus.Infof("OnDeltaStreamOpen %d of type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	logrus.Infof("OnDeltaStreamClosed %d for node %s", id, node.Id)
}

func (cb *Callbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	logrus.Infof("OnStreamRequest %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Requests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	logrus.Infof("OnStreamResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.Report()
}

func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	logrus.Infof("OnStreamDeltaResponse... %d Request [%v], Response [%v]", id, req.TypeUrl, res.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}

func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	logrus.Infof("OnStreamDeltaRequest... %d Request [%v]", id, req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaRequests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *discovery.DiscoveryRequest) error {
	logrus.Infof("OnFetchRequest... Request [%v]", req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchResponse(req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	logrus.Infof("OnFetchResponse... Request [%v], Response [%v]", req.TypeUrl, res.TypeUrl)
}

func RunManagementServer(ctx context.Context, srv server.Server, port uint) {
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

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	logrus.WithFields(logrus.Fields{"port": port}).Info("Management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			logrus.Error(err)
		}
	}()
	<-ctx.Done()
	grpcServer.GracefulStop()
}

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 18000, "Management server port")
	flag.UintVar(&gatewayPort, "gateway", 18001, "Management server port for HTTP gateway")
	flag.StringVar(&mode, "ads", Ads, "Management server type (ads only now)")
	flag.Var(&upstreamPorts, "upstream_port", "List of upstream gRPC ports")
}

func main() {
	flag.Parse()
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
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

	config = cache.NewSnapshotCache(true, cache.IDHash{}, nil)
	srv := server.NewServer(ctx, config, cb)
	go RunManagementServer(ctx, srv, port)

	<-signal

	cb.Report()

	nodeId := config.GetStatusKeys()[0]
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot for nodeID " + fmt.Sprint(nodeId))

	var lbEndpoints []*endpoint.LbEndpoint
	var index int = 0

	for {
		if index+1 <= len(upstreamPorts) {
			p := upstreamPorts[index]
			index++

			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating endpoint for %s:%d", backendHostName, p)
			hst := &core.Address{Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address:  backendHostName,
					Protocol: core.SocketAddress_TCP,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(p),
					},
				},
			}}

			ep := &endpoint.LbEndpoint{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: hst,
					}},
				HealthStatus: core.HealthStatus_HEALTHY,
			}
			lbEndpoints = append(lbEndpoints, ep)

			e := []types.Resource{
				&endpoint.ClusterLoadAssignment{
					ClusterName: clusterName,
					Endpoints: []*endpoint.LocalityLbEndpoints{{
						Locality: &core.Locality{
							Region: "us-central1",
							Zone:   "us-central1-a",
						},
						Priority:            0,
						LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(1000)},
						LbEndpoints:         lbEndpoints,
					}},
				},
			}

			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating cluster " + clusterName)
			c := []types.Resource{
				&cluster.Cluster{
					Name:                 clusterName,
					LbPolicy:             cluster.Cluster_ROUND_ROBIN,
					ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
					EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
						EdsConfig: &core.ConfigSource{
							ConfigSourceSpecifier: &core.ConfigSource_Ads{},
						},
					},
				},
			}

			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating routing for " + virtualHostName)
			r := []types.Resource{
				&route.RouteConfiguration{
					Name:             routeConfigName,
					ValidateClusters: &wrapperspb.BoolValue{Value: true},
					VirtualHosts: []*route.VirtualHost{{
						Name:    virtualHostName,
						Domains: []string{listenerName}, // must match what is specified in xDS
						Routes: []*route.Route{{
							Match: &route.RouteMatch{
								PathSpecifier: &route.RouteMatch_Prefix{
									Prefix: "",
								},
							},
							Action: &route.Route_Route{
								Route: &route.RouteAction{
									ClusterSpecifier: &route.RouteAction_Cluster{
										Cluster: clusterName,
									},
								},
							},
						},
						},
					}},
				},
			}

			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating listener " + listenerName)
			hcRds := &hcm.HttpConnectionManager_Rds{
				Rds: &hcm.Rds{
					RouteConfigName: routeConfigName,
					ConfigSource: &core.ConfigSource{
						ResourceApiVersion: core.ApiVersion_V3,
						ConfigSourceSpecifier: &core.ConfigSource_Ads{
							Ads: &core.AggregatedConfigSource{},
						},
					},
				},
			}

			hff := &router.Router{}
			tctx, err := anypb.New(hff)
			if err != nil {
				logrus.Errorf("could not unmarshall router: %v\n", err)
				os.Exit(1)
			}

			manager := &hcm.HttpConnectionManager{
				CodecType:      hcm.HttpConnectionManager_AUTO,
				RouteSpecifier: hcRds,
				HttpFilters: []*hcm.HttpFilter{{
					Name: wellknown.Router,
					ConfigType: &hcm.HttpFilter_TypedConfig{
						TypedConfig: tctx,
					},
				}},
			}

			pbst, err := anypb.New(manager)
			if err != nil {
				logrus.Fatal(err)
			}

			l := []types.Resource{&listener.Listener{
				Name: listenerName,
				ApiListener: &listener.ApiListener{
					ApiListener: pbst,
				},
			}}

			atomic.AddInt32(&version, 1)
			logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))
			resources := make(map[resource.Type][]types.Resource, 4)
			resources[resource.ClusterType] = c
			resources[resource.ListenerType] = l
			resources[resource.RouteType] = r
			resources[resource.EndpointType] = e

			snap, _ := cache.NewSnapshot(fmt.Sprint(version), resources)
			/*
				if err := snap.Consistent(); err != nil {
					logrus.Errorf("Snapshot inconsistency: %+v\n%+v", snap, err)
					os.Exit(1)
				}
			*/

			if err = config.SetSnapshot(ctx, nodeId, snap); err != nil {
				logrus.Fatalf("Snapshot error %q for %+v", err, snap)
			}

			logrus.Infof("Snapshot served: %+v", snap)
			time.Sleep(30 * time.Second)
		}
	}
}

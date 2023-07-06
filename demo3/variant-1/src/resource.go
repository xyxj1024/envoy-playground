package main

import (
	// Standard library
	"context"
	"fmt"
	"os"
	"time"

	// Third-party library
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	// Envoy go-control-plane
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	// clusterName     = "my_envoy_control"
	listenerName    = "listener_0"
	secretName      = "server_cert"
	virtualHostName = "local_service"
	routeConfigName = "local_route"
	// upstreamHost    = "www.google.com"
)

// The following two functions are taken from https://github.com/istio/istio/blob/master/pilot/pkg/util/protoconv/protoconv.go
// messageToAnyWithError converts from proto message to proto Any
func messageToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		// nolint: staticcheck
		TypeUrl: "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

// messageToAny converts from proto message to proto Any
func messageToAny(msg proto.Message) *anypb.Any {
	out, err := messageToAnyWithError(msg)
	if err != nil {
		logrus.Error(fmt.Sprintf("Error marshaling Any %s: %v", prototext.Format(msg), err))
		return nil
	}
	return out
}

func makeCluster(clusterName string, upstreamHost string) *cluster.Cluster {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating cluster with clusterName %s, upstreamHost %s", clusterName, upstreamHost)

	hst := &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  upstreamHost,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(443),
				},
			},
		},
	}
	uctx := &tls.UpstreamTlsContext{}
	tctx, err := anypb.New(uctx)
	if err != nil {
		logrus.Fatal(err)
	}

	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(2 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{{
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: hst,
							}},
					},
				},
			}},
		},
		TransportSocket: &core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tctx,
			},
		},
	}
}

func makeHTTPListener(pub []byte, priv []byte, clusterName string, upstreamHost string) *listener.Listener {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating listener with listenerName " + listenerName)

	rte := &route.RouteConfiguration{
		Name: routeConfigName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    virtualHostName,
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
						PrefixRewrite: "/robots.txt",
						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: upstreamHost,
						},
					},
				},
			}},
		}},
	}

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: rte,
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: messageToAny(&router.Router{}),
			},
		}},
	}

	pbst, err := anypb.New(manager)
	if err != nil {
		logrus.Fatal(err)
	}

	// 1. send TLS certs filename back directly
	sdsTls := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificates: []*tls.TlsCertificate{{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(pub)},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(priv)},
				},
			}},
		},
	}

	/* or
	// 2. send TLS SDS Reference value
	sdsTls := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tls.SdsSecretConfig{{
				Name: "server_cert",
			}},
		},
	}

	// 3. SDS via ADS
	sdsTls := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tls.SdsSecretConfig{{
				Name: "server_cert",
				SdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
					ResourceApiVersion: core.ApiVersion_V3,
				},
			}},
		},
	}
	*/

	scfg, err := anypb.New(sdsTls)
	if err != nil {
		logrus.Fatal(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "127.0.0.1",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 10000,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
			TransportSocket: &core.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: scfg,
				},
			},
		}},
	}
}

func makeSecret(pub []byte, priv []byte) *tls.Secret {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating secret with secretName " + secretName)
	return &tls.Secret{
		Name: secretName,
		Type: &tls.Secret_TlsCertificate{
			TlsCertificate: &tls.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(pub)},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(priv)},
				},
			},
		},
	}
}

func GenerateSnapshot(ctx context.Context, config cache.SnapshotCache, clusterName string, upstreamHost string, version int32) {
	// the first connected Envoy instance
	nodeId := config.GetStatusKeys()[0]
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot " + fmt.Sprint(version) + ", nodeID " + fmt.Sprint(nodeId))

	pub, err := os.ReadFile("../certs/envoy-proxy-server.crt")
	if err != nil {
		logrus.Fatal(err)
	}
	priv, err := os.ReadFile("../certs/envoy-proxy-server.key")
	if err != nil {
		logrus.Fatal(err)
	}

	resources := make(map[string][]types.Resource, 3)
	resources[resource.ClusterType] = []types.Resource{makeCluster(clusterName, upstreamHost)}
	resources[resource.ListenerType] = []types.Resource{makeHTTPListener(pub, priv, clusterName, upstreamHost)}
	resources[resource.SecretType] = []types.Resource{makeSecret(pub, priv)}

	// create the snapshot that Envoy will serve
	snap, _ := cache.NewSnapshot(fmt.Sprint(version), resources)
	if err := snap.Consistent(); err != nil {
		logrus.Errorf("Snapshot inconsistency: %+v\n%+v", snap, err)
		os.Exit(1)
	}
	logrus.Infof("Serve snapshot %+v", snap)

	// add the snapshot to the cache
	if err = config.SetSnapshot(ctx, nodeId, snap); err != nil {
		logrus.Fatalf("Snapshot error %q for %+v", err, snap)
	}
}

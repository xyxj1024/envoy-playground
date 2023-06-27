package convert

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type FilterChainBuilder struct {
	name                 string
	sniServerNames       []string
	sdsCertificateConfig *auth.SdsSecretConfig
	vhosts               []*route.VirtualHost
}

func NewFilterChainBuilder(name string) *FilterChainBuilder {
	return &FilterChainBuilder{
		name:   name,
		vhosts: []*route.VirtualHost{},
	}
}

func (b *FilterChainBuilder) EnableTLS(serverNames []string, sdsConfig *auth.SdsSecretConfig) *FilterChainBuilder {
	b.sniServerNames = serverNames
	b.sdsCertificateConfig = sdsConfig

	return b
}

func (b *FilterChainBuilder) ForVhost(vhost *route.VirtualHost) *FilterChainBuilder {
	b.vhosts = append(b.vhosts, vhost)

	return b
}

func (b *FilterChainBuilder) Build(configureTLS bool) *listener.FilterChain {
	filterChain := listener.FilterChain{Name: b.name}
	if len(b.vhosts) > 0 {
		filterChain.Filters = append(filterChain.Filters, b.buildHTTPFilterForVhosts(configureTLS))
	}

	if configureTLS {
		filterChain.FilterChainMatch = &listener.FilterChainMatch{
			ServerNames: b.sniServerNames,
		}

		filterChain.TransportSocket = b.buildDownstreamTransportSocket()
	}

	return &filterChain
}

func (b *FilterChainBuilder) buildHTTPFilterForVhosts(configureTLS bool) *listener.Filter {
	const (
		ServerName                                 = "envoy-swarm-control/0.1"
		HTTPIdleTimeout                            = 1 * time.Hour
		RequestTimeout                             = 5 * time.Minute
		MaxConcurrentHTTP2Streams                  = 100
		InitialDownstreamHTTP2StreamWindowSize     = 65536   // 64 KiB
		InitialDownstreamHTTP2ConnectionWindowSize = 1048576 // 1 MiB
	)

	routeType := "http"
	if configureTLS {
		routeType = "https"
	}

	routes := &route.RouteConfiguration{
		Name:         fmt.Sprintf("%s_%s_routes", b.name, routeType),
		VirtualHosts: b.vhosts,
	}
	conManager := &hcm.HttpConnectionManager{
		ServerName:       ServerName,
		CodecType:        hcm.HttpConnectionManager_AUTO,
		StatPrefix:       b.name,
		UseRemoteAddress: &wrappers.BoolValue{Value: true},
		RouteSpecifier:   &hcm.HttpConnectionManager_RouteConfig{RouteConfig: routes},
		HttpFilters:      []*hcm.HttpFilter{{Name: "envoy.filters.http.router"}},
		CommonHttpProtocolOptions: &core.HttpProtocolOptions{
			IdleTimeout:                  durationpb.New(HTTPIdleTimeout),
			HeadersWithUnderscoresAction: core.HttpProtocolOptions_REJECT_REQUEST,
		},
		Http2ProtocolOptions: &core.Http2ProtocolOptions{
			MaxConcurrentStreams:        &wrappers.UInt32Value{Value: uint32(MaxConcurrentHTTP2Streams)},
			InitialStreamWindowSize:     &wrappers.UInt32Value{Value: uint32(InitialDownstreamHTTP2StreamWindowSize)},
			InitialConnectionWindowSize: &wrappers.UInt32Value{Value: uint32(InitialDownstreamHTTP2ConnectionWindowSize)},
		},
		StreamIdleTimeout: durationpb.New(RequestTimeout),
		RequestTimeout:    durationpb.New(RequestTimeout),
	}
	managerConfig, err := anypb.New(conManager)
	if err != nil {
		panic(err)
	}

	return &listener.Filter{
		Name: wellknown.HTTPConnectionManager,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: managerConfig,
		},
	}
}

func (b *FilterChainBuilder) buildDownstreamTransportSocket() *core.TransportSocket {
	c := &auth.DownstreamTlsContext{
		CommonTlsContext: &auth.CommonTlsContext{
			AlpnProtocols:                  []string{"h2", "http/1.1"},
			TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{b.sdsCertificateConfig},
			TlsParams: &auth.TlsParameters{
				TlsMinimumProtocolVersion: auth.TlsParameters_TLSv1_2,
			},
		},
	}
	tlsc, _ := anypb.New(c)

	return &core.TransportSocket{
		Name: "envoy.transport_sockets.tls",
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsc,
		},
	}
}

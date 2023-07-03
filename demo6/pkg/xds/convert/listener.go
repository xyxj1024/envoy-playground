package convert

import (
	"github.com/golang/protobuf/ptypes/wrappers"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
)

type ListenerBuilder struct {
	name         string
	configureTLS bool // true for HTTPS, false for HTTP
	filterChains []*FilterChainBuilder
}

func NewListenerBuilder(name string) *ListenerBuilder {
	return &ListenerBuilder{
		name:         name,
		configureTLS: false, // default is HTTP
		filterChains: []*FilterChainBuilder{},
	}
}

func (b *ListenerBuilder) EnableTLS() *ListenerBuilder {
	b.configureTLS = true

	return b
}

func (b *ListenerBuilder) AddFilterChain(f *FilterChainBuilder) *ListenerBuilder {
	b.filterChains = append(b.filterChains, f)

	return b
}

func (b *ListenerBuilder) Build(listenerPort uint32) *listener.Listener {
	const PerConnectionBufferLimit = 32768 // 32 KiB

	var listenerFilters []*listener.ListenerFilter
	if b.configureTLS {
		listenerFilters = []*listener.ListenerFilter{{
			Name: "envoy.filters.listener.tls_inspector",
		}}
	}

	chains := []*listener.FilterChain{}
	for i := range b.filterChains {
		chains = append(chains, b.filterChains[i].Build(b.configureTLS))
	}

	return &listener.Listener{
		Name: b.name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: listenerPort,
					},
				},
			},
		},
		ListenerFilters:               listenerFilters,
		FilterChains:                  chains,
		PerConnectionBufferLimitBytes: &wrappers.UInt32Value{Value: uint32(PerConnectionBufferLimit)},
	}
}

package provider

import (
	// Standard library
	"context"

	// Envoy go-control-plane
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type ADS interface {
	Provide(ctx context.Context) (clusters, listeners []types.Resource, err error)
}

type SDS interface {
	Provide(ctx context.Context) (secrets []types.Resource, err error)
	HasValidCertificate(vhst *route.VirtualHost) bool
	GetCertificateConfig(vhst *route.VirtualHost) *auth.SdsSecretConfig
}

package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"
	"time"

	"envoy-swarm-control/pkg/logger"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type CertificateSecretsProvider struct {
	configSource     *core.ConfigSource
	configKeyPrefix  string
	requestedConfigs map[string]*route.VirtualHost
	storage          Certificate
	logger           logger.Logger
}

func NewCertificateSecretsProvider(
	controlPlaneClusterName string,
	certificateStorage *Certificate,
	log logger.Logger,
) *CertificateSecretsProvider {
	c := &core.ConfigSource{
		ResourceApiVersion: core.ApiVersion_V3,
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:             core.ApiConfigSource_GRPC,
				TransportApiVersion: core.ApiVersion_V3,
				GrpcServices: []*core.GrpcService{{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
							ClusterName: controlPlaneClusterName,
						},
					},
				}},
			},
		},
	}

	return &CertificateSecretsProvider{
		configSource:     c,
		configKeyPrefix:  "downstream_tls_",
		requestedConfigs: make(map[string]*route.VirtualHost),
		storage:          *certificateStorage,
		logger:           log,
	}
}

func (p *CertificateSecretsProvider) Provide(_ context.Context) (secrets []types.Resource, err error) {
	for key := range p.requestedConfigs {
		vhost := p.requestedConfigs[key]

		public, private, err := p.getCertificateFromStorage(vhost)
		if err != nil {
			p.logger.Warnf("can't find promised certificate for %s", key)
			continue
		}

		secrets = append(secrets, &auth.Secret{
			Name: key,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: public},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: private},
					},
				},
			},
		})
	}

	return secrets, nil
}

func (p *CertificateSecretsProvider) HasValidCertificate(vhost *route.VirtualHost) bool {
	cert, err := p.getParsedCertificate(vhost)
	if err != nil {
		return false
	}

	return IsCertificateUsable(cert)
}

func (p *CertificateSecretsProvider) GetCertificateConfig(vhost *route.VirtualHost) *auth.SdsSecretConfig {
	key := p.getSecretConfigKey(vhost)
	p.requestedConfigs[key] = vhost

	return &auth.SdsSecretConfig{
		Name:      key,
		SdsConfig: p.configSource,
	}
}

func (p *CertificateSecretsProvider) getSecretConfigKey(vhost *route.VirtualHost) string {
	return p.configKeyPrefix + strings.ToLower(vhost.Name)
}

func (p *CertificateSecretsProvider) getParsedCertificate(vhost *route.VirtualHost) (*tls.Certificate, error) {
	certBytes, keyBytes, err := p.getCertificateFromStorage(vhost)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		p.logger.Infof("decoding certificate from storage failed: %", err.Error())
		return nil, err
	}

	return &cert, err
}

func (p *CertificateSecretsProvider) getCertificateFromStorage(vhost *route.VirtualHost) ([]byte, []byte, error) {
	// Return a list of domains (host/authority header) matched to this virtual host
	domains := vhost.GetDomains()
	if len(domains) == 0 {
		return nil, nil, errors.New("vhost contains no domains")
	}
	// The first domain is the primary one
	return p.storage.GetCertificate(vhost.GetDomains()[0], vhost.GetDomains())
}

func IsCertificateUsable(cert *tls.Certificate) bool {
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return false
	}

	return (time.Now()).Before(leaf.NotAfter)
}

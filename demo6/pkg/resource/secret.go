package resource

import (
	"github.com/sirupsen/logrus"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

const (
	secretName = "server_cert"
	certPath   = "/Users/xuanyuanxingjian/Documents/projects/Repos/github-envoy-playground/demo6/deploy/certs"
)

var (
	envoyServerCert = certPath + "/envoy-server.crt"
	envoyServerKey  = certPath + "/envoy-server.key"
)

func ProvideSecret() *auth.Secret {
	logrus.Infof(">>>>>>>>>>>>>>>>>>> creating secret with secretName " + secretName)
	return &auth.Secret{
		Name: secretName,
		Type: &auth.Secret_TlsCertificate{
			TlsCertificate: &auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(envoyServerCert)},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(envoyServerKey)},
				},
			},
		},
	}
}

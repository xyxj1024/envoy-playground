package xds

import (
	"envoy-swarm-control/pkg/xds/convert"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type ListenerProvider struct {
	sdsProvider SDS
}

func NewListenerProvider(sdsProvider SDS) *ListenerProvider {
	return &ListenerProvider{
		sdsProvider: sdsProvider,
	}
}

/* Function ProvideListener:
 * returns a HTTPS listener for port 443 if TLS is configured;
 * returns a HTTP listener for port 80, otherwise.
 */
func (l *ListenerProvider) ProvideListeners(v *convert.VhostCollection, listenerPort uint32) ([]types.Resource, error) {
	httpListener, httpsListener := l.createListenersFromVhosts(v, listenerPort)
	if err := httpListener.Validate(); err != nil {
		return nil, err
	}

	if len(httpsListener.FilterChains) == 0 || httpsListener.Validate() != nil {
		return []types.Resource{httpListener}, nil
	}

	return []types.Resource{httpListener}, nil
}

func (l *ListenerProvider) createListenersFromVhosts(vhosts *convert.VhostCollection, listenerPort uint32) (http, https *listener.Listener) {
	httpFilter := convert.NewFilterChainBuilder("httpFilter")
	httpListenerBuilder := convert.NewListenerBuilder("httpListener")
	httpsListenerBuilder := convert.NewListenerBuilder("httpsListener").EnableTLS()

	for i := range vhosts.Vhosts {
		v := vhosts.Vhosts[i]
		hasValidCertificate := false
		if l.sdsProvider != nil {
			hasValidCertificate = l.sdsProvider.HasValidCertificate(v)
		}

		if hasValidCertificate {
			httpsFilter := l.createFilterChainWithTLS(v)
			httpsFilter.ForVhost(v)
			httpsListenerBuilder.AddFilterChain(httpsFilter)
			httpFilter.ForVhost(createNewHTTPSRedirectVhost(v))
		} else {
			httpFilter.ForVhost(v)
		}
	}

	httpListenerBuilder.AddFilterChain(httpFilter)
	return httpListenerBuilder.Build(listenerPort), httpsListenerBuilder.Build(listenerPort)
}

func (l *ListenerProvider) createFilterChainWithTLS(vhost *route.VirtualHost) *convert.FilterChainBuilder {
	return convert.NewFilterChainBuilder(vhost.Name).EnableTLS(vhost.Domains, l.sdsProvider.GetCertificateConfig(vhost))
}

func createNewHTTPSRedirectVhost(originalVhost *route.VirtualHost) *route.VirtualHost {
	return &route.VirtualHost{
		Name:    originalVhost.Name,
		Domains: originalVhost.Domains,
		Routes: []*route.Route{{
			Name: "https_redirect",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/",
				},
			},
			Action: &route.Route_Redirect{
				Redirect: &route.RedirectAction{
					SchemeRewriteSpecifier: &route.RedirectAction_HttpsRedirect{
						HttpsRedirect: true,
					},
				},
			},
		}},
	}
}

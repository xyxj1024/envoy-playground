package resource

import route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

func ProvideRoute(routeConfigName, virtualHostName, clusterName, upstreamHost string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeConfigName, // e.g., "local_route"
		VirtualHosts: []*route.VirtualHost{{
			Name:    virtualHostName, // e.g., "local_service"
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
						// PrefixRewrite: "/robots.txt", /* removing this line causes no harm */
						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: upstreamHost,
						},
					},
				},
			}},
		}},
	}
}

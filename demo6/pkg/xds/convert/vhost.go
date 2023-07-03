package convert

import (
	"fmt"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	"google.golang.org/protobuf/types/known/durationpb"
)

type VhostCollection struct {
	Vhosts      map[string]*route.VirtualHost
	UsedDomains map[string]*route.VirtualHost
}

func NewVhostCollection() *VhostCollection {
	return &VhostCollection{
		Vhosts:      make(map[string]*route.VirtualHost),
		UsedDomains: make(map[string]*route.VirtualHost),
	}
}

func (v VhostCollection) AddService(clusterIdentifier string, labels *ServiceLabel) (err error) {
	primaryDomain := labels.Route.Domain

	virtualHost, exist := v.Vhosts[primaryDomain]
	if !exist {
		if _, exist := v.UsedDomains[primaryDomain]; exist {
			return fmt.Errorf("domain %s is already used in another vhost", primaryDomain)
		}

		virtualHost = &route.VirtualHost{
			Name:    primaryDomain,
			Domains: []string{primaryDomain},
			Routes:  []*route.Route{},
		}
	}

	var extraDomains []string
	for _, extraDomain := range labels.Route.ExtraDomains {
		if extraDomain == primaryDomain {
			continue
		}

		if v, exist := v.UsedDomains[extraDomain]; exist {
			if v != virtualHost {
				return fmt.Errorf("domain %s is already used in another vhost", extraDomain)
			}
			continue
		}
		extraDomains = append(extraDomains, extraDomain)
	}

	v.Vhosts[primaryDomain] = virtualHost
	v.UsedDomains[primaryDomain] = virtualHost

	newRoute := v.createRoute(clusterIdentifier, labels)
	if labels.Route.PathPrefix == "/" {
		virtualHost.Routes = append(virtualHost.Routes, newRoute)
	} else {
		virtualHost.Routes = append([]*route.Route{newRoute}, virtualHost.Routes...)
	}

	for i := range extraDomains {
		virtualHost.Domains = append(virtualHost.Domains, extraDomains[i])
		v.UsedDomains[extraDomains[i]] = virtualHost
	}

	return nil
}

func (v VhostCollection) createRoute(clusterIdentifier string, labels *ServiceLabel) *route.Route {
	return &route.Route{
		Name: clusterIdentifier + "_route",
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: labels.Route.PathPrefix,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterIdentifier,
				},
				PrefixRewrite: "/robots.txt",
				HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
					HostRewriteLiteral: labels.Route.UpstreamHost,
				},
				// https://github.com/envoyproxy/envoy/issues/8517#issuecomment-540225144
				IdleTimeout: durationpb.New(labels.Endpoint.RequestTimeout),
				Timeout:     durationpb.New(labels.Endpoint.RequestTimeout),
			},
		},
	}
}

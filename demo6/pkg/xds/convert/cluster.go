package convert

import (
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/durationpb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
)

func SwarmServiceToCDS(service *swarm.Service, labels *ServiceLabel) (c *cluster.Cluster, err error) {
	e := SwarmServiceToEndpoint(service, labels)
	if err = e.Validate(); err != nil {
		return
	}

	c = SwarmServiceToCluster(service, e)
	if err = c.Validate(); err != nil {
		return
	}

	return
}

func SwarmServiceToEndpoint(service *swarm.Service, labels *ServiceLabel) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: service.Spec.Annotations.Name, // Swarm service name
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol:      labels.Endpoint.Protocol,  // TCP or UDP
									Address:       labels.Route.UpstreamHost, // IP address
									PortSpecifier: &labels.Endpoint.Port,     // 80, 443
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func SwarmServiceToCluster(service *swarm.Service, loadAssignment *endpoint.ClusterLoadAssignment) *cluster.Cluster {
	const (
		/* When updating services, Docker swarm's default delay is 5 seconds,
		 * setting this to 4 leaves us with a 1 second drain time (worst case)
		 */
		DNSRefreshRate               = 4 * time.Second
		PerConnectionBufferLimit     = 32768 // 32 KiB
		UpstreamConnectTimeout       = 2 * time.Second
		UpstreamTCPKeepaliveProbes   = 3
		UpstreamTCPKeepaliveTime     = 3600
		UpstreamTCPKeepaliveInterval = 60
	)

	return &cluster.Cluster{
		Name:                          service.Spec.Annotations.Name, // Swarm service name
		ConnectTimeout:                durationpb.New(UpstreamConnectTimeout),
		ClusterDiscoveryType:          &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		RespectDnsTtl:                 false, // Default TTL is 600, which is too long in the case of scaling down
		DnsRefreshRate:                durationpb.New(DNSRefreshRate),
		LoadAssignment:                loadAssignment,
		PerConnectionBufferLimitBytes: &wrappers.UInt32Value{Value: uint32(PerConnectionBufferLimit)},
		UpstreamConnectionOptions: &cluster.UpstreamConnectionOptions{
			TcpKeepalive: &core.TcpKeepalive{
				KeepaliveProbes:   &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveProbes)},
				KeepaliveTime:     &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveTime)},
				KeepaliveInterval: &wrappers.UInt32Value{Value: uint32(UpstreamTCPKeepaliveInterval)},
			},
		},
	}
}

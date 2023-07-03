package xds

import (
	"context"
	"errors"

	"envoy-swarm-control/pkg/logger"
	"envoy-swarm-control/pkg/xds/convert"

	swarmtypes "github.com/docker/docker/api/types"
	swarm "github.com/docker/docker/api/types/swarm"
	docker "github.com/docker/docker/client"

	types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

type ADSProvider struct {
	ingressNetwork  string
	dockerClient    docker.APIClient
	listenerBuilder *ListenerProvider
	logger          logger.Logger
}

func NewADSProvider(ingressNetwork string, builder *ListenerProvider, log logger.Logger) *ADSProvider {
	return &ADSProvider{
		ingressNetwork:  ingressNetwork,
		dockerClient:    NewDockerClient(),
		listenerBuilder: builder,
		logger:          log,
	}
}

func (s *ADSProvider) Provide(ctx context.Context, listenerPort uint32) (clusters, listeners []types.Resource, err error) {
	clusters, vhosts, err := s.provideClustersAndVhosts(ctx)
	if err != nil {
		s.logger.Errorf("Failed creating clusters and virtual host configurations")
		return nil, nil, err
	}

	listeners, err = s.listenerBuilder.ProvideListeners(vhosts, listenerPort)
	if err != nil {
		s.logger.Errorf("Failed converting virtual hosts into a listener configuration")
		return nil, nil, err
	}

	return clusters, listeners, nil
}

func (s *ADSProvider) provideClustersAndVhosts(ctx context.Context) (clusters []types.Resource, vhosts *convert.VhostCollection, err error) {
	ingress, err := s.getIngressNetwork(ctx)
	if err != nil {
		return
	}

	services, err := s.dockerClient.ServiceList(ctx, swarmtypes.ServiceListOptions{})
	if err != nil {
		return
	}

	vhosts = convert.NewVhostCollection()
	for i := range services {
		service := &services[i]
		log := s.logger.WithFields(logger.Fields{"swarm-service-name": service.Spec.Annotations.Name})

		labels := convert.ParseServiceLabels(service.Spec.Annotations.Labels)
		if err = labels.Validate(); err != nil {
			log.Debugf("Skipping service because labels are invalid: %s", err.Error())
			continue
		}

		if !isInIngressNetwork(service, &ingress) {
			log.Warnf("Service is not connected to the ingress network, stopping processing")
			continue
		}

		cluster, err := convert.SwarmServiceToCDS(service, labels)
		if err != nil {
			log.Warnf("Skipped generating CDS for service because %s", err.Error())
			continue
		}

		err = vhosts.AddService(cluster.Name, labels)
		if err != nil {
			log.Warnf("Skipped creating virtual host for service because %s", err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	return clusters, vhosts, nil
}

func (s *ADSProvider) getIngressNetwork(ctx context.Context) (network swarmtypes.NetworkResource, err error) {
	network, err = s.dockerClient.NetworkInspect(ctx, s.ingressNetwork, swarmtypes.NetworkInspectOptions{})
	if err != nil {
		return
	}

	if network.Scope != "swarm" {
		return network, errors.New("the provided ingress network is not scoped for the entire cluster (swarm)")
	}

	return
}

func isInIngressNetwork(service *swarm.Service, ingress *swarmtypes.NetworkResource) bool {
	for _, vip := range service.Endpoint.VirtualIPs {
		if vip.NetworkID == ingress.ID {
			return true
		}
	}

	return false
}

func NewDockerClient() *docker.Client {
	httpHeaders := map[string]string{
		"User-Agent": "envoy-swarm-control",
	}

	c, err := docker.NewClientWithOpts(
		docker.FromEnv,
		docker.WithHTTPHeaders(httpHeaders),
		docker.WithAPIVersionNegotiation(), // For "Maximum supported API version is 1.41"
	)
	if err != nil {
		panic(err)
	}

	return c
}

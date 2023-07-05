package watcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"envoy-swarm-control/pkg/snapshot"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	swarm "github.com/docker/docker/api/types/swarm"
	docker "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

/* Function InitUpdateChannel:
 * sets the update channel to an empty structure.
 */
func InitUpdateChannel(updateChannel chan snapshot.ServiceLabels) {
	updateChannel <- snapshot.ServiceLabels{}
}

func StartWatcher(ctx context.Context, cli docker.APIClient, ingressNetwork string, updateChannel chan snapshot.ServiceLabels) {
	events, errorEvent := cli.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "type", Value: "service"}),
	})
	/* Docker services report the following events:
	 * - create
	 * - remove
	 * - update
	 */
	for {
		select {
		case err := <-errorEvent:
			logrus.Errorf(err.Error())
			StartWatcher(ctx, cli, ingressNetwork, updateChannel)

		case event := <-events:
			logrus.WithFields(logrus.Fields{"type": event.Type, "action": event.Action}).Debugf("Docker swarm service event received")

			if event.Action == "create" {
				continue
			}

			ingress, err := getIngressNetwork(ctx, cli, ingressNetwork)
			if err != nil {
				return
			}

			fmt.Println("Please enter the service name to be updated: ")
			userInput := bufio.NewScanner(os.Stdin)
			userInput.Scan()
			serviceName := string(userInput.Text())

			args := filters.NewArgs()
			args.Add("name", serviceName)
			services, err := cli.ServiceList(context.Background(), types.ServiceListOptions{Filters: args})
			if err != nil {
				return
			}
			if len(services) != 1 {
				logrus.Println("Something went wrong")
				logrus.Println("Count:", len(services))
				logrus.Printf("Service %s did not found", serviceName)
				return
			}

			for _, service := range services {
				if !isInIngressNetwork(&service, &ingress) {
					logrus.Warnf("Service is not connected to the ingress network, stopping processing")
					return
				}

				labels := snapshot.ParseServiceLabels(service.Spec.Annotations.Labels)
				if err = labels.Validate(); err != nil {
					logrus.Debugf("Skipping service because labels are invalid: %s", err.Error())
					return
				}

				updateChannel <- *labels
			}
		}
	}
}

func getIngressNetwork(ctx context.Context, cli docker.APIClient, ingressNetwork string) (network types.NetworkResource, err error) {
	network, err = cli.NetworkInspect(ctx, ingressNetwork, types.NetworkInspectOptions{})
	if err != nil {
		return
	}

	if network.Scope != "swarm" {
		return network, errors.New("the provided ingress network is not scoped for the entire cluster (swarm)")
	}

	return
}

func isInIngressNetwork(service *swarm.Service, ingress *types.NetworkResource) bool {
	for _, vip := range service.Endpoint.VirtualIPs {
		if vip.NetworkID == ingress.ID {
			return true
		}
	}

	return false
}

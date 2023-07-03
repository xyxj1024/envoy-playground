package watcher

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	"envoy-swarm-control/pkg/logger"
	"envoy-swarm-control/pkg/snapshot"
	"envoy-swarm-control/pkg/xds"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
)

type SwarmEvent struct {
	client docker.APIClient
	logger logger.Logger
}

func NewSwarmEvent(log logger.Logger) *SwarmEvent {
	return &SwarmEvent{
		client: xds.NewDockerClient(),
		logger: log,
	}
}

/* Function InitUpdateChannel:
 * initializes the two fields of an update channel.
 */
func InitUpdateChannel(updateChannel chan snapshot.UpdateReason) {
	updateChannel <- snapshot.UpdateReason{
		EnvoyNodeId:       "test-id",
		EnvoyListenerPort: 10000,
	}
}

/* Function StartWatcher:
 * reads events reported by Docker and processes them;
 * accepts user input (node ID and listener port) as update channel.
 * Please follow the guidance from comments in the file:
 * https://github.com/docker/engine/blob/master/client/events.go
 */
func (s SwarmEvent) StartWatcher(ctx context.Context, updateChannel chan snapshot.UpdateReason) {
	events, errorEvent := s.client.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "type", Value: "service"}),
	})
	/* Docker services report the following events:
	 * - create
	 * - remove
	 * - update
	 */
	for {
		select {
		case event := <-events:
			s.logger.WithFields(logger.Fields{"type": event.Type, "action": event.Action}).Debugf("Docker swarm service event received")
			if event.Action == "create" {
				continue
			}

			fmt.Println("Please enter Envoy node ID: ")
			nodeId := bufio.NewScanner(os.Stdin)
			nodeId.Scan()

			fmt.Println("Please enter listener port: ")
			listenerPortStr := bufio.NewScanner(os.Stdin)
			listenerPortStr.Scan()
			listenerPort, _ := strconv.ParseUint(listenerPortStr.Text(), 10, 64)

			updateChannel <- snapshot.UpdateReason{
				EnvoyNodeId:       string(nodeId.Text()),
				EnvoyListenerPort: uint32(listenerPort),
			}
		case err := <-errorEvent:
			s.logger.Errorf(err.Error())
			s.StartWatcher(ctx, updateChannel)
		}
	}
}

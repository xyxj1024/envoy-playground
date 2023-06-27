package watcher

import (
	"context"

	"envoy-swarm-control/pkg/logger"
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

func InitUpdateChannel(updateChannel chan string) {
	updateChannel <- "initial Docker events watcher"
}

/* Function StartWatcher:
 * reads events reported by Docker and processes them.
 * Please follow the guidance from comments in the file:
 * https://github.com/docker/engine/blob/master/client/events.go
 */
func (s SwarmEvent) StartWatcher(ctx context.Context, updateChannel chan string) {
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
			s.logger.WithFields(logger.Fields{"type": event.Type, "action": event.Action}).Debugf("swarm service event from Docker received")
			if event.Action == "create" {
				continue
			}
			updateChannel <- "a swarm service just changed"
		case err := <-errorEvent:
			s.logger.Errorf(err.Error())
			s.StartWatcher(ctx, updateChannel)
		}
	}
}

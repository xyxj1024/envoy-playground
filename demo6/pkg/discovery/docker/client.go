package docker

import (
	// Standard library
	"context"

	// Third-party libraries
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

func inspectContainers(ctx context.Context, dockerClient client.ContainerAPIClient, containerID string) dockerData {
	containerInspected, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		logrus.Errorf("Failed to inspect container %s: %+v", containerID, err)
		return dockerData{}
	}

	return dockerData{}
}

func parseContainer(container dockertypes.ContainerJSON) dockerData {

}

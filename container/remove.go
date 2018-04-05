package container

import (
	docker "github.com/fsouza/go-dockerclient"
)

func RemoveContainer(client *docker.Client, container *docker.Container) error {
	return client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
}

package command

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) removeContainer(container *docker.Container) error {
	return command.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
}

package command

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) createContainer(image string) (*docker.Container, error) {
	return command.Client.CreateContainer(docker.CreateContainerOptions{
		Name:   command.Config.ContainerName,
		Config: &docker.Config{
			User:         command.Config.User,
			Env:          command.Config.Environment, // TODO: support env_file
			Cmd:          []string{}, // TODO: command
			Image:        image,
			WorkingDir:   command.Config.WorkingDir,
			Entrypoint:   []string{"/bin/sh"}, // TODO: entrypoint
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    true,
		},
		HostConfig: &docker.HostConfig{
			Binds:        []string{}, // TODO: bind mounts
			VolumesFrom:  command.Config.VolumesFrom,
		},
	})
}

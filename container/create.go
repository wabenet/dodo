package container

import (
	"github.com/oclaussen/dodo/config"
	docker "github.com/fsouza/go-dockerclient"
)

func CreateContainer(client *docker.Client, image string, config *config.CommandConfig) (*docker.Container, error) {
	return client.CreateContainer(docker.CreateContainerOptions{
		Name:   config.ContainerName,
		Config: &docker.Config{
			User:         config.User,
			Env:          config.Environment, // TODO: support env_file
			Cmd:          []string{}, // TODO: command
			Image:        image,
			WorkingDir:   config.WorkingDir,
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
			VolumesFrom:  config.VolumesFrom,
		},
	})
}

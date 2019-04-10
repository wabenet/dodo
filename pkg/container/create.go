package container

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/net/context"
)

func createContainer(ctx context.Context, options Options, tty bool) (string, error) {
	response, err := options.Client.ContainerCreate(
		ctx,
		&container.Config{
			User:         options.User,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          tty,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          options.Environment,
			Cmd:          options.Command,
			Image:        options.Image,
			WorkingDir:   options.WorkingDir,
			Entrypoint:   options.Entrypoint,
			ExposedPorts: options.PortBindings.PortSet(),
		},
		&container.HostConfig{
			AutoRemove:   options.Remove,
			Binds:        options.Volumes,
			VolumesFrom:  options.VolumesFrom,
			PortBindings: options.PortBindings.PortMap(),
		},
		&network.NetworkingConfig{},
		options.Name,
	)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

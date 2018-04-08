package container

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/net/context"
)

func createContainer(ctx context.Context, options Options) (string, error) {
	response, err := options.Client.ContainerCreate(
		ctx,
		&container.Config{
			User:         options.User,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          options.Environment,
			Cmd:          options.Command,
			Image:        options.Image,
			WorkingDir:   options.WorkingDir,
			Entrypoint:   getEntrypoint(options),
		},
		&container.HostConfig{
			AutoRemove:  options.Remove,
			Binds:       options.Volumes,
			VolumesFrom: options.VolumesFrom,
		},
		&network.NetworkingConfig{},
		options.Name,
	)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

func getEntrypoint(options Options) []string {
	entrypoint := []string{"/bin/sh"}
	if len(options.Interpreter) > 0 {
		entrypoint = options.Interpreter
	}
	if !options.Interactive {
		entrypoint = append(entrypoint, options.Entrypoint)
	}
	return entrypoint
}

package state

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/net/context"
)

// EnsureContainer will soon be moved anyway
func (state *State) EnsureContainer(ctx context.Context) (string, error) {
	config := state.Config
	client := state.Client
	image := state.Image
	if state.ContainerID != "" {
		return state.ContainerID, nil
	}

	response, err := client.ContainerCreate(
		ctx,
		&container.Config{
			User:         config.User,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          config.Environment,
			Cmd:          config.Command,
			Image:        image,
			WorkingDir:   config.WorkingDir,
			Entrypoint:   state.getEntrypoint(),
		},
		&container.HostConfig{
			AutoRemove:  true,
			Binds:       config.Volumes,
			VolumesFrom: config.VolumesFrom,
		},
		&network.NetworkingConfig{},
		config.ContainerName,
	)
	if err != nil {
		return "", err
	}
	state.ContainerID = response.ID
	return state.ContainerID, nil
}

func (state *State) getEntrypoint() []string {
	entrypoint := []string{"/bin/sh"}
	if len(state.Config.Interpreter) > 0 {
		entrypoint = state.Config.Interpreter
	}
	if !state.Config.Interactive {
		entrypoint = append(entrypoint, state.Entrypoint)
	}
	return entrypoint
}

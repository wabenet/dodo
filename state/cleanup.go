package state

import (
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func (state *State) EnsureCleanup(ctx context.Context) {
	if state.ContainerID == "" {
		return
	}
	client, err := state.EnsureClient(ctx)
	if err != nil {
		return
	}
	config, err := state.EnsureConfig(ctx)
	if err != nil {
		return
	}
	if config.Remove != nil && !*config.Remove {
		return
	}

	client.ContainerRemove(
		ctx,
		state.ContainerID,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   true,
			Force:         true,
		},
	)
	state.ContainerID = ""
}

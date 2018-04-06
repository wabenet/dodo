package state

import (
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// EnsureCleanup makes sure the docker container is removed after run.
func (state *State) EnsureCleanup(ctx context.Context) {
	if state.ContainerID == "" {
		return
	}
	client, err := state.EnsureClient()
	if err != nil {
		return
	}
	if state.Config.Remove != nil && !*state.Config.Remove {
		return
	}

	err = client.ContainerRemove(
		ctx,
		state.ContainerID,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   true,
			Force:         true,
		},
	)
	if err != nil {
		log.Error(err)
	}
	state.ContainerID = ""
}

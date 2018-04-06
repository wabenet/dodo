package state

import (
	"github.com/docker/docker/client"
)

// TODO: read docker configuration

// EnsureClient makes sure a client is present on the state.
func (state *State) EnsureClient() (*client.Client, error) {
	if state.Client != nil {
		return state.Client, nil
	}

	newClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	state.Client = newClient
	return state.Client, nil
}

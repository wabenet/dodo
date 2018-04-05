package state

import (
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func (state *State) EnsureClient(ctx context.Context) (*client.Client, error) {
	if state.Client != nil {
		return state.Client, nil
	}
	client, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	state.Client = client
	return client, nil
}

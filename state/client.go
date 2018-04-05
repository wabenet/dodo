package state

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (state *state) ensureClient() error {
	if state.Client != nil {
		return nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	state.Client = client
	return nil
}

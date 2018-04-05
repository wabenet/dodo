package state

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/config"
	"golang.org/x/net/context"
)

type State struct {
	Config      *config.BackdropConfig
	Client      *client.Client
	Entrypoint  string
	Image       string
	ContainerID string
}

func NewState(config *config.BackdropConfig) *State {
	// TODO: generate a temp file in the container for the entrypoint
	return &State{
		Config:     config,
		Entrypoint: "/tmp/dodo-entrypoint",
	}
}

func (state *State) Run() error {
	ctx := context.Background()
	return state.EnsureRun(ctx)
}

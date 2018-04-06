package state

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/config"
	"golang.org/x/net/context"
)

// State represents the state of a command run.
type State struct {
	Config      *config.BackdropConfig
	Client      *client.Client
	Entrypoint  string
	Image       string
	ContainerID string
}

// NewState creates a new state base on a backdrop configuration.
func NewState(config *config.BackdropConfig) *State {
	// TODO: generate a temp file in the container for the entrypoint
	return &State{
		Config:     config,
		Entrypoint: "/tmp/dodo-entrypoint",
	}
}

// Run runs the command.
func (state *State) Run() error {
	ctx := context.Background()
	return state.EnsureRun(ctx)
}

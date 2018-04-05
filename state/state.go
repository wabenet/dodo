package state

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/config"
	"github.com/oclaussen/dodo/options"
	"golang.org/x/net/context"
)

type State struct {
	Name        string
	Options     *options.Options
	Config      *config.BackdropConfig
	Client      *client.Client
	Entrypoint  string
	Image       string
	ContainerID string
}

func NewState(name string, options *options.Options) *State {
	// TODO: generate a temp file in the container for the entrypoint
	return &State{
		Name:       name,
		Options:    options,
		Entrypoint: "/tmp/dodo-entrypoint",
	}
}

func (state *State) Run() error {
	ctx := context.Background()
	return state.EnsureRun(ctx)
}

package container

import (
	"errors"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Options represents the configuration for running a docker container to
// be used as backdrop.
type Options struct {
	Client      *client.Client
	Image       string
	Name        string
	Remove      bool
	Interactive bool
	Interpreter []string
	Entrypoint  string
	Script      string
	Command     []string
	Environment []string
	Volumes     []string
	VolumesFrom []string
	User        string
	WorkingDir  string
}

// Run runs a docker container as backdrop.
func Run(ctx context.Context, options Options) error {
	if options.Client == nil {
		return errors.New("client may not be nil")
	}

	containerID, err := createContainer(ctx, options)
	if err != nil {
		return err
	}

	err = uploadEntrypoint(ctx, containerID, options)
	if err != nil {
		return err
	}

	return runContainer(ctx, containerID, options)
}

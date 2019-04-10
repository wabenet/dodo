package container

import (
	"errors"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/types"
	"golang.org/x/net/context"
)

// Options represents the configuration for running a docker container to
// be used as backdrop.
type Options struct {
	Client       *client.Client
	Image        string
	Name         string
	Remove       bool
	Entrypoint   []string
	Script       string
	ScriptPath   string
	Command      []string
	Environment  []string
	Volumes      []string
	VolumesFrom  []string
	PortBindings types.Ports
	User         string
	WorkingDir   string
}

type ScriptError struct {
	Message  string
	ExitCode int
}

func (e *ScriptError) Error() string {
	return e.Message
}

// Run runs a docker container as backdrop.
func Run(ctx context.Context, options Options) error {
	if options.Client == nil {
		return errors.New("client may not be nil")
	}

	_, inTerm := term.GetFdInfo(os.Stdin)
	_, outTerm := term.GetFdInfo(os.Stdout)
	tty := inTerm && outTerm

	containerID, err := createContainer(ctx, options, tty)
	if err != nil {
		return err
	}

	err = uploadEntrypoint(ctx, containerID, options)
	if err != nil {
		return err
	}

	return runContainer(ctx, containerID, options, tty)
}

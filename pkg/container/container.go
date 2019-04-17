package container

import (
	"errors"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/types"
	"golang.org/x/net/context"
)

type Container struct {
	config     *types.Backdrop
	client     *client.Client
	context    context.Context
	scriptPath string
}

func NewContainer(client *client.Client, config *types.Backdrop) (*Container, error) {
	if client == nil {
		return nil, errors.New("client may not be nil")
	}
	return &Container{
		config:     config,
		client:     client,
		context:    context.Background(),
		scriptPath: "/tmp/dodo-dockerfile-" + stringid.GenerateRandomID()[:20],
	}, nil
}

// Run runs a docker container as backdrop.
func (c *Container) Run(image string) error {
	_, inTerm := term.GetFdInfo(os.Stdin)
	_, outTerm := term.GetFdInfo(os.Stdout)
	tty := inTerm && outTerm

	containerID, err := c.create(image, tty)
	if err != nil {
		return err
	}

	if len(c.config.Script) > 0 {
		if err = c.uploadEntrypoint(containerID); err != nil {
			return err
		}
	}

	return c.run(containerID, tty)
}

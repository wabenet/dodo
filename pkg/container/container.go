package container

import (
	"fmt"
	"os"

	dockerapi "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/image"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"golang.org/x/net/context"
)

type Container struct {
	name        string
	daemon      bool
	config      *types.Backdrop
	stage       stage.Stage
	client      *client.Client
	context     context.Context
	tmpPath     string
	authConfigs map[string]dockerapi.AuthConfig
}

func NewContainer(config *types.Backdrop, s stage.Stage, authConfigs map[string]dockerapi.AuthConfig, daemon bool) (*Container, error) {
	dockerClient, err := stage.GetDockerClient(s)
	if err != nil {
		return nil, err
	}

	name := config.ContainerName
	if daemon {
		name = config.Name
	} else if len(name) == 0 {
		name = fmt.Sprintf("%s-%s", config.Name, stringid.GenerateRandomID()[:8])
	}

	return &Container{
		name:        name,
		daemon:      daemon,
		config:      config,
		stage:       s,
		client:      dockerClient,
		context:     context.Background(),
		tmpPath:     fmt.Sprintf("/tmp/dodo-%s/", stringid.GenerateRandomID()[:20]),
		authConfigs: authConfigs,
	}, nil
}

func (c *Container) Build() error {
	c.config.Image.ForceRebuild = true

	img, err := image.NewImage(c.client, c.authConfigs, c.config.Image)
	if err != nil {
		return err
	}

	if _, err := img.Get(); err != nil {
		return err
	}

	return nil
}

func (c *Container) Run() error {
	img, err := image.NewImage(c.client, c.authConfigs, c.config.Image)
	if err != nil {
		return err
	}

	imageID, err := img.Get()
	if err != nil {
		return err
	}

	containerID, err := c.create(imageID)
	if err != nil {
		return err
	}

	if c.daemon {
		return c.client.ContainerStart(
			c.context,
			containerID,
			dockerapi.ContainerStartOptions{},
		)
	} else {
		return c.run(containerID, hasTTY())
	}
}

func (c *Container) Stop() error {
	if err := c.client.ContainerStop(c.context, c.name, nil); err != nil {
		return err
	}

	if err := c.client.ContainerRemove(c.context, c.name, dockerapi.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func hasTTY() bool {
	_, inTerm := term.GetFdInfo(os.Stdin)
	_, outTerm := term.GetFdInfo(os.Stdout)
	return inTerm && outTerm
}

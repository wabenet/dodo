package container

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"golang.org/x/net/context"
)

type Container struct {
	config  *types.Backdrop
	stage   stage.Stage
	client  *client.Client
	context context.Context
	tmpPath string
}

func NewContainer(config *types.Backdrop, s stage.Stage) (*Container, error) {
	dockerClient, err := stage.GetDockerClient(s)
	if err != nil {
		return nil, err
	}
	return &Container{
		config:  config,
		stage:   s,
		client:  dockerClient,
		context: context.Background(),
		tmpPath: fmt.Sprintf("/tmp/dodo-%s/", stringid.GenerateRandomID()[:20]),
	}, nil
}

// Run runs a docker container as backdrop.
func (c *Container) Run(image string) error {
	_, inTerm := term.GetFdInfo(os.Stdin)
	_, outTerm := term.GetFdInfo(os.Stdout)
	tty := inTerm && outTerm

	opts, err := c.stage.GetDockerOptions()
	if err != nil {
		return err
	}

	// TODO: it's weird to do this in two steps
	if c.config.ForwardStage {
		yes := "1"
		env := c.config.Environment
		env = append(env, types.KeyValue{"DOCKER_HOST", &opts.Host})
		env = append(env, types.KeyValue{"DOCKER_API_VERSION", &opts.Version})
		env = append(env, types.KeyValue{"DOCKER_CERT_PATH", &c.tmpPath})
		env = append(env, types.KeyValue{"DOCKER_TLS_VERIFY", &yes})
		c.config.Environment = env
	}

	containerID, err := c.create(image, tty)
	if err != nil {
		return err
	}

	if len(c.config.Script) > 0 {
		if err = c.uploadEntrypoint(containerID); err != nil {
			return err
		}
	}

	if c.config.ForwardStage {
		if err = c.uploadStageConfig(containerID, opts); err != nil {
			return err
		}
	}

	return c.run(containerID, tty)
}

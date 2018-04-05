package command

import (
	"errors"

	"github.com/oclaussen/dodo/config"
	docker "github.com/fsouza/go-dockerclient"
)

type Command struct {
	Config *config.CommandConfig
	Client *docker.Client
}

func NewCommand(config config.CommandConfig) *Command {
	return &Command{
		Config: &config,
	}
}

func (command *Command) Run() error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	command.Client = client

	// TODO: check if pulling/building is necessary, implement force pull
	if command.Config.Build != nil {
		err = command.buildImage()
		if err != nil {
			return err
		}
	} else if command.Config.Image != "" {
		err = command.pullImage()
		if err != nil {
			return err
		}
	} else {
		// TODO: add validation that exactly one is set
		// TODO: nicer errors
		errors.New("No build and no image!")
	}

	container, err := command.createContainer("dodo-testing") // TODO: get actual image id
	if err != nil {
		return err
	}

	defer command.removeContainer(container)

	err = command.startContainer(container)
	if err != nil {
		return err
	}

	return nil
}

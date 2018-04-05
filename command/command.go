package command

import (
	"errors"

	"github.com/oclaussen/dodo/config"
	"github.com/oclaussen/dodo/command/image"
	"github.com/oclaussen/dodo/command/container"
	docker "github.com/fsouza/go-dockerclient"
)

type Command struct {
	Config    *config.CommandConfig
	Client    *docker.Client
	Image	  string
	Container *docker.Container
}

func NewCommand(config config.CommandConfig) *Command {
	return &Command{
		Config: &config,
	}
}

func (command *Command) ensureConfig() error {
	return nil
}

func (command *Command) ensureClient() error {
	if command.Client != nil {
		return nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	command.Client = client
	return nil
}

func (command *Command) ensureImage() error {
	if command.Image != "" {
		return nil
	}
	if err := command.ensureConfig(); err != nil {
		return err;
	}
	if err := command.ensureClient(); err != nil {
		return err;
	}
	// TODO: check if pulling/building is necessary, implement force pull

	if command.Config.Build != nil && command.Config.Image != "" {
		return errors.New("You can specifiy either image or build, not both.")

	} else if command.Config.Build != nil {
		image, err := image.BuildImage(command.Client, command.Config)
		if err != nil {
			return err
		}
		command.Image = image
		return nil

	} else if command.Config.Image != "" {
		image, err := image.PullImage(command.Client, command.Config)
		if err != nil {
			return err
		}
		command.Image = image
		return nil

	} else {
		return errors.New("You need to specify either image or build.")
	}
}

func (command *Command) ensureContainer() error {
	if command.Container != nil {
		return nil
	}
	if err := command.ensureConfig(); err != nil {
		return err
	}
	if err := command.ensureClient(); err != nil {
		return err
	}
	if err := command.ensureImage(); err != nil {
		return err
	}

	container, err := container.CreateContainer(command.Client, command.Image, command.Config)
	if err != nil {
		return err
	}
	command.Container = container
	return nil
}

func (command *Command) Run() error {
	if err := command.ensureContainer(); err != nil {
		return err
	}

	defer container.RemoveContainer(command.Client, command.Container)

	if err := container.RunContainer(command.Client, command.Container); err != nil {
		return err
	}

	return nil
}

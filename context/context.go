package context

import (
	"github.com/oclaussen/dodo/config"
	"github.com/oclaussen/dodo/container"
	"github.com/oclaussen/dodo/options"
	docker "github.com/fsouza/go-dockerclient"
)

type Context struct {
	Name      string
	Options   *options.Options
	Config    *config.ContextConfig
	Client    *docker.Client
	Image     string
	Container *docker.Container
}

func NewContext(name string, options *options.Options) *Context {
	return &Context{
		Name: name,
		Options: options,
	}
}

func (context *Context) Run(arguments []string) error {
	if err := context.ensureContainer(arguments); err != nil {
		return err
	}

	defer container.RemoveContainer(context.Client, context.Container)

	if err := container.RunContainer(context.Client, context.Container); err != nil {
		return err
	}

	return nil
}

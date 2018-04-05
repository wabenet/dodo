package context

import (
	"github.com/oclaussen/dodo/container"
)

func (context *Context) ensureContainer(command []string) error {
	if context.Container != nil {
		return nil
	}
	if err := context.ensureConfig(); err != nil {
		return err
	}
	if err := context.ensureClient(); err != nil {
		return err
	}
	if err := context.ensureImage(); err != nil {
		return err
	}

	container, err := container.CreateContainer(context.Client, context.Image, command, context.Config)
	if err != nil {
		return err
	}
	context.Container = container
	return nil
}

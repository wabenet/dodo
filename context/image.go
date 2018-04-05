package context

import (
	"errors"
	"github.com/oclaussen/dodo/image"
)

func (context *Context) ensureImage() error {
	if context.Image != "" {
		return nil
	}
	if err := context.ensureConfig(); err != nil {
		return err;
	}
	if err := context.ensureClient(); err != nil {
		return err;
	}
	// TODO: check if pulling/building is necessary, implement force pull

	if context.Config.Build != nil && context.Config.Image != "" {
		return errors.New("You can specifiy either image or build, not both.")

	} else if context.Config.Build != nil {
		image, err := image.BuildImage(context.Client, context.Config)
		if err != nil {
			return err
		}
		context.Image = image
		return nil

	} else if context.Config.Image != "" {
		image, err := image.PullImage(context.Client, context.Config)
		if err != nil {
			return err
		}
		context.Image = image
		return nil

	} else {
		return errors.New("You need to specify either image or build.")
	}
}

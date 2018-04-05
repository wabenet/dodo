package context

import (
	"github.com/oclaussen/dodo/config"
)

func (context *Context) ensureConfig() error {
	if context.Config != nil {
		return nil
	}
	config, err := config.Load(context.Options.Filename)
	if err != nil {
		return err
	}
	contextConfig := config.Contexts[context.Name]
	context.Config = &contextConfig
	return nil
}

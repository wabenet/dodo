package context

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (context *Context) ensureClient() error {
	if context.Client != nil {
		return nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}
	context.Client = client
	return nil
}

package container

import (
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func (c *Container) create(image string, tty bool) (string, error) {
	entrypoint := []string{"/bin/sh"}
	command := c.config.Command

	if c.config.Interpreter != nil {
		entrypoint = c.config.Interpreter
	}
	if c.config.Interactive {
		command = nil
	} else if len(c.config.Script) > 0 {
		entrypoint = append(entrypoint, path.Join(c.tmpPath, "entrypoint"))
	}

	response, err := c.client.ContainerCreate(
		c.context,
		&container.Config{
			User:         c.config.User,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          tty,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          c.config.Environment.Strings(),
			Cmd:          command,
			Image:        image,
			WorkingDir:   c.config.WorkingDir,
			Entrypoint:   entrypoint,
			ExposedPorts: c.config.Ports.PortSet(),
		},
		&container.HostConfig{
			AutoRemove: func() bool {
				if c.config.Remove == nil {
					return true
				}
				return *c.config.Remove
			}(),
			Binds:        c.config.Volumes.Strings(),
			VolumesFrom:  c.config.VolumesFrom,
			PortBindings: c.config.Ports.PortMap(),
		},
		&network.NetworkingConfig{},
		c.config.ContainerName,
	)
	if err != nil {
		return "", err
	}
	return response.ID, nil
}

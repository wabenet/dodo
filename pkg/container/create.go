package container

import (
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
)

func (c *Container) create(image string) (string, error) {
	opts, err := c.stage.GetDockerOptions()
	if err != nil {
		return "", err
	}

	entrypoint, command := c.dockerEntrypoint()
	response, err := c.client.ContainerCreate(
		c.context,
		&container.Config{
			User:         c.config.User,
			AttachStdin:  !c.daemon,
			AttachStdout: !c.daemon,
			AttachStderr: !c.daemon,
			Tty:          hasTTY() && !c.daemon,
			OpenStdin:    !c.daemon,
			StdinOnce:    !c.daemon,
			Env:          c.dockerEnvironment(opts),
			Cmd:          command,
			Image:        image,
			WorkingDir:   c.config.WorkingDir,
			Entrypoint:   entrypoint,
			ExposedPorts: c.config.Ports.PortSet(),
		},
		&container.HostConfig{
			AutoRemove: func() bool {
				if c.daemon {
					return false
				}
				if c.config.Remove == nil {
					return true
				}
				return *c.config.Remove
			}(),
			Binds:         c.config.Volumes.Strings(),
			VolumesFrom:   c.config.VolumesFrom,
			PortBindings:  c.config.Ports.PortMap(),
			RestartPolicy: c.dockerRestartPolicy(),
			Resources: container.Resources{
				Devices: c.dockerDevices(),
			},
		},
		&network.NetworkingConfig{},
		c.name,
	)
	if err != nil {
		return "", err
	}

	if len(c.config.Script) > 0 {
		if err = c.uploadEntrypoint(response.ID); err != nil {
			return "", err
		}
	}

	if c.config.ForwardStage {
		if err = c.uploadStageConfig(response.ID, opts); err != nil {
			return "", err
		}
	}
	return response.ID, nil
}

func (c *Container) dockerEntrypoint() ([]string, []string) {
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

	return entrypoint, command
}

func (c *Container) dockerEnvironment(opts *stage.DockerOptions) []string {
	env := c.config.Environment
	if c.config.ForwardStage {
		yes := "1"
		env = append(env, types.KeyValue{"DOCKER_HOST", &opts.Host})
		env = append(env, types.KeyValue{"DOCKER_API_VERSION", &opts.Version})
		env = append(env, types.KeyValue{"DOCKER_CERT_PATH", &c.tmpPath})
		env = append(env, types.KeyValue{"DOCKER_TLS_VERIFY", &yes})
	}
	return env.Strings()
}

func (c *Container) dockerRestartPolicy() container.RestartPolicy {
	if c.daemon {
		return container.RestartPolicy{Name: "always"}
	} else {
		return container.RestartPolicy{Name: "no"}
	}
}

func (c *Container) dockerDevices() []container.DeviceMapping {
	result := make([]container.DeviceMapping, len(c.config.Devices))
	for i, device := range c.config.Devices {
		result[i] = container.DeviceMapping{
			PathOnHost:        device.Source,
			PathInContainer:   device.Target,
			CgroupPermissions: "mrw",
		}
	}
	return result
}

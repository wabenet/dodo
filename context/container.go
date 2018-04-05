package context

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (context *Context) ensureContainer() error {
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

	// TODO: volumes: allow variables (HOME, PWD, ...)
	volumes := []string{}
	if context.Config.Volumes != nil {
		for _, volume := range context.Config.Volumes.Volumes {
			volumes = append(volumes, volume.String())
		}
	}

	if len(context.Config.Interpreter) == 0 {
		context.Config.Interpreter = []string{"/bin/sh", "-eux", context.Entrypoint}
	}

	container, err := context.Client.CreateContainer(docker.CreateContainerOptions{
		Name:   context.Config.ContainerName,
		Config: &docker.Config{
			User:         context.Config.User,
			Env:          context.Config.Environment, // TODO: support env_file
			Image:        context.Image,
			WorkingDir:   context.Config.WorkingDir,
			Entrypoint:   context.Config.Interpreter,
			Cmd:          context.Options.Arguments,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    true,
		},
		HostConfig: &docker.HostConfig{
			Binds:        volumes,
			VolumesFrom:  context.Config.VolumesFrom,
		},
	})
	if err != nil {
		return err
	}
	context.Container = container
	return nil
}

func (context *Context) ensureCleanup() {
	if context.Container == nil {
		return
	}
	if err := context.ensureClient(); err != nil {
		return
	}
	if context.Config.Remove != nil && !*context.Config.Remove {
		return
	}

	context.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            context.Container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	context.Container = nil
}

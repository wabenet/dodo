package state

import (
	docker "github.com/fsouza/go-dockerclient"
)

func (state *state) ensureContainer() error {
	if state.Container != nil {
		return nil
	}
	if err := state.ensureConfig(); err != nil {
		return err
	}
	if err := state.ensureClient(); err != nil {
		return err
	}
	if err := state.ensureImage(); err != nil {
		return err
	}

	container, err := state.Client.CreateContainer(docker.CreateContainerOptions{
		Name:   state.Config.ContainerName,
		Config: &docker.Config{
			User:         state.Config.User,
			Env:          state.Config.Environment,
			Image:        state.Image,
			WorkingDir:   state.Config.WorkingDir,
			Entrypoint:   state.getEntrypoint(),
			Cmd:          state.Options.Arguments,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			OpenStdin:    true,
			StdinOnce:    true,
		},
		HostConfig: &docker.HostConfig{
			Binds:        state.Config.Volumes,
			VolumesFrom:  state.Config.VolumesFrom,
		},
	})
	if err != nil {
		return err
	}
	state.Container = container
	return nil
}

func (state *state) ensureCleanup() {
	if state.Container == nil {
		return
	}
	if err := state.ensureClient(); err != nil {
		return
	}
	if state.Config.Remove != nil && !*state.Config.Remove {
		return
	}

	state.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            state.Container.ID,
		RemoveVolumes: true,
		Force:         true,
	})
	state.Container = nil
}

func (state *state) getEntrypoint() []string {
	entrypoint := []string{"/bin/sh"}
	if len(state.Config.Interpreter) > 0 {
		entrypoint = state.Config.Interpreter
	}
	if !state.Options.Interactive {
		entrypoint = append(entrypoint, state.Entrypoint)
	}
	return entrypoint
}

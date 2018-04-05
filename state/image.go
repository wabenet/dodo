package state

import (
	"errors"
	"io"
	"os"
	"encoding/json"
	"strings"

	"github.com/oclaussen/dodo/config"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	docker "github.com/fsouza/go-dockerclient"
)

func (state *state) ensureImage() error {
	if state.Image != "" {
		return nil
	}
	if err := state.ensureConfig(); err != nil {
		return err;
	}
	if err := state.ensureClient(); err != nil {
		return err;
	}

	if state.Config.Build != nil {
		image, err := buildImage(state.Client, state.Config)
		if err != nil {
			return err
		}
		state.Image = image
		return nil

	} else if state.Config.Image != "" {
		image, err := pullImage(state.Client, state.Config)
		if err != nil {
			return err
		}
		state.Image = image
		return nil

	} else {
		return errors.New("You need to specify either image or build.")
	}
}

func buildImage(client *docker.Client, config *config.BackdropConfig) (string, error) {
	if config.Image != "" && !config.Build.ForceRebuild {
		images, err := client.ListImages(docker.ListImagesOptions{
			Filter: config.Image,
		})
		if err == nil && len(images) > 0 {
			// TODO: log error
			return config.Image, nil
		}
	}

	args := []docker.BuildArg{}
	for _, arg := range config.Build.Args {
		switch values := strings.SplitN(arg, "=", 2); len(values) {
		case 1:
			args = append(args, docker.BuildArg{Name: values[0], Value: "\x00"})
		case 2:
			args = append(args, docker.BuildArg{Name: values[0], Value: values[1]})
		}
	}

	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return "", err
	}

	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	image := ""
	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		// TODO: handle parse error
		if err := json.Unmarshal(*auxJSON, &result); err == nil {
			image = result.ID
		}
	}

	errChan := make(chan error)
	go func() {
		outFd, isTerminal := term.GetFdInfo(os.Stdout)
		errChan <- jsonmessage.DisplayJSONMessagesStream(rpipe, os.Stdout, outFd, isTerminal, aux)
	}()

	err = client.BuildImage(docker.BuildImageOptions{
		Name:           config.Image,
		Dockerfile:     config.Build.Dockerfile,
		NoCache:        config.Build.NoCache,
		CacheFrom:      []string{}, // TODO implement cache_from
		SuppressOutput: false, // TODO: quiet mode
		Pull:           config.Pull,
		RmTmpContainer: true,
		RawJSONStream:  true,
		OutputStream:   wpipe,
		AuthConfigs:    *authConfigs,
		ContextDir:     config.Build.Context,
		BuildArgs:      args,
	})

	wpipe.Close()
	if err != nil {
		<-errChan
		return "", err
	}

	err = <-errChan
	if err != nil {
		return "", err
	}
	if image == "" {
		return "", errors.New("Build complete, but the server did not send an image id.")
	}
	return image, nil
}

func pullImage(client *docker.Client, config *config.BackdropConfig) (string, error) {
	if !config.Pull {
		images, err := client.ListImages(docker.ListImagesOptions{
			Filter: config.Image,
		})
		if err == nil && len(images) > 0 {
			// TODO: log error
			return config.Image, nil
		}
	}

	ref, err := reference.ParseNormalizedNamed(config.Image)
	if err != nil {
		return "", err
	}
	tagged := reference.TagNameOnly(ref).(reference.Tagged)

	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return "", err
	}
	authConfig := authConfigs.Configs[reference.Domain(ref)]

	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	errChan := make(chan error)
	go func() {
		outFd, isTerminal := term.GetFdInfo(os.Stdout)
		errChan <- jsonmessage.DisplayJSONMessagesStream(rpipe, os.Stdout, outFd, isTerminal, nil)
	}()

	err = client.PullImage(docker.PullImageOptions{
		Repository:     ref.Name(),
		Tag:            tagged.Tag(),
		OutputStream:   wpipe,
		RawJSONStream:  true,
	}, authConfig)

	wpipe.Close()
	if err != nil {
		<-errChan
		return "", err
	}

	return tagged.String(), <-errChan
}

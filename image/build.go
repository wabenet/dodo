package image

import (
	"io"
	"os"
	"encoding/json"
	"errors"
	"strings"

	"github.com/oclaussen/dodo/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

func BuildImage(client *docker.Client, config *config.ContextConfig) (string, error) {
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

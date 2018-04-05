package command

import (
	"io"
	"os"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) buildImage() (string, error) {
	config := command.Config.Build

	args := []docker.BuildArg{}
	for key, value := range config.Args {
		args = append(args, docker.BuildArg{Name: key, Value: *value})
	}

	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return "", err
	}

	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	imageID := ""
	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		// TODO: handle parse error
		if err := json.Unmarshal(*auxJSON, &result); err == nil {
			imageID = result.ID
		}
	}

	errChan := make(chan error)
	go func() {
		outFd, isTerminal := term.GetFdInfo(os.Stdout)
		errChan <- jsonmessage.DisplayJSONMessagesStream(rpipe, os.Stdout, outFd, isTerminal, aux)
	}()

	err = command.Client.BuildImage(docker.BuildImageOptions{
		Dockerfile:     config.Dockerfile,
		NoCache:        false, // TODO no cache mode
		CacheFrom:      []string{}, // TODO implement cache_from
		SuppressOutput: false, // TODO: quiet mode
		Pull:           true, // TODO: force pull option
		RmTmpContainer: true,
		RawJSONStream:  true,
		OutputStream:   wpipe,
		AuthConfigs:    *authConfigs,
		ContextDir:     config.Context,
		BuildArgs:      args,
	})

	wpipe.Close()
	if err != nil {
		<-errChan
		return "", err
	}
	return imageID, <-errChan
}

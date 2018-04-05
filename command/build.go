package command

import (
	"io"
	"os"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) buildImage() error {
	config := command.Config.Build

	args := []docker.BuildArg{}
	for key, value := range config.Args {
		args = append(args, docker.BuildArg{Name: key, Value: *value})
	}

	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}

	outFd, isTTY := term.GetFdInfo(os.Stdout)
	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	errChan := make(chan error)
	go func() {
		err := jsonmessage.DisplayJSONMessagesStream(rpipe, os.Stdout, outFd, isTTY, nil)
		errChan <- err
	}()

	err = command.Client.BuildImage(docker.BuildImageOptions{
		Name:           "dodo-testing", // TODO: should have no name, figure out id
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
		return err
	}
	return <-errChan
}

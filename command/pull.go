package command

import (
	"io"
	"os"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/distribution/reference"
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) pullImage() error {
	// TODO: validate that the image is actually normalized named
	ref, err := reference.ParseNormalizedNamed(command.Config.Image)
	if err != nil {
		return err
	}
	tagged := reference.TagNameOnly(ref).(reference.Tagged)

	authConfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}
	authConfig := authConfigs.Configs[reference.Domain(ref)]

	rpipe, wpipe := io.Pipe()
	defer rpipe.Close()

	errChan := make(chan error)
	go func() {
		outFd, isTerminal := term.GetFdInfo(os.Stdout)
		errChan <- jsonmessage.DisplayJSONMessagesStream(rpipe, os.Stdout, outFd, isTerminal, nil)
	}()

	err = command.Client.PullImage(docker.PullImageOptions{
		Repository:     ref.Name(),
		Tag:            tagged.Tag(),
		OutputStream:   wpipe,
		RawJSONStream:  true,
	}, authConfig)

	wpipe.Close()
	if err != nil {
		<-errChan
		return err
	}
	return <-errChan
}

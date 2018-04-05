package image

import (
	"io"
	"os"

	"github.com/oclaussen/dodo/config"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/distribution/reference"
	docker "github.com/fsouza/go-dockerclient"
)

func PullImage(client *docker.Client, config *config.ContextConfig) (string, error) {
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

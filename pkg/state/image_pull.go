package state

import (
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func pullImage(ctx context.Context, client *client.Client, config *config.BackdropConfig) (string, error) {
	if !config.Pull {
		if image := useExistingImage(ctx, client, config); image != "" {
			return config.Image, nil
		}
	}

	response, err := client.ImagePull(
		ctx,
		config.Image,
		types.ImagePullOptions{},
	)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := response.Close(); err != nil {
			log.Error(err)
		}
	}()

	outFd, isTerminal := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(response, os.Stdout, outFd, isTerminal, nil)
	if err != nil {
		return "", err
	}

	return config.Image, nil
}

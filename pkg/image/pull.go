package image

import (
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func pull(ctx context.Context, options Options) (string, error) {
	response, err := options.Client.ImagePull(
		ctx,
		options.Name,
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

	return options.Name, nil
}

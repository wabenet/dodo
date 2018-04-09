package image

import (
	"github.com/docker/docker/api/types"
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

	_, err = handleImageResult(response, false)
	if err != nil {
		return "", err
	}

	return options.Name, nil
}

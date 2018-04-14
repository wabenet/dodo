package image

import (
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func pull(ctx context.Context, options Options) (string, error) {
	parsed, err := reference.ParseNormalizedNamed(options.Name)
	if err != nil {
		return "", err
	}

	response, err := options.Client.ImagePull(
		ctx,
		parsed.String(),
		types.ImagePullOptions{},
	)
	if err != nil {
		return "", err
	}

	_, err = handleImageResult(response, false)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

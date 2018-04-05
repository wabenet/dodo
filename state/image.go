package state

import (
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/config"
	"golang.org/x/net/context"
)

// TODO: authentication

func (state *State) EnsureImage(ctx context.Context) (string, error) {
	if state.Image != "" {
		return state.Image, nil
	}
	client, err := state.EnsureClient(ctx)
	if err != nil {
		return "", err
	}
	config, err := state.EnsureConfig(ctx)
	if err != nil {
		return "", err
	}

	if config.Build != nil {
		image, err := buildImage(ctx, client, config)
		if err != nil {
			return "", err
		}
		state.Image = image
		return state.Image, nil

	} else if config.Image != "" {
		image, err := pullImage(ctx, client, config)
		if err != nil {
			return "", err
		}
		state.Image = image
		return state.Image, nil

	} else {
		return "", errors.New("You need to specify either image or build.")
	}
}

func useExistingImage(ctx context.Context, client *client.Client, config *config.BackdropConfig) string {
	images, err := client.ImageList(
		ctx,
		types.ImageListOptions{
			Filters: filters.NewArgs(filters.Arg("reference", config.Image)),
		},
	)
	if err == nil && len(images) > 0 {
		// TODO: log error
		return config.Image
	} else {
		return ""
	}
}

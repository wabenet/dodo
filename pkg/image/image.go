package image

import (
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/config"
	"golang.org/x/net/context"
)

// TODO: authentication

// Options represents the configuration for a docker image that can be
// either built or pulled.
type Options struct {
	Client    *client.Client
	Name      string
	Build     *config.BuildConfig
	ForcePull bool
}

// Get gets a valid image id, and builds or pulls the image if necessary.
func Get(ctx context.Context, options Options) (string, error) {
	if options.Client == nil {
		return "", errors.New("client may not be nil")
	} else if name, ok := existsLocally(ctx, options); ok {
		return name, nil
	} else if options.Build != nil {
		return build(ctx, options)
	} else if options.Name != "" {
		return pull(ctx, options)
	} else {
		return "", errors.New("you need to specify either image name or build")
	}
}

func existsLocally(ctx context.Context, options Options) (string, bool) {
	if options.Name == "" {
		return "", false
	} else if options.Build != nil && options.Build.ForceRebuild {
		return "", false
	} else if options.Build == nil && options.ForcePull {
		return "", false
	}

	filter := filters.NewArgs(filters.Arg("reference", options.Name))
	images, err := options.Client.ImageList(
		ctx,
		types.ImageListOptions{Filters: filter},
	)
	if err != nil || len(images) == 0 {
		return "", false
	}
	return options.Name, true
}

package image

import (
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// TODO: authentication

// Options represents the configuration for a docker image that can be
// either built or pulled.
type Options struct {
	Client     Client
	Name       string
	ForcePull  bool
	DoBuild    bool
	ForceBuild bool
	NoCache    bool
	Context    string
	Dockerfile string
	Steps      []string
	Args       []string
}

// Client represents a docker client that can do everything this package
// needs
type Client interface {
	ImageList(context.Context, types.ImageListOptions) ([]types.ImageSummary, error)
	ImagePull(context.Context, string, types.ImagePullOptions) (io.ReadCloser, error)
	ImageBuild(context.Context, io.Reader, types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

// Get gets a valid image id, and builds or pulls the image if necessary.
func Get(ctx context.Context, options Options) (string, error) {
	if options.Client == nil {
		return "", errors.New("client may not be nil")
	} else if name, ok := existsLocally(ctx, options); ok {
		log.Info(fmt.Sprintf("Using image %s", name))
		return name, nil
	} else if options.DoBuild {
		if options.Name != "" {
			log.Info(fmt.Sprintf("Image %s not found, building...", options.Name))
		} else {
			log.Info("Building image...")
		}
		return build(ctx, options)
	} else if options.Name != "" {
		log.Info(fmt.Sprintf("Image %s not found locally, pulling...", options.Name))
		return pull(ctx, options)
	} else {
		return "", errors.New("you need to specify either image name or build")
	}
}

func existsLocally(ctx context.Context, options Options) (string, bool) {
	if options.Name == "" {
		return "", false
	} else if options.DoBuild && options.ForceBuild {
		return "", false
	} else if options.ForcePull {
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

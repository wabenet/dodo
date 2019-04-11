package image

import (
	"io"
	"net"

	"github.com/docker/docker/api/types"
	dodotypes "github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	errMissingImageID = errors.New(
		"build complete, but the server did not send an image id")
)

// Image represents the data necessary to build a docker image
type Image struct {
	config      *dodotypes.Image
	client      Client
	authConfigs map[string]types.AuthConfig
	session     session
}

// Client represents a docker client that can do everything this package needs
type Client interface {
	DialHijack(context.Context, string, string, map[string][]string) (net.Conn, error)
	ImageBuild(context.Context, io.Reader, types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

// NewImage initializes and validates a new Image object
func NewImage(client Client, authConfigs map[string]types.AuthConfig, config *dodotypes.Image) (*Image, error) {
	if client == nil {
		return nil, errors.New("client may not be nil")
	}

	session, err := prepareSession(config.Context)
	if err != nil {
		return nil, err
	}

	return &Image{
		client:      client,
		authConfigs: authConfigs,
		config:      config,
		session:     session,
	}, nil
}

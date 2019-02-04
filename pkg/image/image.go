package image

import (
	"io"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/versions"
	dodotypes "github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var (
	errMissingImageID = errors.New(
		"build complete, but the server did not send an image id")
)

type image struct {
	config      *ImageConfig
	client      Client
	authConfigs map[string]types.AuthConfig
}

// ImageConfig represents the build configuration for a docker image
type ImageConfig struct {
	Name         string
	Context      string
	Dockerfile   string
	Steps        []string
	Args         dodotypes.KeyValueList
	NoCache      bool
	ForceRebuild bool
	ForcePull    bool
}

// Client represents a docker client that can do everything this package
// needs
type Client interface {
	ClientVersion() string
	Ping(context.Context) (types.Ping, error)
	DialSession(context.Context, string, map[string][]string) (net.Conn, error)
	ImageBuild(context.Context, io.Reader, types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

func NewImage(client Client, authConfigs map[string]types.AuthConfig, config *ImageConfig) (*image, error) {
	if client == nil {
		return nil, errors.New("client may not be nil")
	}
	ping, err := client.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	if !ping.Experimental || versions.LessThan(client.ClientVersion(), "1.31") {
		return nil, errors.Errorf("buildkit not supported by daemon")
	}

	return &image{
		client:      client,
		authConfigs: authConfigs,
		config:      config,
	}, nil
}

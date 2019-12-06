package stage

import (
	"github.com/docker/docker/client"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
)

const (
	DefaultAPIVersion = "1.39"
)

type Stage interface {
	Initialize(string, *types.Stage) error
	Cleanup()
	Create() error
	Start() error
	Stop() error
	Remove(bool) error
	Exist() (bool, error)
	Available() (bool, error)
	GetSSHOptions() (*SSHOptions, error)
	GetDockerOptions() (*DockerOptions, error)
}

type SSHOptions struct {
	Hostname       string
	Port           int
	Username       string
	PrivateKeyFile string
}

type DockerOptions struct {
	Version  string
	Host     string
	CAFile   string
	CertFile string
	KeyFile  string
}

func GetDockerClient(s Stage) (*client.Client, error) {
	available, err := s.Available()
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("stage is not up")
	}
	opts, err := s.GetDockerOptions()
	if err != nil {
		return nil, err
	}
	mutators := []client.Opt{}
	if len(opts.Version) > 0 {
		mutators = append(mutators, client.WithVersion(opts.Version))
	} else {
		mutators = append(mutators, client.WithVersion(DefaultAPIVersion))
	}
	if len(opts.Host) > 0 {
		mutators = append(mutators, client.WithHost(opts.Host))
	}
	if len(opts.CAFile)+len(opts.CertFile)+len(opts.KeyFile) > 0 {
		mutators = append(mutators, client.WithTLSClientConfig(opts.CAFile, opts.CertFile, opts.KeyFile))
	}
	return client.NewClientWithOpts(mutators...)
}

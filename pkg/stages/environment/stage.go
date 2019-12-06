package environment

import (
	"os"
	"path/filepath"

	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
)

type Stage struct {
	Options *stage.DockerOptions
}

func (s *Stage) Initialize(_ string, _ *types.Stage) (bool, error) {
	opts := &stage.DockerOptions{
		Host: os.Getenv("DOCKER_HOST"),
	}
	if version := os.Getenv("DOCKER_API_VERSION"); len(version) > 0 {
		opts.Version = version
	}
	if certPath := os.Getenv("DOCKER_CERT_PATH"); len(certPath) > 0 {
		opts.CAFile = filepath.Join(certPath, "ca.pem")
		opts.CertFile = filepath.Join(certPath, "cert.pem")
		opts.KeyFile = filepath.Join(certPath, "key.pem")
	}
	s.Options = opts
	return true, nil
}

func (s *Stage) Create() error {
	return nil
}

func (s *Stage) Start() error {
	return nil
}

func (s *Stage) Stop() error {
	return nil
}

func (s *Stage) Remove(_ bool) error {
	return nil
}

func (s *Stage) Exist() (bool, error) {
	return true, nil
}

func (s *Stage) Available() (bool, error) {
	return true, nil
}

func (s *Stage) GetSSHOptions() (*stage.SSHOptions, error) {
	return nil, nil
}

func (s *Stage) GetDockerOptions() (*stage.DockerOptions, error) {
	return s.Options, nil
}

package stage

import (
	"os"
	"path/filepath"

	"github.com/oclaussen/dodo/pkg/types"
)

type EnvStage struct {
	Options *DockerOptions
}

func (stage *EnvStage) Initialize(_ string, _ *types.Stage) (bool, error) {
	opts := &DockerOptions{
		Version: DefaultAPIVersion,
		Host:    os.Getenv("DOCKER_HOST"),
	}
	if version := os.Getenv("DOCKER_API_VERSION"); len(version) > 0 {
		opts.Version = version
	}
	if certPath := os.Getenv("DOCKER_CERT_PATH"); len(certPath) > 0 {
		opts.CAFile = filepath.Join(certPath, "ca.pem")
		opts.CertFile = filepath.Join(certPath, "cert.pem")
		opts.KeyFile = filepath.Join(certPath, "key.pem")
	}
	stage.Options = opts
	return true, nil
}

func (stage *EnvStage) Create() error {
	return nil
}

func (stage *EnvStage) Start() error {
	return nil
}

func (stage *EnvStage) Stop() error {
	return nil
}

func (stage *EnvStage) Remove(_ bool) error {
	return nil
}

func (stage *EnvStage) Exist() (bool, error) {
	return true, nil
}

func (stage *EnvStage) Available() (bool, error) {
	return true, nil
}

func (stage *EnvStage) GetSSHOptions() (*SSHOptions, error) {
	return nil, nil
}

func (stage *EnvStage) GetDockerOptions() (*DockerOptions, error) {
	return stage.Options, nil
}

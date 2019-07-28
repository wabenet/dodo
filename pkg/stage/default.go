package stage

import (
	"os"
	"path/filepath"
)

const (
	DefaultStageName  = "default"
	DefaultAPIVersion = "1.39"
)

// TODO: replace with a generic and environment provider

type DefaultStage struct{}

func (stage *DefaultStage) Initialize(_ string, _ map[string]string) (bool, error) {
	return true, nil
}

func (stage *DefaultStage) Create() error {
	return nil
}

func (stage *DefaultStage) Start() error {
	return nil
}

func (stage *DefaultStage) Stop() error {
	return nil
}

func (stage *DefaultStage) Remove(_ bool) error {
	return nil
}

func (stage *DefaultStage) Exist() (bool, error) {
	return true, nil
}

func (stage *DefaultStage) Available() (bool, error) {
	return true, nil // TODO: actually check for this
}

func (stage *DefaultStage) GetSSHOptions() (*SSHOptions, error) {
	return nil, nil
}

func (stage *DefaultStage) GetDockerOptions() (*DockerOptions, error) {
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
	return opts, nil
}

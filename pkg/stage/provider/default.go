package provider

import (
	"os"
	"path/filepath"
)

const (
	DefaultProviderName = "default"
	DefaultAPIVersion   = "1.39"
)

// TODO: replace with a generic and environment provider

type DefaultProvider struct{}

func (prov *DefaultProvider) Initialize(_ map[string]string) (bool, error) {
	return true, nil
}

func (prov *DefaultProvider) Create() error {
	return nil
}

func (prov *DefaultProvider) Start() error {
	return nil
}

func (prov *DefaultProvider) Stop() error {
	return nil
}

func (prov *DefaultProvider) Remove(_ bool) error {
	return nil
}

func (prov *DefaultProvider) Exist() (bool, error) {
	return true, nil
}

func (prov *DefaultProvider) Available() (bool, error) {
	return true, nil // TODO: actually check for this
}

func (prov *DefaultProvider) GetSSHOptions() (*SSHOptions, error) {
	return nil, nil
}

func (prov *DefaultProvider) GetDockerOptions() (*DockerOptions, error) {
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

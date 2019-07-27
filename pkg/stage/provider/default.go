package provider

import (
	"os"
	"path/filepath"
)

const (
	DefaultProviderName = "default"
	DefaultAPIVersion   = "1.39"
)

type DefaultProvider struct{}

func (prov *DefaultProvider) Initialize(_ map[string]string) (bool, error) {
	return true, nil
}

func (prov *DefaultProvider) Status() (Status, error) {
	return Up, nil
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

func (prov *DefaultProvider) Remove() error {
	return nil
}

func (prov *DefaultProvider) GetURL() (string, error) {
	return "", nil
}

func (prov *DefaultProvider) GetIP() (string, error) {
	return "", nil
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

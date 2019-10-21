package stage

import (
	"github.com/oclaussen/dodo/pkg/types"
)

type GenericStage struct {
	Options *DockerOptions
}

func (stage *GenericStage) Initialize(_ string, config *types.Stage) (bool, error) {
	stage.Options = &DockerOptions{
		Version:  config.Options["api_version"],
		Host:     config.Options["host"],
		CAFile:   config.Options["ca_file"],
		CertFile: config.Options["cert_file"],
		KeyFile:  config.Options["key_file"],
	}
	return true, nil
}

func (stage *GenericStage) Create() error {
	return nil
}

func (stage *GenericStage) Start() error {
	return nil
}

func (stage *GenericStage) Stop() error {
	return nil
}

func (stage *GenericStage) Remove(_ bool) error {
	return nil
}

func (stage *GenericStage) Exist() (bool, error) {
	return true, nil
}

func (stage *GenericStage) Available() (bool, error) {
	return true, nil
}

func (stage *GenericStage) GetSSHOptions() (*SSHOptions, error) {
	return nil, nil
}

func (stage *GenericStage) GetDockerOptions() (*DockerOptions, error) {
	return stage.Options, nil
}

package stage

import (
	"github.com/mitchellh/mapstructure"
	"github.com/oclaussen/dodo/pkg/types"
)

type GenericStage struct {
	Options *DockerOptions
}

func (stage *GenericStage) Initialize(_ string, config *types.Stage) (bool, error) {
	stage.Options = &DockerOptions{}
	if err := mapstructure.Decode(config.Options, stage.Options); err != nil {
		return false, err
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

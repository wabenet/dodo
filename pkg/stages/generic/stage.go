package generic

import (
	"github.com/mitchellh/mapstructure"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
)

type Stage struct {
	Options *stage.DockerOptions
}

func (s *Stage) Initialize(_ string, config *types.Stage) error {
	s.Options = &stage.DockerOptions{}
	if err := mapstructure.Decode(config.Options, s.Options); err != nil {
		return err
	}
	return nil
}

func (s *Stage) Cleanup() {}

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

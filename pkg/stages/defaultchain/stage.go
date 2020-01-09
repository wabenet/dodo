package defaultchain

import (
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/stages/environment"
	"github.com/oclaussen/dodo/pkg/stages/generic"
	"github.com/oclaussen/dodo/pkg/stages/grpc"
	"github.com/oclaussen/dodo/pkg/types"

	"github.com/pkg/errors"
)

type Stage struct {
	current stage.Stage
}

func (s *Stage) Initialize(name string, conf *types.Stage) error {
	candidates := []*types.Stage{
		conf,
		&types.Stage{Type: "environment"},
	}

	for _, currentConfig := range candidates {
		if currentConfig == nil {
			continue
		}

		switch currentConfig.Type {
		case "environment":
			s.current = &environment.Stage{}
		case "generic":
			s.current = &generic.Stage{}
		default:
			s.current = &grpc.Stage{}
		}

		if err := s.current.Initialize(name, currentConfig); err == nil {
			return nil
		}

		s.current.Cleanup()
		s.current = nil
	}

	return errors.New("no valid stage exist for configuration")
}

func (s *Stage) Cleanup() {
	if s.current != nil {
		s.current.Cleanup()
	}
}

func (s *Stage) Create() error {
	return s.current.Create()
}

func (s *Stage) Start() error {
	return s.current.Start()
}

func (s *Stage) Stop() error {
	return s.current.Stop()
}

func (s *Stage) Remove(force bool, volumes bool) error {
	return s.current.Remove(force, volumes)
}

func (s *Stage) Exist() (bool, error) {
	return s.current.Exist()
}

func (s *Stage) Available() (bool, error) {
	return s.current.Available()
}

func (s *Stage) GetSSHOptions() (*stage.SSHOptions, error) {
	return s.current.GetSSHOptions()
}

func (s *Stage) GetDockerOptions() (*stage.DockerOptions, error) {
	return s.current.GetDockerOptions()
}

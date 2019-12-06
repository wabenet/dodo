package grpc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/pkg/errors"
)

type Stage struct {
	wrapped stage.Stage
	client  *plugin.Client
}

func (s *Stage) Initialize(name string, conf *types.Stage) error {
	path, err := findPluginExecutable(conf.Type)
	if err != nil {
		return err
	}

	s.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig(conf.Type),
		Plugins:          PluginMap,
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger:           stage.NewPluginLogger(),
	})

	c, err := s.client.Client()
	if err != nil {
		return err
	}
	raw, err := c.Dispense("stage")
	if err != nil {
		return err
	}

	s.wrapped = raw.(stage.Stage)
	return s.wrapped.Initialize(name, conf)
}

func findPluginExecutable(name string) (string, error) {
	directories, err := configfiles.GimmeConfigDirectories(&configfiles.Options{
		Name:                      "dodo",
		IncludeWorkingDirectories: true,
	})
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("plugin-%s_%s_%s", name, runtime.GOOS, runtime.GOARCH)
	for _, dir := range directories {
		path := filepath.Join(dir, ".dodo", "plugins", filename)
		if stat, err := os.Stat(path); err == nil && stat.Mode().Perm()&0111 != 0 {
			return path, nil
		}
	}

	return "", errors.New("could not find a suitable plugin for the stage anywhere")
}

func (s *Stage) Cleanup() {
	if s.wrapped != nil {
		s.wrapped.Cleanup()
	}
	if s.client != nil {
		s.client.Kill()
	}
}

func (s *Stage) Create() error {
	return s.wrapped.Create()
}

func (s *Stage) Start() error {
	return s.wrapped.Start()
}

func (s *Stage) Stop() error {
	return s.wrapped.Stop()
}

func (s *Stage) Remove(force bool) error {
	return s.wrapped.Remove(force)
}

func (s *Stage) Exist() (bool, error) {
	return s.wrapped.Exist()
}

func (s *Stage) Available() (bool, error) {
	return s.wrapped.Available()
}

func (s *Stage) GetSSHOptions() (*stage.SSHOptions, error) {
	return s.wrapped.GetSSHOptions()
}

func (s *Stage) GetDockerOptions() (*stage.DockerOptions, error) {
	return s.wrapped.GetDockerOptions()
}

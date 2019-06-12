package stage

import (
	"encoding/json"
	"os/user"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
)

// TODO: make machine dir configurable and default somewhere not docker-machine

type Stage struct {
	name   string
	config *types.Stage
	api    libmachine.API
	host   *host.Host
	exists bool
}

func LoadStage(name string, config *types.Stage) (*Stage, error) {
	machineDir := filepath.Join(home(), ".docker", "machine")
	api := libmachine.NewClient(machineDir, filepath.Join(machineDir, "certs"))
	stage := &Stage{name: name, config: config, api: api}

	var err error
	stage.exists, err = api.Exists(name)
	if err != nil {
		return stage, errors.Wrap(err, "could not check if stage exists")
	}

	if stage.exists {
		stage.host, err = api.Load(name)
		if err != nil {
			return stage, errors.Wrap(err, "could not load stage")
		}
	} else {
		driverConfig, _ := json.Marshal(&drivers.BaseDriver{
			MachineName: name,
			StorePath:   machineDir,
		})
		stage.host, err = api.NewHost(stage.config.Type, driverConfig)
		if err != nil {
			return stage, errors.Wrap(err, "could not create stage")
		}
	}

	return stage, nil
}

func home() string {
	user, err := user.Current()
	if err != nil {
		return filepath.FromSlash("/")
	}
	if user.HomeDir == "" {
		return filepath.FromSlash("/")
	}
	return user.HomeDir
}

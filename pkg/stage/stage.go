package stage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/state"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// TODO: make machine dir configurable and default somewhere not docker-machine

type Stage struct {
	name     string
	config   *types.Stage
	stateDir string
	driver   drivers.Driver
	exists   bool
}

func LoadStage(name string, config *types.Stage) (*Stage, error) {
	stage := &Stage{
		name:     name,
		config:   config,
		stateDir: filepath.Join(home(), ".docker", "machine"),
	}

	_, err := os.Stat(stage.hostDir())
	if os.IsNotExist(err) {
		stage.exists = false
	} else if err == nil {
		stage.exists = true
	} else {
		return stage, errors.Wrap(err, "could not check if stage exists")
	}

	driverConfig, _ := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
		StorePath:   stage.stateDir,
	})

	driver, err := rpcdriver.NewRPCClientDriverFactory().NewRPCClientDriver(stage.config.Type, driverConfig)
	if err != nil {
		return stage, errors.Wrap(err, "could not create stage")
	}
	stage.driver = drivers.NewSerialDriver(driver)

	return stage, nil
}

func (stage *Stage) waitForState(desiredState state.State) error {
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		currentState, err := stage.driver.GetState()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get machine state")
		}
		if currentState == desiredState {
			return nil
		}
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("maximum number of retries (%d) exceeded", maxAttempts)
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

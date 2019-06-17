package stage

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// TODO: make machine dir configurable and default somewhere not docker-machine

type Stage struct {
	name     string
	config   *types.Stage
	stateDir string
	host     *host.Host
	exists   bool
}

func LoadStage(name string, config *types.Stage) (*Stage, error) {
	machineDir := filepath.Join(home(), ".docker", "machine")
	certsDir := filepath.Join(machineDir, "certs")
	stage := &Stage{
		name:     name,
		config:   config,
		stateDir: machineDir,
	}

	_, err := os.Stat(stage.hostDir())
	if os.IsNotExist(err) {
		stage.exists = false
	} else if err == nil {
		stage.exists = true
	} else {
		return stage, errors.Wrap(err, "could not check if stage exists")
	}

	driverFactory := rpcdriver.NewRPCClientDriverFactory()

	if stage.exists {
		if err := stage.loadState(); err != nil {
			return stage, errors.Wrap(err, "could not load stage")
		}

		driver, err := driverFactory.NewRPCClientDriver(stage.host.DriverName, stage.host.RawDriver)
		if err != nil {
			return stage, errors.Wrap(err, "could not load stage")
		}
		stage.host.Driver = drivers.NewSerialDriver(driver)
	} else {
		driverConfig, _ := json.Marshal(&drivers.BaseDriver{
			MachineName: name,
			StorePath:   machineDir,
		})

		driver, err := driverFactory.NewRPCClientDriver(stage.config.Type, driverConfig)
		if err != nil {
			return stage, errors.Wrap(err, "could not create stage")
		}

		stage.host = &host.Host{
			ConfigVersion: version.ConfigVersion,
			Name:          driver.GetMachineName(),
			Driver:        driver,
			DriverName:    driver.DriverName(),
			HostOptions: &host.Options{
				AuthOptions: &auth.Options{
					StorePath:        stage.hostDir(),
					CertDir:          certsDir,
					CaCertPath:       filepath.Join(certsDir, "ca.pem"),
					CaPrivateKeyPath: filepath.Join(certsDir, "ca-key.pem"),
					ClientCertPath:   filepath.Join(certsDir, "cert.pem"),
					ClientKeyPath:    filepath.Join(certsDir, "key.pem"),
					ServerCertPath:   filepath.Join(stage.hostDir(), "server.pem"),
					ServerKeyPath:    filepath.Join(stage.hostDir(), "server-key.pem"),
				},
				EngineOptions: &engine.Options{
					InstallURL:       drivers.DefaultEngineInstallURL,
					StorageDriver:    "overlay2",
					TLSVerify:        true,
					ArbitraryFlags:   []string{},
					Env:              []string{},
					InsecureRegistry: []string{},
					Labels:           []string{},
					RegistryMirror:   []string{},
				},
				SwarmOptions: &swarm.Options{
					Host:     "tcp://0.0.0.0:3376",
					Image:    "swarm:latest",
					Strategy: "spread",
				},
			},
		}
	}

	return stage, nil
}

func (stage *Stage) waitForState(desiredState state.State) error {
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		currentState, err := stage.host.Driver.GetState()
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

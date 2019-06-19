package stage

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	machineauth "github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/oclaussen/dodo/pkg/stage/auth"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Up() error {
	if stage.exists {
		return stage.start()
	} else {
		return stage.create()
	}
}

func (stage *Stage) start() error {
	log.WithFields(log.Fields{"name": stage.name}).Info("starting stage...")

	currentState, err := stage.driver.GetState()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Debug("could not get machine state")
	}
	if currentState == state.Running {
		log.WithFields(log.Fields{"name": stage.name}).Info("stage is already up")
		return nil
	}

	if err := stage.driver.Start(); err != nil {
		return errors.Wrap(err, "could not start stage")
	}

	if err := stage.waitForDocker(); err != nil {
		return errors.Wrap(err, "docker did start successfully")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil

}

func (stage *Stage) create() error {
	driverOpts, err := stage.driverOptions()
	if err != nil {
		return err
	}
	if err = stage.driver.SetConfigFromFlags(driverOpts); err != nil {
		return errors.Wrap(err, "could not configure stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("running pre-create checks...")

	if err := stage.driver.PreCreateCheck(); err != nil {
		return err
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not save stage before creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("creating stage...")

	if err := stage.driver.Create(); err != nil {
		return errors.Wrap(err, "could not run driver")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not save stage after creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("waiting for stage...")
	if err := stage.waitForState(state.Running); err != nil {
		return err
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("provisioning...")
	if err := stage.Provision(); err != nil {
		return errors.Wrap(err, "could not provision stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("checking connection...")

	dockerURL, err := stage.driver.GetURL()
	if err != nil {
		return errors.Wrap(err, "could get Docker URL")
	}

	parsedURL, err := url.Parse(dockerURL)
	if err != nil {
		return errors.Wrap(err, "could not parse Docker URL")
	}

	if valid, err := auth.ValidateCertificate(parsedURL.Host, filepath.Join(stage.hostDir())); !valid || err != nil {
		return errors.Wrap(err, "invalid certificate")
	}

	if err := stage.exportState(); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil
}

func authOptions(baseDir string) machineauth.Options {
	return machineauth.Options{
		StorePath:        baseDir,
		CertDir:          baseDir,
		CaCertPath:       filepath.Join(baseDir, "ca.pem"),
		CaPrivateKeyPath: filepath.Join(baseDir, "ca-key.pem"),
		ClientCertPath:   filepath.Join(baseDir, "cert.pem"),
		ClientKeyPath:    filepath.Join(baseDir, "key.pem"),
		ServerCertPath:   filepath.Join(baseDir, "server.pem"),
		ServerKeyPath:    filepath.Join(baseDir, "server-key.pem"),
	}
}

func swarmOptions() swarm.Options {
	return swarm.Options{
		Host:     "tcp://0.0.0.0:3376",
		Image:    "swarm:latest",
		Strategy: "spread",
	}
}

func engineOptions() engine.Options {
	return engine.Options{
		InstallURL:       drivers.DefaultEngineInstallURL,
		StorageDriver:    "overlay2",
		TLSVerify:        true,
		ArbitraryFlags:   []string{},
		Env:              []string{},
		InsecureRegistry: []string{},
		Labels:           []string{},
		RegistryMirror:   []string{},
	}
}

func (stage *Stage) driverOptions() (drivers.DriverOptions, error) {
	options := make(map[string]interface{})

	// Somehow, these are expected by the driver
	options["swarm-master"] = false
	options["swarm-host"] = ""
	options["swarm-discovery"] = ""

	for _, flag := range stage.driver.GetCreateFlags() {
		if flag.Default() == nil {
			options[flag.String()] = false
		} else {
			options[flag.String()] = flag.Default()
		}
	}

	for name, option := range stage.config.Options {
		validOption := false
		driverName := stage.driver.DriverName()
		for _, fuzzyName := range fuzzyOptionNames(driverName, name) {
			if _, ok := options[fuzzyName]; ok {
				options[fuzzyName] = option
				validOption = true
			}
		}
		if !validOption {
			return nil, fmt.Errorf("unsupported option for stage type '%s': '%s'", driverName, name)
		}
	}

	return rpcdriver.RPCFlags{Values: options}, nil
}

func fuzzyOptionNames(driver string, name string) []string {
	prefixed := fmt.Sprintf("%s-%s", driver, name)
	return []string{
		name,
		prefixed,
		strings.ReplaceAll(name, "_", "-"),
		strings.ReplaceAll(prefixed, "_", "-"),
	}
}

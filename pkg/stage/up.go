package stage

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Up() error {
	if stage.exists {
		if err := stage.host.Start(); err != nil {
			return errors.Wrap(err, "could not start stage")
		}
	} else {
		stage.host.HostOptions = &host.Options{
			AuthOptions:   authOptions(stage.name),
			EngineOptions: stage.config.EngineOptions(),
			SwarmOptions:  &swarm.Options{},
		}

		driverOpts, err := stage.driverOptions()
		if err != nil {
			return err
		}
		if err = stage.host.Driver.SetConfigFromFlags(driverOpts); err != nil {
			return errors.Wrap(err, "could not configure stage")
		}

		if err = stage.api.Create(stage.host); err != nil {
			return errors.Wrap(err, "could not create stage")
		}
	}

	if err := stage.api.Save(stage.host); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil
}

func authOptions(name string) *auth.Options {
	configDir := filepath.Join(home(), ".docker", "machine")
	certDir := filepath.Join(configDir, "certs")
	machineDir := filepath.Join(configDir, "machines", name)
	return &auth.Options{
		StorePath:        machineDir,
		CertDir:          certDir,
		CaCertPath:       filepath.Join(certDir, "ca.pem"),
		CaPrivateKeyPath: filepath.Join(certDir, "ca-key.pem"),
		ClientCertPath:   filepath.Join(certDir, "cert.pem"),
		ClientKeyPath:    filepath.Join(certDir, "key.pem"),
		ServerCertPath:   filepath.Join(machineDir, "server.pem"),
		ServerKeyPath:    filepath.Join(machineDir, "server-key.pem"),
	}
}

func (stage *Stage) driverOptions() (drivers.DriverOptions, error) {
	options := make(map[string]interface{})

	// Somehow, these are expected by the driver
	options["swarm-master"] = false
	options["swarm-host"] = ""
	options["swarm-discovery"] = ""

	for _, flag := range stage.host.Driver.GetCreateFlags() {
		if flag.Default() == nil {
			options[flag.String()] = false
		} else {
			options[flag.String()] = flag.Default()
		}
	}

	for name, option := range stage.config.Options {
		validOption := false
		driverName := stage.host.Driver.DriverName()
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

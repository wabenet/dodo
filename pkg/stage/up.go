package stage

import (
	"fmt"
	"strings"

	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (stage *Stage) Up() error {
	if stage.exists {
		if err := stage.host.Start(); err != nil {
			return errors.Wrap(err, "could not start stage")
		}
	} else {
		driverOpts, err := stage.driverOptions()
		if err != nil {
			return err
		}
		if err = stage.host.Driver.SetConfigFromFlags(driverOpts); err != nil {
			return errors.Wrap(err, "could not configure stage")
		}

		if err = stage.create(); err != nil {
			return errors.Wrap(err, "could not create stage")
		}
	}

	if err := stage.saveState(); err != nil {
		return errors.Wrap(err, "could not store stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("stage is now up")
	return nil
}

func (stage *Stage) create() error {
	if err := cert.BootstrapCertificates(stage.host.AuthOptions()); err != nil {
		return errors.Wrap(err, "could not generate certificates")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("running pre-create checks...")

	if err := stage.host.Driver.PreCreateCheck(); err != nil {
		return err
	}

	if err := stage.saveState(); err != nil {
		return errors.Wrap(err, "could not save stage before creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("creating stage...")

	if err := stage.host.Driver.Create(); err != nil {
		return errors.Wrap(err, "could not run driver")
	}

	if err := stage.saveState(); err != nil {
		return errors.Wrap(err, "could not save stage after creation")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("waiting for stage...")
	if err := mcnutils.WaitFor(drivers.MachineInState(stage.host.Driver, state.Running)); err != nil {
		return err
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("detecting operating system...")
	provisioner, err := provision.DetectProvisioner(stage.host.Driver)
	if err != nil {
		return errors.Wrap(err, "could not detect operating system")
	}

	log.WithFields(log.Fields{"name": stage.name, "provisioner": provisioner.String()}).Info("provisioning...")
	if err := provisioner.Provision(*stage.host.HostOptions.SwarmOptions, *stage.host.HostOptions.AuthOptions, *stage.host.HostOptions.EngineOptions); err != nil {
		return errors.Wrap(err, "could not provision stage")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("checking connection...")
	if _, _, err = check.DefaultConnChecker.Check(stage.host, false); err != nil {
		return errors.Wrap(err, "could not connect to host")
	}

	log.WithFields(log.Fields{"name": stage.name}).Info("Docker is up and running")
	return nil
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

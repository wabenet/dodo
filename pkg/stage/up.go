package stage

import (
	"path/filepath"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/host"
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
		}

		driverOptions := rpcdriver.RPCFlags{Values: stage.config.Options}
		if err := stage.host.Driver.SetConfigFromFlags(driverOptions); err != nil {
			return errors.Wrap(err, "could not configure stage")
		}

		if err := stage.api.Create(stage.host); err != nil {
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

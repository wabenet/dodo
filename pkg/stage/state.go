package stage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/version"
)

func (stage *Stage) hostDir() string {
	return filepath.Join(stage.stateDir, "machines", stage.name)
}

func (stage *Stage) stateFile() string {
	return filepath.Join(stage.stateDir, "machines", stage.name, "config.json")
}

func (stage *Stage) deleteState() error {
	return os.RemoveAll(stage.hostDir())
}

func (stage *Stage) exportState() error {
	file := stage.stateFile()

	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(stage.dockerMachineHost(), "", "    ")
	if err != nil {
		return err
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return ioutil.WriteFile(file, data, 0600)
	}

	tmpfile, err := ioutil.TempFile(filepath.Dir(file), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if err = ioutil.WriteFile(tmpfile.Name(), data, 0600); err != nil {
		return err
	}

	if err = tmpfile.Close(); err != nil {
		return err
	}

	if err = os.Remove(file); err != nil {
		return err
	}

	return os.Rename(tmpfile.Name(), file)
}

// TODO: fix paths to certificates
func (stage *Stage) dockerMachineHost() *host.Host {
	baseDir := stage.hostDir()
	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          stage.driver.GetMachineName(),
		Driver:        stage.driver,
		DriverName:    stage.driver.DriverName(),
		HostOptions: &host.Options{
			AuthOptions: &auth.Options{
				StorePath:        baseDir,
				CertDir:          baseDir,
				CaCertPath:       filepath.Join(baseDir, "ca.pem"),
				CaPrivateKeyPath: filepath.Join(baseDir, "ca-key.pem"),
				ClientCertPath:   filepath.Join(baseDir, "client.pem"),
				ClientKeyPath:    filepath.Join(baseDir, "client-key.pem"),
				ServerCertPath:   filepath.Join(baseDir, "server.pem"),
				ServerKeyPath:    filepath.Join(baseDir, "server-key.pem"),
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

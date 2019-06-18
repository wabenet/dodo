package stage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/host"
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

	authOptions := authOptions(stage.hostDir())
	swarmOptions := swarmOptions()
	engineOptions := engineOptions()
	machineHost := &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          stage.driver.GetMachineName(),
		Driver:        stage.driver,
		DriverName:    stage.driver.DriverName(),
		HostOptions: &host.Options{
			AuthOptions:   &authOptions,
			EngineOptions: &engineOptions,
			SwarmOptions:  &swarmOptions,
		},
	}

	data, err := json.MarshalIndent(machineHost, "", "    ")
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

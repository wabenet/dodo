package stage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/host"
	"github.com/pkg/errors"
)

func (stage *Stage) hostDir() string {
	return filepath.Join(stage.stateDir, "machines", stage.name)
}

func (stage *Stage) stateFile() string {
	return filepath.Join(stage.stateDir, "machines", stage.name, "config.json")
}

func (stage *Stage) saveState() error {
	file := stage.stateFile()

	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(stage.host, "", "    ")
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

func (stage *Stage) loadState() error {
	file := stage.stateFile()

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.New("stage does not exist")
	}

	stage.host = &host.Host{Name: stage.name}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	migratedHost, migrationPerformed, err := host.MigrateHost(stage.host, data)
	if err != nil {
		return errors.Wrap(err, "could not migrate stage")
	}
	stage.host = migratedHost
	stage.host.Name = stage.name

	if migrationPerformed {
		if err := stage.saveState(); err != nil {
			return errors.Wrap(err, "could not save stage after migration")
		}
	}

	return nil
}

func (stage *Stage) deleteState() error {
	return os.RemoveAll(stage.hostDir())
}

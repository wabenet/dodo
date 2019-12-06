package virtualbox

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	stateFilename = "state.json"
)

type State struct {
	IPAddress      string
	Username       string
	PrivateKeyFile string
}

func (vbox *Stage) loadState() error {
	filename := filepath.Join(vbox.StoragePath, stateFilename)

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			vbox.State = &State{}
			return nil
		} else {
			return err
		}
	}

	stateFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var state State
	if err := json.Unmarshal(stateFile, &state); err != nil {
		return err
	}
	vbox.State = &state

	return nil
}

func (vbox *Stage) saveState() error {
	filename := filepath.Join(vbox.StoragePath, stateFilename)

	if vbox.State == nil {
		return nil
	}

	stateFile, err := json.Marshal(vbox.State)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, stateFile, 0644)
}

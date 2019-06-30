package stage

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	vbox "github.com/oclaussen/dodo/pkg/stage/virtualbox"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// TODO: make machine dir configurable and default somewhere not docker-machine

type Stage struct {
	name     string
	config   *types.Stage
	stateDir string
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

	return stage, nil
}

func (stage *Stage) hostDir() string {
	return filepath.Join(stage.stateDir, "machines", stage.name)
}

func (stage *Stage) waitForStatus(desiredStatus vbox.Status) error {
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		currentStatus, err := vbox.GetStatus(stage.name)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Debug("could not get machine status")
		}
		if currentStatus == desiredStatus {
			return nil
		}
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("maximum number of retries (%d) exceeded", maxAttempts)
}

func (stage *Stage) waitForDocker() error {
	maxAttempts := 60
	reDaemonListening := fmt.Sprintf(":%d\\s+.*:.*", defaultPort)
	cmd := "if ! type netstat 1>/dev/null; then ss -tln; else netstat -tln; fi"

	for i := 0; i < maxAttempts; i++ {
		output, err := stage.RunSSHCommand(cmd)
		if err != nil {
			log.WithFields(log.Fields{"cmd": cmd}).Debug("error running SSH command")
		}

		for _, line := range strings.Split(output, "\n") {
			match, err := regexp.MatchString(reDaemonListening, line)
			if err != nil {
				log.Warnf("Regex warning: %s", err)
			}
			if match && line != "" {
				return nil
			}
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

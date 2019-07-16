package stage

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/oclaussen/dodo/pkg/stage/provider"
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
	client   *plugin.Client
	provider provider.Provider
}

func LoadStage(name string, config *types.Stage) (*Stage, error) {
	stage := &Stage{
		name:     name,
		config:   config,
		stateDir: filepath.Join(home(), ".docker", "machine"),
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  provider.HandshakeConfig("virtualbox"),
		Plugins:          provider.PluginMap,
		Cmd:              exec.Command("./virtualbox"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	stage.client = client

	c, err := client.Client()
	if err != nil {
		return nil, err
	}
	raw, err := c.Dispense("provider")
	if err != nil {
		return nil, err
	}

	stage.provider = raw.(provider.Provider)

	_, err = os.Stat(stage.hostDir())
	if os.IsNotExist(err) {
		stage.exists = false
	} else if err == nil {
		stage.exists = true
	} else {
		return stage, errors.Wrap(err, "could not check if stage exists")
	}

	success, err := stage.provider.Initialize(map[string]string{
		"vmName":      name,
		"storagePath": stage.hostDir(),
	})
	if err != nil || !success {
		return nil, errors.Wrap(err, "initialization failed")
	}

	return stage, nil
}

func (stage *Stage) Save() {
	stage.client.Kill()
}

func (stage *Stage) hostDir() string {
	return filepath.Join(stage.stateDir, "machines", stage.name)
}

func (stage *Stage) waitForStatus(desiredStatus provider.Status) error {
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		currentStatus, err := stage.provider.Status()
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

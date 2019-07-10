package stage

import (
	"path/filepath"

	vbox "github.com/oclaussen/dodo/pkg/stage/virtualbox"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
)

func (stage *Stage) RunSSHCommand(command string) (string, error) {
	opts, err := vbox.GetSSHOptions(stage.name)
	if err != nil {
		return "", err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{filepath.Join(stage.stateDir, "machines", stage.name, "id_rsa")},
		NonInteractive:    true,
	})
	if err != nil {
		return "", nil
	}
	defer executor.Close()

	return executor.Execute(command)
}

func (stage *Stage) SSH() error {
	currentStatus, err := vbox.GetStatus(stage.name)
	if err != nil {
		return err
	}

	if currentStatus != vbox.Running {
		return errors.New("stage is not up")
	}

	opts, err := vbox.GetSSHOptions(stage.name)
	if err != nil {
		return err
	}

	return ssh.GimmeShell(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{filepath.Join(stage.stateDir, "machines", stage.name, "id_rsa")},
		NonInteractive:    true,
	})
}

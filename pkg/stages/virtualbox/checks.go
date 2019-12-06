package virtualbox

import (
	"time"

	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
)

const (
	retryAttempts  = 60
	retrySleeptime = 4
)

func await(test func() (bool, error)) error {
	for attempts := 0; attempts < retryAttempts; attempts++ {
		success, err := test()
		if err != nil {
			return err
		}
		if success {
			return nil
		}
		time.Sleep(retrySleeptime * time.Second)
	}
	return errors.New("max retries reached")
}

func (vbox *Stage) isSSHAvailable() (bool, error) {
	sshOpts, err := vbox.GetSSHOptions()
	if err != nil {
		return false, err
	}
	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              sshOpts.Hostname,
		Port:              sshOpts.Port,
		User:              sshOpts.Username,
		IdentityFileGlobs: []string{sshOpts.PrivateKeyFile},
		NonInteractive:    true,
	})
	if err != nil {
		return false, err
	}
	defer executor.Close()

	_, err = executor.Execute("id")
	return err == nil, nil
}

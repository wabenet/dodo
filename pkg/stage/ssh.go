package stage

import (
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
)

func (stage *Stage) SSH() error {
	available, err := stage.provider.Available()
	if err != nil {
		return err
	}

	if !available {
		return errors.New("stage is not up")
	}

	opts, err := stage.provider.GetSSHOptions()
	if err != nil {
		return err
	}

	return ssh.GimmeShell(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{opts.PrivateKeyFile},
		NonInteractive:    true,
	})
}

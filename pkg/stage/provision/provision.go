// +build !designer
//go:generate env GOOS=linux GOARCH=amd64 go build -tags 'designer' -o assets/stagedesigner
//go:generate go run assets_generate.go

package provision

import (
	"github.com/oclaussen/dodo/pkg/stage"
	"github.com/oclaussen/dodo/pkg/stage/designer"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Provision(sshOpts *stage.SSHOptions, config *designer.Config) error {
	file, err := Assets.Open("/stagedesigner")
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              sshOpts.Hostname,
		Port:              sshOpts.Port,
		User:              sshOpts.Username,
		IdentityFileGlobs: []string{sshOpts.PrivateKeyFile},
		NonInteractive:    true,
	})
	if err != nil {
		return err
	}
	defer executor.Close()

	log.Debug("write stage designer to stage")
	if err := executor.WriteFile(&ssh.FileOptions{
		Path:   "/tmp/stagedesigner",
		Reader: file,
		Size:   stat.Size(),
		Mode:   0755,
	}); err != nil {
		return err
	}

	encoded, err := designer.EncodeConfig(config)
	if err != nil {
		log.Error(err)
		return err
	}
	if err := executor.WriteFile(&ssh.FileOptions{
		Path:    "/tmp/stagedesigner-config",
		Content: string(encoded),
		Mode:    0644,
	}); err != nil {
		return err
	}

	log.WithFields(log.Fields{"config": encoded}).Debug("executing stage designer")
	// TODO: figure out whether we have/need sudo
	if out, err := executor.Execute("sudo /tmp/stagedesigner /tmp/stagedesigner-config"); err != nil {
		return errors.Wrap(err, string(out))
	}

	return nil
}

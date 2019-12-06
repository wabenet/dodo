//go:generate env GOOS=linux GOARCH=amd64 go build -o assets/stage-designer github.com/oclaussen/dodo/cmd/stage-designer
//go:generate go run assets_generate.go

package stage

import (
	"bytes"

	"github.com/oclaussen/dodo/pkg/stagedesigner"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Provision(sshOpts *SSHOptions, config *stagedesigner.Config) (*stagedesigner.ProvisionResult, error) {
	file, err := Assets.Open("/stage-designer")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              sshOpts.Hostname,
		Port:              sshOpts.Port,
		User:              sshOpts.Username,
		IdentityFileGlobs: []string{sshOpts.PrivateKeyFile},
		NonInteractive:    true,
	})
	if err != nil {
		return nil, err
	}
	defer executor.Close()

	log.Debug("write stage designer to stage")
	if err := executor.WriteFile(&ssh.FileOptions{
		Path:   "/tmp/stage-designer",
		Reader: file,
		Size:   stat.Size(),
		Mode:   0755,
	}); err != nil {
		return nil, err
	}

	encoded, err := stagedesigner.EncodeConfig(config)
	if err != nil {
		return nil, err
	}
	if err := executor.WriteFile(&ssh.FileOptions{
		Path: "/tmp/stage-designer-config",
		// TODO: this is a bug in gimme. Fix it!
		//Content: encoded,
		Reader: bytes.NewReader(encoded),
		Size:   int64(len(encoded)),
		Mode:   0644,
	}); err != nil {
		return nil, err
	}

	// TODO: figure out whether we have/need sudo
	log.Debug("executing stage designer")
	out, err := executor.Execute("sudo /tmp/stage-designer /tmp/stage-designer-config")
	if err != nil {
		return nil, errors.Wrap(err, out)
	}

	result, err := stagedesigner.DecodeResult([]byte(out))
	if err != nil {
		return nil, err
	}

	return result, nil
}

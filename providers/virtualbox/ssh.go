package main

import (
	"path/filepath"

	"github.com/oclaussen/dodo/pkg/stage/provider"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
)

func (vbox *VirtualBoxProvider) GetSSHOptions() (*provider.SSHOptions, error) {
	portForwardings, err := ListPortForwardings(vbox.VMName)
	if err != nil {
		return nil, err
	}

	port := 0
	for _, forward := range portForwardings {
		if forward.Name == "ssh" {
			port = forward.HostPort
			break
		}
	}
	if port == 0 {
		return nil, errors.New("no port forwarding matching ssh port found")
	}

	return &provider.SSHOptions{
		Hostname:       "127.0.0.1",
		Port:           port,
		Username:       "docker",
		PrivateKeyFile: filepath.Join(vbox.StoragePath, "id_rsa"),
	}, nil
}

func (vbox *VirtualBoxProvider) ssh(command string) (string, error) {
	opts, err := vbox.GetSSHOptions()
	if err != nil {
		return "", err
	}

	executor, err := ssh.GimmeExecutor(&ssh.Options{
		Host:              opts.Hostname,
		Port:              opts.Port,
		User:              opts.Username,
		IdentityFileGlobs: []string{opts.PrivateKeyFile},
		NonInteractive:    true,
	})
	if err != nil {
		return "", nil
	}
	defer executor.Close()

	return executor.Execute(command)
}

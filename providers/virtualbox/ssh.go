package main

import (
	"github.com/oclaussen/dodo/pkg/stage/provider"
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
		Hostname: "127.0.0.1",
		Port:     port,
		Username: "docker",
	}, nil
}

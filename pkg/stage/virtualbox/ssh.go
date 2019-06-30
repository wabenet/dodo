package virtualbox

import (
	"github.com/pkg/errors"
)

type SSHOptions struct {
	Hostname string
	Port     int
	Username string
}

func GetSSHOptions(name string) (*SSHOptions, error) {
	portForwardings, err := ListPortForwardings(name)
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

	return &SSHOptions{
		Hostname: "127.0.0.1",
		Port:     port,
		Username: "docker",
	}, nil
}

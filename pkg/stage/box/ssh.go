package box

import (
	"path/filepath"

	"github.com/cavaliercoder/grab"
)

const (
	vagrantDefaultUser    = "vagrant"
	vagrantPrivateKeyName = "insecure_private_key"
	vagrantPrivateKeyURL  = "https://raw.githubusercontent.com/hashicorp/vagrant/master/keys/vagrant"
)

type SSHOptions struct {
	Username       string
	PrivateKeyFile string
}

func (box *Box) GetSSHOptions() (*SSHOptions, error) {
	privateKeyFile := filepath.Join(box.storagePath, vagrantPrivateKeyName)

	client := grab.NewClient()
	req, err := grab.NewRequest(privateKeyFile, vagrantPrivateKeyURL)
	if err != nil {
		return nil, err
	}

	resp := client.Do(req)
	if err := resp.Err(); err != nil {
		return nil, err
	}

	// TODO: support boxes with custom users and keys
	return &SSHOptions{
		Username:       vagrantDefaultUser,
		PrivateKeyFile: privateKeyFile,
	}, nil
}

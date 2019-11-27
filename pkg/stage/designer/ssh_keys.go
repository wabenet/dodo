package designer

import (
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
)

const (
	defaultUser        = "vagrant"
	authorizedKeysFile = "/home/vagrant/.ssh/authorized_keys"
)

func ConfigureSSHKeys(keys []string) error {
	u, err := user.Lookup(defaultUser)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}

	content := strings.Join(keys, "\n")
	if err := ioutil.WriteFile(authorizedKeysFile, []byte(content), 0600); err != nil {
		return err
	}

	if err := os.Chown(authorizedKeysFile, uid, gid); err != nil {
		return err
	}
	return nil
}

package designer

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

func ConfigureSSHKeys(config *Config) error {
	u, err := user.Lookup(config.DefaultUser)
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

	file := filepath.Join(u.HomeDir, ".ssh", "authorized_keys")
	content := strings.Join(config.AuthorizedSSHKeys, "\n")
	if err := ioutil.WriteFile(file, []byte(content), 0600); err != nil {
		return err
	}
	if err := os.Chown(file, uid, gid); err != nil {
		return err
	}
	return nil
}

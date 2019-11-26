package virtualbox

import (
	"os"

	"github.com/pkg/errors"
)

func (vm *VM) ConfigureSharedFolders(folders map[string]string) error {
	if _, err := vbm(
		"guestproperty", "set", vm.Name,
		"/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/",
	); err != nil {
		return errors.Wrap(err, "could not set mount prefx")
	}
	if _, err := vbm(
		"guestproperty", "set", vm.Name,
		"/VirtualBox/GuestAdd/SharedFolders/MountDir", "/",
	); err != nil {
		return errors.Wrap(err, "could not set mount dir")
	}

	for shareName, shareDir := range folders {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if os.IsNotExist(err) {
			continue
		}

		if _, err := vbm(
			"sharedfolder", "add", vm.Name,
			"--name", shareName,
			"--hostpath", shareDir,
			"--automount",
		); err != nil {
			return errors.Wrap(err, "could not mount shared folder")
		}

		if _, err := vbm(
			"setextradata", vm.Name,
			"VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1",
		); err != nil {
			return errors.Wrap(err, "could not set shared folder extra data")
		}
	}
	return nil
}

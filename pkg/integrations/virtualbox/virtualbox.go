package virtualbox

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type VM struct {
	Name string
}

func (vm *VM) Import(args ...string) error {
	_, err := vbm(append([]string{"import"}, args...)...)
	return err
}

func (vm *VM) Modify(args ...string) error {
	_, err := vbm(append([]string{"modifyvm", vm.Name}, args...)...)
	return err
}

func (vm *VM) Start() error {
	_, err := vbm("startvm", vm.Name, "--type", "headless")
	return err
}

func (vm *VM) Stop(force bool) error {
	if force {
		_, err := vbm("controlvm", vm.Name, "poweroff")
		return err
	} else {
		_, err := vbm("controlvm", vm.Name, "acpipowerbutton")
		return err
	}
}

func (vm *VM) Delete() error {
	_, err := vbm("unregistervm", "--delete", vm.Name)
	return err
}

func vbm(args ...string) (string, error) {
	return vbmRetry(5, args...)
}

func vbmRetry(retry int, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	vboxManage, err := exec.LookPath("VBoxManage")
	if err != nil {
		return "", errors.New("could not find VBoxManage command")
	}

	cmd := exec.Command(vboxManage, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}

	if retry > 1 {
		if strings.Contains(stderr.String(), "error: The object is not ready") {
			time.Sleep(100 * time.Millisecond)
			return vbmRetry(retry-1, args...)
		}
	}

	return stdout.String(), nil
}

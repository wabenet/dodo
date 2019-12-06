package stagedesigner

import (
	"io/ioutil"
	"os/exec"
)

func ConfigureHostname(name string) error {
	if err := ioutil.WriteFile("/etc/hostname", []byte(name), 0644); err != nil {
		return err
	}
	if hostnamectl, err := exec.LookPath("hostnamectl"); err == nil {
		return exec.Command(hostnamectl, "set-hostname", name).Run()
	} else if hostname, err := exec.LookPath("hostname"); err == nil {
		return exec.Command(hostname, name).Run()
	} else {
		// TODO what to do?
		return nil
	}
}

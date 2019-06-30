package virtualbox

import (
	"regexp"
)

type Status int

const (
	None Status = iota
	Running
	Paused
	Saved
	Stopped
	Stopping
	Starting
	Error
	Timeout
)

func GetStatus(name string) (Status, error) {
	stdout, err := vbm("showvminfo", name, "--machinereadable")
	if err != nil {
		return Error, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return None, nil
	}
	switch groups[1] {
	case "running":
		return Running, nil
	case "paused":
		return Paused, nil
	case "saved":
		return Saved, nil
	case "poweroff", "aborted":
		return Stopped, nil
	}
	return None, nil
}

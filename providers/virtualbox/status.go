package main

import (
	"regexp"

	"github.com/oclaussen/dodo/pkg/stage/provider"
)

func (vbox *VirtualBoxProvider) Status() (provider.Status, error) {
	stdout, err := vbm("showvminfo", vbox.VMName, "--machinereadable")
	if err != nil {
		return provider.Down, nil
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return provider.Unknown, nil
	}
	switch groups[1] {
	case "running":
		return provider.Up, nil
	case "paused", "saved", "poweroff", "aborted":
		return provider.Paused, nil
	default:
		return provider.Unknown, nil
	}
}

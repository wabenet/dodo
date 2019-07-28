package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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
		return "", err
	}

	if retry > 1 {
		if strings.Contains(stderr.String(), "error: The object is not ready") {
			time.Sleep(100 * time.Millisecond)
			return vbmRetry(retry-1, args...)
		}
	}

	return stdout.String(), nil
}

func checkVBoxManageVersion() error {
	version, err := vbm("--version")
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid version: %q", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid version: %q", version)
	}

	if major < 5 {
		return fmt.Errorf("VirtualBox version %q is not supported", version)
	}

	return nil
}

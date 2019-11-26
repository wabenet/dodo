package virtualbox

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type VersionInfo struct {
	Major   int
	Minor   int
	Release string
}

func Version() (VersionInfo, error) {
	info := VersionInfo{}

	out, err := vbm("--version")
	if err != nil {
		return info, err
	}

	parts := strings.Split(strings.TrimSpace(out), ".")
	switch len(parts) {
	case 2:
		info.Major, err = strconv.Atoi(parts[0])
		if err != nil {
			return info, fmt.Errorf("invalid version: %q", out)
		}
		info.Release = parts[1]
	case 3:
		info.Major, err = strconv.Atoi(parts[0])
		if err != nil {
			return info, fmt.Errorf("invalid version: %q", out)
		}
		info.Minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return info, fmt.Errorf("invalid version: %q", out)
		}
		info.Release = parts[1]
	default:
		return info, fmt.Errorf("invalid version: %q", out)
	}

	return info, nil
}

func (vm *VM) Info() (map[string]string, error) {
	result := map[string]string{}

	out, err := vbm("showvminfo", vm.Name, "--machinereadable")
	if err != nil {
		return result, err
	}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		key := removeQuotes(parts[0])
		value := removeQuotes(parts[1])
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func removeQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

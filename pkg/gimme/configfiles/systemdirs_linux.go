// +build !windows,!darwin

package configfiles

import (
	"os/user"
)

func getUserDirectories(name string) []string {
	user, err := user.Current()
	if err != nil {
		return []string{}
	}
	if user.HomeDir == "" {
		return []string{}
	}
	return []string{user.HomeDir}
}

func getSystemDirectories() []string {
	return []string{"/etc"}
}

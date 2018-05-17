// +build darwin

package config

import (
	"os"
	"path/filepath"
)

var XDGDefaultDir = filepath.Join("/", "etc", "xdg")
var SpecialSystemDirectories = []string{
	filepath.Join("/", "etc"),
	filepath.Join("/", "Library", "Application Support"),
	filepath.Join(os.Getenv("HOME"), "Library", "Application Support"),
}

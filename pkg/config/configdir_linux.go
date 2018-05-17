// +build !windows,!darwin

package config

import "path/filepath"

var XDGDefaultDir = filepath.Join("/", "etc", "xdg")
var SpecialSystemDirectories = []string{
	filepath.Join("/", "etc"),
}

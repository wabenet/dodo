// +build darwin

package configfiles

import (
	"os"
	"path/filepath"
)

var xdgDefaultDir = filepath.Join("/", "etc", "xdg")
var specialSystemDirectories = []string{
	filepath.Join("/", "etc"),
	filepath.Join("/", "Library", "Application Support"),
	filepath.Join(os.Getenv("HOME"), "Library", "Application Support"),
}

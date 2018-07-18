// +build !windows,!darwin

package configfiles

import "path/filepath"

var xdgDefaultDir = filepath.Join("/", "etc", "xdg")
var specialSystemDirectories = []string{
	filepath.Join("/", "etc"),
}

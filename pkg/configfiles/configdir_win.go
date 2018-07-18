// +build windows

package configfiles

import "os"

var xdgDefaultDir = ""
var specialSystemDirectories = []string{
	os.Getenv("PROGRAMDATA"),
	os.Getenv("APPDATA"),
}

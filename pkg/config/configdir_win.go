// +build windows

package config

import "os"

var XDGDefaultDir = ""
var SpecialSystemDirectories = []string{
	os.Getenv("PROGRAMDATA"),
	os.Getenv("APPDATA"),
}

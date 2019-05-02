// +build windows

package configfiles

import "os"

const (
	envProgramData = "PROGRAMDATA"
	envAppData     = "APPDATA"
)

func getSystemDirectories(_ string) ([]string, error) {
	var directories []string
	if programData := os.Getenv(envProgramData); programData != "" {
		direcories = append(directories, programData)
	}
	if appData := os.Getenv(envAppData); appData != "" {
		directories = append(directories, appData)
	}
	return directories
}

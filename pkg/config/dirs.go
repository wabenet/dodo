package config

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	systemAppDir = "/var/lib/dodo"
)

func GetAppDir() string {
	dir := filepath.FromSlash(systemAppDir)
	if user, err := user.Current(); err == nil && user.HomeDir != "" {
		dir = filepath.Join(user.HomeDir, ".dodo")
	}
	os.MkdirAll(dir, 0700)
	return dir
}

func getDir(subdir string) string {
	dir := filepath.Join(GetAppDir(), subdir)
	os.MkdirAll(dir, 0700)
	return dir
}

func GetPluginDir() string {
	return getDir("plugins")
}

func GetTmpDir() string {
	return getDir("tmp")
}

func GetStagesDir() string {
	return getDir("stages")
}

func GetBoxesDir() string {
	return getDir("boxes")
}

package image

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/urlutil"
	"github.com/moby/buildkit/session/filesync"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
	fstypes "github.com/tonistiigi/fsutil/types"
)

type contextData struct {
	remote         string
	dockerfileName string
	contextDir     string
}

func (data *contextData) tempdir() (string, error) {
	if len(data.contextDir) == 0 {
		dir, err := ioutil.TempDir("", "dodo-temp-")
		if err != nil {
			return "", err
		}
		data.contextDir = dir
	}
	return data.contextDir, nil
}

func (data *contextData) cleanup() {
	if data.contextDir != "" {
		os.RemoveAll(data.contextDir)
	}
}

func prepareContext(config *types.Image, session session) (*contextData, error) {
	data := contextData{
		remote:         "",
		dockerfileName: config.Dockerfile,
	}
	syncedDirs := []filesync.SyncedDir{}

	if config.Context == "" {
		data.remote = "client-session"
		dir, err := data.tempdir()
		if err != nil {
			data.cleanup()
			return nil, err
		}
		syncedDirs = append(syncedDirs, filesync.SyncedDir{Name: "context", Dir: dir})

	} else if _, err := os.Stat(config.Context); err == nil {
		data.remote = "client-session"
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "context",
			Dir:  config.Context,
			Map: func(stat *fstypes.Stat) bool {
				stat.Uid = 0
				stat.Gid = 0
				return true
			},
		})

	} else if urlutil.IsURL(config.Context) {
		data.remote = config.Context

	} else {
		return nil, errors.Errorf("Context directory does not exist: %v", config.Context)
	}

	if len(config.Steps) > 0 {
		steps := ""
		for _, step := range config.Steps {
			steps = steps + "\n" + step
		}

		dir, err := data.tempdir()
		if err != nil {
			data.cleanup()
			return nil, err
		}
		tempfile := filepath.Join(dir, "Dockerfile")
		if err := writeDockerfile(tempfile, steps); err != nil {
			data.cleanup()
			return nil, err
		}

		data.dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	} else if config.Dockerfile != "" && data.remote == "client-session" {
		data.dockerfileName = filepath.Base(config.Dockerfile)
		dockerfileDir := filepath.Dir(config.Dockerfile)
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	} else if config.Name != "" && data.remote == "client-session" {
		dir, err := data.tempdir()
		if err != nil {
			data.cleanup()
			return nil, err
		}
		tempfile := filepath.Join(dir, "Dockerfile")
		if err := writeDockerfile(tempfile, fmt.Sprintf("FROM %s", config.Name)); err != nil {
			data.cleanup()
			return nil, err
		}
		data.dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	}

	if len(syncedDirs) > 0 {
		session.Allow(filesync.NewFSSyncProvider(syncedDirs))
	}

	return &data, nil
}

func writeDockerfile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	rc := ioutil.NopCloser(bytes.NewReader([]byte(content)))
	_, err = io.Copy(file, rc)
	if err != nil {
		return err
	}

	err = rc.Close()
	if err != nil {
		return err
	}

	return nil
}

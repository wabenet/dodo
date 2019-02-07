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
	"github.com/pkg/errors"
	fstypes "github.com/tonistiigi/fsutil/types"
)

type contextData struct {
	remote         string
	dockerfileName string
	cleanup        func()
}

func prepareContext(config *ImageConfig, session session) (*contextData, error) {
	data := contextData{
		remote:         "",
		dockerfileName: config.Dockerfile,
		cleanup:        func() {},
	}
	syncedDirs := []filesync.SyncedDir{}

	if config.Context == "" {
		data.remote = "client-session"

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

		tempfile, err := writeDockerfile("Dockerfile", steps)
		if err != nil {
			return nil, err
		}

		data.dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		data.cleanup = func() { os.RemoveAll(dockerfileDir) }
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
		tempfile, err := writeDockerfile("Dockerfile", fmt.Sprintf("FROM %s", config.Name))
		if err != nil {
			return nil, err
		}
		data.dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		data.cleanup = func() { os.RemoveAll(dockerfileDir) }
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

func writeDockerfile(filename string, content string) (dockerfile string, err error) {
	tempdir, err := ioutil.TempDir("", "dodo-temp-dockerfile-")
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempdir)
		}
	}()

	dockerfile = filepath.Join(tempdir, filename)
	file, err := os.Create(dockerfile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	rc := ioutil.NopCloser(bytes.NewReader([]byte(content)))
	_, err = io.Copy(file, rc)
	if err != nil {
		return "", err
	}

	err = rc.Close()
	if err != nil {
		return "", err
	}

	return dockerfile, nil
}

package image

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/urlutil"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/filesync"
	"github.com/pkg/errors"
	fstypes "github.com/tonistiigi/fsutil/types"
)

func prepareContext(
	context string, dockerfile string, steps string, name string, session *session.Session,
) (string, string, func(), error) {
	var (
		remote         string
		dockerfileName = dockerfile
		syncedDirs     = []filesync.SyncedDir{}
		cleanup        = func() {}
	)

	if context == "" {
		remote = "client-session"

	} else if _, err := os.Stat(context); err == nil {
		remote = "client-session"
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "context",
			Dir:  context,
			Map: func(stat *fstypes.Stat) bool {
				stat.Uid = 0
				stat.Gid = 0
				return true
			},
		})

	} else if urlutil.IsURL(context) {
		remote = context

	} else {
		return "", "", nil, errors.Errorf("Context directory does not exist: %v", context)
	}

	if steps != "" {
		tempfile, err := writeDockerfile("Dockerfile", steps)
		if err != nil {
			return "", "", nil, err
		}
		dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		cleanup = func() { os.RemoveAll(dockerfileDir) }
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	} else if dockerfile != "" && remote == "client-session" {
		dockerfileName = filepath.Base(dockerfile)
		dockerfileDir := filepath.Dir(dockerfile)
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	} else if name != "" && remote == "client-session" {
		tempfile, err := writeDockerfile("Dockerfile", fmt.Sprintf("FROM %s", name))
		if err != nil {
			return "", "", nil, err
		}
		dockerfileName = filepath.Base(tempfile)
		dockerfileDir := filepath.Dir(tempfile)
		cleanup = func() { os.RemoveAll(dockerfileDir) }
		syncedDirs = append(syncedDirs, filesync.SyncedDir{
			Name: "dockerfile",
			Dir:  dockerfileDir,
		})

	}

	if len(syncedDirs) > 0 {
		session.Allow(filesync.NewFSSyncProvider(syncedDirs))
	}

	return remote, dockerfileName, cleanup, nil
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

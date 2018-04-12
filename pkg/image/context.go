package image

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/idtools"
	log "github.com/sirupsen/logrus"
)

// TODO: validate that all files in the context are ok
func getContext(options Options, dockerfile string) (io.ReadCloser, error) {
	contextDir := options.Build.Context
	if contextDir == "" {
		contextDir = "."
	}
	contextDir, err := filepath.Abs(contextDir)
	if err != nil {
		return nil, err
	}
	contextDir, err = filepath.EvalSymlinks(contextDir)
	if err != nil {
		return nil, err
	}
	stat, err := os.Lstat(contextDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf(
			"context must be a directory: %s", contextDir)
	}
	excludes, err := getDockerignore(contextDir, dockerfile)
	if err != nil {
		return nil, err
	}
	dockerfileStream, err := getDockerfile(options, contextDir)
	if err != nil {
		return nil, err
	}
	tarStream, err := archive.TarWithOptions(
		contextDir,
		&archive.TarOptions{
			ExcludePatterns: excludes,
			ChownOpts:       &idtools.IDPair{UID: 0, GID: 0},
		},
	)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	header := &tar.Header{
		Mode:       0600,
		Uid:        0,
		Gid:        0,
		ModTime:    now,
		Typeflag:   tar.TypeReg,
		AccessTime: now,
		ChangeTime: now,
	}
	dockerfileFunc := func(
		_ string, _ *tar.Header, _ io.Reader,
	) (*tar.Header, []byte, error) {
		return header, dockerfileStream, nil
	}
	tarStream = archive.ReplaceFileTarWrapper(
		tarStream,
		map[string]archive.TarModifierFunc{
			dockerfile: dockerfileFunc,
		},
	)
	return tarStream, nil
}

func getDockerignore(contextDir string, dockerfile string) ([]string, error) {
	file, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error(closeErr)
		}
	}()

	excludes, err := dockerignore.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if keep, _ := fileutils.Matches(".dockerignore", excludes); keep {
		excludes = append(excludes, "!.dockerignore")
	}
	if keep, _ := fileutils.Matches(dockerfile, excludes); keep {
		excludes = append(excludes, "!"+dockerfile)
	}

	return excludes, nil
}

func getDockerfile(options Options, contextDir string) ([]byte, error) {
	dockerfile := options.Build.Dockerfile
	steps := ""
	for _, step := range options.Build.Steps {
		steps = steps + "\n" + step
	}

	if dockerfile == "" && len(steps) > 0 {
		return []byte(steps), nil
	}

	if dockerfile == "" {
		dockerfile = filepath.Join(contextDir, "Dockerfile")
	}
	if !filepath.IsAbs(dockerfile) {
		dockerfile = filepath.Join(contextDir, dockerfile)
	}
	dockerfile, err := filepath.Abs(dockerfile)
	if err != nil {
		return nil, err
	}
	dockerfile, err = filepath.EvalSymlinks(dockerfile)
	if err != nil {
		return nil, err
	}
	_, err = os.Lstat(dockerfile)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(dockerfile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); err != nil {
			log.Error(closeErr)
		}
	}()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return append(bytes, []byte(steps)...), nil
}

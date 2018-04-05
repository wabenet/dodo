package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/oclaussen/dodo/pkg/config"
	"golang.org/x/net/context"
)

func buildImage(ctx context.Context, client *client.Client, config *config.BackdropConfig) (string, error) {
	if config.Image != "" && !config.Build.ForceRebuild {
		if image := useExistingImage(ctx, client, config); image != "" {
			return config.Image, nil
		}
	}

	args := map[string]*string{}
	for _, arg := range config.Build.Args {
		switch values := strings.SplitN(arg, "=", 2); len(values) {
		case 1:
			args[values[0]] = nil
		case 2:
			args[values[0]] = &values[1]
		}
	}

	contextDir, err := getContextDir(config.Build.Context)
	if err != nil {
		return "", err
	}
	dockerfile, err := getDockerfile(config.Build.Dockerfile, contextDir)
	if err != nil {
		return "", err
	}
	excludes, err := getDockerignore(contextDir, dockerfile)
	if err != nil {
		return "", err
	}

	// TODO: validate that all files in the context are ok
	tarStream, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.IDPair{UID: 0, GID: 0},
	})
	if err != nil {
		return "", err
	}

	response, err := client.ImageBuild(
		ctx,
		tarStream,
		types.ImageBuildOptions{
			SuppressOutput: false, // TODO: quiet mode
			NoCache:        config.Build.NoCache,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     config.Pull,
			Dockerfile:     config.Build.Dockerfile,
			BuildArgs:      args,
		},
	)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	image := ""
	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		// TODO: handle parse error
		if err := json.Unmarshal(*auxJSON, &result); err == nil {
			image = result.ID
		}
	}

	outFd, isTerminal := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, outFd, isTerminal, aux)
	if err != nil {
		return "", err
	}
	if image == "" {
		return "", errors.New("Build complete, but the server did not send an image id.")
	}
	return image, nil
}

func getContextDir(givenContext string) (string, error) {
	contextDir := givenContext
	if contextDir == "" {
		contextDir = "."
	}
	contextDir, err := filepath.Abs(contextDir)
	if err != nil {
		return "", err
	}
	contextDir, err = filepath.EvalSymlinks(contextDir)
	if err != nil {
		return "", err
	}
	stat, err := os.Lstat(contextDir)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("context must be a directory: %s", contextDir)
	}
	return contextDir, nil
}

func getDockerfile(givenDockerfile string, contextDir string) (string, error) {
	dockerfile := givenDockerfile
	if dockerfile == "" {
		dockerfile = filepath.Join(contextDir, "Dockerfile")
	}
	if !filepath.IsAbs(dockerfile) {
		dockerfile = filepath.Join(contextDir, dockerfile)
	}
	dockerfile, err := filepath.EvalSymlinks(dockerfile)
	if err != nil {
		return "", err
	}
	_, err = os.Lstat(dockerfile)
	if err != nil {
		return "", err
	}
	dockerfile, err = filepath.Rel(contextDir, dockerfile)
	if err != nil {
		return "", err
	}
	dockerfile, err = archive.CanonicalTarNameForPath(dockerfile)
	if err != nil {
		return "", err
	}
	return dockerfile, nil
}

func getDockerignore(contextDir string, dockerfile string) ([]string, error) {
	file, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

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

package image

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/pkg/stringid"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func buildToDo(options Options) (string, error) {
	if versions.LessThan(options.Client.ClientVersion(), "1.31") {
		return "", errors.Errorf("buildkit not supported by daemon")
	}

	args := map[string]*string{}
	for _, arg := range options.Args {
		switch values := strings.SplitN(arg, "=", 2); len(values) {
		case 1:
			args[values[0]] = nil
		case 2:
			args[values[0]] = &values[1]
		}
	}

	var tags []string
	if options.Name != "" {
		tags = append(tags, options.Name)
	}

	ctx := appcontext.Context()

	sessionID, err := readOrCreateSessionID()
	if err != nil {
		return "", err
	}
	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", sessionID, options.Context)))
	sharedKey := hex.EncodeToString(s[:])

	session, err := session.NewSession(context.Background(), filepath.Base(options.Context), sharedKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}

	steps := ""
	for _, step := range options.Steps {
		steps = steps + "\n" + step
	}

	remote, dockerfile, cleanup, err := prepareContext(options.Context, options.Dockerfile, steps, session)
	if err != nil {
		return "", err
	}
	defer cleanup()

	buildID := stringid.GenerateRandomID()
	imageID := ""

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return session.Run(context.TODO(), options.Client.DialSession)
	})

	eg.Go(func() error {
		defer func() {
			session.Close()
		}()

		response, err := options.Client.ImageBuild(
			context.Background(),
			nil,
			types.ImageBuildOptions{
				Tags:           tags,
				SuppressOutput: false,
				NoCache:        options.NoCache,
				Remove:         true,
				ForceRemove:    true,
				PullParent:     options.ForcePull,
				Dockerfile:     dockerfile,
				BuildArgs:      args,
				AuthConfigs:    options.AuthConfigs,
				Version:        types.BuilderBuildKit,
				RemoteContext:  remote,
				SessionID:      session.ID(),
				BuildID:        buildID,
			},
		)
		if err != nil {
			return err
		}
		imageID, err = handleImageResult(response.Body, true)
		return err
	})

	return imageID, eg.Wait()
}

func readOrCreateSessionID() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	sessionDir := filepath.Join(user.HomeDir, ".dodo")
	sessionFile := filepath.Join(sessionDir, "sessionID")

	err = os.MkdirAll(sessionDir, 0700)
	if err != nil {
		return "", err
	}

	if _, err = os.Lstat(sessionFile); err == nil {
		sessionID, err := ioutil.ReadFile(sessionFile)
		if err != nil {
			return "", err
		}
		return string(sessionID), nil
	}

	sessionID := make([]byte, 32)
	_, err = rand.Read(sessionID)
	if err != nil {
		return "", err
	}
	sessionID = []byte(hex.EncodeToString(sessionID))
	err = ioutil.WriteFile(sessionFile, sessionID, 0600)
	if err != nil {
		return "", err
	}

	return string(sessionID), nil
}

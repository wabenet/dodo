package image

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/containerd/console"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stringid"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// Build produces a runnable image and returns an image id.
func (image *image) Build() (string, error) {
	args := map[string]*string{}
	for _, arg := range image.config.Args.Strings() {
		switch values := strings.SplitN(arg, "=", 2); len(values) {
		case 1:
			args[values[0]] = nil
		case 2:
			args[values[0]] = &values[1]
		}
	}

	var tags []string
	if image.config.Name != "" {
		tags = append(tags, image.config.Name)
	}

	var contextDir string
	if image.config.Context == "" {
		contextDir = "."
	} else {
		contextDir = image.config.Context
	}

	ctx := appcontext.Context()

	sessionID, err := readOrCreateSessionID()
	if err != nil {
		return "", err
	}
	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", sessionID, contextDir)))
	sharedKey := hex.EncodeToString(s[:])

	session, err := session.NewSession(context.Background(), filepath.Base(contextDir), sharedKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}

	steps := ""
	for _, step := range image.config.Steps {
		steps = steps + "\n" + step
	}

	remote, dockerfile, cleanup, err := prepareContext(contextDir, image.config.Dockerfile, steps, image.config.Name, session)
	if err != nil {
		return "", err
	}
	defer cleanup()

	buildID := stringid.GenerateRandomID()
	imageID := ""

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return session.Run(context.TODO(), image.client.DialSession)
	})

	eg.Go(func() error {
		defer func() {
			session.Close()
		}()

		response, err := image.client.ImageBuild(
			context.Background(),
			nil,
			types.ImageBuildOptions{
				Tags:           tags,
				SuppressOutput: false,
				NoCache:        image.config.NoCache,
				Remove:         true,
				ForceRemove:    true,
				PullParent:     image.config.ForcePull,
				Dockerfile:     dockerfile,
				BuildArgs:      args,
				AuthConfigs:    image.authConfigs,
				Version:        types.BuilderBuildKit,
				RemoteContext:  remote,
				SessionID:      session.ID(),
				BuildID:        buildID,
			},
		)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		displayCh := make(chan *client.SolveStatus)
		defer close(displayCh)

		cons, err := console.ConsoleFromFile(os.Stdout)
		if err != nil {
			return err
		}

		eg.Go(func() error {
			return progressui.DisplaySolveStatus(context.TODO(), "", cons, os.Stdout, displayCh)
		})

		decoder := json.NewDecoder(response.Body)
		for {
			var msg jsonmessage.JSONMessage
			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			if msg.Error != nil {
				return msg.Error
			}

			if msg.Aux != nil {
				if msg.ID == "moby.image.id" {
					var result types.BuildResult
					if err := json.Unmarshal(*msg.Aux, &result); err != nil {
						continue
					}
					imageID = result.ID
				} else if msg.ID == "moby.buildkit.trace" {
					var resp controlapi.StatusResponse
					var dt []byte
					if err := json.Unmarshal(*msg.Aux, &dt); err != nil {
						continue
					}
					if err := (&resp).Unmarshal(dt); err != nil {
						continue
					}

					s := client.SolveStatus{}
					for _, v := range resp.Vertexes {
						s.Vertexes = append(s.Vertexes, &client.Vertex{
							Digest:    v.Digest,
							Inputs:    v.Inputs,
							Name:      v.Name,
							Started:   v.Started,
							Completed: v.Completed,
							Error:     v.Error,
							Cached:    v.Cached,
						})
					}
					for _, v := range resp.Statuses {
						s.Statuses = append(s.Statuses, &client.VertexStatus{
							ID:        v.ID,
							Vertex:    v.Vertex,
							Name:      v.Name,
							Total:     v.Total,
							Current:   v.Current,
							Timestamp: v.Timestamp,
							Started:   v.Started,
							Completed: v.Completed,
						})
					}
					for _, v := range resp.Logs {
						s.Logs = append(s.Logs, &client.VertexLog{
							Vertex:    v.Vertex,
							Stream:    int(v.Stream),
							Data:      v.Msg,
							Timestamp: v.Timestamp,
						})
					}

					displayCh <- &s
				}
			}
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return "", err
	}

	if imageID == "" {
		return "", errMissingImageID
	}

	return imageID, nil
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

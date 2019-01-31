package image

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/containerd/console"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/versions"
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

var (
	errMissingImageID = errors.New(
		"build complete, but the server did not send an image id")
)

// Options represents the configuration for a docker image that can be
// either built or pulled.
type Options struct {
	Client      Client
	AuthConfigs map[string]types.AuthConfig
	Name        string
	ForcePull   bool
	ForceBuild  bool
	NoCache     bool
	Context     string
	Dockerfile  string
	Steps       []string
	Args        []string
}

// Client represents a docker client that can do everything this package
// needs
type Client interface {
	ClientVersion() string
	Ping(context.Context) (types.Ping, error)
	DialSession(context.Context, string, map[string][]string) (net.Conn, error)
	ImageBuild(context.Context, io.Reader, types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

// Get gets a valid image id, and builds or pulls the image if necessary.
func Get(options Options) (string, error) {
	if options.Client == nil {
		return "", errors.New("client may not be nil")
	}
	ping, err := options.Client.Ping(context.Background())
	if err != nil {
		return "", err
	}
	if !ping.Experimental || versions.LessThan(options.Client.ClientVersion(), "1.31") {
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

	var contextDir string
	if options.Context == "" {
		contextDir = "."
	} else {
		contextDir = options.Context
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
	for _, step := range options.Steps {
		steps = steps + "\n" + step
	}

	remote, dockerfile, cleanup, err := prepareContext(contextDir, options.Dockerfile, steps, options.Name, session)
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

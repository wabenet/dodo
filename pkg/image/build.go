package image

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/containerd/console"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stringid"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/moby/buildkit/util/progress/progressui"
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

	if image.config.Context == "" {
		image.config.Context = "."
	}

	ctx := appcontext.Context()

	session, err := prepareSession(image.config.Context)
	if err != nil {
		return "", err
	}

	contextData, err := prepareContext(image.config, session)
	if err != nil {
		return "", err
	}
	defer contextData.cleanup()

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
				Dockerfile:     contextData.dockerfileName,
				BuildArgs:      args,
				AuthConfigs:    image.authConfigs,
				Version:        types.BuilderBuildKit,
				RemoteContext:  contextData.remote,
				SessionID:      session.ID(),
				BuildID:        stringid.GenerateRandomID(),
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

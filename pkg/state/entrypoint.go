package state

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// EnsureEntrypoint makes sure the entrypoint script ist uploaded to the
// container.
func (state *State) EnsureEntrypoint(ctx context.Context) error {
	config := state.Config
	client, err := state.EnsureClient()
	if err != nil {
		return err
	}
	container, err := state.EnsureContainer(ctx)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		err := client.CopyToContainer(
			ctx,
			container,
			"/",
			reader,
			types.CopyToContainerOptions{},
		)
		if err != nil {
			log.Error(err)
		}
	}()

	tarWriter := tar.NewWriter(writer)
	err = tarWriter.WriteHeader(&tar.Header{
		Name: state.Entrypoint,
		Mode: 0600,
		Size: int64(len(config.Script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(state.Config.Script))
	if err != nil {
		return err
	}
	err = tarWriter.Close()
	if err != nil {
		return err
	}
	return writer.Close()
}

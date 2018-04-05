package state

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func (state *State) EnsureEntrypoint(ctx context.Context) error {
	config := state.Config
	client, err := state.EnsureClient(ctx)
	if err != nil {
		return err
	}
	container, err := state.EnsureContainer(ctx)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	// TODO: handle errors
	go client.CopyToContainer(
		ctx,
		container,
		"/",
		reader,
		types.CopyToContainerOptions{},
	)

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
	err = writer.Close()
	if err != nil {
		return err
	}

	return nil
}

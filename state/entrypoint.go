package state

import (
	"archive/tar"
	"io"

	docker "github.com/fsouza/go-dockerclient"
)

func (state *state) ensureEntrypoint() error {
	if err := state.ensureConfig(); err != nil {
		return err
	}
	if err := state.ensureContainer(); err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	// TODO: handle errors
	go state.Client.UploadToContainer(state.Container.ID, docker.UploadToContainerOptions{
		InputStream:  reader,
		Path:         "/",
	})

	tarWriter := tar.NewWriter(writer)
	err := tarWriter.WriteHeader(&tar.Header{
		Name: state.Entrypoint,
		Mode: 0600,
		Size: int64(len(state.Config.Script)),
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

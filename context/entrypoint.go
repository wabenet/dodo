package context

import (
	"archive/tar"
	"io"

	docker "github.com/fsouza/go-dockerclient"
)

func (context *Context) ensureEntrypoint() error {
	if err := context.ensureConfig(); err != nil {
		return err
	}
	if err := context.ensureContainer(); err != nil {
		return err
	}

	reader, writer := io.Pipe()
	defer reader.Close()

	// TODO: handle errors
	go context.Client.UploadToContainer(context.Container.ID, docker.UploadToContainerOptions{
		InputStream:  reader,
		Path:         "/",
	})

	tarWriter := tar.NewWriter(writer)
	err := tarWriter.WriteHeader(&tar.Header{
		Name: context.Entrypoint,
		Mode: 0600,
		Size: int64(len(context.Config.Script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(context.Config.Script))
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

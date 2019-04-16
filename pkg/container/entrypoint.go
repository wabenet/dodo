package container

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
)

func (container *Container) uploadEntrypoint(containerID string) error {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	go container.client.CopyToContainer(
		container.context,
		containerID,
		"/",
		reader,
		types.CopyToContainerOptions{},
	)

	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	script := container.config.Script + "\n"

	err := tarWriter.WriteHeader(&tar.Header{
		Name: container.scriptPath,
		Mode: 0644,
		Size: int64(len(script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(script))
	return err
}

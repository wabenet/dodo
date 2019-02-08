package container

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

func uploadEntrypoint(
	ctx context.Context, containerID string, options Options,
) error {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	go options.Client.CopyToContainer(
		ctx,
		containerID,
		"/",
		reader,
		types.CopyToContainerOptions{},
	)

	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	script := options.Script + "\n"

	err := tarWriter.WriteHeader(&tar.Header{
		Name: options.ScriptPath,
		Mode: 0644,
		Size: int64(len(script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(script))
	return err
}

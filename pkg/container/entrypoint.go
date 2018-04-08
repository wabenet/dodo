package container

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func uploadEntrypoint(ctx context.Context, containerID string, options Options) error {
	reader, writer := io.Pipe()
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		err := options.Client.CopyToContainer(
			ctx,
			containerID,
			"/",
			reader,
			types.CopyToContainerOptions{},
		)
		if err != nil {
			log.Error(err)
		}
	}()

	tarWriter := tar.NewWriter(writer)
	err := tarWriter.WriteHeader(&tar.Header{
		Name: options.Entrypoint,
		Mode: 0600,
		Size: int64(len(options.Script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(options.Script))
	if err != nil {
		return err
	}
	err = tarWriter.Close()
	if err != nil {
		return err
	}
	return writer.Close()
}

package container

import (
	"archive/tar"
	"io"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func uploadEntrypoint(
	ctx context.Context, containerID string, options Options,
) error {
	// TODO: missing eol in entrypoint?
	reader, writer := io.Pipe()
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error(err)
		}
	}()
	defer func() {
		if err := writer.Close(); err != nil {
			log.Error(err)
		}
	}()

	go options.Client.CopyToContainer(
		ctx,
		containerID,
		"/",
		reader,
		types.CopyToContainerOptions{},
	)

	tarWriter := tar.NewWriter(writer)
	defer func() {
		if err := tarWriter.Close(); err != nil {
			log.Error(err)
		}
	}()

	err := tarWriter.WriteHeader(&tar.Header{
		Name: options.Entrypoint,
		Mode: 0644,
		Size: int64(len(options.Script)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write([]byte(options.Script))
	return err
}

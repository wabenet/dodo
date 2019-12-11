package container

import (
	"archive/tar"
	"encoding/json"
	"io"
	"io/ioutil"
	"path"

	dockerapi "github.com/docker/docker/api/types"
	"github.com/oclaussen/dodo/pkg/stage"
)

func (container *Container) uploadEntrypoint(containerID string) error {
	return container.uploadFile(containerID, "entrypoint", []byte(container.config.Script+"\n"))
}

func (container *Container) uploadStageConfig(containerID string, opts *stage.DockerOptions) error {
	newOpts := &stage.DockerOptions{
		Version: opts.Version,
		Host:    opts.Host,
	}

	if len(opts.CAFile) > 0 {
		data, err := ioutil.ReadFile(opts.CAFile)
		if err != nil {
			return err
		}
		if err := container.uploadFile(containerID, "ca.pem", data); err != nil {
			return err
		}
		newOpts.CAFile = path.Join(container.tmpPath, "ca.pem")
	}

	if len(opts.CertFile) > 0 {
		data, err := ioutil.ReadFile(opts.CertFile)
		if err != nil {
			return err
		}
		if err := container.uploadFile(containerID, "cert.pem", data); err != nil {
			return err
		}
		newOpts.CertFile = path.Join(container.tmpPath, "cert.pem")
	}

	if len(opts.KeyFile) > 0 {
		data, err := ioutil.ReadFile(opts.KeyFile)
		if err != nil {
			return err
		}
		if err := container.uploadFile(containerID, "key.pem", data); err != nil {
			return err
		}
		newOpts.KeyFile = path.Join(container.tmpPath, "key.pem")
	}

	data, err := json.Marshal(newOpts)
	if err != nil {
		return err
	}

	if err := container.uploadFile(containerID, "stagecfg.json", data); err != nil {
		return err
	}

	return nil
}

func (container *Container) uploadFile(containerID string, name string, contents []byte) error {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()

	go container.client.CopyToContainer(
		container.context,
		containerID,
		"/",
		reader,
		dockerapi.CopyToContainerOptions{},
	)

	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	err := tarWriter.WriteHeader(&tar.Header{
		Name: path.Join(container.tmpPath, name),
		Mode: 0644,
		Size: int64(len(contents)),
	})
	if err != nil {
		return err
	}
	_, err = tarWriter.Write(contents)
	return err
}

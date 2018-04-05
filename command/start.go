package command

import (
	"os"

	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

func (command *Command) startContainer(container *docker.Container) error {
	_, err := command.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true,
		Stream:       true,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		return err
	}

	inFd, _ := term.GetFdInfo(os.Stdin)
	state, err := term.SetRawTerminal(inFd)
	if err != nil {
		return err
	}
	defer term.RestoreTerminal(inFd, state)

	err = command.Client.StartContainer(container.ID, nil)
	_, err = command.Client.WaitContainer(container.ID)
	// TODO: handle exit code
	if err != nil {
		return err
	}

	return nil
}

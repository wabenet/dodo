package state

import (
	"os"

	"github.com/oclaussen/dodo/config"
	"github.com/oclaussen/dodo/options"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

type Run interface {
	Run() error
}

type state struct {
	Name        string
	Options     *options.Options
	Config      *config.BackdropConfig
	Client      *docker.Client
	Entrypoint  string
	Image       string
	Container   *docker.Container
}

func NewState(name string, options *options.Options) Run {
	// TODO: generate a temp file in the container for the entrypoint
	return &state{
		Name:       name,
		Options:    options,
		Entrypoint: "/tmp/dodo-entrypoint",
	}
}

func (state *state) Run() error {
	if err := state.ensureContainer(); err != nil {
		return err
	}
	defer state.ensureCleanup()
	if err := state.ensureEntrypoint(); err != nil {
		return err
	}

	_, err := state.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    state.Container.ID,
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
	terminalState, err := term.SetRawTerminal(inFd)
	if err != nil {
		return err
	}
	defer term.RestoreTerminal(inFd, terminalState)

	err = state.Client.StartContainer(state.Container.ID, nil)
	_, err = state.Client.WaitContainer(state.Container.ID)
	// TODO: handle exit code
	if err != nil {
		return err
	}

	return nil
}

package context

import (
	"os"

	"github.com/oclaussen/dodo/config"
	"github.com/oclaussen/dodo/options"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

type Context struct {
	Name        string
	Options     *options.Options
	Config      *config.ContextConfig
	Client      *docker.Client
	Entrypoint  string
	Image       string
	Container   *docker.Container
}

func NewContext(name string, options *options.Options) *Context {
	// TODO: generate a temp file in the container for the entrypoint
	return &Context{
		Name:       name,
		Options:    options,
		Entrypoint: "/tmp/dodo-entrypoint",
	}
}

func (context *Context) Run() error {
	if err := context.ensureContainer(); err != nil {
		return err
	}
	defer context.ensureCleanup()
	if err := context.ensureEntrypoint(); err != nil {
		return err
	}

	_, err := context.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    context.Container.ID,
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

	err = context.Client.StartContainer(context.Container.ID, nil)
	_, err = context.Client.WaitContainer(context.Container.ID)
	// TODO: handle exit code
	if err != nil {
		return err
	}

	return nil
}

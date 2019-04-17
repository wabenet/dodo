package container

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	dodotypes "github.com/oclaussen/dodo/pkg/types"
	"golang.org/x/net/context"
)

func (c *Container) run(containerID string, tty bool) error {
	attach, err := c.client.ContainerAttach(
		c.context,
		containerID,
		types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
			Logs:   true,
		},
	)
	if err != nil {
		return err
	}
	defer attach.Close()

	streamErrorChannel := make(chan error, 1)
	go streamContainer(c.context, streamErrorChannel, attach, tty)

	condition := container.WaitConditionNextExit
	if c.config.Remove == nil || *c.config.Remove == true {
		condition = container.WaitConditionRemoved
	}
	waitChannel, waitErrorChannel := c.client.ContainerWait(
		c.context,
		containerID,
		condition,
	)

	err = c.client.ContainerStart(
		c.context,
		containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		return err
	}

	if tty {
		c.resize(containerID)
		resizeChannel := make(chan os.Signal, 1)
		signal.Notify(resizeChannel, syscall.SIGWINCH)
		go func() {
			for range resizeChannel {
				c.resize(containerID)
			}
		}()
	}

	if err := <-streamErrorChannel; err != nil {
		return err
	}

	select {
	case response := <-waitChannel:
		if response.StatusCode != 0 {
			scriptError := &dodotypes.ScriptError{ExitCode: int(response.StatusCode)}
			if response.Error != nil {
				scriptError.Message = response.Error.Message
			}
			return scriptError
		}
		return nil
	case err := <-waitErrorChannel:
		return err
	}
}

func (c *Container) resize(containerID string) {
	outFd, _ := term.GetFdInfo(os.Stdout)

	ws, err := term.GetWinsize(outFd)
	if err != nil {
		return
	}

	height := uint(ws.Height)
	width := uint(ws.Width)
	if height == 0 && width == 0 {
		return
	}

	c.client.ContainerResize(
		c.context,
		containerID,
		types.ResizeOptions{
			Height: height,
			Width:  width,
		},
	)
}

func streamContainer(
	ctx context.Context, errChan chan<- error, attach types.HijackedResponse, tty bool,
) {
	if tty {
		inFd, _ := term.GetFdInfo(os.Stdin)
		inState, err := term.SetRawTerminal(inFd)
		if err != nil {
			errChan <- err
			return
		}
		defer term.RestoreTerminal(inFd, inState)

		outFd, _ := term.GetFdInfo(os.Stdout)
		outState, err := term.SetRawTerminal(outFd)
		if err != nil {
			errChan <- err
			return
		}
		defer term.RestoreTerminal(outFd, outState)
	}

	outputDone := make(chan error)
	go func() {
		if tty {
			_, err := io.Copy(os.Stdout, attach.Reader)
			outputDone <- err
		} else {
			_, err := stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)
			outputDone <- err
		}
	}()

	inputDone := make(chan struct{})
	go func() {
		io.Copy(attach.Conn, os.Stdin)
		attach.CloseWrite()
		close(inputDone)
	}()

	select {
	case err := <-outputDone:
		errChan <- err
	case <-inputDone:
		select {
		case err := <-outputDone:
			errChan <- err
		case <-ctx.Done():
			errChan <- ctx.Err()
		}
	case <-ctx.Done():
		errChan <- ctx.Err()
	}
}

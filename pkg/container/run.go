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
	"golang.org/x/net/context"
)

func runContainer(
	ctx context.Context, containerID string, options Options, tty bool,
) error {
	attach, err := options.Client.ContainerAttach(
		ctx,
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
	go streamContainer(ctx, streamErrorChannel, attach, tty)

	condition := container.WaitConditionNextExit
	if options.Remove {
		condition = container.WaitConditionRemoved
	}
	waitChannel, waitErrorChannel := options.Client.ContainerWait(
		ctx,
		containerID,
		condition,
	)

	err = options.Client.ContainerStart(
		ctx,
		containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		return err
	}

	if tty {
		resizeContainer(ctx, containerID, options)
		resizeChannel := make(chan os.Signal, 1)
		signal.Notify(resizeChannel, syscall.SIGWINCH)
		go func() {
			for range resizeChannel {
				resizeContainer(ctx, containerID, options)
			}
		}()
	}

	if err := <-streamErrorChannel; err != nil {
		return err
	}

	select {
	case response := <-waitChannel:
		if response.StatusCode != 0 {
			scriptError := &ScriptError{ExitCode: int(response.StatusCode)}
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

func resizeContainer(
	ctx context.Context, containerID string, options Options,
) {
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

	options.Client.ContainerResize(
		ctx,
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

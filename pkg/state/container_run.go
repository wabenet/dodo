package state

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

func (state *State) EnsureRun(ctx context.Context) error {
	client, err := state.EnsureClient(ctx)
	if err != nil {
		return err
	}
	containerID, err := state.EnsureContainer(ctx)
	if err != nil {
		return err
	}
	defer state.EnsureCleanup(ctx)
	err = state.EnsureEntrypoint(ctx)
	if err != nil {
		return err
	}

	attach, err := client.ContainerAttach(
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
	go func() {
		inFd, _ := term.GetFdInfo(os.Stdin)
		inState, err := term.SetRawTerminal(inFd)
		if err != nil {
			streamErrorChannel <- err
			return
		}
		defer term.RestoreTerminal(inFd, inState)

		outFd, _ := term.GetFdInfo(os.Stdout)
		outState, err := term.SetRawTerminal(outFd)
		if err != nil {
			streamErrorChannel <- err
			return
		}
		defer term.RestoreTerminal(outFd, outState)

		outputDone := make(chan error)
		go func() {
			_, err := io.Copy(os.Stdout, attach.Reader)
			outputDone <- err
		}()

		inputDone := make(chan struct{})
		go func() {
			io.Copy(attach.Conn, os.Stdin)
			attach.CloseWrite()
			close(inputDone)
		}()

		select {
		case err := <-outputDone:
			streamErrorChannel <- err
		case <-inputDone:
			select {
			case err := <-outputDone:
				streamErrorChannel <- err
			case <-ctx.Done():
				streamErrorChannel <- ctx.Err()
			}
		case <-ctx.Done():
			streamErrorChannel <- ctx.Err()
		}
	}()

	waitChannel, waitErrorChannel := client.ContainerWait(
		ctx,
		containerID,
		container.WaitConditionRemoved,
	)

	err = client.ContainerStart(
		ctx,
		containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		return err
	}

	if err := <-streamErrorChannel; err != nil {
		return err
	}

	select {
	case _ = <-waitChannel:
		return nil
	case err := <-waitErrorChannel:
		return err
	}
}

package state

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// EnsureRun makes sure the command is performed.
func (state *State) EnsureRun(ctx context.Context) error {
	client, err := state.EnsureClient()
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
		defer func() {
			if err := term.RestoreTerminal(inFd, inState); err != nil {
				log.Error(err)
			}
		}()

		outFd, _ := term.GetFdInfo(os.Stdout)
		outState, err := term.SetRawTerminal(outFd)
		if err != nil {
			streamErrorChannel <- err
			return
		}
		defer func() {
			if err := term.RestoreTerminal(outFd, outState); err != nil {
				log.Error(err)
			}
		}()

		outputDone := make(chan error)
		go func() {
			_, err := io.Copy(os.Stdout, attach.Reader)
			outputDone <- err
		}()

		inputDone := make(chan struct{})
		go func() {
			if _, err := io.Copy(attach.Conn, os.Stdin); err != nil {
				log.Error(err)
			}
			if err := attach.CloseWrite(); err != nil {
				log.Error(err)
			}
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
	case <-waitChannel:
		return nil
	case err := <-waitErrorChannel:
		return err
	}
}

package container

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/term"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func runContainer(
	ctx context.Context, containerID string, options Options,
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
	go streamContainer(ctx, streamErrorChannel, attach)

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

func streamContainer(
	ctx context.Context, errChan chan<- error, attach types.HijackedResponse,
) {
	inFd, _ := term.GetFdInfo(os.Stdin)
	inState, err := term.SetRawTerminal(inFd)
	if err != nil {
		errChan <- err
		return
	}
	defer func() {
		if restErr := term.RestoreTerminal(inFd, inState); err != nil {
			log.Error(restErr)
		}
	}()

	outFd, _ := term.GetFdInfo(os.Stdout)
	outState, err := term.SetRawTerminal(outFd)
	if err != nil {
		errChan <- err
		return
	}
	defer func() {
		if restErr := term.RestoreTerminal(outFd, outState); err != nil {
			log.Error(restErr)
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

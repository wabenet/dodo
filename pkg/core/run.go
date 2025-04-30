package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	log "github.com/hashicorp/go-hclog"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/configuration"
	"github.com/wabenet/dodo-core/pkg/plugin/runtime"
	"github.com/wabenet/dodo-core/pkg/ui"
)

const (
	ExitCodeInternalError = 1
	DefaultCommand        = "run"
)

func RunByName(m plugin.Manager, name string, overrides ...configuration.Backdrop) (int, error) {
	b, err := AssembleBackdropConfig(m, name, overrides...)
	if err != nil {
		return ExitCodeInternalError, err
	}

	if len(b.ContainerConfig.Name) == 0 {
		id := make([]byte, 8)
		if _, err := rand.Read(id); err != nil {
			panic(err)
		}

		b.ContainerConfig.Name = fmt.Sprintf("%s-%s", b.Name, hex.EncodeToString(id))
	}

	if len(b.ContainerConfig.Image) == 0 {
		for _, dep := range b.BuildConfig.Dependencies {
			if _, err := BuildByName(m, dep); err != nil {
				return ExitCodeInternalError, err
			}
		}

		imageID, err := BuildImage(m, b)
		if err != nil {
			return ExitCodeInternalError, err
		}

		b.ContainerConfig.Image = imageID
	}

	return RunBackdrop(m, b)
}

func RunBackdrop(m plugin.Manager, b configuration.Backdrop) (int, error) {
	rt, err := runtime.GetByName(m, b.Runtime)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find runtime plugin for %s: %w", b.Runtime, err)
	}

	imageID, err := rt.ResolveImage(b.ContainerConfig.Image)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find %s: %w", b.ContainerConfig.Image, err)
	}

	b.ContainerConfig.Image = imageID
	b.ContainerConfig.Terminal = runtime.TerminalConfig{
		StdIO: true,
		TTY:   ui.IsTTY(),
	}

	containerID, err := rt.CreateContainer(b.ContainerConfig)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not create container: %w", err)
	}

	for _, upload := range b.RequiredFiles {
		if err := rt.WriteFile(containerID, upload.FilePath, ([]byte)(upload.Contents)); err != nil {
			return ExitCodeInternalError, fmt.Errorf("could not upload file: %w", err)
		}
	}

	exitCode := 0

	t := ui.NewTerminal()
	t = t.OnSignal(func(s os.Signal, t *ui.Terminal) {
		log.L().Debug("handling signal", "s", s)

		switch s {
		case syscall.SIGWINCH:
			if err := rt.ResizeContainer(containerID, t.Height, t.Width); err != nil {
				log.L().Warn("could not resize terminal", "error", err)
			}
		default:
			if err := rt.KillContainer(containerID, s); err != nil {
				log.L().Warn("could not kill container", "error", err)
			}
		}
	})

	err = t.RunInRaw(func(t *ui.Terminal) error {
		if r, err := rt.StreamContainer(containerID, &plugin.StreamConfig{
			Stdin:          t.Stdin,
			Stdout:         t.Stdout,
			Stderr:         t.Stderr,
			TerminalHeight: t.Height,
			TerminalWidth:  t.Width,
		}); err != nil {
			return fmt.Errorf("error in container I/O stream: %w", err)
		} else {
			exitCode = r.ExitCode
		}

		return nil
	})

	return exitCode, err
}

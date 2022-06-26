package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	log "github.com/hashicorp/go-hclog"
	api "github.com/wabenet/dodo-core/api/v1alpha3"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/configuration"
	"github.com/wabenet/dodo-core/pkg/plugin/runtime"
	"github.com/wabenet/dodo-core/pkg/ui"
)

const (
	ExitCodeInternalError = 1
	DefaultCommand        = "run"
)

func RunByName(m plugin.Manager, overrides *api.Backdrop) (int, error) {
	b, err := configuration.AssembleBackdropConfig(m, overrides.Name, overrides)
	if err != nil {
		return ExitCodeInternalError, err
	}

	if len(b.ContainerName) == 0 {
		id := make([]byte, 8)
		if _, err := rand.Read(id); err != nil {
			panic(err)
		}

		b.ContainerName = fmt.Sprintf("%s-%s", b.Name, hex.EncodeToString(id))
	}

	if len(b.ImageId) == 0 {
		for _, dep := range b.BuildInfo.Dependencies {
			if _, err := BuildByName(m, &api.BuildInfo{ImageName: dep}); err != nil {
				return ExitCodeInternalError, err
			}
		}

		imageID, err := BuildImage(m, b.BuildInfo)
		if err != nil {
			return ExitCodeInternalError, err
		}

		b.ImageId = imageID
	}

	return RunBackdrop(m, b)
}

func RunBackdrop(m plugin.Manager, b *api.Backdrop) (int, error) {
	rt, err := runtime.GetByName(m, b.Runtime)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find runtime plugin for %s: %w", b.Runtime, err)
	}

	imageID, err := rt.ResolveImage(b.ImageId)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find %s: %w", b.ImageId, err)
	}

	b.ImageId = imageID

	containerID, err := rt.CreateContainer(b, ui.IsTTY(), true)
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not create container: %w", err)
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

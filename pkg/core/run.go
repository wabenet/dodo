package core

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	api "github.com/dodo-cli/dodo-core/api/v1alpha2"
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo-core/pkg/plugin/configuration"
	"github.com/dodo-cli/dodo-core/pkg/plugin/runtime"
	"github.com/dodo-cli/dodo-core/pkg/ui"
	log "github.com/hashicorp/go-hclog"
)

const (
	ExitCodeInternalError = 1
	DefaultCommand        = "run"
)

var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
)

func RunByName(m plugin.Manager, overrides *api.Backdrop) (int, error) {
	b := configuration.AssembleBackdropConfig(m, overrides.Name, overrides)

	if len(b.ContainerName) == 0 {
		id := make([]byte, 8)
		if _, err := rand.Read(id); err != nil {
			panic(err)
		}

		b.ContainerName = fmt.Sprintf("%s-%s", b.Name, hex.EncodeToString(id))
	}

	if len(b.ImageId) == 0 {
		if b.BuildInfo == nil {
			return ExitCodeInternalError, fmt.Errorf("neither image nor build configured for backdrop %s: %w", overrides.Name, ErrInvalidConfiguration)
		}

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
	t = t.WithResizeHook(func(t *ui.Terminal) {
		if err := rt.ResizeContainer(containerID, t.Height, t.Width); err != nil {
			log.L().Warn("could not resize terminal", "error", err)
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

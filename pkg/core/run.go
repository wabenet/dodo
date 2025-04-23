package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	log "github.com/hashicorp/go-hclog"
	configapi "github.com/wabenet/dodo-core/api/configuration/v1alpha2"
	runtimeapi "github.com/wabenet/dodo-core/api/runtime/v1alpha2"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/runtime"
	"github.com/wabenet/dodo-core/pkg/ui"
)

const (
	ExitCodeInternalError = 1
	DefaultCommand        = "run"
)

func RunByName(m plugin.Manager, name string, overrides ...*configapi.Backdrop) (int, error) {
	b, err := AssembleBackdropConfig(m, name, overrides...)
	if err != nil {
		return ExitCodeInternalError, err
	}

	if len(b.GetContainerConfig().GetName()) == 0 {
		id := make([]byte, 8)
		if _, err := rand.Read(id); err != nil {
			panic(err)
		}

		b.GetContainerConfig().Name = fmt.Sprintf("%s-%s", b.Name, hex.EncodeToString(id))
	}

	if len(b.GetContainerConfig().GetImage()) == 0 {
		for _, dep := range b.GetBuildConfig().GetDependencies() {
			if _, err := BuildByName(m, dep); err != nil {
				return ExitCodeInternalError, err
			}
		}

		imageID, err := BuildImage(m, b)
		if err != nil {
			return ExitCodeInternalError, err
		}

		b.GetContainerConfig().Image = imageID
	}

	return RunBackdrop(m, b)
}

func RunBackdrop(m plugin.Manager, b *configapi.Backdrop) (int, error) {
	rt, err := runtime.GetByName(m, b.GetRuntime())
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find runtime plugin for %s: %w", b.GetRuntime(), err)
	}

	imageID, err := rt.ResolveImage(b.GetContainerConfig().GetImage())
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not find %s: %w", b.GetContainerConfig().GetImage(), err)
	}

	b.GetContainerConfig().Image = imageID
	b.GetContainerConfig().Terminal = &runtimeapi.TerminalConfig{
		Stdio: true,
		Tty:   ui.IsTTY(),
	}

	containerID, err := rt.CreateContainer(b.GetContainerConfig())
	if err != nil {
		return ExitCodeInternalError, fmt.Errorf("could not create container: %w", err)
	}

	for _, upload := range b.GetRequiredFiles() {
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

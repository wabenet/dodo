package core

import (
	"fmt"

	api "github.com/dodo-cli/dodo-core/api/v1alpha2"
	"github.com/dodo-cli/dodo-core/pkg/plugin"
	"github.com/dodo-cli/dodo-core/pkg/plugin/builder"
	"github.com/dodo-cli/dodo-core/pkg/plugin/configuration"
	"github.com/dodo-cli/dodo-core/pkg/ui"
)

func BuildByName(m plugin.Manager, overrides *api.BuildInfo) (string, error) {
	config, err := configuration.FindBuildConfig(m, overrides.ImageName, overrides)
	if err != nil {
		return "", fmt.Errorf("error finding build config for %s: %w", overrides.ImageName, err)
	}

	for _, dep := range config.Dependencies {
		conf := &api.BuildInfo{}
		configuration.MergeBuildInfo(conf, overrides)
		conf.ImageName = dep

		if _, err := BuildByName(m, conf); err != nil {
			return "", err
		}
	}

	return BuildImage(m, config)
}

func BuildImage(m plugin.Manager, config *api.BuildInfo) (string, error) {
	b, err := builder.GetByName(m, config.Builder)
	if err != nil {
		return "", fmt.Errorf("could not find build plugin for %s: %w", config.Builder, err)
	}

	if !ui.IsTTY() {
		imageID, err := b.CreateImage(config, nil)
		if err != nil {
			return "", fmt.Errorf("error during image build: %w", err)
		}

		return imageID, nil
	}

	imageID := ""

	err = ui.NewTerminal().RunInRaw(
		func(t *ui.Terminal) error {
			if id, err := b.CreateImage(config, &plugin.StreamConfig{
				Stdin:          t.Stdin,
				Stdout:         t.Stdout,
				Stderr:         t.Stderr,
				TerminalHeight: t.Height,
				TerminalWidth:  t.Width,
			}); err != nil {
				return fmt.Errorf("error in container I/O stream: %w", err)
			} else {
				imageID = id
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return imageID, nil
}

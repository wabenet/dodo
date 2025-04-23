package core

import (
	"fmt"

	configapi "github.com/wabenet/dodo-core/api/configuration/v1alpha2"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/builder"
	"github.com/wabenet/dodo-core/pkg/ui"
)

func BuildByName(m plugin.Manager, name string, overrides ...*configapi.Backdrop) (string, error) {
	config, err := FindBuildConfig(m, name, overrides...)
	if err != nil {
		return "", fmt.Errorf("error finding build config for %s: %w", name, err)
	}

	for _, dep := range config.GetBuildConfig().GetDependencies() {
		conf := &configapi.Backdrop{}
		for _, override := range overrides {
			MergeBackdrop(conf, override)
		}
		conf.BuildConfig.ImageName = dep

		if _, err := BuildByName(m, dep, conf); err != nil {
			return "", err
		}
	}

	return BuildImage(m, config)
}

func BuildImage(m plugin.Manager, config *configapi.Backdrop) (string, error) {
	b, err := builder.GetByName(m, config.GetBuilder())
	if err != nil {
		return "", fmt.Errorf("could not find build plugin for %s: %w", config.GetBuilder(), err)
	}

	if !ui.IsTTY() {
		imageID, err := b.CreateImage(config.GetBuildConfig(), nil)
		if err != nil {
			return "", fmt.Errorf("error during image build: %w", err)
		}

		return imageID, nil
	}

	imageID := ""

	err = ui.NewTerminal().RunInRaw(
		func(t *ui.Terminal) error {
			if id, err := b.CreateImage(config.GetBuildConfig(), &plugin.StreamConfig{
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

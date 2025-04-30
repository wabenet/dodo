package core

import (
	"fmt"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/configuration"
)

type NotFoundError struct {
	Name   string
	Reason error
}

func (e NotFoundError) Error() string {
	if e.Reason == nil {
		return fmt.Sprintf(
			"could not find any configuration for '%s'", e.Name,
		)
	}

	return fmt.Sprintf(
		"could not find any configuration for '%s': %s",
		e.Name,
		e.Reason.Error(),
	)
}

func AssembleBackdropConfig(m plugin.Manager, name string, overrides ...configuration.Backdrop) (configuration.Backdrop, error) {
	var errs error

	config := configuration.Backdrop{}
	foundSomething := false

	for n, p := range m.GetPlugins(configuration.Type.String()) {
		log.L().Debug("Fetching configuration from plugin", "name", n)

		conf, err := p.(configuration.Configuration).GetBackdrop(name)
		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		foundSomething = true

		config = configuration.MergeBackdrops(config, conf)
	}

	if !foundSomething {
		return config, NotFoundError{Name: name, Reason: errs}
	}

	for _, override := range overrides {
		config = configuration.MergeBackdrops(config, override)
	}
	log.L().Debug("assembled configuration", "backdrop", config)

	err := config.Validate()

	return config, err
}

func FindBuildConfig(m plugin.Manager, name string, overrides ...configuration.Backdrop) (configuration.Backdrop, error) {
	var errs error

	for n, p := range m.GetPlugins(configuration.Type.String()) {
		log.L().Debug("Fetching configuration from plugin", "name", n)

		confs, err := p.(configuration.Configuration).ListBackdrops()
		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		for _, conf := range confs {
			if conf.BuildConfig.ImageName == name {
				config := configuration.Backdrop{}
				config = configuration.MergeBackdrops(config, conf)
				for _, override := range overrides {
					config = configuration.MergeBackdrops(config, override)
				}

				if err := config.Validate(); err != nil {
					errs = multierror.Append(errs, err)

					continue
				}

				return config, nil
			}
		}
	}

	return configuration.Backdrop{}, NotFoundError{Name: name, Reason: errs}
}

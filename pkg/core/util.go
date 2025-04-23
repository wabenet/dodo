package core

import (
	"fmt"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-multierror"
	buildapi "github.com/wabenet/dodo-core/api/build/v1alpha2"
	configapi "github.com/wabenet/dodo-core/api/configuration/v1alpha2"
	runtimeapi "github.com/wabenet/dodo-core/api/runtime/v1alpha2"
	"github.com/wabenet/dodo-core/pkg/plugin"
	configuration "github.com/wabenet/dodo-core/pkg/plugin/configuration"
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

type InvalidError struct {
	Name    string
	Message string
}

func (e InvalidError) Error() string {
	return fmt.Sprintf(
		"invalid configuration for '%s': %s",
		e.Name,
		e.Message,
	)
}

func AssembleBackdropConfig(m plugin.Manager, name string, overrides ...*configapi.Backdrop) (*configapi.Backdrop, error) {
	var errs error

	config := &configapi.Backdrop{}
	foundSomething := false

	for n, p := range m.GetPlugins(configuration.Type.String()) {
		log.L().Debug("Fetching configuration from plugin", "name", n)

		conf, err := p.(configuration.Configuration).GetBackdrop(name)
		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		foundSomething = true

		MergeBackdrop(config, conf)
	}

	if !foundSomething {
		return nil, NotFoundError{Name: name, Reason: errs}
	}

	for _, override := range overrides {
		MergeBackdrop(config, override)
	}
	log.L().Debug("assembled configuration", "backdrop", config)

	err := ValidateBackdrop(config)

	return config, err
}

func FindBuildConfig(m plugin.Manager, name string, overrides ...*configapi.Backdrop) (*configapi.Backdrop, error) {
	var errs error

	for n, p := range m.GetPlugins(configuration.Type.String()) {
		log.L().Debug("Fetching configuration from plugin", "name", n)

		confs, err := p.(configuration.Configuration).ListBackdrops()
		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		for _, conf := range confs {
			if conf.GetBuildConfig() != nil && conf.GetBuildConfig().GetImageName() == name {
				config := &configapi.Backdrop{}
				MergeBackdrop(config, conf)
				for _, override := range overrides {
					MergeBackdrop(config, override)
				}

				if err := ValidateBackdrop(config); err != nil {
					errs = multierror.Append(errs, err)

					continue
				}

				return config, nil
			}
		}
	}

	return nil, NotFoundError{Name: name, Reason: errs}
}

func MergeBackdrop(target, source *configapi.Backdrop) {
	if len(source.GetName()) > 0 {
		target.Name = source.GetName()
	}

	target.Aliases = append(target.GetAliases(), source.GetAliases()...)

	if len(source.GetRuntime()) > 0 {
		target.Runtime = source.GetRuntime()
	}

	if len(source.GetBuilder()) > 0 {
		target.Builder = source.GetBuilder()
	}

	if source.GetContainerConfig() != nil {
		if target.GetContainerConfig() == nil {
			target.ContainerConfig = source.GetContainerConfig()
		} else {
			MergeContainerConfig(target.GetContainerConfig(), source.GetContainerConfig())
		}
	}

	if source.GetBuildConfig() != nil {
		if target.GetBuildConfig() == nil {
			target.BuildConfig = source.GetBuildConfig()
		} else {
			MergeBuildConfig(target.GetBuildConfig(), source.GetBuildConfig())
		}
	}

	target.RequiredFiles = append(target.GetRequiredFiles(), source.GetRequiredFiles()...)
}

func MergeContainerConfig(target, source *runtimeapi.ContainerConfig) {
	if len(source.GetName()) > 0 {
		target.Name = source.GetName()
	}

	if len(source.GetImage()) > 0 {
		target.Image = source.GetImage()
	}

	if source.GetProcess() != nil {
		if len(source.GetProcess().GetUser()) > 0 {
			target.Process.User = source.GetProcess().GetUser()
		}

		if len(source.GetProcess().GetWorkingDir()) > 0 {
			target.Process.WorkingDir = source.GetProcess().GetWorkingDir()
		}

		if len(source.GetProcess().GetEntrypoint()) > 0 {
			target.Process.Entrypoint = source.GetProcess().GetEntrypoint()
		}

		if len(source.GetProcess().GetCommand()) > 0 {
			target.Process.Command = source.GetProcess().GetCommand()
		}
	}

	target.Environment = append(target.GetEnvironment(), source.GetEnvironment()...)
	target.Ports = append(target.GetPorts(), source.GetPorts()...)
	target.Mounts = append(target.GetMounts(), source.GetMounts()...)
	target.Capabilities = append(target.GetCapabilities(), source.GetCapabilities()...)
}

func MergeBuildConfig(target, source *buildapi.BuildConfig) {
	if len(source.GetImageName()) > 0 {
		target.ImageName = source.GetImageName()
	}

	if len(source.GetContext()) > 0 {
		target.Context = source.GetContext()
	}

	if len(source.GetDockerfile()) > 0 {
		target.Dockerfile = source.GetDockerfile()
	}

	if len(source.GetInlineDockerfile()) > 0 {
		target.InlineDockerfile = source.GetInlineDockerfile()
	}

	target.Arguments = append(target.GetArguments(), source.GetArguments()...)
	target.Secrets = append(target.GetSecrets(), source.GetSecrets()...)
	target.SshAgents = append(target.GetSshAgents(), source.GetSshAgents()...)

	if source.GetNoCache() {
		target.NoCache = true
	}

	if source.GetForceRebuild() {
		target.ForceRebuild = true
	}

	if source.GetForcePull() {
		target.ForcePull = true
	}

	target.Dependencies = append(target.GetDependencies(), source.GetDependencies()...)
}

func ValidateBackdrop(b *configapi.Backdrop) error {
	if b.GetContainerConfig().GetImage() == "" && b.GetBuildConfig() == nil {
		return InvalidError{Name: b.GetName(), Message: "neither image nor build configured"}
	}

	if b.GetBuildConfig() != nil {
		if err := ValidateBuildInfo(b.GetBuildConfig()); err != nil {
			return err
		}
	}

	return nil
}

func ValidateBuildInfo(b *buildapi.BuildConfig) error {
	return nil
}

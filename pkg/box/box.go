package box

import (
	"fmt"
	"path/filepath"

	"github.com/oclaussen/dodo/pkg/config"
	"github.com/oclaussen/dodo/pkg/integrations/vagrantcloud"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/pkg/errors"
)

type Box struct {
	config      *types.Box
	storagePath string
	tmpPath     string
	metadata    *vagrantcloud.Box
	version     *vagrantcloud.Version
	provider    *vagrantcloud.Provider
}

func Load(conf *types.Box, provider string) (*Box, error) {
	box := &Box{config: conf}
	api := vagrantcloud.New(conf.AccessToken)
	metadata, err := api.GetBox(&vagrantcloud.BoxOptions{Username: conf.User, Name: conf.Name})
	if err != nil {
		return nil, errors.Wrap(err, "could not get box metadata")
	}
	box.metadata = metadata
	box.storagePath = config.GetBoxesDir()
	box.tmpPath = config.GetTmpDir()

	v, err := findVersion(conf.Version, metadata)
	if err != nil {
		return box, errors.Wrap(err, "could not find a valid box version")
	}
	box.version = v

	p, err := findProvider(provider, metadata, v)
	if err != nil {
		return box, errors.Wrap(err, "could not find a valid box provider")
	}
	box.provider = p

	return box, nil
}

func findVersion(version string, box *vagrantcloud.Box) (*vagrantcloud.Version, error) {
	if len(version) == 0 {
		return &box.CurrentVersion, nil
	}
	for _, v := range box.Versions {
		if v.Version == version {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("could not find version '%s' for box '%s/%s'", version, box.Username, box.Name)
}

func findProvider(name string, box *vagrantcloud.Box, version *vagrantcloud.Version) (*vagrantcloud.Provider, error) {
	for _, p := range version.Providers {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("provider '%s' is not supported in version '%s' of box '%s/%s'", name, version.Version, box.Username, box.Name)
}

func (box *Box) Path() string {
	return filepath.Join(
		box.storagePath,
		box.metadata.Username,
		box.metadata.Name,
		box.version.Version,
		box.provider.Name,
	)
}

package image

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
	"golang.org/x/net/context"
)

func pull(ctx context.Context, options Options) (string, error) {
	parsed, err := reference.ParseNormalizedNamed(options.Name)
	if err != nil {
		return "", err
	}
	if reference.IsNameOnly(parsed) {
		parsed = reference.TagNameOnly(parsed)
	}

	repoInfo, err := registry.ParseRepositoryInfo(parsed)
	if err != nil {
		return "", err
	}

	configKey := repoInfo.Index.Name
	if repoInfo.Index.Official {
		info, err := options.Client.Info(ctx)
		if err != nil && info.IndexServerAddress != "" {
			configKey = info.IndexServerAddress
		} else {
			configKey = registry.IndexServer
		}
	}

	buf, err := json.Marshal(options.AuthConfigs[configKey])
	if err != nil {
		return "", err
	}

	response, err := options.Client.ImagePull(
		ctx,
		parsed.String(),
		types.ImagePullOptions{
			RegistryAuth: base64.URLEncoding.EncodeToString(buf),
		},
	)
	if err != nil {
		return "", err
	}

	_, err = handleImageResult(response, false)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

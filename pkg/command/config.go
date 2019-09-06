package command

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/oclaussen/dodo/pkg/types"
	"github.com/oclaussen/go-gimme/configfiles"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// LoadConfiguration tries to find a backdrop configuration by name in any of
// the supported locations. If given, will only look in the supplied config
// file.
func LoadConfiguration(backdrop string, filename string) (*types.Backdrop, error) {
	var opts *configfiles.Options
	if len(filename) > 0 {
		opts = &configfiles.Options{
			FileGlobs:        []string{filename},
			UseFileGlobsOnly: true,
		}
	} else {
		opts = &configfiles.Options{
			Name:                      "dodo",
			Extensions:                []string{"yaml", "yml", "json"},
			IncludeWorkingDirectories: true,
			Filter: func(configFile *configfiles.ConfigFile) bool {
				return containsBackdrop(configFile, backdrop)
			},
		}
	}

	configFile, err := configfiles.GimmeConfigFiles(opts)
	if err != nil {
		return nil, err
	}

	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		return nil, err
	}

	decoder := types.NewDecoder(configFile.Path, backdrop)
	config, err := decoder.DecodeGroup(configFile.Path, mapType)
	if err != nil {
		return nil, err
	}

	if result, ok := config.Backdrops[backdrop]; ok {
		return &result, nil
	}

	return nil, fmt.Errorf("could not find backdrop %s in file %s", backdrop, configFile.Path)
}

func containsBackdrop(configFile *configfiles.ConfigFile, backdrop string) bool {
	var mapType map[interface{}]interface{}
	if err := yaml.Unmarshal(configFile.Content, &mapType); err != nil {
		log.WithFields(log.Fields{"file": configFile.Path}).Warn("invalid YAML syntax in file")
		return false
	}

	decoder := types.NewDecoder(configFile.Path, backdrop)
	config, err := decoder.DecodeNames(configFile.Path, "", mapType)
	if err != nil {
		log.WithFields(log.Fields{"file": configFile.Path, "reason": err}).Warn("invalid config file")
		return false
	}

	_, ok := config.Backdrops[backdrop]
	return ok
}

func LoadAuthConfig() map[string]dockertypes.AuthConfig {
	var authConfigs map[string]dockertypes.AuthConfig

	configFile, err := configfiles.GimmeConfigFiles(&configfiles.Options{
		Name:       "docker",
		Extensions: []string{"json"},
		Filter: func(configFile *configfiles.ConfigFile) bool {
			var config map[string]*json.RawMessage
			err := json.Unmarshal(configFile.Content, &config)
			return err == nil && config["auths"] != nil
		},
	})
	if err != nil {
		return authConfigs
	}

	var config map[string]*json.RawMessage
	if err = json.Unmarshal(configFile.Content, &config); err != nil || config["auths"] == nil {
		return authConfigs
	}
	if err = json.Unmarshal(*config["auths"], &authConfigs); err != nil {
		return authConfigs
	}

	for addr, ac := range authConfigs {
		ac.Username, ac.Password, err = decodeAuth(ac.Auth)
		if err == nil {
			ac.Auth = ""
			ac.ServerAddress = addr
			authConfigs[addr] = ac
		}
	}

	return authConfigs
}

func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", errors.New("something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.New("invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}

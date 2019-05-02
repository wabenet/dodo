package command

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/homedir"
	"github.com/oclaussen/dodo/pkg/gimme/configfiles"
	"github.com/oclaussen/dodo/pkg/types"
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
			Filter: func(c *configfiles.ConfigFile) bool {
				// TODO: only decode names, groups and includes
				if config, err := loadConfig(c); err == nil {
					_, ok := config.Backdrops[backdrop]
					return ok
				}
				return false
			},
		}
	}

	configFile, err := configfiles.GimmeConfigFiles(opts)
	if err != nil {
		return nil, err
	}

	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	if result, ok := config.Backdrops[backdrop]; ok {
		return &result, nil
	}

	return nil, fmt.Errorf("could not find backdrop %s in file %s", backdrop, configFile.Path)
}

// ListConfigurations prints out all available backdrop names and the file
// it was found in.
func ListConfigurations() error {
	names := map[string]bool{}
	configfiles.GimmeConfigFiles(&configfiles.Options{
		Name:                      "dodo",
		Extensions:                []string{"yaml", "yml", "json"},
		IncludeWorkingDirectories: true,
		Filter: func(configFile *configfiles.ConfigFile) bool {
			// TODO: only decode names, groups and includes
			config, err := loadConfig(configFile)
			if err != nil {
				return false
			}
			for name := range config.Backdrops {
				// TODO: handle groups here correctly
				if !names[name] {
					fmt.Printf("%s (%s)\n", name, configFile.Path)
					names[name] = true
				}
			}
			return false
		},
	})
	return nil
}

func loadConfig(configFile *configfiles.ConfigFile) (types.Group, error) {
	var mapType map[interface{}]interface{}
	err := yaml.Unmarshal(configFile.Content, &mapType)
	if err != nil {
		return types.Group{}, err
	}

	config, err := types.DecodeGroup(configFile.Path, mapType)
	if err != nil {
		return types.Group{}, err
	}

	return config, nil
}

func LoadAuthConfig() map[string]dockertypes.AuthConfig {
	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(homedir.Get(), ".docker")
	}
	filename := filepath.Join(configDir, "config.json")

	var authConfigs map[string]dockertypes.AuthConfig

	if _, err := os.Stat(filename); err != nil {
		return authConfigs
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return authConfigs
	}

	var config map[string]*json.RawMessage
	if err = json.Unmarshal(bytes, &config); err != nil {
		return authConfigs
	}
	if config["auths"] == nil {
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

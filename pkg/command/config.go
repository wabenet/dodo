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
	"github.com/oclaussen/dodo/pkg/configfiles"
	"github.com/oclaussen/dodo/pkg/types"
	"gopkg.in/yaml.v2"
)

// LoadConfiguration tries to find a backdrop configuration by name in any of
// the supported locations. If given, will only look in the supplied config
// file.
func LoadConfiguration(
	backdrop string, configfile string,
) (*types.Backdrop, error) {
	if configfile != "" {
		config, err := ParseConfigurationFile(configfile)
		if err != nil {
			return nil, err
		}
		result, ok := config.Backdrops[backdrop]
		if !ok {
			return nil, fmt.Errorf("could not find backdrop %s in file %s", backdrop, configfile)
		}
		return &result, nil
	}

	candidates, err := configfiles.FindConfigFiles("dodo", []string{"yaml", "yml", "json"})
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		config, err := ParseConfigurationFile(candidate)
		if err != nil {
			return nil, err
		}
		if result, ok := config.Backdrops[backdrop]; ok {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("could not find backdrop %s in any configuration file", backdrop)
}

// ListConfigurations prints out all available backdrop names and the file
// it was found in.
func ListConfigurations() error {
	result := map[string]string{}
	candidates, err := configfiles.FindConfigFiles("dodo", []string{"yaml", "yml", "json"})
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		config, err := ParseConfigurationFile(candidate)
		if err != nil {
			return err
		}
		for name := range config.Backdrops {
			if result[name] == "" {
				fmt.Printf("%s (%s)\n", name, candidate)
				result[name] = candidate
			}
		}
	}
	return nil
}

// ParseConfigurationFile reads a full dodo configuration from a file.
func ParseConfigurationFile(filename string) (types.Group, error) {
	if !filepath.IsAbs(filename) {
		directory, err := os.Getwd()
		if err != nil {
			return types.Group{}, err
		}
		filename, err = filepath.Abs(filepath.Join(directory, filename))
		if err != nil {
			return types.Group{}, err
		}
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return types.Group{}, fmt.Errorf("could not read file '%s'", filename)
	}

	var mapType map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return types.Group{}, err
	}

	config, err := types.DecodeGroup(filename, mapType)
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

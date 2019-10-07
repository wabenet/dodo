package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/oclaussen/go-gimme/configfiles"
	"github.com/pkg/errors"
)

func LoadAuthConfig() map[string]types.AuthConfig {
	var authConfigs map[string]types.AuthConfig

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

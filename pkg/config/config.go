package config

import (
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// TODO: allow the following:
// - only the context as string in build
// - environment as key=value list or as map
// - volumes as source:dest:type list or as special structs
// - builds args as key=value list or as map
// - steps as string or slice

// TODO: support env_file as well

// TODO: validation
// TODO: check if there are unknown keys

// Config represents a full configuration file
type Config struct {
	Backdrops map[string]BackdropConfig `mapstructure:"backdrops"`
}

func ParseConfiguration(bytes []byte) (*Config, error) {
	var mapType map[string]interface{}
	err := yaml.Unmarshal(bytes, &mapType)
	if err != nil {
		return nil, err
	}

	var config Config
	err = mapstructure.Decode(mapType, &config)
	return &config, err
}

// BackdropConfig represents the configuration for a backdrop
// (possible target for running a command)
type BackdropConfig struct {
	Build         *BuildConfig `mapstructure:"build"`
	ContainerName string       `mapstructure:"container_name"`
	Remove        *bool        `mapstructure:"remove"`
	Pull          bool         `mapstructure:"pull"`
	Interactive   bool         `mapstructure:"interactive"`
	Environment   []string     `mapstructure:"environment"`
	Image         string       `mapstructure:"image"`
	User          string       `mapstructure:"user"`
	Volumes       []string     `mapstructure:"volumes"`
	VolumesFrom   []string     `mapstructure:"volumes_from"`
	WorkingDir    string       `mapstructure:"working_dir"`
	Interpreter   []string     `mapstructure:"interpreter"`
	Script        string       `mapstructure:"script"`
	Command       []string     `mapstructure:"command"`
}

// BuildConfig represents the build configuration for a docker image
type BuildConfig struct {
	Context      string   `mapstructure:"context"`
	Dockerfile   string   `mapstructure:"dockerfile"`
	Steps        []string `mapstructure:"steps"`
	Args         []string `mapstructure:"args"`
	NoCache      bool     `mapstructure:"no_cache"`
	ForceRebuild bool     `mapstructure:"force_rebuild"`
}

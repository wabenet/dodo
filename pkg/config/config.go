package config

// TODO: allow the following:
// - only the context as string in build
// - environment as key=value list or as map
// - volumes as source:dest:type list or as special structs
// - builds args as key=value list or as map

// Config represents a full configuration file
type Config struct {
	Backdrops map[string]BackdropConfig `yaml:"backdrops,omitempty"`
}

// TODO: support env_file as well

// BackdropConfig represents the configuration for a backdrop
// (possible target for running a command)
type BackdropConfig struct {
	Build         *BuildConfig `yaml:"build,omitempty"`
	ContainerName string       `yaml:"container_name,omitempty"`
	Remove        *bool        `yaml:"remove,omitempty"`
	Pull          bool         `yaml:"pull,omitempty"`
	Interactive   bool         `yaml:"interactive,omitempty"`
	Environment   []string     `yaml:"environment,omitempty"`
	Image         string       `yaml:"image,omitempty"`
	User          string       `yaml:"user,omitempty"`
	Volumes       []string     `yaml:"volumes,omitempty"`
	VolumesFrom   []string     `yaml:"volumes_from,omitempty"`
	WorkingDir    string       `yaml:"working_dir,omitempty"`
	Interpreter   []string     `yaml:"interpreter,omitempty"`
	Script        string       `yaml:"script,omitempty"`
	Command       []string     `yaml:"command,omitempty"`
}

// TODO: add inline dockerfile steps

// BuildConfig represents the build configuration for a docker image
type BuildConfig struct {
	Context      string   `yaml:"context,omitempty"`
	Dockerfile   string   `yaml:"dockerfile,omitempty"`
	Args         []string `yaml:"args,omitempty"`
	NoCache      bool     `yaml:"no_cache,omitempty"`
	ForceRebuild bool     `yaml:"force_rebuild,omitempty"`
}

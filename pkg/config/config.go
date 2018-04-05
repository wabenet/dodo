package config

// TODO: allow the following:
// - only the context as string in build
// - environment as key=value list or as map
// - volumes as source:dest:type list or as special structs
// - builds args as key=value list or as map

type Config struct {
	Backdrops map[string]BackdropConfig `yaml:"backdrops,omitempty"`
}

// TODO: support env_file as well
// TODO: the type *bool sounds wrong. Is there optional or something?
type BackdropConfig struct {
	Build         *BuildConfig `yaml:"build,omitempty"`
	ContainerName string       `yaml:"container_name,omitempty"`
	Remove        *bool        `yaml:"remove,omitempty"`
	Pull          bool         `yaml:"pull,omitempty"`
	Environment   []string     `yaml:"environment,omitempty"`
	Image         string       `yaml:"image,omitempty"`
	User          string       `yaml:"user,omitempty"`
	Volumes       []string     `yaml:"volumes,omitempty"`
	VolumesFrom   []string     `yaml:"volumes_from,omitempty"`
	WorkingDir    string       `yaml:"working_dir,omitempty"`
	Interpreter   []string     `yaml:"interpreter,omitempty"`
	Script        string       `yaml:"script,omitempty"`
	Command       []string     `yaml:"command,omitempty"`
	Interactive   bool         `yaml:"interactive,omitempty"`
}

// TODO: add inline dockerfile steps
type BuildConfig struct {
	Context      string   `yaml:"context,omitempty"`
	Dockerfile   string   `yaml:"dockerfile,omitempty"`
	Args         []string `yaml:"args,omitempty"`
	NoCache      bool     `yaml:"no_cache,omitempty"`
	ForceRebuild bool     `yaml:"force_rebuild,omitempty"`
}

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleYaml = `
backdrops:
  simplePull:
    image: testimage

  fullExample:
    build:
      context: .
      dockerfile: Dockerfile
      steps:
        - RUN hello
        - RUN world
      args:
        - FOO=BAR
      no_cache: true
      force_rebuild: true
    container_name: testcontainer
    remove: false
    pull: true
    interactive: true
    environment:
      - BAR=FOO
    image: testimage
    volumes:
      - /foo:/bar:ro
    volumes_from: ['somevolume']
    interpreter: ['/bin/sh']
    script: |
      echo "$@"
    command: ['Hello', 'World']
`

func getExampleConfig(t *testing.T, name string) BackdropConfig {
	config, err := ParseConfiguration([]byte(exampleYaml))
	assert.Nil(t, err)
	assert.Contains(t, config.Backdrops, name)
	assert.NotNil(t, config.Backdrops[name])
	return config.Backdrops[name]
}

func TestSimplePull(t *testing.T) {
	config := getExampleConfig(t, "simplePull")
	assert.Equal(t, "testimage", config.Image)
}

func TestFullExample(t *testing.T) {
	config := getExampleConfig(t, "fullExample")
	assert.NotNil(t, config.Build)
	assert.Equal(t, ".", config.Build.Context)
	assert.Equal(t, "Dockerfile", config.Build.Dockerfile)
	assert.Equal(t, []string{"RUN hello", "RUN world"}, config.Build.Steps)
	assert.Contains(t, config.Build.Args, "FOO=BAR")
	assert.True(t, config.Build.NoCache)
	assert.True(t, config.Build.ForceRebuild)
	assert.Equal(t, "testcontainer", config.ContainerName)
	assert.NotNil(t, config.Remove)
	assert.False(t, *config.Remove)
	assert.True(t, config.Pull)
	assert.True(t, config.Interactive)
	assert.Contains(t, config.Environment, "BAR=FOO")
	assert.Equal(t, "testimage", config.Image)
	assert.Contains(t, config.Volumes, "/foo:/bar:ro")
	assert.Contains(t, config.VolumesFrom, "somevolume")
	assert.Contains(t, config.Interpreter, "/bin/sh")
	assert.Equal(t, "echo \"$@\"\n", config.Script)
	assert.Equal(t, []string{"Hello", "World"}, config.Command)
}

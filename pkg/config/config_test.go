package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleYaml = `
backdrops:
  simplePull:
    image: testimage

  contextOnly:
    build: ./path/to/context

  environments:
    environment:
      - FOO=BAR
      - SOMETHING

  simpleVolume:
    volumes: foo:bar:ro

  mixedVolumes:
    volumes:
      - test
      - source: from
        target: to
        read_only: true
      - bar:baz

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
    image: testimage
    volumes_from: 'somevolume'
    interpreter: '/bin/sh'
    script: |
      echo "$@"
    command: ['Hello', 'World']
`

var exampleGroupedYaml = `
groups:
  first:
    backdrops:
      example1:
        image: testimage

    groups:
      second:
        backdrops:
          example2:
            image: testimage

  third:
    backdrops:
      example3:
        image: testimage
`

func getExampleConfig(t *testing.T, name string) BackdropConfig {
	config, err := ParseConfiguration("exampleYaml", []byte(exampleYaml))
	assert.Nil(t, err)
	assert.Contains(t, config.Backdrops, name)
	assert.NotNil(t, config.Backdrops[name])
	return config.Backdrops[name]
}

func TestSimplePull(t *testing.T) {
	config := getExampleConfig(t, "simplePull")
	assert.Equal(t, "testimage", config.Image)
}

func TestContextOnly(t *testing.T) {
	config := getExampleConfig(t, "contextOnly")
	assert.NotNil(t, config.Build)
	assert.Equal(t, "./path/to/context", config.Build.Context)
	assert.Equal(t, "", config.Build.Dockerfile)
}

func TestEnvironments(t *testing.T) {
	config := getExampleConfig(t, "environments")
	assert.Equal(t, 2, len(config.Environment))

	foobar := config.Environment[0]
	assert.Equal(t, foobar.Key, "FOO")
	assert.NotNil(t, foobar.Value)
	assert.Equal(t, "BAR", *foobar.Value)

	something := config.Environment[1]
	assert.Equal(t, "SOMETHING", something.Key)
	assert.Nil(t, something.Value)
}

func TestSimpleVolume(t *testing.T) {
	config := getExampleConfig(t, "simpleVolume")
	assert.Equal(t, 1, len(config.Volumes))
	assert.Equal(t, "foo", config.Volumes[0].Source)
	assert.Equal(t, "bar", config.Volumes[0].Target)
	assert.True(t, config.Volumes[0].ReadOnly)
}

func TestMixedVolumes(t *testing.T) {
	config := getExampleConfig(t, "mixedVolumes")
	assert.Equal(t, 3, len(config.Volumes))

	sourceOnly := config.Volumes[0]
	assert.Equal(t, "test", sourceOnly.Source)
	assert.Equal(t, "", sourceOnly.Target)
	assert.False(t, sourceOnly.ReadOnly)

	fullSpec := config.Volumes[1]
	assert.Equal(t, "from", fullSpec.Source)
	assert.Equal(t, "to", fullSpec.Target)
	assert.True(t, fullSpec.ReadOnly)

	readWrite := config.Volumes[2]
	assert.Equal(t, "bar", readWrite.Source)
	assert.Equal(t, "baz", readWrite.Target)
	assert.False(t, readWrite.ReadOnly)
}

func TestFullExample(t *testing.T) {
	config := getExampleConfig(t, "fullExample")
	assert.NotNil(t, config.Build)
	assert.Equal(t, ".", config.Build.Context)
	assert.Equal(t, "Dockerfile", config.Build.Dockerfile)
	assert.Equal(t, []string{"RUN hello", "RUN world"}, config.Build.Steps)
	assert.Equal(t, 1, len(config.Build.Args))
	assert.Equal(t, "FOO", config.Build.Args[0].Key)
	assert.Equal(t, "BAR", *config.Build.Args[0].Value)
	assert.True(t, config.Build.NoCache)
	assert.True(t, config.Build.ForceRebuild)
	assert.Equal(t, "testcontainer", config.ContainerName)
	assert.NotNil(t, config.Remove)
	assert.False(t, *config.Remove)
	assert.True(t, config.Pull)
	assert.True(t, config.Interactive)
	assert.Equal(t, "testimage", config.Image)
	assert.Contains(t, config.VolumesFrom, "somevolume")
	assert.Contains(t, config.Interpreter, "/bin/sh")
	assert.Equal(t, "echo \"$@\"\n", config.Script)
	assert.Equal(t, []string{"Hello", "World"}, config.Command)
}

func TestNestedGroups(t *testing.T) {
	config, err := ParseConfiguration("exampleGroupedYaml", []byte(exampleGroupedYaml))
	assert.Nil(t, err)
	assert.Contains(t, config.Groups, "first")
	assert.Contains(t, config.Groups["first"].Backdrops, "example1")
	assert.Contains(t, config.Groups["first"].Groups, "second")
	assert.Contains(t, config.Groups["first"].Groups["second"].Backdrops, "example2")
	assert.Contains(t, config.Groups, "third")
	assert.Contains(t, config.Groups["third"].Backdrops, "example3")
}

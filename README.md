# dodo - like sudo, but for docker

```bash
$ terraform help
bash: terraform: command not found

$ dodo terraform help
Usage: terraform [--version] [--help] <command> [args]
```

Do you have all your applications neatly packed into docker images, but no really
convenient way to run them? Do you have a bashrc file full of aliases that just
wrap docker run commands? Makefiles that just run make inside a build container?
Did you abuse docker-compose as a task runner but always felt that it wasn't
quite the right tool for the job? Then dodo might be for you.

Dodo is at its heart just a stripped down docker-compose. In fact, it was mostly
inspired by docker-compose issue [#1896](https://github.com/docker/compose/issues/1896),
because I used to use compose as a replacement for make but I didn't want all
of my tasks executed when I run `docker-compose up`.

I like to think of dodo as a "sudo for docker", because it constructs the desired
environment for a shell command you want to execute. But instead of wrapping
your command with the correct `su` calls, it wraps it in a `docker run`.

## state of development

Currently, I maintain dodo as a small side project, whenever I feel like it. I
use it quite heavily, and therefore keep it in a stable and usable state. But any
features or use cases outside of my usual workflows will probably get overlooked.

I run docker mostly on OSX, with the docker daemon inside a VirtualBox VM, so
that case works perfectly. I have no idea, however, how dodo behaves on native
OSX or even Windows.

As long as dodo does not have many users outside myself, I will probably introduce
breaking changes easily.

Any contributions (especially OS compatibility or tests) are very welcome.

## usage

```bash
$ dodo --help
Run commands in a Docker context.

Dodo operates on a set of backdrops, that must be configured in configuration
files (in the current directory or one of the config directories). Backdrops
are similar to docker-composes services, but they define one-shot commands
instead of long-running services. More specifically, each backdrop defines a
docker container in which a script should be executed. Dodo simply passes all
CMD arguments to the first backdrop with NAME that is found. Additional FLAGS
can be used to overwrite the backdrop configuration.

Usage:
  dodo [flags] [name] [cmd...]
  dodo [command]

Available Commands:
  build       Build all required images for backdrop without running it
  help        Help about any command
  list        List available all backdrop configurations
  run         Same as running 'dodo [name]', can be used when a backdrop name collides with a top-level command
  stage       Manage stages
  validate    Validate configuration files for syntax errors

Flags:
      --build                      always build an image, even if already exists
      --forward-stage              forward stage information into container, so dodo can be used inside the container
  -e, --env stringArray            Set environment variables
  -h, --help                       help for dodo
  -i, --interactive                run an interactive session
      --no-cache                   do not use cache when building the image
      --no-rm                      keep the container after it exits
  -p, --publish stringArray        Publish a container's port(s) to the host
      --pull                       always attempt to pull a newer version of the image
      --rm                         automatically remove the container when it exits
  -s, --stage string               stage to user for docker daemon
  -u, --user string                Username or UID (format: <name|uid>[:<group|gid>])
  -v, --volume stringArray         Bind mount a volume
      --volumes-from stringArray   Mount volumes from the specified container(s)
  -w, --workdir string             working directory inside the container

Use "dodo [command] --help" for more information about a command.
```

### configuration

The configuration for dodo is written in YAML. Dodo generally requires a backdrop
name to run. All configuration files are then searched for a backdrop configuration
with the matching name. Configuration files are searched in the current directory
any parent directory up to the filesystem root, as well as all the usual
places for config files (`$HOME`, `$XDG_CONFIG_HOME/dodo`, `%APPDATA%`, etc).
The configuration file name is some variation of `dodo.yml`. Either with a `.yml`
or `.yaml` or even `.json` extension, and an optional leading dot.

The configuration format is very similar to docker-compose. The top level item
is usually `backdrops`, which is a mapping from backdrop names to configurations,
similar to compose services. The biggest difference is how the entrypoint (and
command) works. Instead of `entrypoint`, there is usually a `script` block,
that will be copied over to the container. The docker entrypoint will then
be set to `["${interpreter}","/path/to/script"]`, where the interpreter defaults
to `/bin/sh`.

Configuration files allow some templating. See the [templating](#templating)
section for details.

For details on the backdrop configuration, check the following [examples](#examples)
or the full [reference](#config-reference).

### examples

For example, I use the following configuration in most of my Ruby projects based
on the usual bundler + rake combo, which allows me to just run `dodo rake` to
build everything:

```yaml
backdrops:
  rake:
    image:
      context: .
      steps:
        - FROM ruby:2.4-alpine
        - COPY *.gemspec Gemfile* ./
        - RUN bundle install
    volumes:
      - '{{ projectRoot }}:/build'
    working_dir: '/build'
    script: exec bundle exec rake "$@"
    command: all
```

In my home directory, I have a dodo configuration with the following config for
terraform:

```yaml
backdrops:
  terraform:
    image: hashicorp/terraform:latest
    environment:
      - AWS_PROFILE
      - AWS_ACCESS_KEY_ID
      - AWS_SECRET_ACCESS_KEY
      - TF_LOG=DEBUG
      - TF_LOG_PATH=/terraform/terraform.log
    volumes:
      - {{ projectRoot }}:/terraform
    working_dir: '/terraform/{{ projectPath }}'
    script: |
      test -f terraform.log && rm -f terraform.log
      terraform init
      exec terraform "$@"
    command: apply
```

Another example is the `dodo.yaml` in this very repository, which allows you
to build the tool without any requirements other than dodo itself and of
course docker.

### stages (experimental)

An experimental support for managing `stages` is implemented. Stages are places
where backdrops can run, in other words environments with a docker daemon. By
default, only the `environment` stage is available, which means that dodo will
pick up the docker host configuration from environment variables, exactly like
the normal docker client would. In additions, a plugin system exists that allows
for more stage options. To install a stage plugin, fetch one of the binaries from
the [releases](/releases) page and drop it into your `~/.dodo/plugins` directory.
Currently these plugins are available:

* `docker-machine`: Simply delegates all actions to a docker-machine instance
  with the same name as the stage. Requires docker-machine and the respective
  drivers to be installed on the system.

* `virtualbox`: Manages a VirtualBox VM based on a Vagrant box as a stage. This
  is basically a weird cross between a custom docker-machine and vagrant
  implementation, to allow for a bit more customization for your local setup.

The whole stage API is very likely to change quite a bit in the future. For now,
it is just a simple proof of concept. In future releases, it is planned to run
dodo backdrops on remote systems by using stages based on EC2 or Kubernetes pods.

## config reference

### backdrops

Backdrops are the main components of dodo. Each backdrop is a template for a
docker container that acts as runtime environment for a script. The top-level
configuration object is `backdrops`, which is a map of backdrop names to objects
with the following options:

* `aliases`: a list of aliases that can be used instead of the backdrop name to run it
* `image`, `build`: defines configuration for building the docker image. Can be either
  a string containing an existing docker image, or an object with:
  * `name`: a name for the resulting image, will be used for dependency resolution
  * `context`: path to the build context
  * `dockerfile`: path to the dockerfile (relative to the build context)
  * `steps`, `inline`: list of additional steps to perform on the docker image. Can be
    used as an inline dockerfile.
  * `args`, `arguments`: build arguments for the image
  * `secrets`: secrets used for building
  * `ssh`: ssh agent connfiguration used for building
  * `no_cache`: set to true to disable the docker cache during build
  * `force_rebuild`: always rebuild the image, even if an image with the
    specified name already exists
  * `force_pull`: always pull the base image, even if it already exists
  * `requires`, `dependencies`: list of image names that are required to build
    this image. Backdrop configurations are searched for image declarations with
    this name and build before this image.
* `container_name`: set the container name
* `remove`, `rm`: always remove the container after running (defaults to `true`)
* `environment`, `env`: set environment variables
* `volumes`: list of additional volumes to mount. Only bind-mount volumes are
  currently supported.
* `volumes_from`: mount volumes from an existing container
* `ports`: expose ports from the container
* `user`: set the uid inside the container
* `workdir`, `working_dir`: set the working directory inside the container
* `script`: the script that should be executed
* `interpreter`: set the interpreter that should execute the script (defaults to
  `/bin/sh`)
* `command`: arguments that will be passed to the script by default. Will be
  overwritten by any command line arguments.
* `interactive`: try to start an interactive session by setting the docker
  entrypoint to the interpreter only, skipping the script and command

### stages

Stages are places to put backdrops, in other words an environment running a docker
daemon. Stages are still experimental, so the configuration is not defined too
well yet. The top-level configuration object is `stages`, which is a map of
stage names to objects with the following options:

* `type`: the type of stage, usually the name of a plugin to invoke to manage it
* `box`: configuration that defines a Vagrant that should be used for the stage,
  provided the stage plugin supports Vagrant. The following options are allowed:
  * `user`: the Vagrant cloud user that provides the box
  * `name`: name of the Vagrant box
  * `version`: version of the Vagrant box
  * `access_token`: Vagrant cloud access token to access private boxes
* `options`: additional options, that are passed directly to the plugin

### includes

Includes allow merging additional files or output of commands into the current
configuration. This is often useful with templating to generate configuration
from other tools. The top-level configuration object is `include`, which is a
list of objects with the following options:

* `file`: Absolute path to a valid dodo configuration file. The file will be
  parsed and merged into the current file.
* `text`: A YAML document that is a valid dodo configuration file. The document
  will be parsed and merged into the current file.

### templating

All strings in the YAML configuration are processed by the [golang templating
engine](https://golang.org/pkg/text/template/). The following additional methods
are available:

 * Everything from the [sprig library](http://masterminds.github.io/sprig/)
 * `{{ cwd }}` evaluates to the current working directory
 * `{{ currentFile }}` is the path to the current YAML file that is evaluated
 * `{{ currentDir }}` is the path to the directory where the current file is located
 * `{{ projectRoot }}` is the path to the current Git project (determined by the
   first `.git` directory found by walking the current working directory upwards).
   Useful in combination with `{{ projectPath }]` if you don't only want to
   bind-mount the current directory but the whole project.
 * `{{ projectPath }}` the path of the current working directory relative to `{{
   projectRoot }}`
 * `{{ env <variable> }}` evaluates to the contents of environment variable
   `<variable>`
 * `{{ user }}` evaluates to the current user, in form of a
   [golang user](https://golang.org/pkg/os/user/). From this, you can access
   fields like `{{ user.HomeDir }}` or `{{ user.Uid }}`.
 * `{{ sh <command> }}` executes `<command>` via `/bin/sh` and evaluates
   to its stdout

## dodo compared to other software

### docker-compose

Dodo is basically just a stripped down docker-compose. It is perfectly possible
to do everything that dodo does with docker-compose instead. The difference is
that compose is not specifically designed for the purpose, which is the reason why
[#1896](https://github.com/docker/compose/issues/1896) is open for so long.

### dobi

[Dobi](https://dnephin.github.io/dobi/) is a tool that was inspired by the same
issue as dodo. But dobi focuses on being a full fledged task runner / build tool,
while dodo is designed to work together with existing tools like make or rake
instead.

### toast

[Toast](https://github.com/stepchowfun/toast) is another tool very similar to dodo.
Like dobi, its focus is stronger on defining specific tasks for reproducible
builds locally and in CI, while dodo has evolved in a generic application wrapper.

## license & authors

```text
Copyright 2019 Ole Claussen

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

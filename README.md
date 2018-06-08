# dodo - like sudo, but for docker

```bash
$ terraform help
bash: terraform: command not found
$ dodo terraform help
Using image hashicorp/terraform:latest
Usage: terraform [--version] [--help] <command> [args]
```

Do you have all your dev tools neatly packed into docker images, but no really
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
  dodo [FLAGS] NAME [CMD...]

Flags:
      --build                      always build an image, even if already exists
      --debug                      show additional debug output
  -e, --env stringArray            Set environment variables
  -f, --file string                specify a dodo configuration file
  -h, --help                       help for dodo
  -i, --interactive                run an interactive session
      --list                       list all available backdrop configurations
      --no-cache                   do not use cache when building the image
      --no-rm                      keep the container after it exits
      --pull                       always attempt to pull a newer version of the image
  -q, --quiet                      suppress informational output
      --rm                         automatically remove the container when it exits
  -u, --user string                Username or UID (format: <name|uid>[:<group|gid>])
  -v, --volume stringArray         Bind mount a volume
      --volumes-from stringArray   Mount volumes from the specified container(s)
  -w, --workdir string             working directory inside the container
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
is always `backdrops`, which is a mapping from backdrop names to configurations,
similar to compose services. The biggest difference is how the entrypoint (and
command) works. Instead of `entrypoint`, there is usually a `script` block,
that will be copied over to the container. The docker entrypoint will then
be set to `["${interpreter}","/path/to/script"]`, where the interpreter defaults
to `/bin/sh`.

For details on the backdrop configuration, check the following [examples](#examples)
or the full [reference](#config-reference).

### examples

For example, I use the following configuration in most of my Ruby projects based
on the usual bundler + rake combo, which allows me to just run `dodo rake` to
build everything:

```yaml
backdrops:
  rake:
    build:
      context: .
      steps:
        - FROM ruby:2.4-alpine
        - COPY *.gemspec Gemfile* ./
        - RUN bundle install
    volumes:
      - ${PWD}:/build
    working_dir: /build
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
      - ${PWD}:/terraform
    working_dir: /terraform
    script: |
      test -f terraform.log && rm -f terraform.log
      terraform init
      exec terraform "$@"
    command: [plan, -out=terraform.tfplan]
```

Another example is the `dodo.yaml` in this very repository, which allows you
to build the tool without any requirements other than dodo itself and of
course docker.

### config reference

The following configuration options are supported for backdrops. They mostly
behave the same as their equivalent in a docker-compose file, unless otherwise
noted.

* `image`: the docker image to run. either this or `build` is required. If both
  are given, the built image is tagged with this name and reused on subsequent
  runs.
* `pull`: always pull the specified image, even if it already exists. Applies to
  the base image while building as well.
* `build`: defines configuration for building a docker image. Can be either
  a string containing the path to the build context, or an object with:
  * `context`: path to the build context
  * `dockerfile`: path to the dockerfile inside the build context
  * `steps`: list of additional steps to perform on the docker image. Can be
    used as an inline dockerfile.
  * `args`: build arguments
  * `no_cache`: set to true to disable the docker cache during build
  * `force_rebuild`: always rebuild the image, even if an image with the
    specified name already exists
* `container_name`: set the container name
* `remove`: always remove the container after running (defaults to `true`)
* `environment`: set environment variables
* `volumes`: list of additional volumes to mount. Only bind-mount volumes are
  currently supported.
* `volumes_from`: mount volumes from an existing container
* `user`: set the uid inside the container
* `working_dir`: set the working directory inside the container
* `script`: the script that should be executed
* `interpreter`: set the interpreter that should execute the script (defaults to
  `/bin/sh`)
* `command`: arguments that will be passed to the script by default. Will be
  overwritten by any command line arguments.
* `interactive`: try to start an interactive session by setting the docker
  entrypoint to the interpreter only, skipping the script and command

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

## license & authors

```text
Copyright 2018 Ole Claussen

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

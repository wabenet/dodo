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

Flags:
  -e, --env stringArray       set environment variables
  -h, --help                  help for dodo
  -i, --interactive           run an interactive session
  -p, --publish stringArray   publish a container's port(s) to the host
  -u, --user string           username or UID (format: <name|uid>[:<group|gid>])
  -v, --volume stringArray    bind mount a volume
  -w, --workdir string        working directory inside the container
```

### configuration

By default, dodo is bundled with the config plugin, which searches the working
directory and user home directory for YAML config files.

Take a look at the [plugin repository](github.com/wabenet/dodo-config) for a
defailed description of the configuration format.

### examples

For example, I use the following configuration in most of my Ruby projects based
on the usual bundler + rake combo, which allows me to just run `dodo rake` to
build everything:

```yaml
backdrops:
  rake:
    image:
      context: "{{ projectRoot }}"
      steps:
        - FROM ruby:latest
        - COPY *.gemspec Gemfile* ./
        - RUN bundle install
    volumes:
      - '{{ projectRoot }}:/build'
    working_dir: '/build'
    script: exec bundle exec rake "$@"
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
```

Another example is the `dodo.yaml` in this very repository, which allows you
to build the tool without any requirements other than dodo itself and of
course docker.

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
Copyright 2021 Ole Claussen

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

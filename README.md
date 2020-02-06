kompoze
=======

> Render Docker Compose / Stack files with the power of [go templates](https://golang.org/pkg/text/template/) with [dockerize](https://github.com/jwilder/dockerize) and [sprig](https://masterminds.github.io/sprig/) useful template functions.

![version v0.1.0](https://img.shields.io/badge/version-v1.0.0-brightgreen.svg) 
![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)

Docker Compose / Stack files are very static in nature as you only can use [YAML](https://yaml.org/) to define them. 

Yes, there are several nice tricks to make YAML feels more dynamic [like anchors or block merging](https://www.hadeploy.com/more/yaml_tricks/) but in the end, you can't add conditionals, neither iterations, scoped blocks...

This is where `kompoze` appears to the rescue!

## Installation

Download the latest version in your container:

* [linux/amd64](https://github.com/41North/kompoze/releases/download/v1.0.0/kompoze-linux-amd64-v1.0.0.tar.gz)
* [alpine/amd64](https://github.com/41North/kompoze/releases/download/v1.0.0/kompoze-alpine-linux-amd64-v1.0.0.tar.gz)
* [darwin/amd64](https://github.com/41North/kompoze/releases/download/v1.0.0/kompoze-darwin-amd64-v1.0.0.tar.gz)

### Docker Base Image

The `41north/kompoze` image is a base image based on `alpine linux`. `kompoze` is installed in the `$PATH` and can be used directly.

```
FROM 41north/kompoze
...
ENTRYPOINT kompoze ...
```

### Ubuntu Images

``` Dockerfile
RUN apt-get update && apt-get install -y wget

ENV KOMPOZE_VERSION v1.0.0

RUN wget https://github.com/41North/kompoze/releases/download/$KOMPOZE_VERSION/kompoze-linux-amd64--KOMPOZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf kompoze-linux-amd64-$KOMPOZE_VERSION.tar.gz \
    && rm kompoze-linux-amd64-$KOMPOZE_VERSION.tar.gz
```

### For Alpine Images:

``` Dockerfile
RUN apk add --no-cache openssl

ENV KOMPOZE_VERSION v1.0.0

RUN wget https://github.com/41North/kompoze/releases/download/$KOMPOZE_VERSION/kompoze-alpine-linux-amd64-KOMPOZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf kompoze-alpine-linux-amd64-$KOMPOZE_VERSION.tar.gz \
    && rm kompoze-alpine-linux-amd64-$KOMPOZE_VERSION.tar.gz
```


## Usage

`kompoze` utilizes a `definition.toml` file (read [definition section](#definition-file) for more information) that defines how to render the templates.

By default, if you don't pass anything to `kompoze`, it will search for a `definition.toml` file in current directory. So:

```sh
$ kompoze
```

And this, are equal:

```sh
$ kompoze definition.toml
```

You can specify multiple definition files by passing their paths directly:

```sh
$ kompoze definition.toml another-definition.toml /another/path/definition2.toml
```

You can tail multiple files to `STDOUT` and `STDERR` by passing the options multiple times:

```sh
$ kompoze -stdout definition.toml
```

If your file uses `{{` and `}}` as part of it's syntax, you can change the template escape characters using the `-delims` option:

```sh
$ kompoze -delims "<%:%>"
```

By default, the `base-path` for rendering will be the one on which you run `kompoze` (so any relative paths that are specified inside the definition file can be resolved). You can change it with the following option:

```sh
$ kompoze --base-path /another/path
```

### Definition File

The definition file uses [TOML](https://github.com/toml-lang/toml) format and tries to be very minimal and concise. It's composed by two main sections:

- Global vars: Those common variables that will be applied to every template (if any).
- Templates: Where it defines which templates to render and which variables are overridden from the global scope.

Take a look on the example below:

```toml
# Example definition file

# defines global variables that will be applied to each template definition (can be null)
[vars]

  # you can define the variables directly here (higher priority when merging same entries)
  [vars.global]
    network_enabled = true
    network_name = "net"
    network_subnet = "172.25.0.0/16"

  # or you can include other global variables files (lower priority when merging same entries)
  include = ["vars/global.toml"]

# defines a list of templates to render
[[templates]]
  src  = "templates/stack.yml.tpl"
  dest = "out/stack-1.yml"
  include_vars = ["vars/local.toml"]
  [templates.local_vars]
    mariadb_version = "10.2.21"
    mariadb_volume_enabled = true

[[templates]]
  src  = "templates/stack.yml.tpl"
  dest = "out/stack-2.yml"
  [templates.local_vars]
    mariadb_version = "11"
    mariadb_volume_enabled = false
```

As you can see above the syntax is pretty straightforward. You can define relative paths for `src` and `dest` and they will be resolved to the defined `base-path` option.

The format for including external variables are as follows:

```toml
[vars]
  this_is_another_var = 'var'
```

The different sources of variables are merged together in the following order:

1. global `vars`
2. global `include`
3. template `include_vars`
4. template `vars`

### Templates

Templates are rendered by using Golang's [text/template](http://golang.org/pkg/text/template/) package with the mix of two powerful additions:

- [sprig](https://masterminds.github.io/sprig/) functions. 
- Some of [dockerize](https://github.com/jwilder/dockerize) set of functions.

You can access environment variables within a template with `.Env` like `dockerize` or those defined in the definition file with plain `.` (like `.some_global_var`).

```
{{ .Env.PATH }} is my path
```

The set of stolen built in functions stolen from [dockerize](https://github.com/jwilder/dockerize) are the following:

  * `exists $path` - Determines if a file path exists or not. `{{ exists "/etc/default/myapp" }}`
  * `parseUrl $url` - Parses a URL into it's [protocol, scheme, host, etc. parts](https://golang.org/pkg/net/url/#URL). Alias for [`url.Parse`](https://golang.org/pkg/net/url/#Parse)
  * `isTrue $value` - Parses a string $value to a boolean value. `{{ if isTrue .Env.ENABLED }}`
  * `isFalse $value` - Parses a string $value to a boolean value. `{{ if isFalse .Env.ENABLED }}`
  * `loop` - Create for loops.
  
On the sprig side, everything is included by default, so you have access to all defined functions. 

## Contributions

Contributions to this project are very **welcome** and will be fully **credited**.

Feel free to send a PR to correct any possible bug or improvement you may want to add. 

Just make sure you follow these rules:

- **Create feature branches**: It's important to be concise with your commits, so don't ask us to pull from your master branch.
- **Document any change in behaviour**: Make sure the `README.md` is kept up-to-date.
- **One pull request per feature**: If you want to do more than one thing, send multiple pull requests.
- **Send coherent history**: Make sure each individual commit in your pull request is meaningful. If you had to make multiple intermediate commits while developing, please [squash them](http://www.git-scm.com/book/en/v2/Git-Tools-Rewriting-History#Changing-Multiple-Commit-Messages) before submitting.

## Acknowledgements

Many thanks to:

 - [jwilder](https://github.com/jwilder)
 - [Aisbergg](https://github.com/Aisbergg) 
 
Both of them for creating [dockerize](https://github.com/jwilder/dockerize) and [python-docker-compose-templer](https://github.com/Aisbergg/python-docker-compose-templer) respectively, from which this project draws 99% inspiration!

## License

*kompoze* is released under the MIT License. See [LICENSE.md](LICENSE.md) for more information.

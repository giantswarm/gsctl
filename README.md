[![Travis-CI Build Badge](https://api.travis-ci.org/giantswarm/gsctl.svg?branch=master)](https://travis-ci.org/giantswarm/gsctl)
[![codecov](https://codecov.io/gh/giantswarm/gsctl/branch/master/graph/badge.svg)](https://codecov.io/gh/giantswarm/gsctl)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/gsctl)](https://goreportcard.com/report/github.com/giantswarm/gsctl)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

# `gsctl` - The Giant Swarm CLI

`gsctl` is the cross-platform command line utility to manage your Kubernetes clusters at Giant Swarm.

## Usage

Call `gsctl` without any arguments to get an overview on commands. Some usage examples:

#### Log in using your Giant Swarm credentials

```nohighlight
$ gsctl login demo@example.com
Password:
Successfully logged in!
```

#### Show your clusters

```nohighlight
$ gsctl list clusters
Id     Name                Created             Organization
9gxjo  Production Cluster  2016 Oct 19, 15:43  giantswarm
xl8t1  Staging Cluster     2016 Sep 16, 09:30  giantswarm
```

#### Configure `kubectl` to access a cluster

```nohighlight
$ gsctl create kubeconfig -c xl8t1
Creating new key-pair…
New key-pair created with ID 153a93201… and expiry of 720 hours
Certificate and key files written to:
/Users/demo/.gsctl/certs/xl8t1-ca.crt
/Users/demo/.gsctl/certs/xl8t1-153a932010-client.crt
/Users/demo/.gsctl/certs/xl8t1-153a932010-client.key
Switched to kubectl context 'giantswarm-xl8t1'

kubectl is set up. Check it using this command:

    kubectl cluster-info

Whenever you want to switch to using this context:

    kubectl config set-context giantswarm-xl8t1
```

#### Sign out

```nohighlight
$ gsctl logout
Successfully logged out
```

## Install

See the [`gsctl` reference docs](https://docs.giantswarm.io/reference/gsctl/#install)

## Configuration

See the [`gsctl` reference docs](https://docs.giantswarm.io/reference/gsctl/#configuration)

## Changelog

See [Releases](https://github.com/giantswarm/gsctl/releases)

## Development

### What you might need

- Go environment (`brew install go`)
- [`glide`](https://github.com/Masterminds/glide) (`brew install glide`)
- GNU Make
- `git`
- Docker

### Cloning to the right location

Make sure you have the `$GOPATH` environment variable set.

```nohighlight
$ echo $GOPATH
/Users/johndoe/go
```

Go to right location, then check out:

```nohighlight
$ mkdir -p $GOPATH/src/github.com/giantswarm
$ cd $GOPATH/src/github.com/giantswarm
$ git clone https://github.com/giantswarm/gsctl.git
$ cd gsctl
```

So the repo content will end up in `$GOPATH/src/github.com/giantswarm/gsctl`.

### Dependencies

Dependencies are managed using [`glide`](https://github.com/Masterminds/glide).

**CAUTION:** The `vendor` directory is _flattened_. Make sure to use glide with the `-v` (or `--strip-vendor`) flag.

### Executing via `go run`

Check everything by running the code via `go run`:

```nohighlight
$ go run main.go info
```

### Running tests

make gotest
make test

### Cleaning up code

Please do a `golint .` check and act on recommendations before pushing.

### Building binaries

A simple build can be done using `go build`.

To run a more reproducible build, the Makefile defines targets that build in a Docker container:

- `make` will build a binary for the current platform and place it in `./build/bin`
- `make install` will install this as `/usr/local/bin/gsctl`
- `make crosscompile` will build binaries for multiple platforms and place them into `./build/bin`

## Contributing

We welcome contributions! Please read our additional information on [how to contribute](https://github.com/giantswarm/gsctl/blob/master/CONTRIBUTING.md) for details.

## Publishing a Release

See [RELEASE.md](https://github.com/giantswarm/gsctl/blob/master/RELEASE.md)

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

#### Mac OS using Homebrew

```nohighlight
brew tap giantswarm/giantswarm && brew update && brew install gsctl
```

#### Mac OS without Homebrew

```nohighlight
curl -O http://downloads.giantswarm.io/gsctl/0.1.0/gsctl-0.1.0-darwin-amd64.tar.gz
tar xzf gsctl-0.1.0-darwin-amd64.tar.gz
sudo cp gsctl-0.1.0-darwin-amd64/gsctl /usr/local/bin/
```

#### Linux

```nohighlight
curl -O http://downloads.giantswarm.io/gsctl/0.1.0/gsctl-0.1.0-linux-amd64.tar.gz
tar xzf gsctl-0.1.0-linux-amd64.tar.gz
sudo cp gsctl-0.1.0-linux-amd64/gsctl /usr/local/bin/
```

#### Windows

- Download [`gsctl` for Windows (64 Bit)](http://downloads.giantswarm.io/gsctl/0.1.0/gsctl-0.1.0-windows-amd64.zip) or [32 Bit](http://downloads.giantswarm.io/gsctl/0.1.0/gsctl-0.1.0-windows-386.zip)
- Copy the contained `gsctl.exe` to a convenient location

## Changelog

TODO: Link to changelog

## Development

For productive development you will need a Go environment. In addition, GNU Make will come handy.

Fetching all dependencies can be done this way:

```nohighlight
make get-deps
```

To be able to execute the code using `go run`, set the `GOPATH` to the root directory of the repository:

```nohighlight
export GOPATH=`pwd`/.gobuild
```

Then commands like this should work fine:

```nohighlight
go run main.go info
```

## Contributing

We welcome contributions! Please read our additional information on [how to contribute](https://github.com/giantswarm/gsctl/blob/master/CONTRIBUTING.md) for details.

## Publishing a Release

See [RELEASE.md](https://github.com/giantswarm/gsctl/blob/master/RELEASE.md)

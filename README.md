[![Travis-CI Build Badge](https://api.travis-ci.org/giantswarm/gsctl.svg?branch=master)](https://travis-ci.org/giantswarm/gsctl)


# `gsctl` - The Giant Swarm CLI

`gsctl` is the cross-platform command line utility to manage your Kubernetes clusters at Giant Swarm.

## Usage

TODO: Link to documentation

## Install

TODO

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

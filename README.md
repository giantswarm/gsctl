[![Travis-CI Build Badge](https://api.travis-ci.org/giantswarm/gsctl.svg?branch=master)](https://travis-ci.org/giantswarm/gsctl)
[![codecov](https://codecov.io/gh/giantswarm/gsctl/branch/master/graph/badge.svg)](https://codecov.io/gh/giantswarm/gsctl)


# `gsctl` - The Giant Swarm CLI

`gsctl` is the cross-platform command line utility to manage your Kubernetes clusters at Giant Swarm.

## Usage

Call `gsctl` without any arguments to get an overview on commands. Some usage examples:

#### Log in using your Giant Swarm credentials

```nohighlight
$ gsctl login demo@example.com
Password: ************
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

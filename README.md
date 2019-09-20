![Repo Image](https://repository-images.githubusercontent.com/74132145/691bf000-70de-11e9-89d3-2f4693461d00)

[![Coverage Status](https://coveralls.io/repos/github/giantswarm/gsctl/badge.svg?branch=master)](https://coveralls.io/github/giantswarm/gsctl?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/gsctl)](https://goreportcard.com/report/github.com/giantswarm/gsctl)

# `gsctl` - The Giant Swarm CLI

`gsctl` is the cross-platform command line utility to manage your Kubernetes clusters at Giant Swarm.

## Usage

Call `gsctl` without any arguments to get an overview on commands. Some usage examples:

#### Log in using your Giant Swarm credentials

```nohighlight
$ gsctl login demo@example.com -e <giant-swarm-api-endpoint>
Password for demo@example.com at <giant-swarm-api-endpoint>:
Successfully logged in!
```

#### Show your clusters

```nohighlight
$ gsctl list clusters
ID     NAME                CREATED                 ORGANIZATION
9gxjo  Production Cluster  2016 Apr 30, 15:43 UTC  acme
xl8t1  Staging Cluster     2017 May 11, 09:30 UTC  acme
```

#### Create a cluster

```nohighlight
$ gsctl create cluster --owner acme --name "Test Cluster" --num-workers 5
Requesting new cluster for organization 'acme'
New cluster with ID 'h8d0j' is launching.
```

More in the [docs](https://docs.giantswarm.io/reference/gsctl/create-cluster/)

#### Configure `kubectl` to access a cluster

```nohighlight
$ gsctl create kubeconfig -c h8d0j
Creating new key pair…
New key pair created with ID 153a93201… and expiry of 720 hours
Certificate and key files written to:
/Users/demo/.config/gsctl/certs/h8d0j-ca.crt
/Users/demo/.config/gsctl/certs/h8d0j-153a932010-client.crt
/Users/demo/.config/gsctl/certs/h8d0j-153a932010-client.key
Switched to kubectl context 'giantswarm-xl8t1'

kubectl is set up. Check it using this command:

    kubectl cluster-info

Whenever you want to switch to using this context:

    kubectl config use-context giantswarm-xl8t1
```

#### Cluster acccess via internal networks

Internal Kubernetes API allows you talking to Kubernetes via internal load balancer. That can be useful for peered networks.

In case you want to use internal Kubernetes API, pass `--tenant-internal=true` to gsctl:
```nohighlight
$ gsctl create kubeconfig -c h8d0j
```

This will render kubeconfig with internal Kubernetes API server (`internal-api`).

* Internal API is awailable only for AWS installations.

## Install

See the [`gsctl` reference docs](https://docs.giantswarm.io/reference/gsctl/#install)

## Configuration

See the [`gsctl` reference docs](https://docs.giantswarm.io/reference/gsctl/#configuration)

## Changelog

See [Releases](https://github.com/giantswarm/gsctl/releases)

## Development

See [docs/Development.md](https://github.com/giantswarm/gsctl/blob/master/docs/Development.md)

## Contributing

We welcome contributions! Please read our additional information on [how to contribute](https://github.com/giantswarm/gsctl/blob/master/CONTRIBUTING.md) for details.

## Publishing a Release

See [docs/Release.md](https://github.com/giantswarm/gsctl/blob/master/docs/Release.md)


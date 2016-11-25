# How to publish a release

## Prerequisites

- GNU Make
- Docker environment
- `builder` (see https://github.com/giantswarm/builder)
- AWS CLI (`aws` command line utility)
- AWS S3 create/upload permissions for `downloads.giantswarm.io`
- `git`
- Push permissions for https://github.com/giantswarm/homebrew-giantswarm

## Instructions

```nohighlight
make
make test

builder release patch|minor|major

git checkout <version-tag>

make bin-dist
make release

```

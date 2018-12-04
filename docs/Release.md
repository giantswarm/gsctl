# How to publish a release

TL;DR: Releases are published automatically whenever a new tag of the format X.Y.Z is pushed to the GitHub repository.

## Prerequisites

CircleCI must be set up with certain environment variables:

- `CODE_SIGNING_CERT_BUNDLE_BASE64` - Base64 encoded PKCS#12 key/cert bundle used for signing Windows binaries
- `CODE_SIGNING_CERT_BUNDLE_PASSWORD` - Password for the above bundle
- `RELEASE_TOKEN` - A GitHub token with the permission to write to repositories
  - [giantswarm/gsctl](https://github.com/giantswarm/gsctl/)
  - [giantswarm/scoop-bucket](https://github.com/giantswarm/scoop-bucket)
  - [giantswarm/homebrew-giantswarm](https://github.com/giantswarm/homebrew-giantswarm)

## Create and push a new release tag

Replace `<MAJOR.MINOR.PATCH>` with the actual version number you want to publish.

```
export VERSION=<MAJOR.MINOR.PATCH>
git checkout master
git pull
git tag -a ${VERSION} -m "Release version ${VERSION}"
git push origin ${VERSION}
```

Follow CircleCI's progress in https://circleci.com/gh/giantswarm/gsctl/.

## Edit and publish the release

Open the [release draft](https://github.com/giantswarm/gsctl/releases/) on Github.

Edit the description to inform about what has changed since the last release. Save and publish the release.

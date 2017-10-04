# How to publish a release

## Prerequisites

- GNU Make
- Docker environment
- `builder` (see https://github.com/giantswarm/builder)
- AWS CLI (`aws` command line utility, see http://docs.aws.amazon.com/cli/latest/userguide/installing.html)
- AWS S3 create/upload permissions for `downloads.giantswarm.io`
- env variable `AWS_ACCESS_KEY_ID` set
- env variable `AWS_SECRET_ACCESS_KEY` set
- `git` command line utility
- Push permissions for https://github.com/giantswarm/gsctl
- Push permissions for https://github.com/giantswarm/homebrew-giantswarm
- env variable `BUILDER_GITHUB_TOKEN` set to a valid Github token
- Code signing certificate from keePass

## Instructions

We first build a binary for the current platform and test it.

```nohighlight
make
make test
make clean
```

Now we create a release draft on Github using `builder`. This will add some broken binaries to that release. We take care of that later.

```nohighlight
builder release patch|minor|major
```

Now check out the release tag:

```nohighlight
make clean
git checkout <version-tag>
```

Create the folder `certs` inside the repo, if not there, and place the file `code-signing.p12` from keePass (code signing certificate) there.

Set the environment variable `$CODE_SIGNING_CERT_BUNDLE_PASSWORD` to the PKCS#12 encryption password you find in keePass.

Then run this:

```nohighlight
make release
```

Open the [release draft](https://github.com/giantswarm/gsctl/releases/) on Github.

Upload the windows binaries to the release.

Edit the description to inform about what has changed since the last release. Save and publish the release.

To also upload the binaries to AWS S3 and provide an update for homebrew users, execute this:


Finally, when everything went fine, do another `make clean`.

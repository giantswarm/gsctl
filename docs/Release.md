# How to publish a release

## Prerequisites

- GNU Make
- Docker environment
- AWS CLI (`aws` command line utility, see http://docs.aws.amazon.com/cli/latest/userguide/installing.html)
- AWS S3 create/upload permissions for `downloads.giantswarm.io`
- `git` command line utility
- Push permissions for https://github.com/giantswarm/gsctl,  https://github.com/giantswarm/homebrew-giantswarm, https://github.com/giantswarm/scoop-bucket
- Code signing certificate from keePass
- `CODE_SIGNING_CERT_BUNDLE_PASSWORD` environment variable set to code signing PKCS#12 encryption password (from keePass)

## Authenticating for the AWS CLI

We'll use a specific profile for the CLI called `gsctl-release`. To create this
profile, run

```
aws configure --profile gsctl-release
```

and set the access key ID and the secret key to values matching your IAM account.

Then, if you have ever made a release before, you'll likely have to clean `~/.aws/credentials`
to remove any existing `aws_session_token` entry from the `gsctl-release` profile.

The use of multi factor authentication (MFA) requires us to create short-lived
credentials for AWS client. Have your MFA device and it's ARN (from the web
console) ready.

The ARN looks similar to this:

    arn:aws:iam::084190472784:mfa/marian@giantswarm.io

Next, get your current MFA token and use it like here (yes, you have to be swift):

```bash
ARN="arn:aws:iam::084190472784:mfa/marian@giantswarm.io"
MFA_TOKEN=123456
CREDENTIALS=$(aws --profile gsctl-release \
  sts get-session-token \
  --duration-seconds 20000 \
  --serial-number $ARN \
  --token-code $MFA_TOKEN | jq .Credentials)
```

Now take the output of the following command and place it in the
credentials file `~/.aws/credentials` at the `[gsctl-release]` entry.

```bash
echo "aws_access_key_id = $(echo $CREDENTIALS | jq -r .AccessKeyId )" && \
  echo "aws_secret_access_key = $(echo $CREDENTIALS | jq -r .SecretAccessKey )" && \
  echo "aws_session_token = $(echo $CREDENTIALS | jq -r .SessionToken )"
```

Test your credentials like this:

```nohighlight
aws --profile gsctl-release s3 ls s3://downloads.giantswarm.io/gsctl/
```

If this lists the gsctl versions released so far, this step is done.

## Test the binary locally

We first build a binary for the current platform and test it.

```nohighlight
make
make test
make clean
```

## Prepare signing of Windows binaries

Create the folder `certs` inside the repo, if not there, and place the file `code-signing.p12` from keePass (code signing certificate) there.

Set the environment variable `$CODE_SIGNING_CERT_BUNDLE_PASSWORD` to the PKCS#12 encryption password you find in keePass.

## Set the release version

Replace `<MAJOR.MINOR.PATCH>` with the actual version number.

```
export VERSION=<MAJOR.MINOR.PATCH>
git checkout master
echo "${VERSION}" > ./VERSION
git commit -m "Version bump to ${VERSION}" ./VERSION
git push origin master
git tag -a ${VERSION} -m "Release version ${VERSION}"
git push origin ${VERSION}
```

## Create binaries for distribution

```nohighlight
make bin-dist
```

## Publish the release

```
./release.sh
```

Open the [release draft](https://github.com/giantswarm/gsctl/releases/) on Github.

Edit the description to inform about what has changed since the last release. Save and publish the release.

## Update os-specific distribution channels

Once the release info is published, update homebrew and scoop:

```nohighlight
./update-homebrew.sh
./update-scoop.sh
```

## Cleaning up

When everything went fine, do another `make clean`.

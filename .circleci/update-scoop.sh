#!/bin/bash

# This script creates a scoop app manifest for the Windows binary release
# and pushes it to the repository "scoop-bucket"

set -o errexit
set -o nounset
set -o pipefail

# Our version number
VERSION=$1

REPO_URL="https://${RELEASE_TOKEN}@github.com/giantswarm/scoop-bucket.git"

# SHA256 hashs of the ZIP files
SHA256_64BIT=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-windows-amd64.zip|awk '{print $2}')
SHA256_32BIT=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-windows-386.zip|awk '{print $2}')

git clone --depth 1 $REPO_URL
cd scoop-bucket

# Dump manifest JSON
cat > gsctl.json << EOF
{
  "version": "$VERSION",
  "homepage": "https://github.com/giantswarm/gsctl/",
  "bin": "gsctl.exe",
  "license": "APACHE-2.0",
  "architecture": {
    "64bit": {
      "url": "https://downloads.giantswarm.io/gsctl/$VERSION/gsctl-$VERSION-windows-amd64.zip",
      "hash": "$SHA256_64BIT",
      "extract_dir": "gsctl-$VERSION-windows-amd64"
    },
    "32bit": {
      "url": "https://downloads.giantswarm.io/gsctl/$VERSION/gsctl-$VERSION-windows-386.zip",
      "hash": "$SHA256_32BIT",
      "extract_dir": "gsctl-$VERSION-windows-386"
    }
  }
}
EOF

git config credential.helper 'cache --timeout=120'
git config user.email "${TAYLORBOT_EMAIL}"
git config user.name "Taylor Bot"
git add gsctl.rb
git commit -m "Update gsctl to ${VERSION}"

# Push quietly with -q to prevent showing the token in log
git push -q $REPO_URL master

#!/bin/bash

# This script creates a scoop app manifest for the Windows binary release
# and pushes it to the repository "scoop-bucket"

SCOOP_REPO=https://github.com/giantswarm/scoop-bucket.git

# Our version number
VERSION=$(cat VERSION)

# SHA256 hashs of the ZIP files
SHA256_64BIT=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-windows-amd64.zip|awk '{print $2}')
SHA256_32BIT=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-windows-386.zip|awk '{print $2}')

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

# commit and push formula to our Homebrew tap
cd bin-dist
git clone $SCOOP_REPO
mv ../gsctl.json scoop-bucket/
cd scoop-bucket
git add ./gsctl.json && git commit -m "Updated gsctl to ${VERSION}" && git push origin master
rm -rf bin-dist/scoop-bucket

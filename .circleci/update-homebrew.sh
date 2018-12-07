#!/bin/bash

# This script creates a homebrew formula for the Mac OS binary release
# and pushes it to the right repository "homebrew-giantswarm"

set -o errexit
set -o nounset
set -o pipefail

# Our version number
VERSION=$1

REPO_URL="https://${RELEASE_TOKEN}@github.com/giantswarm/homebrew-giantswarm.git"

# SHA256 hash of the tar.gz file
SHA256=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-darwin-amd64.tar.gz|awk '{print $2}')

git clone --depth 1 $REPO_URL
cd homebrew-giantswarm/Formula

# Dump formula ruby code
cat > gsctl.rb << EOF
require "formula"

# This file is generated automatically by
# https://github.com/giantswarm/gsctl/blob/master/.circleci/update-homebrew.sh

class Gsctl < Formula
  desc "Controls things on Giant Swarm"
  homepage "https://github.com/giantswarm/gsctl"
  url "http://downloads.giantswarm.io/gsctl/$VERSION/gsctl-$VERSION-darwin-amd64.tar.gz"
  version "$VERSION"
  # openssl dgst -sha256 <file>
  sha256 "$SHA256"

  def install
    bin.install "gsctl"
  end
end
EOF

git config credential.helper 'cache --timeout=120'
git config user.email "${GITHUB_USER_EMAIL}"
git config user.name "${GITHUB_USER_NAME}"
git add gsctl.rb
git commit -m "Update gsctl to ${VERSION}"

# Push quietly with -q to prevent showing the token in log
git push -q $REPO_URL master

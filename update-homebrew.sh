#!/bin/bash

# This script creates a homebrew formula for the Mac OS binary release
# and pushes it to the right repository "homebrew-giantswarm"

# Our version number
VERSION=$(cat VERSION)

# SHA256 hash of the tar.gz file
SHA256=$(openssl dgst -sha256 bin-dist/gsctl-$VERSION-darwin-amd64.tar.gz|awk '{print $2}')

# Dump formula ruby code
cat > gsctl.rb << EOF
require "formula"

# This file is generated automatically by
# https://github.com/giantswarm/gsctl/blob/master/update-homebrew.sh

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

# commit and push formula to our Homebrew tap
cd bin-dist
git clone https://github.com/giantswarm/homebrew-giantswarm.git
mv ../gsctl.rb homebrew-giantswarm/Formula/
cd homebrew-giantswarm
git add Formula/gsctl.rb && git commit -m "Updated gsctl to ${VERSION}" && git push origin master
rm -rf bin-dist/homebrew-giantswarm

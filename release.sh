#!/bin/sh

# Script to publish a release

PROJECT=gsctl

VERSION=$(cat ./VERSION)

# test if this version is already released
echo "Checking if this release already exists"
aws --profile giantswarm s3 ls s3://downloads.giantswarm.io/${PROJECT}/ \
  | grep ${VERSION} \
  && echo "Error: A release for this version already exists in S3" \
  && exit 1

# test if bin-dist folder is there
test -d bin-dist || $(echo "Error: please run 'make bin-dist' first" && exit 2)

# Github personal access token of Github user
GITHUB_TOKEN=$(cat ~/.github-token)

echo "Creating Github release ${PROJECT} v${VERSION}"
release_output=$(curl -s \
    -X POST \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{
        \"tag_name\": \"${VERSION}\",
        \"name\": \"${PROJECT} v${VERSION}\",
        \"body\": \"### New features\\n\\n### Minor changes\\n\\n### Bugfixes\\n\\n\",
        \"draft\": true,
        \"prerelease\": false
    }" \
    https://api.github.com/repos/giantswarm/${PROJECT}/releases
)

# fetch the release id for the upload
RELEASE_ID=$(echo $release_output | jq '.id')

echo "Upload binary to GitHub Release"
cd bin-dist
for FILENAME in *.zip *.tar.gz; do
    [ -f "$FILENAME" ] || break
    curl \
      -H "Authorization: token ${GITHUB_TOKEN}" \
      -H "Content-Type: application/octet-stream" \
      --data-binary @${FILENAME} \
        https://uploads.github.com/repos/giantswarm/${PROJECT}/releases/${RELEASE_ID}/assets?name=${FILENAME}
done
cd ..


echo "Uploading release to S3 bucket downloads.giantswarm.io"
aws --profile giantswarm s3 cp bin-dist s3://downloads.giantswarm.io/${PROJECT}/${VERSION}/ --recursive --exclude="*" --include="*.tar.gz" --acl=public-read
aws --profile giantswarm s3 cp bin-dist s3://downloads.giantswarm.io/${PROJECT}/${VERSION}/ --recursive --exclude="*" --include="*.zip" --acl=public-read
aws --profile giantswarm s3 cp VERSION s3://downloads.giantswarm.io/${PROJECT}/VERSION --acl=public-read

echo "Done. The release is now prepared, but not yet published."
echo "You can now edit your release description here:"
echo "https://github.com/giantswarm/${PROJECT}/releases/"

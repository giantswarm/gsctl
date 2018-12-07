#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly TAG=$1
readonly GITHUB_TOKEN=$2

main() {
  echo "Creating GitHub release draft for version ${TAG}..."
  if ! id=$(release_github "${TAG}" "${GITHUB_TOKEN}"); then
    log_error "GitHub Release could not get created."
    exit 1
  fi
  echo "Created new GitHub release draft with ID ${id}"

  for filepath in ./bin-dist/*; do
    [ -f "$filepath" ] || continue
    if [ "$(upload_asset ${TAG}" "${GITHUB_TOKEN}" "${id}" "${filepath}")" == "uploaded" ]; then
      echo "File ${filepath} uploaded successfully"
    else
      log_error "Asset ${filepath} could not be uploaded"
      exit 1
    fi
  done
  
  echo "Please edit the release draft in https://github.com/giantswarm/gsctl/releases/"
}


release_github() {
  local version="${2?Specify version}"
  local token="${3?Specify Github Token}"

  release_output=$(curl -s \
      -X POST \
      -H "Authorization: token ${token}" \
      -H "Content-Type: application/json" \
      -d "{
          \"tag_name\": \"${version}\",
          \"name\": \"gsctl v${version}\",
          \"body\": \"### New features\\n\\n### Minor changes\\n\\n### Bugfixes\\n\\n\",
          \"draft\": true,
          \"prerelease\": false
      }" \
      "https://api.github.com/repos/giantswarm/gsctl/releases"
  )

  # Return release id for the asset upload
  release_id=$(echo "${release_output}" | jq '.id')
  echo "${release_id}"
  return 0
}

upload_asset() {
  local version="${2?Specify version}"
  local token="${3?Specify GitHub token}"
  local release_id="${4?Specify release Id}"
  local file_path="${5?Specify file path}"

  file_name=$(basename "${file_path}")

  echo "Upload file ${file_name} to GitHub Release"
  upload_output=$(curl -s \
        -H "Authorization: token ${token}" \
        -H "Content-Type: application/octet-stream" \
        --data-binary @${file_path} \
          "https://uploads.github.com/repos/giantswarm/gsctl/releases/${release_id}/assets?name=${file_name}"
  )

  echo "${upload_output}" | jq .state
  exit 0
}

log_error() {
    printf '\e[31mERROR: %s\n\e[39m' "$1" >&2
}

main

#!/usr/bin/env bash

# This script runs unit tests in a way that produces a coverage
# report in coverage.txt, to be used by Codecov.io

set -e
echo "" > coverage.txt

for d in $(go list ./...); do
    go test -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

# sort in place, remove redundant lines
sort -u -r coverage.txt -o coverage.txt

#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up-ops/* "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up" || exit 1

go get -u github.com/mattn/go-bindata/...
go generate github.com/EngineerBetter/concourse-up/...
export CONCOURSE_UP_ACME_URL=https://acme-staging.api.letsencrypt.org/directory # Avoid rate limits when testing
ginkgo -r

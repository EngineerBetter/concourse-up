#!/bin/bash

set -eu

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up"

go get -u github.com/mattn/go-bindata/...
go generate github.com/EngineerBetter/concourse-up/...
export CONCOURSE_UP_ACME_URL=https://acme-staging.api.letsencrypt.org/directory # Avoid rate limits when testing
ginkgo -r

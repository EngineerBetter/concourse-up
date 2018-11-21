#!/bin/bash

set -eu

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up-ops/* "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up"

go get -u github.com/mattn/go-bindata/...
go generate github.com/EngineerBetter/concourse-up/...
gometalinter \
--disable-all \
--enable=goconst \
--enable=ineffassign \
--enable=vetshadow \
--enable=golint \
--exclude=bindata \
--exclude=resource/internal/file \
--vendor \
--enable-gc \
--deadline=120s \
./...

find . -name vendor -prune ! -type d -o -name '*.sh' -exec shellcheck {} +

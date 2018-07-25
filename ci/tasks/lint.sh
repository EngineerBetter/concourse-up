#!/bin/bash

set -eu

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up"

go get -u github.com/mattn/go-bindata/...
go generate github.com/EngineerBetter/concourse-up/...
gometalinter.v1 \
--exclude vendor \
--exclude "_test\\.go" \
--disable=gotype \
--disable=gas \
--disable=aligncheck \
--disable=errcheck \
--deadline=500s \
./...

shellcheck -e SC2046,SC2140 $(find . -name '*.sh' | grep -v vendor)
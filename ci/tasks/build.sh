#!/bin/bash

set -eu

build_dir=$PWD/build-$GOOS
mkdir -p build_dir

if [ -e "version/version" ]; then
  version=$(cat version/version)
else
  version="TESTVERSION"
fi

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up"

GOOS=linux go get -u github.com/mattn/go-bindata/...
go generate github.com/EngineerBetter/concourse-up/...
go build -ldflags "
  -X main.ConcourseUpVersion=$version
  -X github.com/EngineerBetter/concourse-up/fly.ConcourseUpVersion=$version
" -o "$build_dir/$OUTPUT_FILE"

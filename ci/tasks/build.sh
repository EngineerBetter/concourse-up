#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

build_dir=$PWD/build-$GOOS
mkdir -p build_dir

if [ -e "version/version" ]; then
  version=$(cat version/version)
else
  version="TESTVERSION"
fi

mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
mv concourse-up/* "$GOPATH/src/github.com/EngineerBetter/concourse-up"
mv concourse-up-ops/* "$GOPATH/src/github.com/EngineerBetter/concourse-up-ops"
cd "$GOPATH/src/github.com/EngineerBetter/concourse-up" || exit 1

GOOS=linux go get -u github.com/mattn/go-bindata/... github.com/maxbrunsfeld/counterfeiter
go generate github.com/EngineerBetter/concourse-up/...
go build -ldflags "
  -X main.ConcourseUpVersion=$version
  -X github.com/EngineerBetter/concourse-up/fly.ConcourseUpVersion=$version
" -o "$build_dir/$OUTPUT_FILE"

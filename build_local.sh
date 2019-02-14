#!/bin/bash

set -eu

version=dev
grep -lr --include=*.go --exclude-dir=vendor "go:generate go-bindata" . | xargs -I {} go generate {}
GO111MODULE=on go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/concourse-up/fly.ConcourseUpVersion=$version
  -X main.ConcourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"

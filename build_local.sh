#!/bin/bash

set -eu

version=dev
go generate github.com/EngineerBetter/concourse-up/...
go build -ldflags "
  -X github.com/EngineerBetter/concourse-up/fly.ConcourseUpVersion=$version
  -X main.ConcourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"

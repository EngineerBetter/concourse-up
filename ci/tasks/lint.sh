#!/bin/bash

set -eu

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

gometalinter -e vendor -e "_test\.go" --deadline=120s ./...
